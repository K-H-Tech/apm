package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Issue represents a Jira issue
type Issue struct {
	ID         string            `json:"id"`
	Key        string            `json:"key"`
	Self       string            `json:"self"`
	Fields     IssueFields       `json:"fields"`
	Changelog  *Changelog        `json:"changelog,omitempty"`
}

// IssueFields represents the fields of a Jira issue
type IssueFields struct {
	Summary       string      `json:"summary"`
	Description   *ADF        `json:"description,omitempty"`
	IssueType     *IssueType  `json:"issuetype,omitempty"`
	Project       *Project    `json:"project,omitempty"`
	Status        *Status     `json:"status,omitempty"`
	Priority      *Priority   `json:"priority,omitempty"`
	Assignee      *User       `json:"assignee,omitempty"`
	Reporter      *User       `json:"reporter,omitempty"`
	Labels        []string    `json:"labels,omitempty"`
	Created       string      `json:"created,omitempty"`
	Updated       string      `json:"updated,omitempty"`
	DueDate       string      `json:"duedate,omitempty"`
	StoryPoints   *float64    `json:"customfield_10016,omitempty"` // Common story points field
	Epic          *Epic       `json:"parent,omitempty"`
	Subtasks      []Issue     `json:"subtasks,omitempty"`
	Components    []Component `json:"components,omitempty"`
	FixVersions   []Version   `json:"fixVersions,omitempty"`
	TimeTracking  *TimeTracking `json:"timetracking,omitempty"`
}

// ADF represents Atlassian Document Format
type ADF struct {
	Version int           `json:"version"`
	Type    string        `json:"type"`
	Content []interface{} `json:"content"`
}

// TextADF creates a simple text ADF document
func TextADF(text string) *ADF {
	return &ADF{
		Version: 1,
		Type:    "doc",
		Content: []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": text,
					},
				},
			},
		},
	}
}

// Epic represents a Jira epic
type Epic struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Self   string `json:"self"`
	Fields struct {
		Summary string `json:"summary"`
	} `json:"fields,omitempty"`
}

// Component represents a Jira component
type Component struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Version represents a Jira version
type Version struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Released    bool   `json:"released"`
	ReleaseDate string `json:"releaseDate,omitempty"`
}

// TimeTracking represents time tracking information
type TimeTracking struct {
	OriginalEstimate        string `json:"originalEstimate,omitempty"`
	RemainingEstimate       string `json:"remainingEstimate,omitempty"`
	TimeSpent               string `json:"timeSpent,omitempty"`
	OriginalEstimateSeconds int    `json:"originalEstimateSeconds,omitempty"`
	RemainingEstimateSeconds int   `json:"remainingEstimateSeconds,omitempty"`
	TimeSpentSeconds        int    `json:"timeSpentSeconds,omitempty"`
}

// Changelog represents issue changelog
type Changelog struct {
	Histories []struct {
		ID      string `json:"id"`
		Author  User   `json:"author"`
		Created string `json:"created"`
		Items   []struct {
			Field      string `json:"field"`
			FieldType  string `json:"fieldtype"`
			FromString string `json:"fromString"`
			ToString   string `json:"toString"`
		} `json:"items"`
	} `json:"histories"`
}

// CreateIssueRequest represents a request to create an issue
type CreateIssueRequest struct {
	Fields CreateIssueFields `json:"fields"`
}

// CreateIssueFields represents fields for creating an issue
type CreateIssueFields struct {
	Project     ProjectRef     `json:"project"`
	Summary     string         `json:"summary"`
	Description *ADF           `json:"description,omitempty"`
	IssueType   IssueTypeRef   `json:"issuetype"`
	Assignee    *UserRef       `json:"assignee,omitempty"`
	Priority    *PriorityRef   `json:"priority,omitempty"`
	Labels      []string       `json:"labels,omitempty"`
	Parent      *ParentRef     `json:"parent,omitempty"`
	Components  []ComponentRef `json:"components,omitempty"`
	FixVersions []VersionRef   `json:"fixVersions,omitempty"`
	DueDate     string         `json:"duedate,omitempty"`
}

// ProjectRef is a reference to a project
type ProjectRef struct {
	Key string `json:"key,omitempty"`
	ID  string `json:"id,omitempty"`
}

// IssueTypeRef is a reference to an issue type
type IssueTypeRef struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

// UserRef is a reference to a user
type UserRef struct {
	AccountID string `json:"accountId"`
}

// PriorityRef is a reference to a priority
type PriorityRef struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

// ParentRef is a reference to a parent issue (for subtasks or issues in epics)
type ParentRef struct {
	Key string `json:"key,omitempty"`
	ID  string `json:"id,omitempty"`
}

// ComponentRef is a reference to a component
type ComponentRef struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

// VersionRef is a reference to a version
type VersionRef struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

