package priority

import (
	"errors"

	"github.com/K-H-Tech/apm/internal/models"
)

var (
	// ErrInvalidICEScore is returned when an ICE component is out of range
	ErrInvalidICEScore = errors.New("ICE components must be between 1 and 10")
)

// ICECalculator calculates ICE scores
type ICECalculator struct{}

// NewICECalculator creates a new ICE calculator
func NewICECalculator() *ICECalculator {
	return &ICECalculator{}
}

// Calculate computes the ICE score from parameters
//
// ICE Score = (Impact + Confidence + Ease) / 3
//
// Where each component is rated 1-10:
//   - Impact: How much will this move the needle?
//   - Confidence: How confident are we in our estimates?
//   - Ease: How easy is this to implement?
func (i *ICECalculator) Calculate(params *models.ICEParams) (*models.ICEScore, error) {
	// Validate inputs
	if err := i.Validate(params); err != nil {
		return nil, err
	}

	// Calculate ICE score (average of the three components)
	score := float64(params.Impact+params.Confidence+params.Ease) / 3.0

	return &models.ICEScore{
		Impact:     params.Impact,
		Confidence: params.Confidence,
		Ease:       params.Ease,
		Score:      score,
	}, nil
}

// Validate validates ICE parameters
func (i *ICECalculator) Validate(params *models.ICEParams) error {
	if params.Impact < 1 || params.Impact > 10 {
		return ErrInvalidICEScore
	}
	if params.Confidence < 1 || params.Confidence > 10 {
		return ErrInvalidICEScore
	}
	if params.Ease < 1 || params.Ease > 10 {
		return ErrInvalidICEScore
	}
	return nil
}

// ICEScoreInterpretation provides context for an ICE score
type ICEScoreInterpretation struct {
	Score       float64
	Category    string
	Description string
}

// InterpretScore provides interpretation for an ICE score
func InterpretICEScore(score float64) *ICEScoreInterpretation {
	switch {
	case score >= 9:
		return &ICEScoreInterpretation{
			Score:       score,
			Category:    "Exceptional",
			Description: "Extremely high priority - pursue immediately",
		}
	case score >= 7:
		return &ICEScoreInterpretation{
			Score:       score,
			Category:    "High",
			Description: "Strong candidate for immediate action",
		}
	case score >= 5:
		return &ICEScoreInterpretation{
			Score:       score,
			Category:    "Medium",
			Description: "Worth pursuing, but not urgent",
		}
	case score >= 3:
		return &ICEScoreInterpretation{
			Score:       score,
			Category:    "Low",
			Description: "Consider only if resources permit",
		}
	default:
		return &ICEScoreInterpretation{
			Score:       score,
			Category:    "Very Low",
			Description: "Likely not worth the effort",
		}
	}
}

// ICEComponentGuidance provides guidance for scoring each component
type ICEComponentGuidance struct {
	Component   string
	Description string
	Scale       []ICEScalePoint
}

// ICEScalePoint represents a point on the 1-10 scale
type ICEScalePoint struct {
	Score       int
	Label       string
	Description string
}

// GetImpactGuidance returns guidance for scoring Impact
func GetImpactGuidance() *ICEComponentGuidance {
	return &ICEComponentGuidance{
		Component:   "Impact",
		Description: "How much will this initiative move the needle on key metrics?",
		Scale: []ICEScalePoint{
			{Score: 10, Label: "Massive", Description: "10x improvement, game-changing"},
			{Score: 8, Label: "Major", Description: "Significant metric improvement (50%+)"},
			{Score: 6, Label: "Moderate", Description: "Noticeable improvement (20-50%)"},
			{Score: 4, Label: "Minor", Description: "Small but measurable improvement (5-20%)"},
			{Score: 2, Label: "Minimal", Description: "Barely noticeable (<5%)"},
			{Score: 1, Label: "None", Description: "No expected impact on metrics"},
		},
	}
}

// GetConfidenceGuidance returns guidance for scoring Confidence
func GetConfidenceGuidance() *ICEComponentGuidance {
	return &ICEComponentGuidance{
		Component:   "Confidence",
		Description: "How confident are we in our Impact and Ease estimates?",
		Scale: []ICEScalePoint{
			{Score: 10, Label: "Certain", Description: "Data-backed, proven approach"},
			{Score: 8, Label: "High", Description: "Strong evidence, low uncertainty"},
			{Score: 6, Label: "Moderate", Description: "Some data, educated estimate"},
			{Score: 4, Label: "Low", Description: "Limited data, significant assumptions"},
			{Score: 2, Label: "Guess", Description: "Mostly speculation"},
			{Score: 1, Label: "Unknown", Description: "Complete uncertainty"},
		},
	}
}

// GetEaseGuidance returns guidance for scoring Ease
func GetEaseGuidance() *ICEComponentGuidance {
	return &ICEComponentGuidance{
		Component:   "Ease",
		Description: "How easy is this to implement? Consider time, resources, and complexity.",
		Scale: []ICEScalePoint{
			{Score: 10, Label: "Trivial", Description: "Less than a day, single person"},
			{Score: 8, Label: "Easy", Description: "About a week, straightforward"},
			{Score: 6, Label: "Moderate", Description: "2-4 weeks, some complexity"},
			{Score: 4, Label: "Hard", Description: "1-2 months, significant effort"},
			{Score: 2, Label: "Very Hard", Description: "Quarter+, major undertaking"},
			{Score: 1, Label: "Extremely Hard", Description: "Massive effort, many unknowns"},
		},
	}
}

// GetAllGuidance returns guidance for all ICE components
func GetAllICEGuidance() []*ICEComponentGuidance {
	return []*ICEComponentGuidance{
		GetImpactGuidance(),
		GetConfidenceGuidance(),
		GetEaseGuidance(),
	}
}

// CompareICEScores compares two ICE scores and returns which is higher
func CompareICEScores(a, b *models.ICEScore) int {
	if a.Score > b.Score {
		return 1
	}
	if a.Score < b.Score {
		return -1
	}
	return 0
}
