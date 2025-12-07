package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the APM system
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	AvatarURL string    `json:"avatar_url,omitempty" db:"avatar_url"`

	// Atlassian OAuth tokens (stored encrypted)
	AtlassianAccountID             string     `json:"-" db:"atlassian_account_id"`
	AtlassianAccessTokenEncrypted  string     `json:"-" db:"atlassian_access_token_encrypted"`
	AtlassianRefreshTokenEncrypted string     `json:"-" db:"atlassian_refresh_token_encrypted"`
	AtlassianTokenExpiry           *time.Time `json:"-" db:"atlassian_token_expiry"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Organization represents a team or company using APM
type Organization struct {
	ID   uuid.UUID `json:"id" db:"id"`
	Name string    `json:"name" db:"name"`

	// Atlassian configuration
	JiraBaseURL       string `json:"jira_base_url,omitempty" db:"jira_base_url"`
	JiraCloudID       string `json:"jira_cloud_id,omitempty" db:"jira_cloud_id"`
	ConfluenceBaseURL string `json:"confluence_base_url,omitempty" db:"confluence_base_url"`
	ConfluenceCloudID string `json:"confluence_cloud_id,omitempty" db:"confluence_cloud_id"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// OrganizationMember represents a user's membership in an organization
type OrganizationMember struct {
	ID             uuid.UUID        `json:"id" db:"id"`
	OrganizationID uuid.UUID        `json:"organization_id" db:"organization_id"`
	UserID         uuid.UUID        `json:"user_id" db:"user_id"`
	Role           OrganizationRole `json:"role" db:"role"`
	CreatedAt      time.Time        `json:"created_at" db:"created_at"`
}

// OrganizationRole represents a user's role in an organization
type OrganizationRole string

const (
	RoleOwner  OrganizationRole = "owner"
	RoleAdmin  OrganizationRole = "admin"
	RoleMember OrganizationRole = "member"
)


// UserWithOrg represents a user with their organization context
type UserWithOrg struct {
	User         *User            `json:"user"`
	Organization *Organization    `json:"organization"`
	Role         OrganizationRole `json:"role"`
}

// AtlassianTokens holds decrypted OAuth tokens for Atlassian API calls
type AtlassianTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// IsExpired checks if the access token has expired
func (t *AtlassianTokens) IsExpired() bool {
	// Consider expired if less than 5 minutes remaining
	return time.Now().Add(5 * time.Minute).After(t.ExpiresAt)
}
