package contract

import (
	"context"

	"github.com/K-H-Tech/apm/internal/models"
	"github.com/google/uuid"
)

// PRDService defines the PRD management operations
type PRDService interface {
	// CRUD operations
	Create(ctx context.Context, orgID uuid.UUID, req *models.CreatePRDRequest, ownerID uuid.UUID) (*models.PRD, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.PRD, error)
	Update(ctx context.Context, id uuid.UUID, req *models.UpdatePRDRequest) (*models.PRD, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter models.PRDFilter) ([]*models.PRD, int, error)

	// AI-powered generation
	GenerateFromNotes(ctx context.Context, orgID uuid.UUID, req *models.GeneratePRDRequest, ownerID uuid.UUID) (*models.PRD, error)
	RefineContent(ctx context.Context, prdID uuid.UUID, req *models.RefinePRDRequest) (*models.PRD, error)
	CheckConsistency(ctx context.Context, prdID uuid.UUID) (*ConsistencyReport, error)

	// Version management
	GetVersionHistory(ctx context.Context, prdID uuid.UUID) ([]*models.PRDVersion, error)
	RestoreVersion(ctx context.Context, prdID uuid.UUID, versionNum int, userID uuid.UUID) (*models.PRD, error)

	// Export operations
	ExportToConfluence(ctx context.Context, prdID uuid.UUID) (string, error) // returns page ID
	SyncToJira(ctx context.Context, prdID uuid.UUID, req *models.SyncPRDToJiraRequest) (*models.SyncPRDToJiraResponse, error)
}

// PRDRepository defines data access operations for PRDs
type PRDRepository interface {
	Create(ctx context.Context, prd *models.PRD) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.PRD, error)
	Update(ctx context.Context, prd *models.PRD) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter models.PRDFilter) ([]*models.PRD, int, error)

	// Version operations
	CreateVersion(ctx context.Context, version *models.PRDVersion) error
	GetVersions(ctx context.Context, prdID uuid.UUID) ([]*models.PRDVersion, error)
	GetVersion(ctx context.Context, prdID uuid.UUID, versionNum int) (*models.PRDVersion, error)

	// Query helpers
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	GetByJiraEpicKey(ctx context.Context, epicKey string) (*models.PRD, error)
}

// ConsistencyReport represents the result of a PRD consistency check
type ConsistencyReport struct {
	IsConsistent bool               `json:"is_consistent"`
	Issues       []ConsistencyIssue `json:"issues"`
	Suggestions  []string           `json:"suggestions"`
	CheckedAt    string             `json:"checked_at"`
}

// ConsistencyIssue represents a single consistency issue
type ConsistencyIssue struct {
	Section     string `json:"section"`
	Issue       string `json:"issue"`
	Severity    string `json:"severity"` // error, warning, info
	Suggestion  string `json:"suggestion,omitempty"`
}

// TemplateService defines template management operations
type TemplateService interface {
	Create(ctx context.Context, orgID uuid.UUID, req *models.CreateTemplateRequest, userID uuid.UUID) (*models.PRDTemplate, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.PRDTemplate, error)
	Update(ctx context.Context, id uuid.UUID, req *models.UpdateTemplateRequest) (*models.PRDTemplate, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter models.TemplateFilter) ([]*models.PRDTemplate, error)
	GetSystemTemplates(ctx context.Context) ([]*models.PRDTemplate, error)
}

// TemplateRepository defines data access operations for templates
type TemplateRepository interface {
	Create(ctx context.Context, template *models.PRDTemplate) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.PRDTemplate, error)
	Update(ctx context.Context, template *models.PRDTemplate) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter models.TemplateFilter) ([]*models.PRDTemplate, error)
	GetSystemTemplates(ctx context.Context) ([]*models.PRDTemplate, error)
}
