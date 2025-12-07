package llm

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/K-H-Tech/apm/internal/contract"
)

var (
	// ErrProviderNotConfigured is returned when the LLM provider is not configured
	ErrProviderNotConfigured = errors.New("LLM provider not configured")

	// ErrInvalidResponse is returned when the LLM response is invalid
	ErrInvalidResponse = errors.New("invalid LLM response")

	// ErrRateLimited is returned when rate limited by the provider
	ErrRateLimited = errors.New("rate limited by provider")

	// ErrContextTooLong is returned when the context exceeds the model's limit
	ErrContextTooLong = errors.New("context exceeds model limit")

	// ErrGenerationFailed is returned when generation fails
	ErrGenerationFailed = errors.New("generation failed")
)

// ProviderType represents the type of LLM provider
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
)

// Config holds configuration for LLM providers
type Config struct {
	// OpenAI configuration
	OpenAIAPIKey    string `yaml:"OPENAI_API_KEY"`
	OpenAIBaseURL   string `yaml:"OPENAI_BASE_URL"`
	OpenAIModel     string `yaml:"OPENAI_MODEL"`

	// Anthropic configuration
	AnthropicAPIKey  string `yaml:"ANTHROPIC_API_KEY"`
	AnthropicBaseURL string `yaml:"ANTHROPIC_BASE_URL"`
	AnthropicModel   string `yaml:"ANTHROPIC_MODEL"`

	// Default provider
	DefaultProvider ProviderType `yaml:"DEFAULT_PROVIDER"`

	// Common settings
	MaxTokens       int           `yaml:"MAX_TOKENS"`
	Temperature     float64       `yaml:"TEMPERATURE"`
	Timeout         time.Duration `yaml:"TIMEOUT"`
	MaxRetries      int           `yaml:"MAX_RETRIES"`
	RetryDelay      time.Duration `yaml:"RETRY_DELAY"`
}

// DefaultConfig returns default LLM configuration
func DefaultConfig() *Config {
	return &Config{
		OpenAIModel:     "gpt-4-turbo-preview",
		AnthropicModel:  "claude-3-sonnet-20240229",
		DefaultProvider: ProviderOpenAI,
		MaxTokens:       4096,
		Temperature:     0.7,
		Timeout:         60 * time.Second,
		MaxRetries:      3,
		RetryDelay:      time.Second,
	}
}

// ProviderFactory creates LLM providers
type ProviderFactory struct {
	config    *Config
	providers map[ProviderType]contract.LLMProvider
	mu        sync.RWMutex
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(config *Config) *ProviderFactory {
	return &ProviderFactory{
		config:    config,
		providers: make(map[ProviderType]contract.LLMProvider),
	}
}

// GetProvider returns the specified provider
func (f *ProviderFactory) GetProvider(providerType ProviderType) (contract.LLMProvider, error) {
	// Check if already created (read lock)
	f.mu.RLock()
	provider, ok := f.providers[providerType]
	f.mu.RUnlock()
	if ok {
		return provider, nil
	}

	// Acquire write lock for creation
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock
	if provider, ok := f.providers[providerType]; ok {
		return provider, nil
	}

	// Create provider
	var err error

	switch providerType {
	case ProviderOpenAI:
		if f.config.OpenAIAPIKey == "" {
			return nil, ErrProviderNotConfigured
		}
		provider, err = NewOpenAIProvider(f.config)
	case ProviderAnthropic:
		if f.config.AnthropicAPIKey == "" {
			return nil, ErrProviderNotConfigured
		}
		provider, err = NewAnthropicProvider(f.config)
	default:
		return nil, errors.New("unknown provider type: " + string(providerType))
	}

	if err != nil {
		return nil, err
	}

	f.providers[providerType] = provider
	return provider, nil
}

// GetDefaultProvider returns the default configured provider
func (f *ProviderFactory) GetDefaultProvider() (contract.LLMProvider, error) {
	return f.GetProvider(f.config.DefaultProvider)
}

// MultiProvider wraps multiple providers with fallback support
type MultiProvider struct {
	primary  contract.LLMProvider
	fallback contract.LLMProvider
}

// NewMultiProvider creates a provider with fallback support
func NewMultiProvider(primary, fallback contract.LLMProvider) *MultiProvider {
	return &MultiProvider{
		primary:  primary,
		fallback: fallback,
	}
}

// GenerateText generates text with fallback support
func (m *MultiProvider) GenerateText(ctx context.Context, req contract.GenerateRequest) (*contract.GenerateResponse, error) {
	// Try primary provider
	resp, err := m.primary.GenerateText(ctx, req)
	if err == nil {
		return resp, nil
	}

	// If fallback is configured, try it
	if m.fallback != nil {
		return m.fallback.GenerateText(ctx, req)
	}

	return nil, err
}

// GenerateStructured generates structured output with fallback support
func (m *MultiProvider) GenerateStructured(ctx context.Context, req contract.GenerateRequest, schema interface{}) (interface{}, error) {
	// Try primary provider
	resp, err := m.primary.GenerateStructured(ctx, req, schema)
	if err == nil {
		return resp, nil
	}

	// If fallback is configured, try it
	if m.fallback != nil {
		return m.fallback.GenerateStructured(ctx, req, schema)
	}

	return nil, err
}

// StreamText streams text from the primary provider (no fallback for streaming)
func (m *MultiProvider) StreamText(ctx context.Context, req contract.GenerateRequest) (<-chan contract.StreamChunk, error) {
	return m.primary.StreamText(ctx, req)
}

// GetProviderName returns the primary provider name
func (m *MultiProvider) GetProviderName() string {
	return m.primary.GetProviderName()
}

// GetModelName returns the primary model name
func (m *MultiProvider) GetModelName() string {
	return m.primary.GetModelName()
}
