package atlassian

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// ErrInvalidState is returned when OAuth state validation fails
	ErrInvalidState = errors.New("invalid OAuth state")

	// ErrTokenExchangeFailed is returned when token exchange fails
	ErrTokenExchangeFailed = errors.New("failed to exchange authorization code for tokens")

	// ErrTokenRefreshFailed is returned when token refresh fails
	ErrTokenRefreshFailed = errors.New("failed to refresh access token")

	// ErrInvalidToken is returned when a token is invalid
	ErrInvalidToken = errors.New("invalid token")
)

const (
	// Atlassian OAuth URLs
	authorizationURL = "https://auth.atlassian.com/authorize"
	tokenURL         = "https://auth.atlassian.com/oauth/token"
	userInfoURL      = "https://api.atlassian.com/me"
	accessibleResourcesURL = "https://api.atlassian.com/oauth/token/accessible-resources"
)

// OAuthConfig holds Atlassian OAuth configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// OAuthClient handles Atlassian OAuth 2.0 flow
type OAuthClient struct {
	config     OAuthConfig
	httpClient *http.Client
}

// NewOAuthClient creates a new Atlassian OAuth client
func NewOAuthClient(config OAuthConfig) *OAuthClient {
	return &OAuthClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAuthorizationURL generates the authorization URL for OAuth flow
func (c *OAuthClient) GetAuthorizationURL(state string) string {
	params := url.Values{
		"audience":      {"api.atlassian.com"},
		"client_id":     {c.config.ClientID},
		"scope":         {strings.Join(c.config.Scopes, " ")},
		"redirect_uri":  {c.config.RedirectURL},
		"state":         {state},
		"response_type": {"code"},
		"prompt":        {"consent"},
	}

	return fmt.Sprintf("%s?%s", authorizationURL, params.Encode())
}

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// ExchangeCode exchanges an authorization code for tokens
func (c *OAuthClient) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {c.config.ClientID},
		"client_secret": {c.config.ClientSecret},
		"code":          {code},
		"redirect_uri":  {c.config.RedirectURL},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s", ErrTokenExchangeFailed, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// RefreshToken refreshes an expired access token
func (c *OAuthClient) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {c.config.ClientID},
		"client_secret": {c.config.ClientSecret},
		"refresh_token": {refreshToken},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s", ErrTokenRefreshFailed, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// UserInfo represents Atlassian user information
type UserInfo struct {
	AccountID   string `json:"account_id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	Picture     string `json:"picture"`
	AccountType string `json:"account_type"`
}

// GetUserInfo retrieves the authenticated user's information
func (c *OAuthClient) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// AccessibleResource represents an Atlassian site the user has access to
type AccessibleResource struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	AvatarURL string   `json:"avatarUrl"`
}

// GetAccessibleResources retrieves the Atlassian sites the user can access
func (c *OAuthClient) GetAccessibleResources(ctx context.Context, accessToken string) ([]AccessibleResource, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", accessibleResourcesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get accessible resources: %s", string(body))
	}

	var resources []AccessibleResource
	if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return nil, err
	}

	return resources, nil
}

// TokenExpiry calculates the token expiry time
func TokenExpiry(expiresIn int) time.Time {
	return time.Now().Add(time.Duration(expiresIn) * time.Second)
}

// DefaultScopes returns the default OAuth scopes for APM
func DefaultScopes() []string {
	return []string{
		// Jira scopes
		"read:jira-work",
		"write:jira-work",
		"read:jira-user",
		// Confluence scopes
		"read:confluence-content.all",
		"write:confluence-content",
		"read:confluence-space.summary",
		// User profile
		"read:me",
		// Offline access for refresh tokens
		"offline_access",
	}
}

// StateGenerator generates OAuth state tokens
type StateGenerator interface {
	Generate() (string, error)
	Validate(state string) bool
}

// SimpleStateStore is a simple in-memory state store (use Redis in production)
type SimpleStateStore struct {
	states map[string]time.Time
	ttl    time.Duration
}

// NewSimpleStateStore creates a new simple state store
func NewSimpleStateStore(ttl time.Duration) *SimpleStateStore {
	return &SimpleStateStore{
		states: make(map[string]time.Time),
		ttl:    ttl,
	}
}

// Store stores a state token
func (s *SimpleStateStore) Store(state string) {
	s.states[state] = time.Now().Add(s.ttl)
}

// Validate validates and consumes a state token
func (s *SimpleStateStore) Validate(state string) bool {
	expiry, ok := s.states[state]
	if !ok {
		return false
	}
	delete(s.states, state)
	return time.Now().Before(expiry)
}

// Cleanup removes expired states
func (s *SimpleStateStore) Cleanup() {
	now := time.Now()
	for state, expiry := range s.states {
		if now.After(expiry) {
			delete(s.states, state)
		}
	}
}
