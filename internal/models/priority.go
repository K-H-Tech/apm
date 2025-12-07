package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// PriorityScore represents a priority score for a PRD
type PriorityScore struct {
	ID        uuid.UUID         `json:"id" db:"id"`
	PRDID     uuid.UUID         `json:"prd_id" db:"prd_id"`
	Framework PriorityFramework `json:"framework" db:"framework"`

	// RICE framework components
	Reach      *int `json:"reach,omitempty" db:"reach"`           // Users reached per quarter
	Impact     *int `json:"impact,omitempty" db:"impact"`         // 25=minimal, 50=low, 100=medium, 200=high, 300=massive
	Confidence *int `json:"confidence,omitempty" db:"confidence"` // Percentage: 50, 80, 100
	Effort     *int `json:"effort,omitempty" db:"effort"`         // Person-weeks

	// MoSCoW category
	MoSCoWCategory *MoSCoWCategory `json:"moscow_category,omitempty" db:"moscow_category"`

	// Custom criteria (JSON string for weighted scoring)
	CustomCriteria string `json:"custom_criteria,omitempty" db:"custom_criteria"`

	// Final calculated score
	FinalScore float64 `json:"final_score" db:"final_score"`

	CalculatedBy uuid.UUID `json:"calculated_by" db:"calculated_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// PriorityFramework represents the scoring framework used
type PriorityFramework string

const (
	FrameworkRICE     PriorityFramework = "rice"
	FrameworkMoSCoW   PriorityFramework = "moscow"
	FrameworkICE      PriorityFramework = "ice"
	FrameworkWeighted PriorityFramework = "weighted"
)

// MoSCoWCategory represents MoSCoW prioritization categories
type MoSCoWCategory string

const (
	MoSCoWMust   MoSCoWCategory = "must"
	MoSCoWShould MoSCoWCategory = "should"
	MoSCoWCould  MoSCoWCategory = "could"
	MoSCoWWont   MoSCoWCategory = "wont"
)

// MoSCoWPriority returns the numeric priority for sorting (lower = higher priority)
func (m MoSCoWCategory) Priority() int {
	switch m {
	case MoSCoWMust:
		return 1
	case MoSCoWShould:
		return 2
	case MoSCoWCould:
		return 3
	case MoSCoWWont:
		return 4
	default:
		return 5
	}
}

// RICEParams represents input parameters for RICE scoring
type RICEParams struct {
	Reach      int `json:"reach" binding:"required,min=1"`       // Users per quarter
	Impact     int `json:"impact" binding:"required,oneof=25 50 100 200 300"` // 25, 50, 100, 200, 300
	Confidence int `json:"confidence" binding:"required,min=50,max=100"` // 50, 80, 100
	Effort     int `json:"effort" binding:"required,min=1"`      // Person-weeks
}

// RICEScore represents the calculated RICE score
type RICEScore struct {
	Reach      int     `json:"reach"`
	Impact     float64 `json:"impact"`      // Converted to decimal: 0.25, 0.5, 1, 2, 3
	Confidence float64 `json:"confidence"`  // Converted to decimal: 0.5, 0.8, 1.0
	Effort     int     `json:"effort"`
	Score      float64 `json:"score"`       // (Reach * Impact * Confidence) / Effort
}

// ImpactToDecimal converts stored impact value to decimal
func ImpactToDecimal(impact int) float64 {
	switch impact {
	case 25:
		return 0.25 // Minimal
	case 50:
		return 0.5  // Low
	case 100:
		return 1.0  // Medium
	case 200:
		return 2.0  // High
	case 300:
		return 3.0  // Massive
	default:
		return float64(impact) / 100.0
	}
}

// ICEParams represents input parameters for ICE scoring
type ICEParams struct {
	Impact     int `json:"impact" binding:"required,min=1,max=10"`     // 1-10
	Confidence int `json:"confidence" binding:"required,min=1,max=10"` // 1-10
	Ease       int `json:"ease" binding:"required,min=1,max=10"`       // 1-10
}

// ICEScore represents the calculated ICE score
type ICEScore struct {
	Impact     int     `json:"impact"`
	Confidence int     `json:"confidence"`
	Ease       int     `json:"ease"`
	Score      float64 `json:"score"` // Average of Impact, Confidence, Ease
}

// WeightedCriteria represents custom weighted scoring criteria
type WeightedCriteria []WeightedCriterion

// Value implements driver.Valuer for database storage
func (w WeightedCriteria) Value() (driver.Value, error) {
	if w == nil {
		return nil, nil
	}
	return json.Marshal(w)
}

// Scan implements sql.Scanner for database retrieval
func (w *WeightedCriteria) Scan(value interface{}) error {
	if value == nil {
		*w = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, w)
}

// WeightedCriterion represents a single weighted criterion
type WeightedCriterion struct {
	Name   string  `json:"name"`                                // Criterion name
	Weight float64 `json:"weight" binding:"required,min=0,max=1"` // Weight (0-1, must sum to 1)
	Score  int     `json:"score" binding:"required,min=1,max=10"` // Score (1-10)
}

// WeightedParams represents input parameters for weighted scoring
type WeightedParams struct {
	Criteria []WeightedCriterion `json:"criteria" binding:"required,dive"`
}

// WeightedScore represents the calculated weighted score
type WeightedScore struct {
	Criteria []WeightedCriterion `json:"criteria"`
	Score    float64             `json:"score"` // Sum of (weight * score)
}

// ValidateWeights validates that weights sum to 1.0
func (w *WeightedParams) ValidateWeights() error {
	var total float64
	for _, c := range w.Criteria {
		total += c.Weight
	}
	// Allow small floating point tolerance
	if total < 0.99 || total > 1.01 {
		return errors.New("weights must sum to 1.0")
	}
	return nil
}

// PRDWithScore represents a PRD with its priority score
type PRDWithScore struct {
	PRD   *PRD           `json:"prd"`
	Score *PriorityScore `json:"score"`
}

// CalculateRICERequest represents the API request to calculate RICE score
type CalculateRICERequest struct {
	PRDID  uuid.UUID  `json:"prd_id" binding:"required"`
	Params RICEParams `json:"params" binding:"required"`
	Notes  string     `json:"notes"`
}

// CalculateICERequest represents the API request to calculate ICE score
type CalculateICERequest struct {
	PRDID  uuid.UUID `json:"prd_id" binding:"required"`
	Params ICEParams `json:"params" binding:"required"`
	Notes  string    `json:"notes"`
}

// SetMoSCoWRequest represents the API request to set MoSCoW category
type SetMoSCoWRequest struct {
	PRDID    uuid.UUID      `json:"prd_id" binding:"required"`
	Category MoSCoWCategory `json:"category" binding:"required"`
	Notes    string         `json:"notes"`
}

// CalculateWeightedRequest represents the API request to calculate weighted score
type CalculateWeightedRequest struct {
	PRDID  uuid.UUID      `json:"prd_id" binding:"required"`
	Params WeightedParams `json:"params" binding:"required"`
	Notes  string         `json:"notes"`
}

// PriorityRankingRequest represents the request to get priority rankings
type PriorityRankingRequest struct {
	OrganizationID uuid.UUID         `json:"organization_id" binding:"required"`
	Framework      PriorityFramework `json:"framework" binding:"required"`
	Limit          int               `json:"limit"`
	Offset         int               `json:"offset"`
}
