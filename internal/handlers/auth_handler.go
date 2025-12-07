package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/K-H-Tech/apm/internal/models"
	"github.com/K-H-Tech/apm/internal/repository"
	"github.com/K-H-Tech/apm/pkg/atlassian"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	oauthClient *atlassian.OAuthClient
	stateStore  *atlassian.SimpleStateStore
	userRepo    *repository.UserRepository
	orgRepo     *repository.OrganizationRepository
	encrypter   Encrypter
}

// Encrypter interface for encrypting tokens
type Encrypter interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// AuthHandlerConfig holds configuration for the auth handler
type AuthHandlerConfig struct {
	OAuthClient *atlassian.OAuthClient
	UserRepo    *repository.UserRepository
	OrgRepo     *repository.OrganizationRepository
	Encrypter   Encrypter
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(config AuthHandlerConfig) *AuthHandler {
	return &AuthHandler{
		oauthClient: config.OAuthClient,
		stateStore:  atlassian.NewSimpleStateStore(10 * time.Minute),
		userRepo:    config.UserRepo,
		orgRepo:     config.OrgRepo,
		encrypter:   config.Encrypter,
	}
}

// RegisterRoutes registers auth routes
func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.GET("/atlassian", h.InitiateOAuth)
		auth.GET("/atlassian/callback", h.HandleCallback)
		auth.POST("/refresh", h.RefreshToken)
		auth.GET("/me", h.GetCurrentUser)
		auth.POST("/logout", h.Logout)
	}
}

// InitiateOAuth initiates the OAuth flow
// @Summary Initiate Atlassian OAuth
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/atlassian [get]
func (h *AuthHandler) InitiateOAuth(c *gin.Context) {
	// Generate state token
	state, err := generateRandomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state"})
		return
	}

	// Store state for validation
	h.stateStore.Store(state)

	// Get authorization URL
	authURL := h.oauthClient.GetAuthorizationURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

// HandleCallback handles the OAuth callback
// @Summary Handle Atlassian OAuth callback
// @Tags Auth
// @Param code query string true "Authorization code"
// @Param state query string true "State token"
// @Success 200 {object} map[string]interface{}
// @Router /auth/atlassian/callback [get]
func (h *AuthHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code or state"})
		return
	}

	// Validate state
	if !h.stateStore.Validate(state) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	ctx := c.Request.Context()

	// Exchange code for tokens
	tokens, err := h.oauthClient.ExchangeCode(ctx, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user info
	userInfo, err := h.oauthClient.GetUserInfo(ctx, tokens.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}

	// Get accessible resources (Atlassian sites)
	resources, err := h.oauthClient.GetAccessibleResources(ctx, tokens.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get accessible resources"})
		return
	}

	// Encrypt tokens for storage
	encryptedAccess, err := h.encrypter.Encrypt(tokens.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt access token"})
		return
	}

	encryptedRefresh, err := h.encrypter.Encrypt(tokens.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt refresh token"})
		return
	}

	expiry := atlassian.TokenExpiry(tokens.ExpiresIn)

	// Check if user exists
	user, err := h.userRepo.GetByAtlassianAccountID(ctx, userInfo.AccountID)
	if err != nil {
		// Create new user
		user = &models.User{
			ID:                             uuid.New(),
			Email:                          userInfo.Email,
			Name:                           userInfo.Name,
			AvatarURL:                      userInfo.Picture,
			AtlassianAccountID:             userInfo.AccountID,
			AtlassianAccessTokenEncrypted:  encryptedAccess,
			AtlassianRefreshTokenEncrypted: encryptedRefresh,
			AtlassianTokenExpiry:           &expiry,
		}

		if err := h.userRepo.Create(ctx, user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	} else {
		// Update existing user's tokens
		user.AtlassianAccessTokenEncrypted = encryptedAccess
		user.AtlassianRefreshTokenEncrypted = encryptedRefresh
		user.AtlassianTokenExpiry = &expiry
		user.Name = userInfo.Name
		user.AvatarURL = userInfo.Picture

		if err := h.userRepo.Update(ctx, user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
			return
		}
	}

	// Process accessible resources (create/update organizations)
	organizations := make([]map[string]string, 0, len(resources))
	for _, resource := range resources {
		org, err := h.orgRepo.GetByJiraCloudID(ctx, resource.ID)
		if err != nil {
			// Create new organization
			org = &models.Organization{
				ID:          uuid.New(),
				Name:        resource.Name,
				JiraCloudID: resource.ID,
				JiraBaseURL: resource.URL,
			}
			if err := h.orgRepo.Create(ctx, org); err != nil {
				continue
			}

			// Add user as admin
			h.orgRepo.AddUserToOrganization(ctx, &models.OrganizationMember{
				OrganizationID: org.ID,
				UserID:         user.ID,
				Role:           models.RoleAdmin,
			})
		}

		organizations = append(organizations, map[string]string{
			"id":   org.ID.String(),
			"name": org.Name,
		})
	}

	// Generate session token (in production, use JWT)
	sessionToken, err := generateRandomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"name":       user.Name,
			"avatar_url": user.AvatarURL,
		},
		"organizations": organizations,
		"session_token": sessionToken,
		"expires_at":    expiry,
	})
}

// RefreshTokenRequest represents the request for refreshing tokens
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken refreshes the access token
// @Summary Refresh access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} map[string]interface{}
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx := c.Request.Context()

	// Get user
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// Decrypt refresh token
	refreshToken, err := h.encrypter.Decrypt(user.AtlassianRefreshTokenEncrypted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt refresh token"})
		return
	}

	// Refresh tokens
	tokens, err := h.oauthClient.RefreshToken(ctx, refreshToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Encrypt new tokens
	encryptedAccess, err := h.encrypter.Encrypt(tokens.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt access token"})
		return
	}

	encryptedRefresh, err := h.encrypter.Encrypt(tokens.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt refresh token"})
		return
	}

	expiry := atlassian.TokenExpiry(tokens.ExpiresIn)

	// Update user tokens
	if err := h.userRepo.UpdateAtlassianTokens(ctx, userID, encryptedAccess, encryptedRefresh, &expiry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"expires_at": expiry,
	})
}

// GetCurrentUser returns the current authenticated user
// @Summary Get current user
// @Tags Auth
// @Produce json
// @Success 200 {object} models.User
// @Router /auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Get user's organizations
	orgs, err := h.orgRepo.GetUserOrganizations(c.Request.Context(), userID)
	if err != nil {
		orgs = []*models.OrganizationMember{}
	}

	c.JSON(http.StatusOK, gin.H{
		"user": map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"name":       user.Name,
			"avatar_url": user.AvatarURL,
			"created_at": user.CreatedAt,
		},
		"organizations": orgs,
	})
}

// Logout logs out the user
// @Summary Logout
// @Tags Auth
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a real implementation, invalidate the session token
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// generateRandomState generates a random state string for OAuth
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
