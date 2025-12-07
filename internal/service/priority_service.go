package service

import (
	"context"
	"encoding/json"

	"github.com/K-H-Tech/apm/internal/models"
	"github.com/K-H-Tech/apm/internal/repository"
	"github.com/K-H-Tech/apm/pkg/priority"
	"github.com/google/uuid"
)

// PriorityService implements priority calculation and management
type PriorityService struct {
	repo             *repository.PriorityRepository
	riceCalculator   *priority.RICECalculator
	iceCalculator    *priority.ICECalculator
	moscowCategorizer *priority.MoSCoWCategorizer
	weightedCalc     *priority.WeightedCalculator
}

// NewPriorityService creates a new priority service
func NewPriorityService(repo *repository.PriorityRepository) *PriorityService {
	return &PriorityService{
		repo:             repo,
		riceCalculator:   priority.NewRICECalculator(),
		iceCalculator:    priority.NewICECalculator(),
		moscowCategorizer: priority.NewMoSCoWCategorizer(),
		weightedCalc:     priority.NewWeightedCalculator(),
	}
}

// CalculateRICEInput holds input for RICE calculation
type CalculateRICEInput struct {
	PRDID        uuid.UUID
	Reach        int // Number of users impacted per quarter
	Impact       int // 25, 50, 100, 200, or 300
	Confidence   int // 50-100 percentage
	Effort       int // Person-weeks
	CalculatedBy uuid.UUID
}

// CalculateRICE calculates a RICE score for a PRD
func (s *PriorityService) CalculateRICE(ctx context.Context, input CalculateRICEInput) (*models.PriorityScore, error) {
	params := &models.RICEParams{
		Reach:      input.Reach,
		Impact:     input.Impact,
		Confidence: input.Confidence,
		Effort:     input.Effort,
	}

	result, err := s.riceCalculator.Calculate(params)
	if err != nil {
		return nil, err
	}

	score := &models.PriorityScore{
		ID:           uuid.New(),
		PRDID:        input.PRDID,
		Framework:    models.FrameworkRICE,
		Reach:        &input.Reach,
		Impact:       &input.Impact,
		Confidence:   &input.Confidence,
		Effort:       &input.Effort,
		FinalScore:   result.Score,
		CalculatedBy: input.CalculatedBy,
	}

	if err := s.repo.Create(ctx, score); err != nil {
		return nil, err
	}

	return score, nil
}

// CalculateICEInput holds input for ICE calculation
type CalculateICEInput struct {
	PRDID        uuid.UUID
	Impact       int // 1-10
	Confidence   int // 1-10
	Ease         int // 1-10
	CalculatedBy uuid.UUID
}

// CalculateICE calculates an ICE score for a PRD
func (s *PriorityService) CalculateICE(ctx context.Context, input CalculateICEInput) (*models.PriorityScore, error) {
	params := &models.ICEParams{
		Impact:     input.Impact,
		Confidence: input.Confidence,
		Ease:       input.Ease,
	}

	result, err := s.iceCalculator.Calculate(params)
	if err != nil {
		return nil, err
	}

	score := &models.PriorityScore{
		ID:           uuid.New(),
		PRDID:        input.PRDID,
		Framework:    models.FrameworkICE,
		Impact:       &input.Impact,
		Confidence:   &input.Confidence,
		Effort:       &input.Ease, // Ease is stored in Effort field
		FinalScore:   result.Score,
		CalculatedBy: input.CalculatedBy,
	}

	if err := s.repo.Create(ctx, score); err != nil {
		return nil, err
	}

	return score, nil
}

// SetMoSCoWInput holds input for MoSCoW categorization
type SetMoSCoWInput struct {
	PRDID        uuid.UUID
	Category     models.MoSCoWCategory
	CalculatedBy uuid.UUID
}

// SetMoSCoW sets the MoSCoW category for a PRD
func (s *PriorityService) SetMoSCoW(ctx context.Context, input SetMoSCoWInput) (*models.PriorityScore, error) {
	if err := s.moscowCategorizer.ValidateCategory(input.Category); err != nil {
		return nil, err
	}

	// Get numeric score for the category
	numericScore := s.moscowCategorizer.GetScore(input.Category)

	score := &models.PriorityScore{
		ID:             uuid.New(),
		PRDID:          input.PRDID,
		Framework:      models.FrameworkMoSCoW,
		MoSCoWCategory: &input.Category,
		FinalScore:     numericScore,
		CalculatedBy:   input.CalculatedBy,
	}

	if err := s.repo.Create(ctx, score); err != nil {
		return nil, err
	}

	return score, nil
}

// SuggestMoSCoWInput holds input for MoSCoW suggestion
type SuggestMoSCoWInput struct {
	IsCritical    bool
	HasWorkaround bool
	IsRegulatory  bool
	IsLegal       bool
	BusinessValue int  // 1-10
	UserRequested bool
}

// SuggestMoSCoW suggests a MoSCoW category based on criteria
func (s *PriorityService) SuggestMoSCoW(input SuggestMoSCoWInput) models.MoSCoWCategory {
	criteria := priority.MoSCoWSuggestionCriteria{
		IsCritical:    input.IsCritical,
		HasWorkaround: input.HasWorkaround,
		IsRegulatory:  input.IsRegulatory,
		IsLegal:       input.IsLegal,
		BusinessValue: input.BusinessValue,
		UserRequested: input.UserRequested,
	}
	return s.moscowCategorizer.SuggestCategory(criteria)
}

// CalculateWeightedInput holds input for weighted scoring
type CalculateWeightedInput struct {
	PRDID        uuid.UUID
	Criteria     []priority.WeightedCriterion
	CalculatedBy uuid.UUID
}

