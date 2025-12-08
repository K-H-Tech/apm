package contract

import (
	"context"

	"github.com/K-H-Tech/apm/internal/models"
	"github.com/google/uuid"
)

// PriorityService defines prioritization framework operations
type PriorityService interface {
	// RICE scoring
	CalculateRICE(ctx context.Context, req *models.CalculateRICERequest, userID uuid.UUID) (*models.PriorityScore, error)

	// MoSCoW categorization
	SetMoSCoW(ctx context.Context, req *models.SetMoSCoWRequest, userID uuid.UUID) (*models.PriorityScore, error)

	// ICE scoring
	CalculateICE(ctx context.Context, req *models.CalculateICERequest, userID uuid.UUID) (*models.PriorityScore, error)

	// Weighted scoring
	CalculateWeighted(ctx context.Context, req *models.CalculateWeightedRequest, userID uuid.UUID) (*models.PriorityScore, error)

	// Score retrieval
	GetScoreByID(ctx context.Context, id uuid.UUID) (*models.PriorityScore, error)
	GetScoresForPRD(ctx context.Context, prdID uuid.UUID) ([]*models.PriorityScore, error)
	DeleteScore(ctx context.Context, id uuid.UUID) error

	// Rankings
	GetRankedPRDs(ctx context.Context, req *models.PriorityRankingRequest) ([]*models.PRDWithScore, int, error)
}

// PriorityRepository defines data access operations for priority scores
type PriorityRepository interface {
	Create(ctx context.Context, score *models.PriorityScore) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.PriorityScore, error)
	Update(ctx context.Context, score *models.PriorityScore) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Get scores for a PRD
	GetByPRDID(ctx context.Context, prdID uuid.UUID) ([]*models.PriorityScore, error)

	// Get score by PRD and framework (unique constraint)
	GetByPRDAndFramework(ctx context.Context, prdID uuid.UUID, framework models.PriorityFramework) (*models.PriorityScore, error)

	// Upsert score (update if exists, create if not)
	Upsert(ctx context.Context, score *models.PriorityScore) error

	// Get ranked PRDs by framework
	GetRankedByFramework(ctx context.Context, orgID uuid.UUID, framework models.PriorityFramework, limit, offset int) ([]*models.PRDWithScore, int, error)
}

// RICECalculator defines RICE score calculation
type RICECalculator interface {
	// Calculate computes the RICE score
	Calculate(params *models.RICEParams) *models.RICEScore
}

// ICECalculator defines ICE score calculation
type ICECalculator interface {
	// Calculate computes the ICE score
	Calculate(params *models.ICEParams) *models.ICEScore
}

// MoSCoWCategorizer defines MoSCoW categorization
type MoSCoWCategorizer interface {
	// GetPriority returns numeric priority for sorting
	GetPriority(category models.MoSCoWCategory) int

	// ValidateCategory validates a MoSCoW category
	ValidateCategory(category models.MoSCoWCategory) error
}

// WeightedCalculator defines weighted score calculation
type WeightedCalculator interface {
	// Calculate computes the weighted score
	Calculate(params *models.WeightedParams) (*models.WeightedScore, error)

	// ValidateWeights validates that weights sum to 1
	ValidateWeights(criteria []models.WeightedCriterion) error
}
