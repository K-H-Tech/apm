package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// PRDTemplate represents a reusable PRD template
type PRDTemplate struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" db:"organization_id"`

	Name        string       `json:"name" db:"name"`
	Description string       `json:"description,omitempty" db:"description"`
	Type        TemplateType `json:"type" db:"type"`

	// Template structure definition
	ContentSchema TemplateSchema `json:"content_schema" db:"content_schema"`

	// Example content for reference
	ExampleContent *PRDContent `json:"example_content,omitempty" db:"example_content"`

	// System templates are available to all organizations
	IsSystemTemplate bool `json:"is_system_template" db:"is_system_template"`

	CreatedBy *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// TemplateType represents the type of template
type TemplateType string

const (
	TemplateTypeFeature     TemplateType = "feature"
	TemplateTypeBugFix      TemplateType = "bug_fix"
	TemplateTypeEnhancement TemplateType = "enhancement"
	TemplateTypeTechnical   TemplateType = "technical"
)

// TemplateSchema defines the structure of a PRD template
type TemplateSchema struct {
	Sections []TemplateSection `json:"sections"`
}

// Value implements driver.Valuer for database storage
func (s TemplateSchema) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements sql.Scanner for database retrieval
func (s *TemplateSchema) Scan(value interface{}) error {
	if value == nil {
		*s = TemplateSchema{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

// TemplateSection represents a section in a PRD template
type TemplateSection struct {
	Name        string      `json:"name"`                   // Internal name (e.g., "overview")
	Label       string      `json:"label"`                  // Display label (e.g., "Overview")
	Type        SectionType `json:"type"`                   // Section type
	Required    bool        `json:"required"`               // Whether this section is required
	Description string      `json:"description,omitempty"`  // Help text for this section
	Placeholder string      `json:"placeholder,omitempty"`  // Placeholder text
	Default     interface{} `json:"default,omitempty"`      // Default value
}

// SectionType represents the type of content for a template section
type SectionType string

const (
	SectionTypeText         SectionType = "text"         // Free-form text
	SectionTypeList         SectionType = "list"         // List of items
	SectionTypePersonas     SectionType = "personas"     // User personas
	SectionTypeStories      SectionType = "stories"      // User stories
	SectionTypeRequirements SectionType = "requirements" // Requirements list
	SectionTypeMetrics      SectionType = "metrics"      // Success metrics
	SectionTypeTimeline     SectionType = "timeline"     // Timeline with milestones
)

// TemplateFilter represents filter options for listing templates
type TemplateFilter struct {
	OrganizationID   *uuid.UUID
	Type             *TemplateType
	IncludeSystem    bool // Include system templates in results
	Search           string
	Limit            int
	Offset           int
}

// CreateTemplateRequest represents the request to create a new template
type CreateTemplateRequest struct {
	Name           string          `json:"name" binding:"required"`
	Description    string          `json:"description"`
	Type           TemplateType    `json:"type" binding:"required"`
	ContentSchema  TemplateSchema  `json:"content_schema" binding:"required"`
	ExampleContent *PRDContent     `json:"example_content"`
}

// UpdateTemplateRequest represents the request to update a template
type UpdateTemplateRequest struct {
	Name           *string         `json:"name"`
	Description    *string         `json:"description"`
	Type           *TemplateType   `json:"type"`
	ContentSchema  *TemplateSchema `json:"content_schema"`
	ExampleContent *PRDContent     `json:"example_content"`
}

// ValidateAgainstSchema validates PRD content against a template schema
func (t *PRDTemplate) ValidateAgainstSchema(content *PRDContent) []string {
	var errors []string

	contentMap := make(map[string]interface{})

	// Convert content to map for validation
	contentBytes, _ := json.Marshal(content)
	json.Unmarshal(contentBytes, &contentMap)

	for _, section := range t.ContentSchema.Sections {
		if section.Required {
			value, exists := contentMap[section.Name]
			if !exists || isEmpty(value) {
				errors = append(errors, "missing required section: "+section.Label)
			}
		}
	}

	return errors
}

// isEmpty checks if a value is empty
func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return val == ""
	case []interface{}:
		return len(val) == 0
	case map[string]interface{}:
		return len(val) == 0
	default:
		return false
	}
}

// DefaultFeatureTemplate returns the default feature PRD template schema
func DefaultFeatureTemplate() TemplateSchema {
	return TemplateSchema{
		Sections: []TemplateSection{
			{Name: "overview", Label: "Overview", Type: SectionTypeText, Required: true, Description: "High-level description of the feature"},
			{Name: "problem_statement", Label: "Problem Statement", Type: SectionTypeText, Required: true, Description: "What problem are we solving?"},
			{Name: "goals", Label: "Goals", Type: SectionTypeList, Required: true, Description: "What are the goals of this feature?"},
			{Name: "user_personas", Label: "User Personas", Type: SectionTypePersonas, Required: true, Description: "Who are the target users?"},
			{Name: "user_stories", Label: "User Stories", Type: SectionTypeStories, Required: true, Description: "User stories with acceptance criteria"},
			{Name: "requirements", Label: "Requirements", Type: SectionTypeRequirements, Required: true, Description: "Functional and non-functional requirements"},
			{Name: "acceptance_criteria", Label: "Acceptance Criteria", Type: SectionTypeList, Required: true, Description: "Overall acceptance criteria"},
			{Name: "out_of_scope", Label: "Out of Scope", Type: SectionTypeList, Required: false, Description: "What is explicitly NOT included"},
			{Name: "assumptions", Label: "Assumptions", Type: SectionTypeList, Required: false, Description: "Assumptions being made"},
			{Name: "dependencies", Label: "Dependencies", Type: SectionTypeList, Required: false, Description: "Dependencies on other teams or systems"},
			{Name: "success_metrics", Label: "Success Metrics", Type: SectionTypeMetrics, Required: true, Description: "How will we measure success?"},
		},
	}
}
