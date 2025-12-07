package priority

import (
	"errors"

	"github.com/K-H-Tech/apm/internal/models"
)

var (
	// ErrInvalidMoSCoWCategory is returned when the category is invalid
	ErrInvalidMoSCoWCategory = errors.New("invalid MoSCoW category: must be 'must', 'should', 'could', or 'wont'")
)

// MoSCoWCategorizer handles MoSCoW prioritization
type MoSCoWCategorizer struct{}

// NewMoSCoWCategorizer creates a new MoSCoW categorizer
func NewMoSCoWCategorizer() *MoSCoWCategorizer {
	return &MoSCoWCategorizer{}
}

// ValidateCategory validates a MoSCoW category
func (m *MoSCoWCategorizer) ValidateCategory(category models.MoSCoWCategory) error {
	switch category {
	case models.MoSCoWMust, models.MoSCoWShould, models.MoSCoWCould, models.MoSCoWWont:
		return nil
	default:
		return ErrInvalidMoSCoWCategory
	}
}

// GetPriority returns the numeric priority for a category (lower = higher priority)
func (m *MoSCoWCategorizer) GetPriority(category models.MoSCoWCategory) int {
	return category.Priority()
}

// GetScore returns a numeric score for the category (for sorting with other frameworks)
// Must = 100, Should = 75, Could = 50, Won't = 0
func (m *MoSCoWCategorizer) GetScore(category models.MoSCoWCategory) float64 {
	switch category {
	case models.MoSCoWMust:
		return 100.0
	case models.MoSCoWShould:
		return 75.0
	case models.MoSCoWCould:
		return 50.0
	case models.MoSCoWWont:
		return 0.0
	default:
		return 0.0
	}
}

// MoSCoWCategoryInfo provides information about a MoSCoW category
type MoSCoWCategoryInfo struct {
	Category    models.MoSCoWCategory
	Name        string
	Description string
	Guidance    string
	Priority    int
	Score       float64
}

// GetCategoryInfo returns detailed information about a category
func (m *MoSCoWCategorizer) GetCategoryInfo(category models.MoSCoWCategory) (*MoSCoWCategoryInfo, error) {
	if err := m.ValidateCategory(category); err != nil {
		return nil, err
	}

	switch category {
	case models.MoSCoWMust:
		return &MoSCoWCategoryInfo{
			Category:    category,
			Name:        "Must Have",
			Description: "Critical requirements that must be delivered for the project to be considered a success",
			Guidance:    "Without these features, the product would not be viable. These are non-negotiable for the current release.",
			Priority:    1,
			Score:       100.0,
		}, nil
	case models.MoSCoWShould:
		return &MoSCoWCategoryInfo{
			Category:    category,
			Name:        "Should Have",
			Description: "Important requirements that should be included if possible",
			Guidance:    "These are important but not vital. The product can launch without them, but they add significant value.",
			Priority:    2,
			Score:       75.0,
		}, nil
	case models.MoSCoWCould:
		return &MoSCoWCategoryInfo{
			Category:    category,
			Name:        "Could Have",
			Description: "Desirable requirements that would be nice to have if time and resources permit",
			Guidance:    "These are 'nice to have' features. Include them only if there's time and budget after Must and Should items.",
			Priority:    3,
			Score:       50.0,
		}, nil
	case models.MoSCoWWont:
		return &MoSCoWCategoryInfo{
			Category:    category,
			Name:        "Won't Have (this time)",
			Description: "Requirements that will not be included in the current release but may be considered for future releases",
			Guidance:    "Explicitly excluded from the current scope. May be revisited in future iterations.",
			Priority:    4,
			Score:       0.0,
		}, nil
	default:
		return nil, ErrInvalidMoSCoWCategory
	}
}

