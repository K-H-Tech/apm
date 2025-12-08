package contract

import (
	"context"

	"github.com/K-H-Tech/apm/internal/models"
)

// JiraClient defines Jira API operations
type JiraClient interface {
	// Issue operations
	CreateIssue(ctx context.Context, issue *models.JiraIssue) (*models.JiraIssue, error)
	UpdateIssue(ctx context.Context, issueKey string, update *models.JiraIssueUpdate) error
	GetIssue(ctx context.Context, issueKey string) (*models.JiraIssue, error)
	DeleteIssue(ctx context.Context, issueKey string) error

	// Search
	SearchIssues(ctx context.Context, jql string, maxResults int, startAt int) (*models.JiraSearchResponse, error)

	// Bulk operations
	CreateBulkIssues(ctx context.Context, issues []*models.JiraIssue) ([]*models.JiraIssue, []error)
	UpdateBulkIssues(ctx context.Context, updates []*models.JiraBulkUpdate) []error

	// Sprint operations
	GetActiveSprints(ctx context.Context, boardID int) ([]*models.JiraSprint, error)
	GetSprintIssues(ctx context.Context, sprintID int) ([]*models.JiraIssue, error)
	MoveToSprint(ctx context.Context, issueKeys []string, sprintID int) error
	MoveToBacklog(ctx context.Context, issueKeys []string) error

	// Backlog management
	GetBacklog(ctx context.Context, boardID int, maxResults int) ([]*models.JiraIssue, error)
	RankIssues(ctx context.Context, issueKeys []string, rankBeforeKey string) error

	// Project operations
	GetProjects(ctx context.Context) ([]*models.JiraProject, error)
	GetProject(ctx context.Context, projectKey string) (*models.JiraProject, error)

	// Board operations
	GetBoards(ctx context.Context, projectKey string) ([]*models.JiraBoard, error)

	// Transitions
	GetTransitions(ctx context.Context, issueKey string) ([]JiraTransition, error)
	TransitionIssue(ctx context.Context, issueKey string, transitionID string) error
}

// JiraTransition represents an available status transition
type JiraTransition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	To   struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"to"`
}

// JiraClientConfig represents configuration for Jira client
type JiraClientConfig struct {
	BaseURL     string `yaml:"BASE_URL"`
	CloudID     string `yaml:"CLOUD_ID"`
	AccessToken string // Set at runtime from OAuth
	Timeout     int    `yaml:"TIMEOUT_SECONDS"`
	MaxRetries  int    `yaml:"MAX_RETRIES"`
}

// JiraService defines high-level Jira integration operations
type JiraService interface {
	// PRD to Jira sync
	SyncPRDToJira(ctx context.Context, prdID string, req *models.SyncPRDToJiraRequest) (*models.SyncPRDToJiraResponse, error)

	// Issue management
	CreateIssueFromStory(ctx context.Context, story *models.UserStory, projectKey string, epicKey string) (*models.JiraIssue, error)

	// Backlog management
	GetPrioritizedBacklog(ctx context.Context, projectKey string, framework string) ([]*models.JiraIssue, error)

	// Sprint planning
	SuggestSprintItems(ctx context.Context, boardID int, capacity int) ([]*models.JiraIssue, error)
}
