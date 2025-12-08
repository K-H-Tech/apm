package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	// ErrUnauthorized is returned when the request is unauthorized
	ErrUnauthorized = errors.New("unauthorized: invalid or expired token")

	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrBadRequest is returned when the request is malformed
	ErrBadRequest = errors.New("bad request")

	// ErrForbidden is returned when access is forbidden
	ErrForbidden = errors.New("forbidden: insufficient permissions")

	// ErrRateLimited is returned when rate limited
	ErrRateLimited = errors.New("rate limited: too many requests")
)

const (
	apiVersion = "3"
	baseAPIPath = "/rest/api/3"
	agileAPIPath = "/rest/agile/1.0"
)

// Client is a Jira REST API client
type Client struct {
	cloudID     string
	accessToken string
	httpClient  *http.Client
	baseURL     string
}

// ClientConfig holds Jira client configuration
type ClientConfig struct {
	CloudID     string
	AccessToken string
	Timeout     time.Duration
}

// NewClient creates a new Jira client
func NewClient(config ClientConfig) *Client {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		cloudID:     config.CloudID,
		accessToken: config.AccessToken,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: fmt.Sprintf("https://api.atlassian.com/ex/jira/%s", config.CloudID),
	}
}

// SetAccessToken updates the access token
func (c *Client) SetAccessToken(token string) {
	c.accessToken = token
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		return respBody, nil
	case http.StatusUnauthorized:
		return nil, ErrUnauthorized
	case http.StatusForbidden:
		return nil, ErrForbidden
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusBadRequest:
		return nil, fmt.Errorf("%w: %s", ErrBadRequest, string(respBody))
	case http.StatusTooManyRequests:
		return nil, ErrRateLimited
	default:
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}
}

// doRequestWithQuery performs an HTTP request with query parameters
func (c *Client) doRequestWithQuery(ctx context.Context, method, path string, query url.Values, body interface{}) ([]byte, error) {
	if len(query) > 0 {
		path = path + "?" + query.Encode()
	}
	return c.doRequest(ctx, method, path, body)
}

// Project represents a Jira project
type Project struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ProjectType string `json:"projectTypeKey"`
	Style       string `json:"style"`
	AvatarURLs  map[string]string `json:"avatarUrls,omitempty"`
}

