package priority

import (
	"errors"

	"github.com/K-H-Tech/apm/internal/models"
)

var (
	// ErrInvalidReach is returned when reach is invalid
	ErrInvalidReach = errors.New("reach must be greater than 0")

	// ErrInvalidImpact is returned when impact is invalid
	ErrInvalidImpact = errors.New("impact must be one of: 25, 50, 100, 200, 300")

	// ErrInvalidConfidence is returned when confidence is invalid
	ErrInvalidConfidence = errors.New("confidence must be one of: 50, 80, 100")

	// ErrInvalidEffort is returned when effort is invalid
	ErrInvalidEffort = errors.New("effort must be greater than 0")
)

// RICECalculator calculates RICE scores
type RICECalculator struct{}

// NewRICECalculator creates a new RICE calculator
func NewRICECalculator() *RICECalculator {
	return &RICECalculator{}
}

// Calculate computes the RICE score from parameters
//
// RICE Score = (Reach × Impact × Confidence) / Effort
//
// Where:
//   - Reach: Number of users/customers impacted per quarter
//   - Impact: 0.25 (minimal), 0.5 (low), 1 (medium), 2 (high), 3 (massive)
//   - Confidence: Percentage (50%, 80%, 100%)
//   - Effort: Person-weeks required
func (r *RICECalculator) Calculate(params *models.RICEParams) (*models.RICEScore, error) {
	// Validate inputs
	if err := r.Validate(params); err != nil {
		return nil, err
	}

	// Convert impact to decimal
	impactDecimal := models.ImpactToDecimal(params.Impact)

	// Convert confidence to decimal
	confidenceDecimal := float64(params.Confidence) / 100.0

	// Calculate RICE score
	score := (float64(params.Reach) * impactDecimal * confidenceDecimal) / float64(params.Effort)

	return &models.RICEScore{
		Reach:      params.Reach,
		Impact:     impactDecimal,
		Confidence: confidenceDecimal,
		Effort:     params.Effort,
		Score:      score,
	}, nil
}

// Validate validates RICE parameters
func (r *RICECalculator) Validate(params *models.RICEParams) error {
	if params.Reach <= 0 {
		return ErrInvalidReach
	}

	// Valid impact values: 25 (0.25), 50 (0.5), 100 (1), 200 (2), 300 (3)
	validImpacts := map[int]bool{25: true, 50: true, 100: true, 200: true, 300: true}
	if !validImpacts[params.Impact] {
		return ErrInvalidImpact
	}

	// Valid confidence values: 50 (low), 80 (medium), 100 (high)
	validConfidence := map[int]bool{50: true, 80: true, 100: true}
	if !validConfidence[params.Confidence] {
		return ErrInvalidConfidence
	}

	if params.Effort <= 0 {
		return ErrInvalidEffort
	}

	return nil
}

// ImpactLevel represents the impact level for RICE scoring
type ImpactLevel int

const (
	ImpactMinimal ImpactLevel = 25  // 0.25x - Minimal impact
	ImpactLow     ImpactLevel = 50  // 0.5x - Low impact
	ImpactMedium  ImpactLevel = 100 // 1x - Medium impact
	ImpactHigh    ImpactLevel = 200 // 2x - High impact
	ImpactMassive ImpactLevel = 300 // 3x - Massive impact
)

// String returns the string representation of impact level
func (i ImpactLevel) String() string {
	switch i {
	case ImpactMinimal:
		return "Minimal (0.25x)"
	case ImpactLow:
		return "Low (0.5x)"
	case ImpactMedium:
		return "Medium (1x)"
	case ImpactHigh:
		return "High (2x)"
	case ImpactMassive:
		return "Massive (3x)"
	default:
		return "Unknown"
	}
}

// ToDecimal converts impact level to decimal multiplier
func (i ImpactLevel) ToDecimal() float64 {
	return float64(i) / 100.0
}

// ConfidenceLevel represents the confidence level for RICE scoring
type ConfidenceLevel int

const (
	ConfidenceLow    ConfidenceLevel = 50  // 50% - Low confidence
	ConfidenceMedium ConfidenceLevel = 80  // 80% - Medium confidence
	ConfidenceHigh   ConfidenceLevel = 100 // 100% - High confidence
)

// String returns the string representation of confidence level
func (c ConfidenceLevel) String() string {
	switch c {
	case ConfidenceLow:
		return "Low (50%)"
	case ConfidenceMedium:
		return "Medium (80%)"
	case ConfidenceHigh:
		return "High (100%)"
	default:
		return "Unknown"
	}
}

// ToDecimal converts confidence level to decimal multiplier
func (c ConfidenceLevel) ToDecimal() float64 {
	return float64(c) / 100.0
}

// RICEScoreInterpretation provides context for a RICE score
type RICEScoreInterpretation struct {
	Score       float64
	Category    string
	Description string
}

// InterpretScore provides interpretation for a RICE score
func InterpretRICEScore(score float64) *RICEScoreInterpretation {
	switch {
	case score >= 100:
		return &RICEScoreInterpretation{
			Score:       score,
			Category:    "Critical",
			Description: "Extremely high impact, should be prioritized immediately",
		}
	case score >= 50:
		return &RICEScoreInterpretation{
			Score:       score,
			Category:    "High",
			Description: "High impact, strong candidate for next sprint",
		}
	case score >= 25:
		return &RICEScoreInterpretation{
			Score:       score,
			Category:    "Medium",
			Description: "Moderate impact, consider for backlog prioritization",
		}
	case score >= 10:
		return &RICEScoreInterpretation{
			Score:       score,
			Category:    "Low",
			Description: "Lower impact, may be worth doing if resources permit",
		}
	default:
		return &RICEScoreInterpretation{
			Score:       score,
			Category:    "Minimal",
			Description: "Very low impact, deprioritize unless strategic",
		}
	}
}
