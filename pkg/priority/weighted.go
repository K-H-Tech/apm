package priority

import (
	"errors"
	"math"
	"sort"
)

var (
	// ErrNoCriteria is returned when no criteria are provided
	ErrNoCriteria = errors.New("at least one criterion is required")

	// ErrInvalidWeight is returned when a weight is invalid
	ErrInvalidWeight = errors.New("weight must be between 0 and 100")

	// ErrInvalidCriterionScore is returned when a criterion score is invalid
	ErrInvalidCriterionScore = errors.New("criterion score must be between 0 and 10")

	// ErrWeightSumInvalid is returned when weights don't sum to 100
	ErrWeightSumInvalid = errors.New("weights must sum to 100")

	// ErrDuplicateCriterion is returned when duplicate criterion names are found
	ErrDuplicateCriterion = errors.New("duplicate criterion names are not allowed")
)

// WeightedCalculator calculates weighted priority scores
type WeightedCalculator struct {
	// NormalizeWeights automatically normalizes weights to sum to 100
	NormalizeWeights bool
}

// NewWeightedCalculator creates a new weighted priority calculator
func NewWeightedCalculator() *WeightedCalculator {
	return &WeightedCalculator{
		NormalizeWeights: false,
	}
}

// NewWeightedCalculatorWithNormalization creates a calculator that auto-normalizes weights
func NewWeightedCalculatorWithNormalization() *WeightedCalculator {
	return &WeightedCalculator{
		NormalizeWeights: true,
	}
}

// WeightedCriterion represents a single criterion with its weight and score
type WeightedCriterion struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Weight      float64 `json:"weight"`      // Weight as percentage (0-100)
	Score       float64 `json:"score"`       // Score (0-10)
	Category    string  `json:"category,omitempty"` // Optional category grouping
}

// WeightedParams holds parameters for weighted scoring
type WeightedParams struct {
	Criteria []WeightedCriterion `json:"criteria"`
}

// WeightedScore represents a calculated weighted score
type WeightedScore struct {
	Criteria        []WeightedCriterionResult `json:"criteria"`
	TotalScore      float64                   `json:"total_score"`
	NormalizedScore float64                   `json:"normalized_score"` // Score on 0-100 scale
	MaxPossible     float64                   `json:"max_possible"`
}

// WeightedCriterionResult includes the contribution of each criterion
type WeightedCriterionResult struct {
	Name         string  `json:"name"`
	Weight       float64 `json:"weight"`
	Score        float64 `json:"score"`
	Contribution float64 `json:"contribution"` // Weight * Score / 10
	Percentage   float64 `json:"percentage"`   // Contribution as % of total
}

// Calculate computes the weighted score from parameters
//
// Weighted Score = Σ(Weight_i × Score_i) / 10
//
// Where:
//   - Each criterion has a weight (0-100%) and score (0-10)
//   - Weights must sum to 100%
//   - Final score is on a 0-100 scale
func (w *WeightedCalculator) Calculate(params *WeightedParams) (*WeightedScore, error) {
	// Validate inputs
	if err := w.Validate(params); err != nil {
		return nil, err
	}

	// Normalize weights if enabled
	criteria := params.Criteria
	if w.NormalizeWeights {
		criteria = w.normalizeWeights(params.Criteria)
	}

	// Calculate weighted score
	var totalScore float64
	results := make([]WeightedCriterionResult, len(criteria))

	for i, c := range criteria {
		contribution := (c.Weight * c.Score) / 10.0
		totalScore += contribution

		results[i] = WeightedCriterionResult{
			Name:         c.Name,
			Weight:       c.Weight,
			Score:        c.Score,
			Contribution: contribution,
		}
	}

	// Calculate percentages
	for i := range results {
		if totalScore > 0 {
			results[i].Percentage = (results[i].Contribution / totalScore) * 100
		}
	}

	// Max possible is 100 (all criteria scored 10/10)
	maxPossible := 100.0

	return &WeightedScore{
		Criteria:        results,
		TotalScore:      totalScore,
		NormalizedScore: totalScore, // Already on 0-100 scale
		MaxPossible:     maxPossible,
	}, nil
}