// GetAllCategories returns information about all MoSCoW categories
func (m *MoSCoWCategorizer) GetAllCategories() []*MoSCoWCategoryInfo {
	categories := []models.MoSCoWCategory{
		models.MoSCoWMust,
		models.MoSCoWShould,
		models.MoSCoWCould,
		models.MoSCoWWont,
	}

	result := make([]*MoSCoWCategoryInfo, 0, len(categories))
	for _, cat := range categories {
		info, _ := m.GetCategoryInfo(cat)
		result = append(result, info)
	}

	return result
}

// SuggestCategory suggests a MoSCoW category based on criteria
func (m *MoSCoWCategorizer) SuggestCategory(criteria MoSCoWSuggestionCriteria) models.MoSCoWCategory {
	// Critical and no workaround = Must
	if criteria.IsCritical && !criteria.HasWorkaround {
		return models.MoSCoWMust
	}

	// High impact and regulatory/legal = Must
	if criteria.IsRegulatory || criteria.IsLegal {
		return models.MoSCoWMust
	}

	// High business value but has workaround = Should
	if criteria.BusinessValue >= 8 && criteria.HasWorkaround {
		return models.MoSCoWShould
	}

	// Medium business value = Should
	if criteria.BusinessValue >= 6 {
		return models.MoSCoWShould
	}

	// Low business value but user requested = Could
	if criteria.BusinessValue >= 4 || criteria.UserRequested {
		return models.MoSCoWCould
	}

	// Everything else = Won't
	return models.MoSCoWWont
}

// MoSCoWSuggestionCriteria holds criteria for category suggestion
type MoSCoWSuggestionCriteria struct {
	IsCritical    bool // Is this critical to core functionality?
	HasWorkaround bool // Is there a workaround if this isn't implemented?
	IsRegulatory  bool // Is this required for regulatory compliance?
	IsLegal       bool // Is this required for legal reasons?
	BusinessValue int  // Business value score (1-10)
	UserRequested bool // Was this explicitly requested by users?
}

// MoSCoWDistribution represents the distribution of items across categories
type MoSCoWDistribution struct {
	MustCount   int     `json:"must_count"`
	ShouldCount int     `json:"should_count"`
	CouldCount  int     `json:"could_count"`
	WontCount   int     `json:"wont_count"`
	Total       int     `json:"total"`
	MustPercent float64 `json:"must_percent"`
}

// CalculateDistribution calculates the distribution of categories
func (m *MoSCoWCategorizer) CalculateDistribution(categories []models.MoSCoWCategory) *MoSCoWDistribution {
	dist := &MoSCoWDistribution{}

	for _, cat := range categories {
		switch cat {
		case models.MoSCoWMust:
			dist.MustCount++
		case models.MoSCoWShould:
			dist.ShouldCount++
		case models.MoSCoWCould:
			dist.CouldCount++
		case models.MoSCoWWont:
			dist.WontCount++
		}
	}

	dist.Total = len(categories)
	if dist.Total > 0 {
		dist.MustPercent = float64(dist.MustCount) / float64(dist.Total) * 100
	}

	return dist
}

// ValidateDistribution checks if the MoSCoW distribution follows best practices
// Best practice: Must-haves should be ~60% of capacity, Could-haves ~20% buffer
func (m *MoSCoWCategorizer) ValidateDistribution(dist *MoSCoWDistribution) []string {
	var warnings []string

	if dist.Total == 0 {
		return warnings
	}

	// Check if Must-haves exceed 60% (DSDM recommendation)
	if dist.MustPercent > 60 {
		warnings = append(warnings, "Must-have items exceed 60% of total. Consider moving some to Should-have for flexibility.")
	}

	// Check if there are no Could-haves (buffer)
	if dist.CouldCount == 0 && dist.Total > 5 {
		warnings = append(warnings, "No Could-have items. Consider adding some buffer items for flexibility.")
	}

	// Check if all items are Must-have
	if dist.MustCount == dist.Total {
		warnings = append(warnings, "All items are Must-have. This leaves no room for scope negotiation.")
	}

	return warnings
}
