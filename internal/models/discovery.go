package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// DiscoveryItem represents a product discovery artifact
type DiscoveryItem struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	OrganizationID uuid.UUID       `json:"organization_id" db:"organization_id"`
	Type           DiscoveryType   `json:"type" db:"type"`
	Title          string          `json:"title" db:"title"`
	Content        DiscoveryContent `json:"content" db:"content"`
	Tags           pq.StringArray  `json:"tags" db:"tags"`
	LinkedPRDIDs   pq.StringArray  `json:"linked_prd_ids" db:"linked_prd_ids"`
	AIInsights     *AIInsights     `json:"ai_insights,omitempty" db:"ai_insights"`
	CreatedBy      *uuid.UUID      `json:"created_by,omitempty" db:"created_by"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// DiscoveryType represents the type of discovery item
type DiscoveryType string

const (
	DiscoveryTypeInterview   DiscoveryType = "interview"
	DiscoveryTypeResearch    DiscoveryType = "research"
	DiscoveryTypeOpportunity DiscoveryType = "opportunity"
	DiscoveryTypeSolution    DiscoveryType = "solution"
	DiscoveryTypeJTBD        DiscoveryType = "jtbd" // Jobs to be Done
)

// DiscoveryContent represents the structured content of a discovery item
type DiscoveryContent struct {
	// Common fields
	Summary     string   `json:"summary,omitempty"`
	Notes       string   `json:"notes,omitempty"`
	KeyFindings []string `json:"key_findings,omitempty"`

	// Interview-specific
	InterviewDetails *InterviewDetails `json:"interview_details,omitempty"`

	// Research-specific
	ResearchDetails *ResearchDetails `json:"research_details,omitempty"`

	// Opportunity-specific
	OpportunityDetails *OpportunityDetails `json:"opportunity_details,omitempty"`

	// Solution-specific
	SolutionDetails *SolutionDetails `json:"solution_details,omitempty"`

	// JTBD-specific
	JTBDDetails *JTBDDetails `json:"jtbd_details,omitempty"`
}

// Value implements driver.Valuer for database storage
func (c DiscoveryContent) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements sql.Scanner for database retrieval
func (c *DiscoveryContent) Scan(value interface{}) error {
	if value == nil {
		*c = DiscoveryContent{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
}

// InterviewDetails represents interview-specific content
type InterviewDetails struct {
	IntervieweeName  string     `json:"interviewee_name,omitempty"`
	IntervieweeRole  string     `json:"interviewee_role,omitempty"`
	IntervieweeCompany string   `json:"interviewee_company,omitempty"`
	InterviewDate    *time.Time `json:"interview_date,omitempty"`
	Duration         int        `json:"duration_minutes,omitempty"`
	Questions        []QAPair   `json:"questions,omitempty"`
	Quotes           []string   `json:"quotes,omitempty"` // Notable quotes
	PainPoints       []string   `json:"pain_points,omitempty"`
	Needs            []string   `json:"needs,omitempty"`
}

// QAPair represents a question-answer pair
type QAPair struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// ResearchDetails represents research-specific content
type ResearchDetails struct {
	ResearchType string   `json:"research_type,omitempty"` // qualitative, quantitative, competitive
	Sources      []string `json:"sources,omitempty"`
	Methodology  string   `json:"methodology,omitempty"`
	SampleSize   int      `json:"sample_size,omitempty"`
	Findings     []ResearchFinding `json:"findings,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// ResearchFinding represents a research finding
type ResearchFinding struct {
	Finding    string `json:"finding"`
	Evidence   string `json:"evidence,omitempty"`
	Confidence string `json:"confidence,omitempty"` // high, medium, low
}

// OpportunityDetails represents opportunity-specific content
type OpportunityDetails struct {
	Description       string   `json:"description"`
	TargetUsers       []string `json:"target_users,omitempty"`
	PotentialImpact   string   `json:"potential_impact,omitempty"`
	LinkedSolutions   []string `json:"linked_solutions,omitempty"`
	ParentOpportunity string   `json:"parent_opportunity,omitempty"` // For opportunity trees
	ChildOpportunities []string `json:"child_opportunities,omitempty"`
}

// SolutionDetails represents solution-specific content
type SolutionDetails struct {
	Description     string   `json:"description"`
	TargetOpportunity string  `json:"target_opportunity,omitempty"`
	Assumptions     []string `json:"assumptions,omitempty"`
	Risks           []string `json:"risks,omitempty"`
	EffortEstimate  string   `json:"effort_estimate,omitempty"`
	Validated       bool     `json:"validated"`
	ValidationNotes string   `json:"validation_notes,omitempty"`
}

// JTBDDetails represents Jobs-to-be-Done specific content
type JTBDDetails struct {
	Job           string   `json:"job"`                     // The job to be done
	JobStatement  string   `json:"job_statement,omitempty"` // Full job statement
	Circumstance  string   `json:"circumstance,omitempty"`  // When/where
	Motivation    string   `json:"motivation,omitempty"`    // Why
	ExpectedOutcome string `json:"expected_outcome,omitempty"`
	CurrentSolutions []string `json:"current_solutions,omitempty"`
	Frustrations    []string `json:"frustrations,omitempty"`
	DesiredOutcomes []DesiredOutcome `json:"desired_outcomes,omitempty"`
}

// DesiredOutcome represents a desired outcome in JTBD
type DesiredOutcome struct {
	Outcome    string `json:"outcome"`
	Direction  string `json:"direction,omitempty"`  // minimize, maximize
	Importance int    `json:"importance,omitempty"` // 1-10
	Satisfaction int  `json:"satisfaction,omitempty"` // 1-10 (current)
}

// AIInsights represents AI-generated insights for a discovery item
type AIInsights struct {
	Summary           string   `json:"summary,omitempty"`
	KeyThemes         []string `json:"key_themes,omitempty"`
	SuggestedActions  []string `json:"suggested_actions,omitempty"`
	RelatedItems      []string `json:"related_items,omitempty"` // IDs of related discovery items
	GeneratedAt       time.Time `json:"generated_at"`
}

// Value implements driver.Valuer for database storage
func (a *AIInsights) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan implements sql.Scanner for database retrieval
func (a *AIInsights) Scan(value interface{}) error {
	if value == nil {
		*a = AIInsights{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a)
}

// DiscoveryFilter represents filter options for listing discovery items
type DiscoveryFilter struct {
	OrganizationID *uuid.UUID
	Type           *DiscoveryType
	Tags           []string
	LinkedPRDID    *uuid.UUID
	Search         string
	Limit          int
	Offset         int
}

// CreateDiscoveryRequest represents the request to create a discovery item
type CreateDiscoveryRequest struct {
	Type    DiscoveryType    `json:"type" binding:"required"`
	Title   string           `json:"title" binding:"required"`
	Content DiscoveryContent `json:"content" binding:"required"`
	Tags    []string         `json:"tags"`
}

// UpdateDiscoveryRequest represents the request to update a discovery item
type UpdateDiscoveryRequest struct {
	Title        *string           `json:"title"`
	Content      *DiscoveryContent `json:"content"`
	Tags         []string          `json:"tags"`
	LinkedPRDIDs []uuid.UUID       `json:"linked_prd_ids"`
}

// GenerateInsightsRequest represents the request to generate AI insights
type GenerateInsightsRequest struct {
	DiscoveryItemID uuid.UUID `json:"discovery_item_id" binding:"required"`
	IncludeRelated  bool      `json:"include_related"` // Analyze related items too
}
