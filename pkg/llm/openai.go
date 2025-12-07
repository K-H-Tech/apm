package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/K-H-Tech/apm/internal/contract"
)

const (
	defaultOpenAIBaseURL = "https://api.openai.com/v1"
)

// OpenAIProvider implements LLMProvider for OpenAI
type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	model      string
	maxTokens  int
	temperature float64
	httpClient *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config *Config) (*OpenAIProvider, error) {
	if config.OpenAIAPIKey == "" {
		return nil, ErrProviderNotConfigured
	}

	baseURL := config.OpenAIBaseURL
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}

	model := config.OpenAIModel
	if model == "" {
		model = "gpt-4-turbo-preview"
	}

	return &OpenAIProvider{
		apiKey:      config.OpenAIAPIKey,
		baseURL:     baseURL,
		model:       model,
		maxTokens:   config.MaxTokens,
		temperature: config.Temperature,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// openAIChatRequest represents the request body for chat completions
type openAIChatRequest struct {
	Model       string             `json:"model"`
	Messages    []openAIMessage    `json:"messages"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature float64            `json:"temperature,omitempty"`
	Stop        []string           `json:"stop,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
	ResponseFormat *openAIResponseFormat `json:"response_format,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponseFormat struct {
	Type string `json:"type"`
}

// openAIChatResponse represents the response from chat completions
type openAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int           `json:"index"`
		Message      openAIMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *openAIError `json:"error,omitempty"`
}

type openAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// GenerateText generates text using OpenAI
func (o *OpenAIProvider) GenerateText(ctx context.Context, req contract.GenerateRequest) (*contract.GenerateResponse, error) {
	startTime := time.Now()

	// Build messages
	messages := []openAIMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, openAIMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}
	messages = append(messages, openAIMessage{
		Role:    "user",
		Content: req.Prompt,
	})

	// Build request
	model := req.Model
	if model == "" {
		model = o.model
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = o.maxTokens
	}

	temperature := req.Temperature
	if temperature == 0 {
		temperature = o.temperature
	}

	chatReq := openAIChatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Stop:        req.StopSequences,
	}

	// Make request
	resp, err := o.doRequest(ctx, "/chat/completions", chatReq)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, ErrInvalidResponse
	}

	return &contract.GenerateResponse{
		Text:         resp.Choices[0].Message.Content,
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
		TotalTokens:  resp.Usage.TotalTokens,
		Model:        resp.Model,
		FinishReason: resp.Choices[0].FinishReason,
		LatencyMs:    time.Since(startTime).Milliseconds(),
	}, nil
}

// GenerateStructured generates structured JSON output
func (o *OpenAIProvider) GenerateStructured(ctx context.Context, req contract.GenerateRequest, schema interface{}) (interface{}, error) {
	startTime := time.Now()

	// Build messages with JSON instruction
	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant that responds in valid JSON format."
	}
	systemPrompt += fmt.Sprintf("\n\nRespond with valid JSON matching this schema:\n%s", string(schemaBytes))

	messages := []openAIMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: req.Prompt},
	}

	// Build request with JSON mode
	model := req.Model
	if model == "" {
		model = o.model
	}

	chatReq := openAIChatRequest{
		Model:          model,
		Messages:       messages,
		MaxTokens:      req.MaxTokens,
		Temperature:    req.Temperature,
		ResponseFormat: &openAIResponseFormat{Type: "json_object"},
	}

	// Make request
	resp, err := o.doRequest(ctx, "/chat/completions", chatReq)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, ErrInvalidResponse
	}

	// Parse JSON response
	var result interface{}
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	_ = startTime // TODO: Add latency tracking to response
	return result, nil
}

// StreamText streams text generation
func (o *OpenAIProvider) StreamText(ctx context.Context, req contract.GenerateRequest) (<-chan contract.StreamChunk, error) {
	ch := make(chan contract.StreamChunk)

	go func() {
		defer close(ch)

		// Build messages
		messages := []openAIMessage{}
		if req.SystemPrompt != "" {
			messages = append(messages, openAIMessage{
				Role:    "system",
				Content: req.SystemPrompt,
			})
		}
		messages = append(messages, openAIMessage{
			Role:    "user",
			Content: req.Prompt,
		})

		chatReq := openAIChatRequest{
			Model:       o.model,
			Messages:    messages,
			MaxTokens:   req.MaxTokens,
			Temperature: req.Temperature,
			Stream:      true,
		}

		body, err := json.Marshal(chatReq)
		if err != nil {
			ch <- contract.StreamChunk{Error: err, Done: true}
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			ch <- contract.StreamChunk{Error: err, Done: true}
			return
		}

		httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Accept", "text/event-stream")

		resp, err := o.httpClient.Do(httpReq)
		if err != nil {
			ch <- contract.StreamChunk{Error: err, Done: true}
			return
		}
		defer resp.Body.Close()

		// Check HTTP status before parsing stream
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			ch <- contract.StreamChunk{
				Error: fmt.Errorf("API error %d: %s", resp.StatusCode, string(body)),
				Done:  true,
			}
			return
		}

		// Read SSE stream with proper line-based parsing
		reader := bufio.NewReader(resp.Body)
		for {
			select {
			case <-ctx.Done():
				ch <- contract.StreamChunk{Error: ctx.Err(), Done: true}
				return
			default:
				line, err := reader.ReadString('\n')
				if err != nil {
					if errors.Is(err, io.EOF) {
						ch <- contract.StreamChunk{Done: true}
						return
					}
					ch <- contract.StreamChunk{Error: err, Done: true}
					return
				}

				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				if !strings.HasPrefix(line, "data: ") {
					continue
				}
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					ch <- contract.StreamChunk{Done: true}
					return
				}

				var event map[string]interface{}
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					ch <- contract.StreamChunk{Error: err, Done: true}
					return
				}

				// Parse delta content with safe type assertions
				if choices, ok := event["choices"].([]interface{}); ok && len(choices) > 0 {
					choice, ok := choices[0].(map[string]interface{})
					if !ok {
						continue
					}
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok {
							ch <- contract.StreamChunk{Text: content}
						}
					}
					if finishReason, ok := choice["finish_reason"].(string); ok && finishReason != "" {
						ch <- contract.StreamChunk{Done: true, FinishReason: finishReason}
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

// GetProviderName returns the provider name
func (o *OpenAIProvider) GetProviderName() string {
	return "openai"
}

// GetModelName returns the model name
func (o *OpenAIProvider) GetModelName() string {
	return o.model
}

// doRequest makes an HTTP request to OpenAI API
func (o *OpenAIProvider) doRequest(ctx context.Context, path string, body interface{}) (*openAIChatResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
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
		var errResp openAIChatResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != nil {
			return nil, fmt.Errorf("%s: %s", errResp.Error.Type, errResp.Error.Message)
		}
		return nil, fmt.Errorf("API error: %s", string(respBody))
	}

	var result openAIChatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}
