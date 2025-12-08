package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// PRD represents a Product Requirements Document
type PRD struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	TemplateID     *uuid.UUID `json:"template_id,omitempty" db:"template_id"`

	Title   string    `json:"title" db:"title"`
	Status  PRDStatus `json:"status" db:"status"`
	Content PRDContent `json:"content" db:"content"`

	// AI generation metadata
	AIGenerated bool   `json:"ai_generated" db:"ai_generated"`
	SourceNotes string `json:"source_notes,omitempty" db:"source_notes"`

	// Atlassian integration
	JiraEpicKey        string `json:"jira_epic_key,omitempty" db:"jira_epic_key"`
	JiraEpicID         string `json:"jira_epic_id,omitempty" db:"jira_epic_id"`
	ConfluencePageID   string `json:"confluence_page_id,omitempty" db:"confluence_page_id"`
	ConfluencePageURL  string `json:"confluence_page_url,omitempty" db:"confluence_page_url"`

	// Ownership
	OwnerID uuid.UUID `json:"owner_id" db:"owner_id"`

	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	PublishedAt *time.Time `json:"published_at,omitempty" db:"published_at"`
}

// PRDStatus represents the status of a PRD
type PRDStatus string

const (
	PRDStatusDraft     PRDStatus = "draft"
	PRDStatusReview    PRDStatus = "review"
	PRDStatusApproved  PRDStatus = "approved"
	PRDStatusPublished PRDStatus = "published"
	PRDStatusArchived  PRDStatus = "archived"
)

// PRDContent represents the structured content of a PRD
type PRDContent struct {
	Overview           string          `json:"overview,omitempty"`
	ProblemStatement   string          `json:"problem_statement,omitempty"`
	Goals              []string        `json:"goals,omitempty"`
	UserPersonas       []UserPersona   `json:"user_personas,omitempty"`
	UserStories        []UserStory     `json:"user_stories,omitempty"`
	Requirements       []Requirement   `json:"requirements,omitempty"`
	AcceptanceCriteria []string        `json:"acceptance_criteria,omitempty"`
	OutOfScope         []string        `json:"out_of_scope,omitempty"`
	Assumptions        []string        `json:"assumptions,omitempty"`
	Dependencies       []string        `json:"dependencies,omitempty"`
	SuccessMetrics     []SuccessMetric `json:"success_metrics,omitempty"`
	Timeline           *Timeline       `json:"timeline,omitempty"`

	// For custom template fields
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// Value implements driver.Valuer for database storage
func (c PRDContent) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements sql.Scanner for database retrieval
func (c *PRDContent) Scan(value interface{}) error {
	if value == nil {
		*c = PRDContent{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
}

// UserPersona represents a target user persona
type UserPersona struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Goals       []string `json:"goals,omitempty"`
	PainPoints  []string `json:"pain_points,omitempty"`
}

// UserStory represents a user story in the PRD
type UserStory struct {
	ID                 string   `json:"id"`
	AsA                string   `json:"as_a"`          // "As a [persona]"
	IWant              string   `json:"i_want"`        // "I want [capability]"
	SoThat             string   `json:"so_that"`       // "So that [benefit]"
	AcceptanceCriteria []string `json:"acceptance_criteria,omitempty"`
	Priority           string   `json:"priority,omitempty"` // high, medium, low
	EstimatePoints     int      `json:"estimate_points,omitempty"`
}

// Format returns the user story in standard format
func (s *UserStory) Format() string {
	story := "As a " + s.AsA + ", I want " + s.IWant
	if s.SoThat != "" {
		story += ", so that " + s.SoThat
	}
	return story
}

// Requirement represents a product requirement
type Requirement struct {
	ID          string `json:"id"`
	Type        string `json:"type"` // functional, non_functional
	Description string `json:"description"`
	Priority    string `json:"priority,omitempty"` // high, medium, low
	Status      string `json:"status,omitempty"`   // pending, approved, rejected
}

// RequirementType constants
const (
	RequirementTypeFunctional    = "functional"
	RequirementTypeNonFunctional = "non_functional"
)

// SuccessMetric represents a measurable success metric
type SuccessMetric struct {
	Name       string `json:"name"`
	Definition string `json:"definition,omitempty"`
	Target     string `json:"target"`
	Source     string `json:"source,omitempty"`
}

// Timeline represents the project timeline
type Timeline struct {
	StartDate  string      `json:"start_date,omitempty"`
	EndDate    string      `json:"end_date,omitempty"`
	Milestones []Milestone `json:"milestones,omitempty"`
}

// Milestone represents a project milestone
type Milestone struct {
	Name        string `json:"name"`
	Date        string `json:"date"`
	Description string `json:"description,omitempty"`
}

// PRDVersion represents a historical version of a PRD
type PRDVersion struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	PRDID         uuid.UUID  `json:"prd_id" db:"prd_id"`
	VersionNumber int        `json:"version_number" db:"version_number"`
	Content       PRDContent `json:"content" db:"content"`
	ChangedBy     *uuid.UUID `json:"changed_by,omitempty" db:"changed_by"`
	ChangeSummary string     `json:"change_summary,omitempty" db:"change_summary"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// PRDFilter represents filter options for listing PRDs
type PRDFilter struct {
	OrganizationID *uuid.UUID
	Status         *PRDStatus
	OwnerID        *uuid.UUID
	TemplateID     *uuid.UUID
	Search         string
	Limit          int
	Offset         int
}

// CreatePRDRequest represents the request to create a new PRD
type CreatePRDRequest struct {
	Title      string      `json:"title" binding:"required"`
	TemplateID *uuid.UUID  `json:"template_id"`
	Content    *PRDContent `json:"content"`
}

// UpdatePRDRequest represents the request to update a PRD
type UpdatePRDRequest struct {
	Title   *string     `json:"title"`
	Status  *PRDStatus  `json:"status"`
	Content *PRDContent `json:"content"`
}

// GeneratePRDRequest represents the request to generate a PRD from notes
type GeneratePRDRequest struct {
	Notes      string     `json:"notes" binding:"required"`
	TemplateID *uuid.UUID `json:"template_id"`
	Title      string     `json:"title"`
}

// RefinePRDRequest represents the request to refine a PRD
type RefinePRDRequest struct {
	Feedback string `json:"feedback" binding:"required"`
	Section  string `json:"section,omitempty"` // Optional: specific section to refine
}
