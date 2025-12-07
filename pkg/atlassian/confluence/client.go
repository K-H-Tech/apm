package confluence

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
)

const (
	baseAPIPath = "/wiki/api/v2"
	legacyAPIPath = "/wiki/rest/api"
)

// Client is a Confluence REST API client
type Client struct {
	cloudID     string
	accessToken string
	httpClient  *http.Client
	baseURL     string
}

// ClientConfig holds Confluence client configuration
type ClientConfig struct {
	CloudID     string
	AccessToken string
	Timeout     time.Duration
}

// NewClient creates a new Confluence client
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
		baseURL: fmt.Sprintf("https://api.atlassian.com/ex/confluence/%s", config.CloudID),
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

// Space represents a Confluence space
type Space struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Type        string `json:"type"` // global, personal
	Status      string `json:"status"`
	Description struct {
		Plain struct {
			Value string `json:"value"`
		} `json:"plain"`
	} `json:"description,omitempty"`
	HomepageID string `json:"homepageId,omitempty"`
}

// GetSpaces retrieves all spaces
func (c *Client) GetSpaces(ctx context.Context, limit int, cursor string) ([]Space, string, error) {
	params := url.Values{
		"limit": {fmt.Sprintf("%d", limit)},
	}
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	resp, err := c.doRequestWithQuery(ctx, "GET", baseAPIPath+"/spaces", params, nil)
	if err != nil {
		return nil, "", err
	}

	var result struct {
		Results []Space `json:"results"`
		Links   struct {
			Next string `json:"next"`
		} `json:"_links"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, "", err
	}

	return result.Results, result.Links.Next, nil
}

// GetSpace retrieves a space by ID
func (c *Client) GetSpace(ctx context.Context, spaceID string) (*Space, error) {
	resp, err := c.doRequest(ctx, "GET", baseAPIPath+"/spaces/"+spaceID, nil)
	if err != nil {
		return nil, err
	}

	var space Space
	if err := json.Unmarshal(resp, &space); err != nil {
		return nil, err
	}

	return &space, nil
}

// GetSpaceByKey retrieves a space by key
func (c *Client) GetSpaceByKey(ctx context.Context, spaceKey string) (*Space, error) {
	params := url.Values{
		"keys": {spaceKey},
	}

	resp, err := c.doRequestWithQuery(ctx, "GET", baseAPIPath+"/spaces", params, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Results []Space `json:"results"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if len(result.Results) == 0 {
		return nil, ErrNotFound
	}

	return &result.Results[0], nil
}

// Page represents a Confluence page
type Page struct {
	ID          string `json:"id"`
	Status      string `json:"status"` // current, draft, historical
	Title       string `json:"title"`
	SpaceID     string `json:"spaceId"`
	ParentID    string `json:"parentId,omitempty"`
	ParentType  string `json:"parentType,omitempty"`
	Position    int    `json:"position,omitempty"`
	AuthorID    string `json:"authorId,omitempty"`
	OwnerId     string `json:"ownerId,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
	Version     *PageVersion `json:"version,omitempty"`
	Body        *PageBody    `json:"body,omitempty"`
	Links       *PageLinks   `json:"_links,omitempty"`
}

// PageVersion represents page version information
type PageVersion struct {
	Number    int    `json:"number"`
	Message   string `json:"message,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	AuthorID  string `json:"authorId,omitempty"`
}

// PageBody represents page content
type PageBody struct {
	Storage        *BodyContent `json:"storage,omitempty"`
	AtlasDocFormat *BodyContent `json:"atlas_doc_format,omitempty"`
}

// BodyContent represents body content in a specific format
type BodyContent struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

// PageLinks contains page links
type PageLinks struct {
	WebUI  string `json:"webui,omitempty"`
	EditUI string `json:"editui,omitempty"`
	TinyUI string `json:"tinyui,omitempty"`
}

// CreatePageRequest represents a request to create a page
type CreatePageRequest struct {
	SpaceID  string    `json:"spaceId"`
	Status   string    `json:"status"` // current, draft
	Title    string    `json:"title"`
	ParentID string    `json:"parentId,omitempty"`
	Body     *PageBody `json:"body"`
}

// CreatePage creates a new page
func (c *Client) CreatePage(ctx context.Context, req *CreatePageRequest) (*Page, error) {
	resp, err := c.doRequest(ctx, "POST", baseAPIPath+"/pages", req)
	if err != nil {
		return nil, err
	}

	var page Page
	if err := json.Unmarshal(resp, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// GetPage retrieves a page by ID
func (c *Client) GetPage(ctx context.Context, pageID string, bodyFormat string) (*Page, error) {
	params := url.Values{}
	if bodyFormat != "" {
		params.Set("body-format", bodyFormat) // storage, atlas_doc_format, view, export_view
	}

	resp, err := c.doRequestWithQuery(ctx, "GET", baseAPIPath+"/pages/"+pageID, params, nil)
	if err != nil {
		return nil, err
	}

	var page Page
	if err := json.Unmarshal(resp, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// UpdatePageRequest represents a request to update a page
type UpdatePageRequest struct {
	ID      string       `json:"id"`
	Status  string       `json:"status"`
	Title   string       `json:"title"`
	Body    *PageBody    `json:"body"`
	Version *PageVersion `json:"version"`
}

// UpdatePage updates an existing page
func (c *Client) UpdatePage(ctx context.Context, pageID string, req *UpdatePageRequest) (*Page, error) {
	resp, err := c.doRequest(ctx, "PUT", baseAPIPath+"/pages/"+pageID, req)
	if err != nil {
		return nil, err
	}

	var page Page
	if err := json.Unmarshal(resp, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// DeletePage deletes a page
func (c *Client) DeletePage(ctx context.Context, pageID string) error {
	_, err := c.doRequest(ctx, "DELETE", baseAPIPath+"/pages/"+pageID, nil)
	return err
}

// GetPagesBySpace retrieves pages in a space
func (c *Client) GetPagesBySpace(ctx context.Context, spaceID string, limit int, cursor string) ([]Page, string, error) {
	params := url.Values{
		"limit": {fmt.Sprintf("%d", limit)},
	}
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	path := fmt.Sprintf("%s/spaces/%s/pages", baseAPIPath, spaceID)
	resp, err := c.doRequestWithQuery(ctx, "GET", path, params, nil)
	if err != nil {
		return nil, "", err
	}

	var result struct {
		Results []Page `json:"results"`
		Links   struct {
			Next string `json:"next"`
		} `json:"_links"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, "", err
	}

	return result.Results, result.Links.Next, nil
}

// GetChildPages retrieves child pages of a page
func (c *Client) GetChildPages(ctx context.Context, pageID string, limit int, cursor string) ([]Page, string, error) {
	params := url.Values{
		"limit": {fmt.Sprintf("%d", limit)},
	}
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	path := fmt.Sprintf("%s/pages/%s/children", baseAPIPath, pageID)
	resp, err := c.doRequestWithQuery(ctx, "GET", path, params, nil)
	if err != nil {
		return nil, "", err
	}

	var result struct {
		Results []Page `json:"results"`
		Links   struct {
			Next string `json:"next"`
		} `json:"_links"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, "", err
	}

	return result.Results, result.Links.Next, nil
}

// SearchPages searches for pages
func (c *Client) SearchPages(ctx context.Context, query string, spaceKey string, limit int) ([]Page, error) {
	cql := fmt.Sprintf("type=page AND text~\"%s\"", query)
	if spaceKey != "" {
		cql = fmt.Sprintf("%s AND space=\"%s\"", cql, spaceKey)
	}

	params := url.Values{
		"cql":   {cql},
		"limit": {fmt.Sprintf("%d", limit)},
	}

	resp, err := c.doRequestWithQuery(ctx, "GET", legacyAPIPath+"/content/search", params, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Results []struct {
			ID    string `json:"id"`
			Type  string `json:"type"`
			Title string `json:"title"`
			Space struct {
				Key string `json:"key"`
			} `json:"space"`
			Links struct {
				WebUI string `json:"webui"`
			} `json:"_links"`
		} `json:"results"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	pages := make([]Page, len(result.Results))
	for i, r := range result.Results {
		pages[i] = Page{
			ID:    r.ID,
			Title: r.Title,
			Links: &PageLinks{WebUI: r.Links.WebUI},
		}
	}

	return pages, nil
}

// Label represents a Confluence label
type Label struct {
	ID     string `json:"id,omitempty"`
	Prefix string `json:"prefix"`
	Name   string `json:"name"`
}

// AddLabels adds labels to a page
func (c *Client) AddLabels(ctx context.Context, pageID string, labels []string) error {
	labelObjects := make([]Label, len(labels))
	for i, l := range labels {
		labelObjects[i] = Label{
			Prefix: "global",
			Name:   l,
		}
	}

	path := fmt.Sprintf("%s/pages/%s/labels", baseAPIPath, pageID)
	_, err := c.doRequest(ctx, "POST", path, labelObjects)
	return err
}

// GetLabels retrieves labels for a page
func (c *Client) GetLabels(ctx context.Context, pageID string) ([]Label, error) {
	path := fmt.Sprintf("%s/pages/%s/labels", baseAPIPath, pageID)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Results []Label `json:"results"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

// RemoveLabel removes a label from a page
func (c *Client) RemoveLabel(ctx context.Context, pageID, labelID string) error {
	path := fmt.Sprintf("%s/pages/%s/labels/%s", baseAPIPath, pageID, labelID)
	_, err := c.doRequest(ctx, "DELETE", path, nil)
	return err
}

// StorageFormatFromMarkdown converts markdown-like content to Confluence storage format
// This is a simple converter - for production use, consider a proper converter
func StorageFormatFromMarkdown(markdown string) string {
	// Simple conversion - in production, use a proper markdown to storage converter
	// This just wraps the content in a basic structure
	return fmt.Sprintf("<p>%s</p>", markdown)
}

// StorageFormatFromHTML wraps HTML content for Confluence storage format
func StorageFormatFromHTML(html string) *PageBody {
	return &PageBody{
		Storage: &BodyContent{
			Value:          html,
			Representation: "storage",
		},
	}
}

// CreatePRDPage creates a PRD page in Confluence with proper formatting
func (c *Client) CreatePRDPage(ctx context.Context, spaceID, title string, content PRDContent) (*Page, error) {
	storageContent := formatPRDToStorage(content)

	req := &CreatePageRequest{
		SpaceID: spaceID,
		Status:  "current",
		Title:   title,
		Body:    StorageFormatFromHTML(storageContent),
	}

	return c.CreatePage(ctx, req)
}

// PRDContent holds structured PRD content for Confluence formatting
type PRDContent struct {
	Overview          string
	ProblemStatement  string
	Goals             []string
	NonGoals          []string
	UserPersonas      []PersonaContent
	UserStories       []UserStoryContent
	SuccessMetrics    []MetricContent
	TechnicalApproach string
	Timeline          string
	OpenQuestions     []string
}

// PersonaContent holds persona information
type PersonaContent struct {
	Name       string
	Description string
	Goals      []string
	PainPoints []string
}

// UserStoryContent holds user story information
type UserStoryContent struct {
	AsA         string
	IWant       string
	SoThat      string
	Criteria    []string
}

// MetricContent holds metric information
type MetricContent struct {
	Name    string
	Target  string
	Current string
}

// formatPRDToStorage converts PRD content to Confluence storage format
func formatPRDToStorage(content PRDContent) string {
	var html string

	// Overview
	if content.Overview != "" {
		html += fmt.Sprintf(`<h2>Overview</h2><p>%s</p>`, content.Overview)
	}

	// Problem Statement
	if content.ProblemStatement != "" {
		html += fmt.Sprintf(`<h2>Problem Statement</h2><p>%s</p>`, content.ProblemStatement)
	}

	// Goals
	if len(content.Goals) > 0 {
		html += `<h2>Goals</h2><ul>`
		for _, g := range content.Goals {
			html += fmt.Sprintf(`<li>%s</li>`, g)
		}
		html += `</ul>`
	}

	// Non-Goals
	if len(content.NonGoals) > 0 {
		html += `<h2>Non-Goals</h2><ul>`
		for _, ng := range content.NonGoals {
			html += fmt.Sprintf(`<li>%s</li>`, ng)
		}
		html += `</ul>`
	}

	// User Personas
	if len(content.UserPersonas) > 0 {
		html += `<h2>User Personas</h2>`
		for _, p := range content.UserPersonas {
			html += fmt.Sprintf(`<h3>%s</h3><p>%s</p>`, p.Name, p.Description)
			if len(p.Goals) > 0 {
				html += `<p><strong>Goals:</strong></p><ul>`
				for _, g := range p.Goals {
					html += fmt.Sprintf(`<li>%s</li>`, g)
				}
				html += `</ul>`
			}
		}
	}

	// User Stories
	if len(content.UserStories) > 0 {
		html += `<h2>User Stories</h2>`
		for i, s := range content.UserStories {
			html += fmt.Sprintf(`<h3>Story %d</h3>`, i+1)
			html += fmt.Sprintf(`<p><strong>As a</strong> %s, <strong>I want</strong> %s, <strong>so that</strong> %s</p>`, s.AsA, s.IWant, s.SoThat)
			if len(s.Criteria) > 0 {
				html += `<p><strong>Acceptance Criteria:</strong></p><ul>`
				for _, c := range s.Criteria {
					html += fmt.Sprintf(`<li>%s</li>`, c)
				}
				html += `</ul>`
			}
		}
	}

	// Success Metrics
	if len(content.SuccessMetrics) > 0 {
		html += `<h2>Success Metrics</h2>`
		html += `<table><thead><tr><th>Metric</th><th>Target</th><th>Current</th></tr></thead><tbody>`
		for _, m := range content.SuccessMetrics {
			html += fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%s</td></tr>`, m.Name, m.Target, m.Current)
		}
		html += `</tbody></table>`
	}

	// Technical Approach
	if content.TechnicalApproach != "" {
		html += fmt.Sprintf(`<h2>Technical Approach</h2><p>%s</p>`, content.TechnicalApproach)
	}

	// Timeline
	if content.Timeline != "" {
		html += fmt.Sprintf(`<h2>Timeline</h2><p>%s</p>`, content.Timeline)
	}

	// Open Questions
	if len(content.OpenQuestions) > 0 {
		html += `<h2>Open Questions</h2><ul>`
		for _, q := range content.OpenQuestions {
			html += fmt.Sprintf(`<li>%s</li>`, q)
		}
		html += `</ul>`
	}

	return html
}
