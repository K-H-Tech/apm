package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/K-H-Tech/apm/internal/models"
	"github.com/google/uuid"
)

var (
	// ErrPRDNotFound is returned when a PRD is not found
	ErrPRDNotFound = errors.New("PRD not found")

	// ErrVersionNotFound is returned when a PRD version is not found
	ErrVersionNotFound = errors.New("PRD version not found")
)

// PRDRepository implements contract.PRDRepository
type PRDRepository struct {
	db *sql.DB
}

// NewPRDRepository creates a new PRD repository
func NewPRDRepository(db *sql.DB) *PRDRepository {
	return &PRDRepository{db: db}
}

// Create creates a new PRD
func (r *PRDRepository) Create(ctx context.Context, prd *models.PRD) error {
	contentJSON, err := json.Marshal(prd.Content)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO prds (
			id, organization_id, template_id, title, status, content,
			jira_epic_key, confluence_page_id, owner_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	if prd.ID == uuid.Nil {
		prd.ID = uuid.New()
	}
	prd.CreatedAt = now
	prd.UpdatedAt = now

	_, err = r.db.ExecContext(ctx, query,
		prd.ID,
		prd.OrganizationID,
		prd.TemplateID,
		prd.Title,
		prd.Status,
		contentJSON,
		prd.JiraEpicKey,
		prd.ConfluencePageID,
		prd.OwnerID,
		prd.CreatedAt,
		prd.UpdatedAt,
	)

	return err
}

// GetByID retrieves a PRD by ID
func (r *PRDRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.PRD, error) {
	query := `
		SELECT id, organization_id, template_id, title, status, content,
			   jira_epic_key, confluence_page_id, owner_id, created_at, updated_at
		FROM prds
		WHERE id = $1
	`

	var prd models.PRD
	var contentJSON []byte
	var templateID, jiraEpicKey, confluencePageID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&prd.ID,
		&prd.OrganizationID,
		&templateID,
		&prd.Title,
		&prd.Status,
		&contentJSON,
		&jiraEpicKey,
		&confluencePageID,
		&prd.OwnerID,
		&prd.CreatedAt,
		&prd.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrPRDNotFound
	}
	if err != nil {
		return nil, err
	}

	if templateID.Valid {
		tid, _ := uuid.Parse(templateID.String)
		prd.TemplateID = &tid
	}
	prd.JiraEpicKey = jiraEpicKey.String
	prd.ConfluencePageID = confluencePageID.String

	if err := json.Unmarshal(contentJSON, &prd.Content); err != nil {
		return nil, err
	}

	return &prd, nil
}

// Update updates an existing PRD
func (r *PRDRepository) Update(ctx context.Context, prd *models.PRD) error {
	contentJSON, err := json.Marshal(prd.Content)
	if err != nil {
		return err
	}

	query := `
		UPDATE prds SET
			template_id = $2,
			title = $3,
			status = $4,
			content = $5,
			jira_epic_key = $6,
			confluence_page_id = $7,
			updated_at = $8
		WHERE id = $1
	`

	prd.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		prd.ID,
		prd.TemplateID,
		prd.Title,
		prd.Status,
		contentJSON,
		prd.JiraEpicKey,
		prd.ConfluencePageID,
		prd.UpdatedAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPRDNotFound
	}

	return nil
}

// Delete deletes a PRD by ID
func (r *PRDRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM prds WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPRDNotFound
	}

	return nil
}

