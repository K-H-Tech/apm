package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/K-H-Tech/apm/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrOrganizationNotFound is returned when an organization is not found
	ErrOrganizationNotFound = errors.New("organization not found")

	// ErrUserAlreadyExists is returned when trying to create a duplicate user
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrOrganizationAlreadyExists is returned when trying to create a duplicate organization
	ErrOrganizationAlreadyExists = errors.New("organization already exists")

	// ErrMemberNotFound is returned when an organization member is not found
	ErrMemberNotFound = errors.New("member not found in organization")
)

// UserRepository implements contract.UserRepository
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, email, name, avatar_url, atlassian_account_id,
			atlassian_access_token_encrypted, atlassian_refresh_token_encrypted,
			atlassian_token_expiry, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.AvatarURL,
		user.AtlassianAccountID,
		user.AtlassianAccessTokenEncrypted,
		user.AtlassianRefreshTokenEncrypted,
		user.AtlassianTokenExpiry,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		// Check for PostgreSQL unique constraint violation (code 23505)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrUserAlreadyExists
		}
		return err
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, name, avatar_url, atlassian_account_id,
			   atlassian_access_token_encrypted, atlassian_refresh_token_encrypted,
			   atlassian_token_expiry, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	var avatarURL, atlassianAccountID sql.NullString
	var accessToken, refreshToken sql.NullString
	var tokenExpiry sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&avatarURL,
		&atlassianAccountID,
		&accessToken,
		&refreshToken,
		&tokenExpiry,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	user.AvatarURL = avatarURL.String
	user.AtlassianAccountID = atlassianAccountID.String
	user.AtlassianAccessTokenEncrypted = accessToken.String
	user.AtlassianRefreshTokenEncrypted = refreshToken.String
	if tokenExpiry.Valid {
		user.AtlassianTokenExpiry = &tokenExpiry.Time
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, name, avatar_url, atlassian_account_id,
			   atlassian_access_token_encrypted, atlassian_refresh_token_encrypted,
			   atlassian_token_expiry, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	var avatarURL, atlassianAccountID sql.NullString
	var accessToken, refreshToken sql.NullString
	var tokenExpiry sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&avatarURL,
		&atlassianAccountID,
		&accessToken,
		&refreshToken,
		&tokenExpiry,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	user.AvatarURL = avatarURL.String
	user.AtlassianAccountID = atlassianAccountID.String
	user.AtlassianAccessTokenEncrypted = accessToken.String
	user.AtlassianRefreshTokenEncrypted = refreshToken.String
	if tokenExpiry.Valid {
		user.AtlassianTokenExpiry = &tokenExpiry.Time
	}

	return &user, nil
}

// GetByAtlassianAccountID retrieves a user by Atlassian account ID
func (r *UserRepository) GetByAtlassianAccountID(ctx context.Context, accountID string) (*models.User, error) {
	query := `
		SELECT id, email, name, avatar_url, atlassian_account_id,
			   atlassian_access_token_encrypted, atlassian_refresh_token_encrypted,
			   atlassian_token_expiry, created_at, updated_at
		FROM users
		WHERE atlassian_account_id = $1
	`

	var user models.User
	var avatarURL, atlassianAccountID sql.NullString
	var accessToken, refreshToken sql.NullString
	var tokenExpiry sql.NullTime

	err := r.db.QueryRowContext(ctx, query, accountID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&avatarURL,
		&atlassianAccountID,
		&accessToken,
		&refreshToken,
		&tokenExpiry,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	user.AvatarURL = avatarURL.String
	user.AtlassianAccountID = atlassianAccountID.String
	user.AtlassianAccessTokenEncrypted = accessToken.String
	user.AtlassianRefreshTokenEncrypted = refreshToken.String
	if tokenExpiry.Valid {
		user.AtlassianTokenExpiry = &tokenExpiry.Time
	}

	return &user, nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET
			name = $2,
			avatar_url = $3,
			atlassian_account_id = $4,
			atlassian_access_token_encrypted = $5,
			atlassian_refresh_token_encrypted = $6,
			atlassian_token_expiry = $7,
			updated_at = $8
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Name,
		user.AvatarURL,
		user.AtlassianAccountID,
		user.AtlassianAccessTokenEncrypted,
		user.AtlassianRefreshTokenEncrypted,
		user.AtlassianTokenExpiry,
		user.UpdatedAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateAtlassianTokens updates just the Atlassian tokens for a user
func (r *UserRepository) UpdateAtlassianTokens(ctx context.Context, userID uuid.UUID, accessToken, refreshToken string, expiry *time.Time) error {
	query := `
		UPDATE users SET
			atlassian_access_token_encrypted = $2,
			atlassian_refresh_token_encrypted = $3,
			atlassian_token_expiry = $4,
			updated_at = $5
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		userID,
		accessToken,
		refreshToken,
		expiry,
		time.Now(),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// Delete deletes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// OrganizationRepository implements organization data access
type OrganizationRepository struct {
	db *sql.DB
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(db *sql.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// Create creates a new organization
func (r *OrganizationRepository) Create(ctx context.Context, org *models.Organization) error {
	query := `
		INSERT INTO organizations (
			id, name, jira_cloud_id, jira_base_url, confluence_cloud_id,
			confluence_base_url, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	if org.ID == uuid.Nil {
		org.ID = uuid.New()
	}
	org.CreatedAt = now
	org.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		org.ID,
		org.Name,
		org.JiraCloudID,
		org.JiraBaseURL,
		org.ConfluenceCloudID,
		org.ConfluenceBaseURL,
		org.CreatedAt,
		org.UpdatedAt,
	)

	if err != nil {
		// Check for PostgreSQL unique constraint violation (code 23505)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrOrganizationAlreadyExists
		}
		return err
	}

	return nil
}

// GetByID retrieves an organization by ID
func (r *OrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	query := `
		SELECT id, name, jira_cloud_id, jira_base_url, confluence_cloud_id,
			   confluence_base_url, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`

	var org models.Organization
	var jiraCloudID, jiraBaseURL, confCloudID, confBaseURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&jiraCloudID,
		&jiraBaseURL,
		&confCloudID,
		&confBaseURL,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrOrganizationNotFound
	}
	if err != nil {
		return nil, err
	}

	org.JiraCloudID = jiraCloudID.String
	org.JiraBaseURL = jiraBaseURL.String
	org.ConfluenceCloudID = confCloudID.String
	org.ConfluenceBaseURL = confBaseURL.String

	return &org, nil
}

