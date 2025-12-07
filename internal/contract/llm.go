package contract

import (
	"context"
)

// LLMProvider defines the interface for LLM integrations
type LLMProvider interface {
	// GenerateText generates text based on a prompt
	GenerateText(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)

	// GenerateStructured generates structured JSON output
	GenerateStructured(ctx context.Context, req GenerateRequest, schema interface{}) (interface{}, error)

	// StreamText streams text generation (returns channel)
	StreamText(ctx context.Context, req GenerateRequest) (<-chan StreamChunk, error)

	// GetProviderName returns the provider name (e.g., "openai", "anthropic")
	GetProviderName() string

	// GetModelName returns the model being used
	GetModelName() string
}

// GenerateRequest represents a request to generate text
type GenerateRequest struct {
	// Prompt is the user prompt
	Prompt string `json:"prompt"`

	// SystemPrompt is the system/context prompt
	SystemPrompt string `json:"system_prompt,omitempty"`

	// MaxTokens limits the response length
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature controls randomness (0-1)
	Temperature float64 `json:"temperature,omitempty"`

	// Model overrides the default model (optional)
	Model string `json:"model,omitempty"`

	// Stop sequences to end generation
	StopSequences []string `json:"stop_sequences,omitempty"`

	// Metadata for logging/tracking
	Metadata map[string]string `json:"metadata,omitempty"`
}

// GenerateResponse represents the response from text generation
type GenerateResponse struct {
	// Text is the generated content
	Text string `json:"text"`

	// InputTokens is the number of input tokens
	InputTokens int `json:"input_tokens"`

	// OutputTokens is the number of output tokens
	OutputTokens int `json:"output_tokens"`

	// TotalTokens is the total tokens used
	TotalTokens int `json:"total_tokens"`

	// Model is the model that was used
	Model string `json:"model"`

	// FinishReason indicates why generation stopped
	FinishReason string `json:"finish_reason"`

	// LatencyMs is the request latency in milliseconds
	LatencyMs int64 `json:"latency_ms"`
}

// StreamChunk represents a chunk in streaming response
type StreamChunk struct {
	// Text is the text content of this chunk
	Text string `json:"text"`

	// Done indicates if this is the final chunk
	Done bool `json:"done"`

	// Error contains any error that occurred
	Error error `json:"error,omitempty"`

	// FinishReason is set on the final chunk
	FinishReason string `json:"finish_reason,omitempty"`
}

// LLMConfig represents configuration for an LLM provider
type LLMConfig struct {
	// Provider is the LLM provider (openai, anthropic)
	Provider string `yaml:"PROVIDER"`

	// APIKey is the API key for the provider
	APIKey string `yaml:"API_KEY"`

	// BaseURL is the API base URL (optional, for custom endpoints)
	BaseURL string `yaml:"BASE_URL"`

	// DefaultModel is the default model to use
	DefaultModel string `yaml:"DEFAULT_MODEL"`

	// MaxTokens is the default max tokens
	MaxTokens int `yaml:"MAX_TOKENS"`

	// Temperature is the default temperature
	Temperature float64 `yaml:"TEMPERATURE"`

	// TimeoutSeconds is the request timeout
	TimeoutSeconds int `yaml:"TIMEOUT_SECONDS"`

	// MaxRetries is the max number of retries
	MaxRetries int `yaml:"MAX_RETRIES"`
}

// AIService defines AI orchestration operations
type AIService interface {
	// PRD Generation
	GeneratePRD(ctx context.Context, notes string, templateID string) (*GeneratedPRD, error)
	RefinePRD(ctx context.Context, prdContent string, feedback string, section string) (*GeneratedPRD, error)
	CheckPRDConsistency(ctx context.Context, prdContent string) (*ConsistencyReport, error)

	// User Stories
	GenerateUserStories(ctx context.Context, prdOverview string, personas []string) ([]GeneratedUserStory, error)

	// Acceptance Criteria
	GenerateAcceptanceCriteria(ctx context.Context, userStory string, context string) ([]string, error)

	// Summarization
	SummarizeMeetingNotes(ctx context.Context, notes string) (*MeetingSummary, error)

	// Discovery
	GenerateDiscoveryInsights(ctx context.Context, content string, itemType string) (*DiscoveryInsights, error)
}

// GeneratedPRD represents AI-generated PRD content
type GeneratedPRD struct {
	Title            string   `json:"title"`
	Overview         string   `json:"overview"`
	ProblemStatement string   `json:"problem_statement"`
	Goals            []string `json:"goals"`
	UserPersonas     []GeneratedPersona `json:"user_personas"`
	UserStories      []GeneratedUserStory `json:"user_stories"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	SuccessMetrics   []GeneratedMetric `json:"success_metrics"`
	OutOfScope       []string `json:"out_of_scope,omitempty"`
	Assumptions      []string `json:"assumptions,omitempty"`
	Dependencies     []string `json:"dependencies,omitempty"`
}

// GeneratedPersona represents an AI-generated user persona
type GeneratedPersona struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Goals       []string `json:"goals"`
	PainPoints  []string `json:"pain_points"`
}

// GeneratedUserStory represents an AI-generated user story
type GeneratedUserStory struct {
	AsA                string   `json:"as_a"`
	IWant              string   `json:"i_want"`
	SoThat             string   `json:"so_that"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	Priority           string   `json:"priority"`
	EstimatePoints     int      `json:"estimate_points,omitempty"`
}

// GeneratedMetric represents an AI-generated success metric
type GeneratedMetric struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
	Target     string `json:"target"`
}

// MeetingSummary represents summarized meeting notes
type MeetingSummary struct {
	Summary     string   `json:"summary"`
	KeyPoints   []string `json:"key_points"`
	ActionItems []ActionItem `json:"action_items"`
	Decisions   []string `json:"decisions"`
	OpenQuestions []string `json:"open_questions"`
}

// ActionItem represents an action item from meeting notes
type ActionItem struct {
	Action   string `json:"action"`
	Owner    string `json:"owner,omitempty"`
	DueDate  string `json:"due_date,omitempty"`
	Priority string `json:"priority,omitempty"`
}

// DiscoveryInsights represents AI-generated insights for discovery
type DiscoveryInsights struct {
	Summary          string   `json:"summary"`
	KeyThemes        []string `json:"key_themes"`
	Patterns         []string `json:"patterns"`
	Opportunities    []string `json:"opportunities"`
	SuggestedActions []string `json:"suggested_actions"`
}