// CalculateWeighted calculates a weighted score for a PRD
func (s *PriorityService) CalculateWeighted(ctx context.Context, input CalculateWeightedInput) (*models.PriorityScore, error) {
	params := &priority.WeightedParams{
		Criteria: input.Criteria,
	}

	result, err := s.weightedCalc.Calculate(params)
	if err != nil {
		return nil, err
	}

	// Serialize criteria to JSON for storage
	criteriaJSON, err := serializeWeightedCriteria(input.Criteria)
	if err != nil {
		return nil, err
	}

	score := &models.PriorityScore{
		ID:             uuid.New(),
		PRDID:          input.PRDID,
		Framework:      models.FrameworkWeighted,
		CustomCriteria: criteriaJSON,
		FinalScore:     result.TotalScore,
		CalculatedBy:   input.CalculatedBy,
	}

	if err := s.repo.Create(ctx, score); err != nil {
		return nil, err
	}

	return score, nil
}

// serializeWeightedCriteria converts weighted criteria to JSON string
func serializeWeightedCriteria(criteria []priority.WeightedCriterion) (string, error) {
	bytes, err := json.Marshal(criteria)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GetByPRD retrieves all priority scores for a PRD
func (s *PriorityService) GetByPRD(ctx context.Context, prdID uuid.UUID) ([]*models.PriorityScore, error) {
	return s.repo.GetByPRDID(ctx, prdID)
}

// GetByPRDAndFramework retrieves a specific framework's score for a PRD
func (s *PriorityService) GetByPRDAndFramework(ctx context.Context, prdID uuid.UUID, framework models.PriorityFramework) (*models.PriorityScore, error) {
	return s.repo.GetByPRDIDAndFramework(ctx, prdID, framework)
}

// GetRankings retrieves ranked PRDs by score for a specific framework
func (s *PriorityService) GetRankings(ctx context.Context, orgID uuid.UUID, framework models.PriorityFramework, limit, offset int) ([]repository.PRDRanking, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.GetRankings(ctx, orgID, framework, limit, offset)
}

// GetMoSCoWDistribution retrieves the MoSCoW distribution for an organization
func (s *PriorityService) GetMoSCoWDistribution(ctx context.Context, orgID uuid.UUID) (*repository.MoSCoWDistribution, error) {
	return s.repo.GetMoSCoWDistribution(ctx, orgID)
}

// ValidateMoSCoWDistribution validates the distribution and returns warnings
func (s *PriorityService) ValidateMoSCoWDistribution(dist *repository.MoSCoWDistribution) []string {
	pDist := &priority.MoSCoWDistribution{
		MustCount:   dist.MustCount,
		ShouldCount: dist.ShouldCount,
		CouldCount:  dist.CouldCount,
		WontCount:   dist.WontCount,
		Total:       dist.Total,
		MustPercent: dist.MustPercent,
	}
	return s.moscowCategorizer.ValidateDistribution(pDist)
}

// GetScoreHistory retrieves score history for a PRD and framework
func (s *PriorityService) GetScoreHistory(ctx context.Context, prdID uuid.UUID, framework models.PriorityFramework) ([]*models.PriorityScore, error) {
	return s.repo.GetScoreHistory(ctx, prdID, framework)
}

// Delete deletes a priority score
func (s *PriorityService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// DeleteByPRD deletes all priority scores for a PRD
func (s *PriorityService) DeleteByPRD(ctx context.Context, prdID uuid.UUID) error {
	return s.repo.DeleteByPRDID(ctx, prdID)
}

// GetRICEInterpretation returns interpretation for a RICE score
func (s *PriorityService) GetRICEInterpretation(score float64) *priority.RICEScoreInterpretation {
	return priority.InterpretRICEScore(score)
}

// GetICEInterpretation returns interpretation for an ICE score
func (s *PriorityService) GetICEInterpretation(score float64) *priority.ICEScoreInterpretation {
	return priority.InterpretICEScore(score)
}

// GetWeightedInterpretation returns interpretation for a weighted score
func (s *PriorityService) GetWeightedInterpretation(score float64) *priority.WeightedScoreInterpretation {
	return priority.InterpretWeightedScore(score)
}

// GetRICEGuidance returns guidance for RICE scoring
func (s *PriorityService) GetRICEGuidance() struct {
	Impact     []priority.ImpactLevel
	Confidence []priority.ConfidenceLevel
} {
	return struct {
		Impact     []priority.ImpactLevel
		Confidence []priority.ConfidenceLevel
	}{
		Impact: []priority.ImpactLevel{
			priority.ImpactMinimal,
			priority.ImpactLow,
			priority.ImpactMedium,
			priority.ImpactHigh,
			priority.ImpactMassive,
		},
		Confidence: []priority.ConfidenceLevel{
			priority.ConfidenceLow,
			priority.ConfidenceMedium,
			priority.ConfidenceHigh,
		},
	}
}

// GetICEGuidance returns guidance for ICE scoring
func (s *PriorityService) GetICEGuidance() []*priority.ICEComponentGuidance {
	return priority.GetAllICEGuidance()
}

// GetMoSCoWInfo returns information about all MoSCoW categories
func (s *PriorityService) GetMoSCoWInfo() []*priority.MoSCoWCategoryInfo {
	return s.moscowCategorizer.GetAllCategories()
}

// GetWeightedTemplates returns available weighted scoring templates
func (s *PriorityService) GetWeightedTemplates() []*priority.CommonCriteriaTemplate {
	return priority.GetAllWeightedTemplates()
}