// CreateIssueResponse represents the response from creating an issue
type CreateIssueResponse struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Self string `json:"self"`
}

// CreateIssue creates a new issue
func (c *Client) CreateIssue(ctx context.Context, req *CreateIssueRequest) (*CreateIssueResponse, error) {
	resp, err := c.doRequest(ctx, "POST", baseAPIPath+"/issue", req)
	if err != nil {
		return nil, err
	}

	var result CreateIssueResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetIssue retrieves an issue by key or ID
func (c *Client) GetIssue(ctx context.Context, issueKeyOrID string, expand []string) (*Issue, error) {
	params := url.Values{}
	if len(expand) > 0 {
		for _, e := range expand {
			params.Add("expand", e)
		}
	}

	path := baseAPIPath + "/issue/" + issueKeyOrID
	resp, err := c.doRequestWithQuery(ctx, "GET", path, params, nil)
	if err != nil {
		return nil, err
	}

	var issue Issue
	if err := json.Unmarshal(resp, &issue); err != nil {
		return nil, err
	}

	return &issue, nil
}

// UpdateIssueRequest represents a request to update an issue
type UpdateIssueRequest struct {
	Fields map[string]interface{} `json:"fields,omitempty"`
	Update map[string][]UpdateOperation `json:"update,omitempty"`
}

// UpdateOperation represents an update operation
type UpdateOperation struct {
	Set    interface{} `json:"set,omitempty"`
	Add    interface{} `json:"add,omitempty"`
	Remove interface{} `json:"remove,omitempty"`
}

// UpdateIssue updates an issue
func (c *Client) UpdateIssue(ctx context.Context, issueKeyOrID string, req *UpdateIssueRequest) error {
	path := baseAPIPath + "/issue/" + issueKeyOrID
	_, err := c.doRequest(ctx, "PUT", path, req)
	return err
}

// DeleteIssue deletes an issue
func (c *Client) DeleteIssue(ctx context.Context, issueKeyOrID string, deleteSubtasks bool) error {
	params := url.Values{
		"deleteSubtasks": {fmt.Sprintf("%t", deleteSubtasks)},
	}

	path := baseAPIPath + "/issue/" + issueKeyOrID
	_, err := c.doRequestWithQuery(ctx, "DELETE", path, params, nil)
	return err
}

// TransitionIssue transitions an issue to a new status
func (c *Client) TransitionIssue(ctx context.Context, issueKeyOrID string, transitionID string) error {
	body := map[string]interface{}{
		"transition": map[string]string{
			"id": transitionID,
		},
	}

	path := baseAPIPath + "/issue/" + issueKeyOrID + "/transitions"
	_, err := c.doRequest(ctx, "POST", path, body)
	return err
}

// GetTransitions retrieves available transitions for an issue
func (c *Client) GetTransitions(ctx context.Context, issueKeyOrID string) ([]Transition, error) {
	path := baseAPIPath + "/issue/" + issueKeyOrID + "/transitions"
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Transitions []Transition `json:"transitions"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result.Transitions, nil
}

// Transition represents an available issue transition
type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	To   Status `json:"to"`
}

// AssignIssue assigns an issue to a user
func (c *Client) AssignIssue(ctx context.Context, issueKeyOrID string, accountID string) error {
	body := map[string]interface{}{
		"accountId": accountID,
	}

	path := baseAPIPath + "/issue/" + issueKeyOrID + "/assignee"
	_, err := c.doRequest(ctx, "PUT", path, body)
	return err
}

// AddComment adds a comment to an issue
func (c *Client) AddComment(ctx context.Context, issueKeyOrID string, body *ADF) (*Comment, error) {
	reqBody := map[string]interface{}{
		"body": body,
	}

	path := baseAPIPath + "/issue/" + issueKeyOrID + "/comment"
	resp, err := c.doRequest(ctx, "POST", path, reqBody)
	if err != nil {
		return nil, err
	}

	var comment Comment
	if err := json.Unmarshal(resp, &comment); err != nil {
		return nil, err
	}

	return &comment, nil
}

// Comment represents a Jira comment
type Comment struct {
	ID      string `json:"id"`
	Self    string `json:"self"`
	Author  User   `json:"author"`
	Body    *ADF   `json:"body"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

// GetComments retrieves comments for an issue
func (c *Client) GetComments(ctx context.Context, issueKeyOrID string, startAt, maxResults int) ([]Comment, int, error) {
	params := url.Values{
		"startAt":    {fmt.Sprintf("%d", startAt)},
		"maxResults": {fmt.Sprintf("%d", maxResults)},
	}

	path := baseAPIPath + "/issue/" + issueKeyOrID + "/comment"
	resp, err := c.doRequestWithQuery(ctx, "GET", path, params, nil)
	if err != nil {
		return nil, 0, err
	}

	var result struct {
		Comments []Comment `json:"comments"`
		Total    int       `json:"total"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, 0, err
	}

	return result.Comments, result.Total, nil
}

// SearchResult represents JQL search results
type SearchResult struct {
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	Issues     []Issue `json:"issues"`
}

// SearchIssues searches for issues using JQL
func (c *Client) SearchIssues(ctx context.Context, jql string, startAt, maxResults int, fields []string) (*SearchResult, error) {
	body := map[string]interface{}{
		"jql":        jql,
		"startAt":    startAt,
		"maxResults": maxResults,
	}
	if len(fields) > 0 {
		body["fields"] = fields
	}

	resp, err := c.doRequest(ctx, "POST", baseAPIPath+"/search", body)
	if err != nil {
		return nil, err
	}

	var result SearchResult
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateEpic creates a new epic
func (c *Client) CreateEpic(ctx context.Context, projectKey, summary, description string) (*CreateIssueResponse, error) {
	req := &CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:   ProjectRef{Key: projectKey},
			Summary:   summary,
			IssueType: IssueTypeRef{Name: "Epic"},
		},
	}

	if description != "" {
		req.Fields.Description = TextADF(description)
	}

	return c.CreateIssue(ctx, req)
}

// CreateStory creates a new story
func (c *Client) CreateStory(ctx context.Context, projectKey, summary, description string, epicKey string) (*CreateIssueResponse, error) {
	req := &CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:   ProjectRef{Key: projectKey},
			Summary:   summary,
			IssueType: IssueTypeRef{Name: "Story"},
		},
	}

	if description != "" {
		req.Fields.Description = TextADF(description)
	}

	if epicKey != "" {
		req.Fields.Parent = &ParentRef{Key: epicKey}
	}

	return c.CreateIssue(ctx, req)
}

// CreateTask creates a new task
func (c *Client) CreateTask(ctx context.Context, projectKey, summary, description string) (*CreateIssueResponse, error) {
	req := &CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:   ProjectRef{Key: projectKey},
			Summary:   summary,
			IssueType: IssueTypeRef{Name: "Task"},
		},
	}

	if description != "" {
		req.Fields.Description = TextADF(description)
	}

	return c.CreateIssue(ctx, req)
}

// CreateSubtask creates a new subtask
func (c *Client) CreateSubtask(ctx context.Context, parentKey, summary, description string) (*CreateIssueResponse, error) {
	// Get parent issue to determine project
	parent, err := c.GetIssue(ctx, parentKey, nil)
	if err != nil {
		return nil, err
	}

	if parent.Fields.Project == nil {
		return nil, fmt.Errorf("parent issue %s has no project information", parentKey)
	}

	req := &CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:   ProjectRef{Key: parent.Fields.Project.Key},
			Summary:   summary,
			IssueType: IssueTypeRef{Name: "Sub-task"},
			Parent:    &ParentRef{Key: parentKey},
		},
	}

	if description != "" {
		req.Fields.Description = TextADF(description)
	}

	return c.CreateIssue(ctx, req)
}

// GetEpicIssues retrieves all issues in an epic
func (c *Client) GetEpicIssues(ctx context.Context, epicKey string, startAt, maxResults int) (*SearchResult, error) {
	jql := fmt.Sprintf("parent = %s ORDER BY rank ASC", epicKey)
	return c.SearchIssues(ctx, jql, startAt, maxResults, nil)
}

// LinkIssues creates a link between two issues
func (c *Client) LinkIssues(ctx context.Context, inwardIssueKey, outwardIssueKey, linkType string) error {
	body := map[string]interface{}{
		"type": map[string]string{
			"name": linkType,
		},
		"inwardIssue": map[string]string{
			"key": inwardIssueKey,
		},
		"outwardIssue": map[string]string{
			"key": outwardIssueKey,
		},
	}

	_, err := c.doRequest(ctx, "POST", baseAPIPath+"/issueLink", body)
	return err
}

// AddLabels adds labels to an issue
func (c *Client) AddLabels(ctx context.Context, issueKeyOrID string, labels []string) error {
	operations := make([]UpdateOperation, len(labels))
	for i, label := range labels {
		operations[i] = UpdateOperation{Add: label}
	}

	req := &UpdateIssueRequest{
		Update: map[string][]UpdateOperation{
			"labels": operations,
		},
	}

	return c.UpdateIssue(ctx, issueKeyOrID, req)
}

// RemoveLabels removes labels from an issue
func (c *Client) RemoveLabels(ctx context.Context, issueKeyOrID string, labels []string) error {
	operations := make([]UpdateOperation, len(labels))
	for i, label := range labels {
		operations[i] = UpdateOperation{Remove: label}
	}

	req := &UpdateIssueRequest{
		Update: map[string][]UpdateOperation{
			"labels": operations,
		},
	}

	return c.UpdateIssue(ctx, issueKeyOrID, req)
}