// GetProjects retrieves all accessible projects
func (c *Client) GetProjects(ctx context.Context) ([]Project, error) {
	resp, err := c.doRequest(ctx, "GET", baseAPIPath+"/project", nil)
	if err != nil {
		return nil, err
	}

	var projects []Project
	if err := json.Unmarshal(resp, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// GetProject retrieves a project by key
func (c *Client) GetProject(ctx context.Context, projectKey string) (*Project, error) {
	resp, err := c.doRequest(ctx, "GET", baseAPIPath+"/project/"+url.PathEscape(projectKey), nil)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal(resp, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// IssueType represents a Jira issue type
type IssueType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Subtask     bool   `json:"subtask"`
	IconURL     string `json:"iconUrl"`
}

// GetIssueTypes retrieves issue types for a project
func (c *Client) GetIssueTypes(ctx context.Context, projectKey string) ([]IssueType, error) {
	resp, err := c.doRequest(ctx, "GET", baseAPIPath+"/project/"+url.PathEscape(projectKey)+"/statuses", nil)
	if err != nil {
		return nil, err
	}

	var result []struct {
		IssueType IssueType `json:"issueType"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	issueTypes := make([]IssueType, len(result))
	for i, r := range result {
		issueTypes[i] = r.IssueType
	}

	return issueTypes, nil
}

// User represents a Jira user
type User struct {
	AccountID    string            `json:"accountId"`
	DisplayName  string            `json:"displayName"`
	EmailAddress string            `json:"emailAddress,omitempty"`
	AvatarURLs   map[string]string `json:"avatarUrls,omitempty"`
	Active       bool              `json:"active"`
}

// GetCurrentUser retrieves the current authenticated user
func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	resp, err := c.doRequest(ctx, "GET", baseAPIPath+"/myself", nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// SearchUsers searches for users
func (c *Client) SearchUsers(ctx context.Context, query string) ([]User, error) {
	params := url.Values{
		"query": {query},
	}
	resp, err := c.doRequestWithQuery(ctx, "GET", baseAPIPath+"/user/search", params, nil)
	if err != nil {
		return nil, err
	}

	var users []User
	if err := json.Unmarshal(resp, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// Priority represents a Jira priority
type Priority struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconURL     string `json:"iconUrl"`
}

// GetPriorities retrieves all priorities
func (c *Client) GetPriorities(ctx context.Context) ([]Priority, error) {
	resp, err := c.doRequest(ctx, "GET", baseAPIPath+"/priority", nil)
	if err != nil {
		return nil, err
	}

	var priorities []Priority
	if err := json.Unmarshal(resp, &priorities); err != nil {
		return nil, err
	}

	return priorities, nil
}

// Status represents a Jira status
type Status struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	StatusCategory struct {
		ID   int    `json:"id"`
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"statusCategory"`
}

// GetStatuses retrieves all statuses for a project
func (c *Client) GetStatuses(ctx context.Context, projectKey string) ([]Status, error) {
	resp, err := c.doRequest(ctx, "GET", baseAPIPath+"/project/"+url.PathEscape(projectKey)+"/statuses", nil)
	if err != nil {
		return nil, err
	}

	var result []struct {
		Statuses []Status `json:"statuses"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	var statuses []Status
	for _, r := range result {
		statuses = append(statuses, r.Statuses...)
	}

	return statuses, nil
}

// Board represents a Jira Agile board
type Board struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Location struct {
		ProjectID  int    `json:"projectId"`
		ProjectKey string `json:"projectKey"`
		Name       string `json:"name"`
	} `json:"location"`
}

// GetBoards retrieves all boards for a project
func (c *Client) GetBoards(ctx context.Context, projectKey string) ([]Board, error) {
	params := url.Values{
		"projectKeyOrId": {projectKey},
	}
	resp, err := c.doRequestWithQuery(ctx, "GET", agileAPIPath+"/board", params, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Values []Board `json:"values"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result.Values, nil
}

// Sprint represents a Jira sprint
type Sprint struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	State         string    `json:"state"` // future, active, closed
	StartDate     string    `json:"startDate,omitempty"`
	EndDate       string    `json:"endDate,omitempty"`
	CompleteDate  string    `json:"completeDate,omitempty"`
	OriginBoardID int       `json:"originBoardId"`
	Goal          string    `json:"goal,omitempty"`
}

// GetSprints retrieves sprints for a board
func (c *Client) GetSprints(ctx context.Context, boardID int, state string) ([]Sprint, error) {
	params := url.Values{}
	if state != "" {
		params.Set("state", state)
	}

	path := fmt.Sprintf("%s/board/%d/sprint", agileAPIPath, boardID)
	resp, err := c.doRequestWithQuery(ctx, "GET", path, params, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Values []Sprint `json:"values"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result.Values, nil
}

// GetActiveSprint retrieves the active sprint for a board
func (c *Client) GetActiveSprint(ctx context.Context, boardID int) (*Sprint, error) {
	sprints, err := c.GetSprints(ctx, boardID, "active")
	if err != nil {
		return nil, err
	}

	if len(sprints) == 0 {
		return nil, ErrNotFound
	}

	return &sprints[0], nil
}

// GetBacklog retrieves backlog issues for a board
func (c *Client) GetBacklog(ctx context.Context, boardID int, startAt, maxResults int) (*SearchResult, error) {
	params := url.Values{
		"startAt":    {fmt.Sprintf("%d", startAt)},
		"maxResults": {fmt.Sprintf("%d", maxResults)},
	}

	path := fmt.Sprintf("%s/board/%d/backlog", agileAPIPath, boardID)
	resp, err := c.doRequestWithQuery(ctx, "GET", path, params, nil)
	if err != nil {
		return nil, err
	}

	var result SearchResult
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// MoveIssuesToSprint moves issues to a sprint
func (c *Client) MoveIssuesToSprint(ctx context.Context, sprintID int, issueKeys []string) error {
	body := map[string]interface{}{
		"issues": issueKeys,
	}

	path := fmt.Sprintf("%s/sprint/%d/issue", agileAPIPath, sprintID)
	_, err := c.doRequest(ctx, "POST", path, body)
	return err
}

// MoveIssuesToBacklog moves issues to the backlog
func (c *Client) MoveIssuesToBacklog(ctx context.Context, issueKeys []string) error {
	body := map[string]interface{}{
		"issues": issueKeys,
	}

	_, err := c.doRequest(ctx, "POST", agileAPIPath+"/backlog/issue", body)
	return err
}

// RankIssues changes the rank of issues
func (c *Client) RankIssues(ctx context.Context, issueKeys []string, rankBeforeIssue, rankAfterIssue string) error {
	body := map[string]interface{}{
		"issues": issueKeys,
	}
	if rankBeforeIssue != "" {
		body["rankBeforeIssue"] = rankBeforeIssue
	}
	if rankAfterIssue != "" {
		body["rankAfterIssue"] = rankAfterIssue
	}

	_, err := c.doRequest(ctx, "PUT", agileAPIPath+"/issue/rank", body)
	return err
}