// GetByJiraCloudID retrieves an organization by Jira Cloud ID
func (r *OrganizationRepository) GetByJiraCloudID(ctx context.Context, cloudID string) (*models.Organization, error) {
	query := `
		SELECT id, name, jira_cloud_id, jira_base_url, confluence_cloud_id,
			   confluence_base_url, created_at, updated_at
		FROM organizations
		WHERE jira_cloud_id = $1
	`

	var org models.Organization
	var jiraCloudID, jiraBaseURL, confCloudID, confBaseURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, cloudID).Scan(
		&org.ID,
		&org.Name,
		&jiraCloudID,
		&jiraBaseURL,
		&confCloudID,
		&confBaseURL,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrOrganizationNotFound
	}
	if err != nil {
		return nil, err
	}

	org.JiraCloudID = jiraCloudID.String
	org.JiraBaseURL = jiraBaseURL.String
	org.ConfluenceCloudID = confCloudID.String
	org.ConfluenceBaseURL = confBaseURL.String

	return &org, nil
}

// Update updates an existing organization
func (r *OrganizationRepository) Update(ctx context.Context, org *models.Organization) error {
	query := `
		UPDATE organizations SET
			name = $2,
			jira_cloud_id = $3,
			jira_base_url = $4,
			confluence_cloud_id = $5,
			confluence_base_url = $6,
			updated_at = $7
		WHERE id = $1
	`

	org.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		org.ID,
		org.Name,
		org.JiraCloudID,
		org.JiraBaseURL,
		org.ConfluenceCloudID,
		org.ConfluenceBaseURL,
		org.UpdatedAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrOrganizationNotFound
	}

	return nil
}

// Delete deletes an organization by ID
func (r *OrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM organizations WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrOrganizationNotFound
	}

	return nil
}

// GetUserOrganizations retrieves all organizations a user belongs to
func (r *OrganizationRepository) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*models.OrganizationMember, error) {
	query := `
		SELECT om.id, om.organization_id, om.user_id, om.role, om.created_at,
			   o.name as org_name
		FROM organization_members om
		JOIN organizations o ON o.id = om.organization_id
		WHERE om.user_id = $1
		ORDER BY o.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]*models.OrganizationMember, 0)
	for rows.Next() {
		var member models.OrganizationMember
		var orgName string

		if err := rows.Scan(
			&member.ID,
			&member.OrganizationID,
			&member.UserID,
			&member.Role,
			&member.CreatedAt,
			&orgName,
		); err != nil {
			return nil, err
		}

		members = append(members, &member)
	}

	return members, rows.Err()
}

// AddUserToOrganization adds a user to an organization
func (r *OrganizationRepository) AddUserToOrganization(ctx context.Context, member *models.OrganizationMember) error {
	query := `
		INSERT INTO organization_members (id, organization_id, user_id, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	if member.ID == uuid.Nil {
		member.ID = uuid.New()
	}
	member.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		member.ID,
		member.OrganizationID,
		member.UserID,
		member.Role,
		member.CreatedAt,
	)

	return err
}

// RemoveUserFromOrganization removes a user from an organization
func (r *OrganizationRepository) RemoveUserFromOrganization(ctx context.Context, orgID, userID uuid.UUID) error {
	query := `DELETE FROM organization_members WHERE organization_id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, orgID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrMemberNotFound
	}

	return nil
}

// GetOrganizationMembers retrieves all members of an organization
func (r *OrganizationRepository) GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]*models.OrganizationMember, error) {
	query := `
		SELECT om.id, om.organization_id, om.user_id, om.role, om.created_at,
			   u.email, u.name
		FROM organization_members om
		JOIN users u ON u.id = om.user_id
		WHERE om.organization_id = $1
		ORDER BY u.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]*models.OrganizationMember, 0)
	for rows.Next() {
		var member models.OrganizationMember
		var email, name string

		if err := rows.Scan(
			&member.ID,
			&member.OrganizationID,
			&member.UserID,
			&member.Role,
			&member.CreatedAt,
			&email,
			&name,
		); err != nil {
			return nil, err
		}

		members = append(members, &member)
	}

	return members, rows.Err()
}

// UpdateMemberRole updates a member's role in an organization
func (r *OrganizationRepository) UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role models.OrganizationRole) error {
	query := `
		UPDATE organization_members
		SET role = $3
		WHERE organization_id = $1 AND user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, orgID, userID, role)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("member not found in organization")
	}

	return nil
}
