package models

import "time"

// JiraIssue represents a Jira issue
type JiraIssue struct {
	ID          string            `json:"id,omitempty"`
	Key         string            `json:"key,omitempty"`
	Self        string            `json:"self,omitempty"`
	ProjectKey  string            `json:"project_key"`
	Summary     string            `json:"summary"`
	Description string            `json:"description,omitempty"`
	IssueType   string            `json:"issue_type"` // epic, story, task, bug
	Status      string            `json:"status,omitempty"`
	Priority    string            `json:"priority,omitempty"`
	Assignee    string            `json:"assignee,omitempty"`
	Reporter    string            `json:"reporter,omitempty"`
	Labels      []string          `json:"labels,omitempty"`
	Components  []string          `json:"components,omitempty"`
	EpicKey     string            `json:"epic_key,omitempty"` // Parent epic key
	StoryPoints *int              `json:"story_points,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	CreatedAt   *time.Time        `json:"created_at,omitempty"`
	UpdatedAt   *time.Time        `json:"updated_at,omitempty"`
}

// JiraIssueType constants
const (
	JiraIssueTypeEpic    = "Epic"
	JiraIssueTypeStory   = "Story"
	JiraIssueTypeTask    = "Task"
	JiraIssueTypeBug     = "Bug"
	JiraIssueTypeSubtask = "Sub-task"
)

// JiraIssueUpdate represents an update to a Jira issue
type JiraIssueUpdate struct {
	Summary      *string           `json:"summary,omitempty"`
	Description  *string           `json:"description,omitempty"`
	Status       *string           `json:"status,omitempty"`
	Priority     *string           `json:"priority,omitempty"`
	Assignee     *string           `json:"assignee,omitempty"`
	Labels       []string          `json:"labels,omitempty"`
	StoryPoints  *int              `json:"story_points,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// JiraBulkUpdate represents a bulk update operation
type JiraBulkUpdate struct {
	IssueKey string          `json:"issue_key"`
	Update   JiraIssueUpdate `json:"update"`
}

// JiraSprint represents a Jira sprint
type JiraSprint struct {
	ID            int       `json:"id"`
	Self          string    `json:"self,omitempty"`
	Name          string    `json:"name"`
	State         string    `json:"state"` // future, active, closed
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	CompleteDate  *time.Time `json:"complete_date,omitempty"`
	OriginBoardID int       `json:"origin_board_id,omitempty"`
	Goal          string    `json:"goal,omitempty"`
}

// JiraProject represents a Jira project
type JiraProject struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ProjectType string `json:"project_type,omitempty"`
	Style       string `json:"style,omitempty"` // classic, next-gen
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// JiraBoard represents a Jira board
type JiraBoard struct {
	ID        int         `json:"id"`
	Self      string      `json:"self,omitempty"`
	Name      string      `json:"name"`
	Type      string      `json:"type"` // scrum, kanban
	ProjectID string      `json:"project_id,omitempty"`
	Location  *JiraProject `json:"location,omitempty"`
}

// CreateJiraIssueRequest represents the API request to create a Jira issue
type CreateJiraIssueRequest struct {
	ProjectKey  string   `json:"project_key" binding:"required"`
	Summary     string   `json:"summary" binding:"required"`
	Description string   `json:"description"`
	IssueType   string   `json:"issue_type" binding:"required"`
	Priority    string   `json:"priority,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	EpicKey     string   `json:"epic_key,omitempty"`
	StoryPoints *int     `json:"story_points,omitempty"`
}

// SyncPRDToJiraRequest represents the request to sync a PRD to Jira
type SyncPRDToJiraRequest struct {
	ProjectKey      string `json:"project_key" binding:"required"`
	CreateEpic      bool   `json:"create_epic"`
	CreateStories   bool   `json:"create_stories"`
	IncludeAcceptanceCriteria bool `json:"include_acceptance_criteria"`
}

// SyncPRDToJiraResponse represents the response from syncing a PRD to Jira
type SyncPRDToJiraResponse struct {
	EpicKey       string        `json:"epic_key,omitempty"`
	EpicURL       string        `json:"epic_url,omitempty"`
	StoriesCreated []JiraIssue  `json:"stories_created,omitempty"`
	Errors        []string      `json:"errors,omitempty"`
}

// JiraSearchRequest represents a JQL search request
type JiraSearchRequest struct {
	JQL        string   `json:"jql" binding:"required"`
	MaxResults int      `json:"max_results"`
	StartAt    int      `json:"start_at"`
	Fields     []string `json:"fields,omitempty"`
}

// JiraSearchResponse represents a JQL search response
type JiraSearchResponse struct {
	StartAt    int          `json:"start_at"`
	MaxResults int          `json:"max_results"`
	Total      int          `json:"total"`
	Issues     []JiraIssue  `json:"issues"`
}

// JiraSyncLog represents a sync operation log entry
type JiraSyncLog struct {
	ID              string    `json:"id" db:"id"`
	PRDID           *string   `json:"prd_id,omitempty" db:"prd_id"`
	JiraIssueKey    string    `json:"jira_issue_key" db:"jira_issue_key"`
	JiraIssueID     string    `json:"jira_issue_id,omitempty" db:"jira_issue_id"`
	IssueType       string    `json:"issue_type,omitempty" db:"issue_type"`
	SyncDirection   string    `json:"sync_direction" db:"sync_direction"` // to_jira, from_jira
	SyncStatus      string    `json:"sync_status" db:"sync_status"`       // pending, success, failed
	RequestPayload  string    `json:"request_payload,omitempty" db:"request_payload"`
	ResponsePayload string    `json:"response_payload,omitempty" db:"response_payload"`
	ErrorMessage    string    `json:"error_message,omitempty" db:"error_message"`
	SyncedAt        time.Time `json:"synced_at" db:"synced_at"`
}

// SyncDirection constants
const (
	SyncDirectionToJira   = "to_jira"
	SyncDirectionFromJira = "from_jira"
)

// SyncStatus constants
const (
	SyncStatusPending = "pending"
	SyncStatusSuccess = "success"
	SyncStatusFailed  = "failed"
)