// Validate validates weighted parameters
func (w *WeightedCalculator) Validate(params *WeightedParams) error {
	if len(params.Criteria) == 0 {
		return ErrNoCriteria
	}

	// Check for duplicates and validate individual criteria
	seen := make(map[string]bool)
	var weightSum float64

	for _, c := range params.Criteria {
		if seen[c.Name] {
			return ErrDuplicateCriterion
		}
		seen[c.Name] = true

		if c.Weight < 0 || c.Weight > 100 {
			return ErrInvalidWeight
		}
		if c.Score < 0 || c.Score > 10 {
			return ErrInvalidCriterionScore
		}

		weightSum += c.Weight
	}

	// Check weight sum (with small tolerance for floating point)
	if !w.NormalizeWeights && math.Abs(weightSum-100) > 0.01 {
		return ErrWeightSumInvalid
	}

	return nil
}

// normalizeWeights adjusts weights to sum to 100
func (w *WeightedCalculator) normalizeWeights(criteria []WeightedCriterion) []WeightedCriterion {
	var sum float64
	for _, c := range criteria {
		sum += c.Weight
	}

	if sum == 0 {
		return criteria
	}

	normalized := make([]WeightedCriterion, len(criteria))
	for i, c := range criteria {
		normalized[i] = WeightedCriterion{
			Name:        c.Name,
			Description: c.Description,
			Weight:      (c.Weight / sum) * 100,
			Score:       c.Score,
			Category:    c.Category,
		}
	}

	return normalized
}

// WeightedScoreInterpretation provides context for a weighted score
type WeightedScoreInterpretation struct {
	Score       float64
	Category    string
	Description string
}

// InterpretWeightedScore provides interpretation for a weighted score
func InterpretWeightedScore(score float64) *WeightedScoreInterpretation {
	switch {
	case score >= 90:
		return &WeightedScoreInterpretation{
			Score:       score,
			Category:    "Exceptional",
			Description: "Outstanding across all criteria - highest priority",
		}
	case score >= 75:
		return &WeightedScoreInterpretation{
			Score:       score,
			Category:    "High",
			Description: "Strong performance - prioritize for near-term",
		}
	case score >= 60:
		return &WeightedScoreInterpretation{
			Score:       score,
			Category:    "Above Average",
			Description: "Good overall - include in planning",
		}
	case score >= 45:
		return &WeightedScoreInterpretation{
			Score:       score,
			Category:    "Average",
			Description: "Moderate priority - consider if resources allow",
		}
	case score >= 30:
		return &WeightedScoreInterpretation{
			Score:       score,
			Category:    "Below Average",
			Description: "Lower priority - may need improvement",
		}
	default:
		return &WeightedScoreInterpretation{
			Score:       score,
			Category:    "Low",
			Description: "Lowest priority - reconsider or deprioritize",
		}
	}
}

// CommonCriteriaTemplate provides common criteria templates for different use cases
type CommonCriteriaTemplate struct {
	Name        string
	Description string
	Criteria    []CriterionTemplate
}

// CriterionTemplate defines a template for a criterion
type CriterionTemplate struct {
	Name           string
	Description    string
	DefaultWeight  float64
	ScoringGuidance string
}

// GetFeaturePrioritizationTemplate returns a template for feature prioritization
func GetFeaturePrioritizationTemplate() *CommonCriteriaTemplate {
	return &CommonCriteriaTemplate{
		Name:        "Feature Prioritization",
		Description: "Standard criteria for prioritizing new features",
		Criteria: []CriterionTemplate{
			{
				Name:           "Business Value",
				Description:    "Revenue impact, strategic alignment, competitive advantage",
				DefaultWeight:  30,
				ScoringGuidance: "10=Critical revenue driver, 5=Moderate value, 1=Minimal business impact",
			},
			{
				Name:           "User Impact",
				Description:    "Number of users affected and degree of impact",
				DefaultWeight:  25,
				ScoringGuidance: "10=All users significantly impacted, 5=Some users moderately impacted, 1=Few users minimally impacted",
			},
			{
				Name:           "Technical Feasibility",
				Description:    "Implementation complexity and technical risk",
				DefaultWeight:  20,
				ScoringGuidance: "10=Simple implementation, 5=Moderate complexity, 1=High complexity/risk",
			},
			{
				Name:           "Time to Value",
				Description:    "How quickly users will see benefits",
				DefaultWeight:  15,
				ScoringGuidance: "10=Immediate value, 5=Weeks to value, 1=Months to value",
			},
			{
				Name:           "Strategic Fit",
				Description:    "Alignment with product vision and roadmap",
				DefaultWeight:  10,
				ScoringGuidance: "10=Core to vision, 5=Supports vision, 1=Tangential to vision",
			},
		},
	}
}

