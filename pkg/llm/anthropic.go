package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/K-H-Tech/apm/internal/contract"
)

const (
	defaultAnthropicBaseURL = "https://api.anthropic.com/v1"
	anthropicAPIVersion     = "2023-06-01"
)

// AnthropicProvider implements LLMProvider for Anthropic Claude
type AnthropicProvider struct {
	apiKey     string
	baseURL    string
	model      string
	maxTokens  int
	temperature float64
	httpClient *http.Client
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(config *Config) (*AnthropicProvider, error) {
	if config.AnthropicAPIKey == "" {
		return nil, ErrProviderNotConfigured
	}

	baseURL := config.AnthropicBaseURL
	if baseURL == "" {
		baseURL = defaultAnthropicBaseURL
	}

	model := config.AnthropicModel
	if model == "" {
		model = "claude-3-sonnet-20240229"
	}

	return &AnthropicProvider{
		apiKey:      config.AnthropicAPIKey,
		baseURL:     baseURL,
		model:       model,
		maxTokens:   config.MaxTokens,
		temperature: config.Temperature,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// anthropicMessageRequest represents the request body for messages API
type anthropicMessageRequest struct {
	Model       string              `json:"model"`
	Messages    []anthropicMessage  `json:"messages"`
	System      string              `json:"system,omitempty"`
	MaxTokens   int                 `json:"max_tokens"`
	Temperature float64             `json:"temperature,omitempty"`
	StopSequences []string          `json:"stop_sequences,omitempty"`
	Stream      bool                `json:"stream,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicMessageResponse represents the response from messages API
type anthropicMessageResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Role         string `json:"role"`
	Content      []anthropicContent `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence,omitempty"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *anthropicError `json:"error,omitempty"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// GenerateText generates text using Anthropic Claude
func (a *AnthropicProvider) GenerateText(ctx context.Context, req contract.GenerateRequest) (*contract.GenerateResponse, error) {
	startTime := time.Now()

	// Build request
	model := req.Model
	if model == "" {
		model = a.model
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = a.maxTokens
	}

	temperature := req.Temperature
	if temperature == 0 {
		temperature = a.temperature
	}

	msgReq := anthropicMessageRequest{
		Model: model,
		Messages: []anthropicMessage{
			{Role: "user", Content: req.Prompt},
		},
		System:        req.SystemPrompt,
		MaxTokens:     maxTokens,
		Temperature:   temperature,
		StopSequences: req.StopSequences,
	}

	// Make request
	resp, err := a.doRequest(ctx, "/messages", msgReq)
	if err != nil {
		return nil, err
	}

	// Extract text from response
	var text string
	for _, content := range resp.Content {
		if content.Type == "text" {
			text += content.Text
		}
	}

	return &contract.GenerateResponse{
		Text:         text,
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
		TotalTokens:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
		Model:        resp.Model,
		FinishReason: resp.StopReason,
		LatencyMs:    time.Since(startTime).Milliseconds(),
	}, nil
}

// GenerateStructured generates structured JSON output
func (a *AnthropicProvider) GenerateStructured(ctx context.Context, req contract.GenerateRequest, schema interface{}) (interface{}, error) {
	// Build messages with JSON instruction
	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant that responds in valid JSON format."
	}
	systemPrompt += fmt.Sprintf("\n\nRespond ONLY with valid JSON matching this schema (no other text):\n%s", string(schemaBytes))

	modifiedReq := contract.GenerateRequest{
		Prompt:       req.Prompt + "\n\nRespond with JSON only, no other text.",
		SystemPrompt: systemPrompt,
		MaxTokens:    req.MaxTokens,
		Temperature:  req.Temperature,
		Model:        req.Model,
	}

	// Generate response
	resp, err := a.GenerateText(ctx, modifiedReq)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var result interface{}
	if err := json.Unmarshal([]byte(resp.Text), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return result, nil
}

// StreamText streams text generation
func (a *AnthropicProvider) StreamText(ctx context.Context, req contract.GenerateRequest) (<-chan contract.StreamChunk, error) {
	ch := make(chan contract.StreamChunk)

	go func() {
		defer close(ch)

		msgReq := anthropicMessageRequest{
			Model: a.model,
			Messages: []anthropicMessage{
				{Role: "user", Content: req.Prompt},
			},
			System:    req.SystemPrompt,
			MaxTokens: req.MaxTokens,
			Stream:    true,
		}

		body, err := json.Marshal(msgReq)
		if err != nil {
			ch <- contract.StreamChunk{Error: err, Done: true}
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/messages", bytes.NewReader(body))
		if err != nil {
			ch <- contract.StreamChunk{Error: err, Done: true}
			return
		}

		httpReq.Header.Set("x-api-key", a.apiKey)
		httpReq.Header.Set("anthropic-version", anthropicAPIVersion)
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Accept", "text/event-stream")

		resp, err := a.httpClient.Do(httpReq)
		if err != nil {
			ch <- contract.StreamChunk{Error: err, Done: true}
			return
		}
		defer resp.Body.Close()

		// Read SSE stream
		buf := make([]byte, 4096)
		for {
			select {
			case <-ctx.Done():
				ch <- contract.StreamChunk{Error: ctx.Err(), Done: true}
				return
			default:
				n, err := resp.Body.Read(buf)
				if err != nil {
					if err == io.EOF {
						ch <- contract.StreamChunk{Done: true}
						return
					}
					ch <- contract.StreamChunk{Error: err, Done: true}
					return
				}

				// Parse SSE data
				data := string(buf[:n])
				// Simple parsing - look for content_block_delta events
				// In production, use proper SSE parsing
				if len(data) > 0 {
					// Extract text deltas from the stream
					// This is simplified - real implementation needs proper SSE parsing
					ch <- contract.StreamChunk{Text: data}
				}
			}
		}
	}()

	return ch, nil
}

// GetProviderName returns the provider name
func (a *AnthropicProvider) GetProviderName() string {
	return "anthropic"
}

// GetModelName returns the model name
func (a *AnthropicProvider) GetModelName() string {
	return a.model
}

// doRequest makes an HTTP request to Anthropic API
func (a *AnthropicProvider) doRequest(ctx context.Context, path string, body interface{}) (*anthropicMessageResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", anthropicAPIVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrRateLimited
	}

	if resp.StatusCode != http.StatusOK {
		var errResp anthropicMessageResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != nil {
			return nil, fmt.Errorf("%s: %s", errResp.Error.Type, errResp.Error.Message)
		}
		return nil, fmt.Errorf("API error: %s", string(respBody))
	}

	var result anthropicMessageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}
