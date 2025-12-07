package contract

import (
	"context"

	"github.com/K-H-Tech/apm/internal/models"
	"github.com/google/uuid"
)

// OAuthService defines OAuth 2.0 operations for Atlassian
type OAuthService interface {
	// GetAuthorizationURL returns the URL to redirect users for authorization
	GetAuthorizationURL(state string) string

	// ExchangeCode exchanges authorization code for tokens
	ExchangeCode(ctx context.Context, code string) (*OAuthTokens, error)

	// RefreshTokens refreshes an expired access token
	RefreshTokens(ctx context.Context, refreshToken string) (*OAuthTokens, error)

	// GetAccessibleResources returns the Atlassian sites the user has access to
	GetAccessibleResources(ctx context.Context, accessToken string) ([]*AtlassianResource, error)

	// ValidateToken checks if an access token is valid
	ValidateToken(ctx context.Context, accessToken string) (bool, error)

	// RevokeToken revokes an access token
	RevokeToken(ctx context.Context, accessToken string) error
}

// OAuthTokens represents OAuth tokens
type OAuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // Seconds until expiration
	Scope        string `json:"scope"`
}

// AtlassianResource represents an accessible Atlassian site
type AtlassianResource struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	AvatarURL string   `json:"avatarUrl"`
}

// OAuthConfig represents OAuth configuration
type OAuthConfig struct {
	ClientID     string   `yaml:"CLIENT_ID"`
	ClientSecret string   `yaml:"CLIENT_SECRET"`
	RedirectURL  string   `yaml:"REDIRECT_URL"`
	Scopes       []string `yaml:"SCOPES"`
	AuthURL      string   `yaml:"AUTH_URL"`
	TokenURL     string   `yaml:"TOKEN_URL"`
}

// DefaultAtlassianScopes returns the default OAuth scopes for Atlassian
func DefaultAtlassianScopes() []string {
	return []string{
		"read:jira-work",
		"write:jira-work",
		"read:jira-user",
		"read:confluence-content.all",
		"write:confluence-content",
		"read:confluence-space.summary",
		"offline_access", // For refresh tokens
	}
}

// UserService defines user management operations
type UserService interface {
	// User operations
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	// OAuth token management
	SaveAtlassianTokens(ctx context.Context, userID uuid.UUID, tokens *OAuthTokens) error
	GetAtlassianTokens(ctx context.Context, userID uuid.UUID) (*models.AtlassianTokens, error)
	RefreshAtlassianTokensIfNeeded(ctx context.Context, userID uuid.UUID) (*models.AtlassianTokens, error)

	// Organization membership
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*models.Organization, error)
	GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]*models.UserWithOrg, error)
	AddUserToOrganization(ctx context.Context, userID uuid.UUID, orgID uuid.UUID, role string) error
	RemoveUserFromOrganization(ctx context.Context, userID uuid.UUID, orgID uuid.UUID) error
	UpdateUserRole(ctx context.Context, userID uuid.UUID, orgID uuid.UUID, role string) error
}

// UserRepository defines data access operations for users
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Token operations
	UpdateAtlassianTokens(ctx context.Context, userID uuid.UUID, accessToken, refreshToken string, expiresAt interface{}) error
}

// OrganizationService defines organization management operations
type OrganizationService interface {
	Create(ctx context.Context, org *models.Organization, ownerID uuid.UUID) (*models.Organization, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Organization, error)
	Update(ctx context.Context, org *models.Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID) ([]*models.Organization, error)

	// Atlassian configuration
	UpdateAtlassianConfig(ctx context.Context, orgID uuid.UUID, jiraBaseURL, jiraCloudID, confluenceBaseURL, confluenceCloudID string) error
}

// OrganizationRepository defines data access operations for organizations
type OrganizationRepository interface {
	Create(ctx context.Context, org *models.Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Organization, error)
	Update(ctx context.Context, org *models.Organization) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Membership operations
	AddMember(ctx context.Context, orgID, userID uuid.UUID, role string) error
	RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error
	UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role string) error
	GetMembers(ctx context.Context, orgID uuid.UUID) ([]*models.OrganizationMember, error)
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*models.Organization, error)
	IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
	GetMemberRole(ctx context.Context, orgID, userID uuid.UUID) (string, error)
}