// GetBugPrioritizationTemplate returns a template for bug prioritization
func GetBugPrioritizationTemplate() *CommonCriteriaTemplate {
	return &CommonCriteriaTemplate{
		Name:        "Bug Prioritization",
		Description: "Standard criteria for prioritizing bug fixes",
		Criteria: []CriterionTemplate{
			{
				Name:           "Severity",
				Description:    "How severe is the bug's impact on functionality",
				DefaultWeight:  35,
				ScoringGuidance: "10=System down/data loss, 7=Major feature broken, 5=Feature degraded, 3=Minor issue, 1=Cosmetic",
			},
			{
				Name:           "Frequency",
				Description:    "How often does the bug occur",
				DefaultWeight:  25,
				ScoringGuidance: "10=Every use, 7=Frequent, 5=Occasional, 3=Rare, 1=Edge case only",
			},
			{
				Name:           "User Impact",
				Description:    "Number and type of users affected",
				DefaultWeight:  25,
				ScoringGuidance: "10=All users, 7=Most users, 5=Many users, 3=Some users, 1=Few users",
			},
			{
				Name:           "Workaround Availability",
				Description:    "Is there an acceptable workaround",
				DefaultWeight:  15,
				ScoringGuidance: "10=No workaround, 7=Difficult workaround, 5=Acceptable workaround, 3=Easy workaround, 1=Simple workaround",
			},
		},
	}
}

// GetTechnicalDebtTemplate returns a template for technical debt prioritization
func GetTechnicalDebtTemplate() *CommonCriteriaTemplate {
	return &CommonCriteriaTemplate{
		Name:        "Technical Debt",
		Description: "Standard criteria for prioritizing technical debt reduction",
		Criteria: []CriterionTemplate{
			{
				Name:           "Risk Reduction",
				Description:    "How much does this reduce system risk",
				DefaultWeight:  30,
				ScoringGuidance: "10=Eliminates critical risk, 5=Reduces moderate risk, 1=Minimal risk reduction",
			},
			{
				Name:           "Development Velocity",
				Description:    "Impact on team's ability to deliver features",
				DefaultWeight:  25,
				ScoringGuidance: "10=Major velocity improvement, 5=Moderate improvement, 1=Minimal improvement",
			},
			{
				Name:           "Maintenance Cost",
				Description:    "Current cost of maintaining this technical debt",
				DefaultWeight:  20,
				ScoringGuidance: "10=High ongoing cost, 5=Moderate cost, 1=Low cost",
			},
			{
				Name:           "Implementation Effort",
				Description:    "Effort required to address (inverse - lower effort = higher score)",
				DefaultWeight:  15,
				ScoringGuidance: "10=Minimal effort, 5=Moderate effort, 1=Major effort",
			},
			{
				Name:           "Dependencies",
				Description:    "How many other improvements depend on this",
				DefaultWeight:  10,
				ScoringGuidance: "10=Many dependencies, 5=Some dependencies, 1=No dependencies",
			},
		},
	}
}

// GetAllTemplates returns all available criteria templates
func GetAllWeightedTemplates() []*CommonCriteriaTemplate {
	return []*CommonCriteriaTemplate{
		GetFeaturePrioritizationTemplate(),
		GetBugPrioritizationTemplate(),
		GetTechnicalDebtTemplate(),
	}
}

// CompareWeightedScores compares two weighted scores
func CompareWeightedScores(a, b *WeightedScore) int {
	if a.TotalScore > b.TotalScore {
		return 1
	}
	if a.TotalScore < b.TotalScore {
		return -1
	}
	return 0
}

