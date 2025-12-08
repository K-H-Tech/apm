package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/K-H-Tech/apm/internal/models"
	"github.com/google/uuid"
)

var (
	// ErrTemplateNotFound is returned when a template is not found
	ErrTemplateNotFound = errors.New("template not found")
)

// TemplateRepository implements contract.TemplateRepository
type TemplateRepository struct {
	db *sql.DB
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(db *sql.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// Create creates a new template
func (r *TemplateRepository) Create(ctx context.Context, template *models.PRDTemplate) error {
	schemaJSON, err := json.Marshal(template.ContentSchema)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO prd_templates (
			id, organization_id, name, description, type, content_schema, is_system_template, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	if template.ID == uuid.Nil {
		template.ID = uuid.New()
	}
	template.CreatedAt = now
	template.UpdatedAt = now

	_, err = r.db.ExecContext(ctx, query,
		template.ID,
		template.OrganizationID,
		template.Name,
		template.Description,
		template.Type,
		schemaJSON,
		template.IsSystemTemplate,
		template.CreatedAt,
		template.UpdatedAt,
	)

	return err
}

// GetByID retrieves a template by ID
func (r *TemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.PRDTemplate, error) {
	query := `
		SELECT id, organization_id, name, description, type, content_schema, is_system_template, created_at, updated_at
		FROM prd_templates
		WHERE id = $1
	`

	var template models.PRDTemplate
	var schemaJSON []byte
	var orgID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&template.ID,
		&orgID,
		&template.Name,
		&template.Description,
		&template.Type,
		&schemaJSON,
		&template.IsSystemTemplate,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTemplateNotFound
	}
	if err != nil {
		return nil, err
	}

	if orgID.Valid {
		oid, err := uuid.Parse(orgID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid organization_id UUID: %w", err)
		}
		template.OrganizationID = &oid
	}

	if err := json.Unmarshal(schemaJSON, &template.ContentSchema); err != nil {
		return nil, err
	}

	return &template, nil
}

// Update updates an existing template
func (r *TemplateRepository) Update(ctx context.Context, template *models.PRDTemplate) error {
	schemaJSON, err := json.Marshal(template.ContentSchema)
	if err != nil {
		return err
	}

	query := `
		UPDATE prd_templates SET
			name = $2,
			description = $3,
			type = $4,
			content_schema = $5,
			is_system_template = $6,
			updated_at = $7
		WHERE id = $1
	`

	template.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		template.ID,
		template.Name,
		template.Description,
		template.Type,
		schemaJSON,
		template.IsSystemTemplate,
		template.UpdatedAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTemplateNotFound
	}

	return nil
}

// Delete deletes a template by ID
func (r *TemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM prd_templates WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTemplateNotFound
	}

	return nil
}

// ListByOrganization lists templates for an organization (including system templates)
func (r *TemplateRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*models.PRDTemplate, error) {
	query := `
		SELECT id, organization_id, name, description, type, content_schema, is_system_template, created_at, updated_at
		FROM prd_templates
		WHERE organization_id = $1 OR organization_id IS NULL
		ORDER BY is_system_template DESC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]*models.PRDTemplate, 0)
	for rows.Next() {
		var template models.PRDTemplate
		var schemaJSON []byte
		var organizationID sql.NullString

		if err := rows.Scan(
			&template.ID,
			&organizationID,
			&template.Name,
			&template.Description,
			&template.Type,
			&schemaJSON,
			&template.IsSystemTemplate,
			&template.CreatedAt,
			&template.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if organizationID.Valid {
			oid, err := uuid.Parse(organizationID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid organization_id UUID: %w", err)
			}
			template.OrganizationID = &oid
		}

		if err := json.Unmarshal(schemaJSON, &template.ContentSchema); err != nil {
			return nil, err
		}

		templates = append(templates, &template)
	}

	return templates, rows.Err()
}

// ListByType lists templates by type
func (r *TemplateRepository) ListByType(ctx context.Context, orgID uuid.UUID, templateType models.TemplateType) ([]*models.PRDTemplate, error) {
	query := `
		SELECT id, organization_id, name, description, type, content_schema, is_system_template, created_at, updated_at
		FROM prd_templates
		WHERE (organization_id = $1 OR organization_id IS NULL) AND type = $2
		ORDER BY is_system_template DESC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, templateType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]*models.PRDTemplate, 0)
	for rows.Next() {
		var template models.PRDTemplate
		var schemaJSON []byte
		var organizationID sql.NullString

		if err := rows.Scan(
			&template.ID,
			&organizationID,
			&template.Name,
			&template.Description,
			&template.Type,
			&schemaJSON,
			&template.IsSystemTemplate,
			&template.CreatedAt,
			&template.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if organizationID.Valid {
			oid, err := uuid.Parse(organizationID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid organization_id UUID: %w", err)
			}
			template.OrganizationID = &oid
		}

		if err := json.Unmarshal(schemaJSON, &template.ContentSchema); err != nil {
			return nil, err
		}

		templates = append(templates, &template)
	}

	return templates, rows.Err()
}

