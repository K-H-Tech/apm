package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/K-H-Tech/apm/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthConfig holds configuration for the auth middleware
type AuthConfig struct {
	UserRepo       *repository.UserRepository
	OrgRepo        *repository.OrganizationRepository
	SkipPaths      []string
	TokenDecrypter TokenDecrypter
}

// TokenDecrypter interface for decrypting tokens
type TokenDecrypter interface {
	Decrypt(ciphertext string) (string, error)
}

// AuthMiddleware creates authentication middleware
func AuthMiddleware(config AuthConfig) gin.HandlerFunc {
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		// Check if path should be skipped
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Check for paths that start with skip patterns
		for path := range skipPaths {
			if strings.HasSuffix(path, "*") {
				prefix := strings.TrimSuffix(path, "*")
				if strings.HasPrefix(c.Request.URL.Path, prefix) {
					c.Next()
					return
				}
			}
		}

		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			return
		}

		// Parse bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			return
		}

		token := parts[1]

		// Parse user ID from token (in production, use JWT validation)
		userID, err := uuid.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}

		// Get user from database
		user, err := config.UserRepo.GetByID(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "user not found",
			})
			return
		}

		// Check if Atlassian tokens are expired
		if user.AtlassianTokenExpiry != nil && user.AtlassianTokenExpiry.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "atlassian token expired",
				"code":    "TOKEN_EXPIRED",
				"message": "Please refresh your Atlassian connection",
			})
			return
		}

		// Set user info in context
		c.Set("user_id", user.ID)
		c.Set("user_email", user.Email)
		c.Set("user", user)

		// Get organization from header or query param
		orgIDStr := c.GetHeader("X-Organization-ID")
		if orgIDStr == "" {
			orgIDStr = c.Query("organization_id")
		}

		if orgIDStr != "" {
			orgID, err := uuid.Parse(orgIDStr)
			if err == nil {
				// Verify user belongs to organization
				orgs, _ := config.OrgRepo.GetUserOrganizations(c.Request.Context(), user.ID)
				for _, org := range orgs {
					if org.OrganizationID == orgID {
						c.Set("organization_id", orgID)
						c.Set("organization_role", org.Role)
						break
					}
				}
			}
		}

		c.Next()
	}
}

// RequireOrganization ensures an organization is selected
func RequireOrganization() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("organization_id"); !exists {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "organization_id required",
			})
			return
		}
		c.Next()
	}
}

// RequireRole ensures the user has the required role
func RequireRole(roles ...string) gin.HandlerFunc {
	roleMap := make(map[string]bool)
	for _, role := range roles {
		roleMap[role] = true
	}

	return func(c *gin.Context) {
		role, exists := c.Get("organization_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "no role in organization",
			})
			return
		}

		roleStr, ok := role.(string)
		if !ok || !roleMap[roleStr] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}

// OptionalAuth allows both authenticated and unauthenticated requests
func OptionalAuth(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Try to authenticate
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Next()
			return
		}

		token := parts[1]
		userID, err := uuid.Parse(token)
		if err != nil {
			c.Next()
			return
		}

		user, err := config.UserRepo.GetByID(c.Request.Context(), userID)
		if err != nil {
			c.Next()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("user_email", user.Email)
		c.Set("user", user)

		c.Next()
	}
}