// RankItems ranks a list of items by their weighted scores
// Items with tied scores receive the same rank (standard competition ranking)
func RankItems(items []WeightedScoredItem) []RankedItem {
	// Sort by score descending
	sort.Slice(items, func(i, j int) bool {
		return items[i].Score.TotalScore > items[j].Score.TotalScore
	})

	ranked := make([]RankedItem, len(items))
	currentRank := 1
	for i, item := range items {
		// Update rank only when score is different from previous item
		if i > 0 && items[i].Score.TotalScore < items[i-1].Score.TotalScore {
			currentRank = i + 1
		}
		ranked[i] = RankedItem{
			Rank:  currentRank,
			ID:    item.ID,
			Name:  item.Name,
			Score: item.Score,
		}
	}

	return ranked
}

// WeightedScoredItem represents an item with its weighted score
type WeightedScoredItem struct {
	ID    string
	Name  string
	Score *WeightedScore
}

// RankedItem represents an item with its rank
type RankedItem struct {
	Rank  int            `json:"rank"`
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Score *WeightedScore `json:"score"`
}

// SensitivityAnalysis analyzes how score changes with weight adjustments
type SensitivityAnalysis struct {
	CriterionName string             `json:"criterion_name"`
	BaseScore     float64            `json:"base_score"`
	Variations    []ScoreVariation   `json:"variations"`
}

// ScoreVariation shows score at different weight levels
type ScoreVariation struct {
	WeightChange float64 `json:"weight_change"` // e.g., -10, -5, +5, +10
	NewWeight    float64 `json:"new_weight"`
	NewScore     float64 `json:"new_score"`
	ScoreChange  float64 `json:"score_change"`
}

// AnalyzeSensitivity performs sensitivity analysis on a criterion
func (w *WeightedCalculator) AnalyzeSensitivity(params *WeightedParams, criterionName string) (*SensitivityAnalysis, error) {
	// Calculate base score
	baseResult, err := w.Calculate(params)
	if err != nil {
		return nil, err
	}

	// Find the criterion
	var targetIdx int = -1
	for i, c := range params.Criteria {
		if c.Name == criterionName {
			targetIdx = i
			break
		}
	}
	if targetIdx == -1 {
		return nil, errors.New("criterion not found")
	}

	baseWeight := params.Criteria[targetIdx].Weight
	variations := []float64{-20, -10, -5, 5, 10, 20}
	results := make([]ScoreVariation, 0, len(variations))

	for _, change := range variations {
		newWeight := baseWeight + change
		if newWeight < 0 || newWeight > 100 {
			continue
		}

		// Create modified params
		modifiedCriteria := make([]WeightedCriterion, len(params.Criteria))
		copy(modifiedCriteria, params.Criteria)
		modifiedCriteria[targetIdx].Weight = newWeight

		// Adjust other weights proportionally to maintain 100% total
		adjustment := change / float64(len(params.Criteria)-1)
		for i := range modifiedCriteria {
			if i != targetIdx {
				modifiedCriteria[i].Weight -= adjustment
				if modifiedCriteria[i].Weight < 0 {
					modifiedCriteria[i].Weight = 0
				}
			}
		}

		// Calculate with normalization to handle weight sum variations
		calcWithNorm := NewWeightedCalculatorWithNormalization()
		modifiedResult, err := calcWithNorm.Calculate(&WeightedParams{Criteria: modifiedCriteria})
		if err != nil {
			continue
		}

		// Find the actual normalized weight from the result
		var actualNewWeight float64
		for _, r := range modifiedResult.Criteria {
			if r.Name == criterionName {
				actualNewWeight = r.Weight
				break
			}
		}

		results = append(results, ScoreVariation{
			WeightChange: change,
			NewWeight:    actualNewWeight,
			NewScore:     modifiedResult.TotalScore,
			ScoreChange:  modifiedResult.TotalScore - baseResult.TotalScore,
		})
	}

	return &SensitivityAnalysis{
		CriterionName: criterionName,
		BaseScore:     baseResult.TotalScore,
		Variations:    results,
	}, nil
}