// GetDefault retrieves the default/system template for a type
func (r *TemplateRepository) GetDefault(ctx context.Context, orgID uuid.UUID, templateType models.TemplateType) (*models.PRDTemplate, error) {
	query := `
		SELECT id, organization_id, name, description, type, content_schema, is_system_template, created_at, updated_at
		FROM prd_templates
		WHERE (organization_id = $1 OR organization_id IS NULL)
		AND type = $2
		AND is_system_template = true
		ORDER BY organization_id DESC NULLS LAST
		LIMIT 1
	`

	var template models.PRDTemplate
	var schemaJSON []byte
	var organizationID sql.NullString

	err := r.db.QueryRowContext(ctx, query, orgID, templateType).Scan(
		&template.ID,
		&organizationID,
		&template.Name,
		&template.Description,
		&template.Type,
		&schemaJSON,
		&template.IsSystemTemplate,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTemplateNotFound
	}
	if err != nil {
		return nil, err
	}

	if organizationID.Valid {
		oid, err := uuid.Parse(organizationID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid organization_id UUID: %w", err)
		}
		template.OrganizationID = &oid
	}

	if err := json.Unmarshal(schemaJSON, &template.ContentSchema); err != nil {
		return nil, err
	}

	return &template, nil
}

// ListSystemTemplates lists all system templates (no organization)
func (r *TemplateRepository) ListSystemTemplates(ctx context.Context) ([]*models.PRDTemplate, error) {
	query := `
		SELECT id, organization_id, name, description, type, content_schema, is_system_template, created_at, updated_at
		FROM prd_templates
		WHERE organization_id IS NULL AND is_system_template = true
		ORDER BY type, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]*models.PRDTemplate, 0)
	for rows.Next() {
		var template models.PRDTemplate
		var schemaJSON []byte
		var organizationID sql.NullString

		if err := rows.Scan(
			&template.ID,
			&organizationID,
			&template.Name,
			&template.Description,
			&template.Type,
			&schemaJSON,
			&template.IsSystemTemplate,
			&template.CreatedAt,
			&template.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(schemaJSON, &template.ContentSchema); err != nil {
			return nil, err
		}

		templates = append(templates, &template)
	}

	return templates, rows.Err()
}

// Clone creates a copy of a template for an organization
func (r *TemplateRepository) Clone(ctx context.Context, sourceID, targetOrgID uuid.UUID) (*models.PRDTemplate, error) {
	source, err := r.GetByID(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	clone := &models.PRDTemplate{
		ID:               uuid.New(),
		OrganizationID:   &targetOrgID,
		Name:             source.Name + " (Copy)",
		Description:      source.Description,
		Type:             source.Type,
		ContentSchema:    source.ContentSchema,
		IsSystemTemplate: false,
	}

	if err := r.Create(ctx, clone); err != nil {
		return nil, err
	}

	return clone, nil
}