// ListByOrganization lists PRDs for an organization with pagination
func (r *PRDRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.PRD, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM prds WHERE organization_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get PRDs
	query := `
		SELECT id, organization_id, template_id, title, status, content,
			   jira_epic_key, confluence_page_id, owner_id, created_at, updated_at
		FROM prds
		WHERE organization_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	prds := make([]*models.PRD, 0)
	for rows.Next() {
		var prd models.PRD
		var contentJSON []byte
		var templateID, jiraEpicKey, confluencePageID sql.NullString

		if err := rows.Scan(
			&prd.ID,
			&prd.OrganizationID,
			&templateID,
			&prd.Title,
			&prd.Status,
			&contentJSON,
			&jiraEpicKey,
			&confluencePageID,
			&prd.OwnerID,
			&prd.CreatedAt,
			&prd.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		if templateID.Valid {
			tid, _ := uuid.Parse(templateID.String)
			prd.TemplateID = &tid
		}
		prd.JiraEpicKey = jiraEpicKey.String
		prd.ConfluencePageID = confluencePageID.String

		if err := json.Unmarshal(contentJSON, &prd.Content); err != nil {
			return nil, 0, err
		}

		prds = append(prds, &prd)
	}

	return prds, total, rows.Err()
}

// ListByOwner lists PRDs owned by a user
func (r *PRDRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*models.PRD, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM prds WHERE owner_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, ownerID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get PRDs
	query := `
		SELECT id, organization_id, template_id, title, status, content,
			   jira_epic_key, confluence_page_id, owner_id, created_at, updated_at
		FROM prds
		WHERE owner_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	prds := make([]*models.PRD, 0)
	for rows.Next() {
		var prd models.PRD
		var contentJSON []byte
		var templateID, jiraEpicKey, confluencePageID sql.NullString

		if err := rows.Scan(
			&prd.ID,
			&prd.OrganizationID,
			&templateID,
			&prd.Title,
			&prd.Status,
			&contentJSON,
			&jiraEpicKey,
			&confluencePageID,
			&prd.OwnerID,
			&prd.CreatedAt,
			&prd.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		if templateID.Valid {
			tid, _ := uuid.Parse(templateID.String)
			prd.TemplateID = &tid
		}
		prd.JiraEpicKey = jiraEpicKey.String
		prd.ConfluencePageID = confluencePageID.String

		if err := json.Unmarshal(contentJSON, &prd.Content); err != nil {
			return nil, 0, err
		}

		prds = append(prds, &prd)
	}

	return prds, total, rows.Err()
}

// ListByStatus lists PRDs with a specific status
func (r *PRDRepository) ListByStatus(ctx context.Context, orgID uuid.UUID, status models.PRDStatus, limit, offset int) ([]*models.PRD, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM prds WHERE organization_id = $1 AND status = $2`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, orgID, status).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get PRDs
	query := `
		SELECT id, organization_id, template_id, title, status, content,
			   jira_epic_key, confluence_page_id, owner_id, created_at, updated_at
		FROM prds
		WHERE organization_id = $1 AND status = $2
		ORDER BY updated_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, status, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	prds := make([]*models.PRD, 0)
	for rows.Next() {
		var prd models.PRD
		var contentJSON []byte
		var templateID, jiraEpicKey, confluencePageID sql.NullString

		if err := rows.Scan(
			&prd.ID,
			&prd.OrganizationID,
			&templateID,
			&prd.Title,
			&prd.Status,
			&contentJSON,
			&jiraEpicKey,
			&confluencePageID,
			&prd.OwnerID,
			&prd.CreatedAt,
			&prd.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		if templateID.Valid {
			tid, _ := uuid.Parse(templateID.String)
			prd.TemplateID = &tid
		}
		prd.JiraEpicKey = jiraEpicKey.String
		prd.ConfluencePageID = confluencePageID.String

		if err := json.Unmarshal(contentJSON, &prd.Content); err != nil {
			return nil, 0, err
		}

		prds = append(prds, &prd)
	}

	return prds, total, rows.Err()
}

// CreateVersion creates a new version of a PRD
func (r *PRDRepository) CreateVersion(ctx context.Context, version *models.PRDVersion) error {
	contentJSON, err := json.Marshal(version.Content)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO prd_versions (id, prd_id, version_number, content, changed_by, change_summary, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	if version.ID == uuid.Nil {
		version.ID = uuid.New()
	}
	version.CreatedAt = time.Now()

	_, err = r.db.ExecContext(ctx, query,
		version.ID,
		version.PRDID,
		version.VersionNumber,
		contentJSON,
		version.ChangedBy,
		version.ChangeSummary,
		version.CreatedAt,
	)

	return err
}

// GetVersion retrieves a specific version of a PRD
func (r *PRDRepository) GetVersion(ctx context.Context, prdID uuid.UUID, versionNumber int) (*models.PRDVersion, error) {
	query := `
		SELECT id, prd_id, version_number, content, changed_by, change_summary, created_at
		FROM prd_versions
		WHERE prd_id = $1 AND version_number = $2
	`

	var version models.PRDVersion
	var contentJSON []byte

	err := r.db.QueryRowContext(ctx, query, prdID, versionNumber).Scan(
		&version.ID,
		&version.PRDID,
		&version.VersionNumber,
		&contentJSON,
		&version.ChangedBy,
		&version.ChangeSummary,
		&version.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrVersionNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(contentJSON, &version.Content); err != nil {
		return nil, err
	}

	return &version, nil
}

// ListVersions lists all versions of a PRD
func (r *PRDRepository) ListVersions(ctx context.Context, prdID uuid.UUID) ([]*models.PRDVersion, error) {
	query := `
		SELECT id, prd_id, version_number, content, changed_by, change_summary, created_at
		FROM prd_versions
		WHERE prd_id = $1
		ORDER BY version_number DESC
	`

	rows, err := r.db.QueryContext(ctx, query, prdID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make([]*models.PRDVersion, 0)
	for rows.Next() {
		var version models.PRDVersion
		var contentJSON []byte

		if err := rows.Scan(
			&version.ID,
			&version.PRDID,
			&version.VersionNumber,
			&contentJSON,
			&version.ChangedBy,
			&version.ChangeSummary,
			&version.CreatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(contentJSON, &version.Content); err != nil {
			return nil, err
		}

		versions = append(versions, &version)
	}

	return versions, rows.Err()
}

// GetLatestVersionNumber gets the latest version number for a PRD
func (r *PRDRepository) GetLatestVersionNumber(ctx context.Context, prdID uuid.UUID) (int, error) {
	query := `SELECT COALESCE(MAX(version_number), 0) FROM prd_versions WHERE prd_id = $1`
	var versionNumber int
	err := r.db.QueryRowContext(ctx, query, prdID).Scan(&versionNumber)
	return versionNumber, err
}

// Search searches PRDs by title or content
func (r *PRDRepository) Search(ctx context.Context, orgID uuid.UUID, query string, limit, offset int) ([]*models.PRD, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*) FROM prds
		WHERE organization_id = $1
		AND (title ILIKE $2 OR content::text ILIKE $2)
	`
	searchPattern := "%" + query + "%"
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, orgID, searchPattern).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get PRDs
	searchQuery := `
		SELECT id, organization_id, template_id, title, status, content,
			   jira_epic_key, confluence_page_id, owner_id, created_at, updated_at
		FROM prds
		WHERE organization_id = $1
		AND (title ILIKE $2 OR content::text ILIKE $2)
		ORDER BY updated_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, searchQuery, orgID, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	prds := make([]*models.PRD, 0)
	for rows.Next() {
		var prd models.PRD
		var contentJSON []byte
		var templateID, jiraEpicKey, confluencePageID sql.NullString

		if err := rows.Scan(
			&prd.ID,
			&prd.OrganizationID,
			&templateID,
			&prd.Title,
			&prd.Status,
			&contentJSON,
			&jiraEpicKey,
			&confluencePageID,
			&prd.OwnerID,
			&prd.CreatedAt,
			&prd.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		if templateID.Valid {
			tid, _ := uuid.Parse(templateID.String)
			prd.TemplateID = &tid
		}
		prd.JiraEpicKey = jiraEpicKey.String
		prd.ConfluencePageID = confluencePageID.String

		if err := json.Unmarshal(contentJSON, &prd.Content); err != nil {
			return nil, 0, err
		}

		prds = append(prds, &prd)
	}

	return prds, total, rows.Err()
}
