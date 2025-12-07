package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"text/template"

	"github.com/K-H-Tech/apm/internal/contract"
	"github.com/K-H-Tech/apm/internal/models"
	"github.com/K-H-Tech/apm/pkg/llm/prompts"
)

// AIService implements contract.AIService
type AIService struct {
	provider contract.LLMProvider
}

// NewAIService creates a new AI service
func NewAIService(provider contract.LLMProvider) *AIService {
	return &AIService{
		provider: provider,
	}
}

// GeneratePRDInput holds input for PRD generation
type GeneratePRDInput struct {
	Notes       string
	MeetingType string
	Template    *models.PRDTemplate
}

// GeneratePRD generates a PRD from meeting notes
func (s *AIService) GeneratePRD(ctx context.Context, input GeneratePRDInput) (*GeneratedPRDResult, error) {
	if s.provider == nil {
		return nil, errors.New("LLM provider not configured")
	}

	// Parse and execute the prompt template
	tmpl, err := template.New("prd").Parse(prompts.PRDGeneratorUserPrompt)
	if err != nil {
		return nil, err
	}

	var promptBuf bytes.Buffer
	err = tmpl.Execute(&promptBuf, map[string]interface{}{
		"Notes":       input.Notes,
		"MeetingType": input.MeetingType,
		"Template":    input.Template,
	})
	if err != nil {
		return nil, err
	}

	// Generate using LLM
	request := contract.GenerateRequest{
		SystemPrompt: prompts.PRDGeneratorSystemPrompt,
		Prompt:       promptBuf.String(),
		MaxTokens:    4096,
		Temperature:  0.7,
	}

	response, err := s.provider.GenerateText(ctx, request)
	if err != nil {
		return nil, err
	}

	// Parse the JSON response - expecting a structure with title and content
	var parsed struct {
		Title            string          `json:"title"`
		Overview         string          `json:"overview"`
		ProblemStatement string          `json:"problem_statement"`
		Goals            []string        `json:"goals"`
		UserPersonas     []models.UserPersona   `json:"user_personas"`
		UserStories      []models.UserStory     `json:"user_stories"`
		AcceptanceCriteria []string      `json:"acceptance_criteria"`
		OutOfScope       []string        `json:"out_of_scope"`
		Assumptions      []string        `json:"assumptions"`
		Dependencies     []string        `json:"dependencies"`
		SuccessMetrics   []models.SuccessMetric `json:"success_metrics"`
	}
	if err := json.Unmarshal([]byte(response.Text), &parsed); err != nil {
		// If JSON parsing fails, use raw content as overview
		return &GeneratedPRDResult{
			Title: "Generated PRD",
			Content: &models.PRDContent{
				Overview: response.Text,
			},
		}, nil
	}

	content := &models.PRDContent{
		Overview:         parsed.Overview,
		ProblemStatement: parsed.ProblemStatement,
		Goals:            parsed.Goals,
		UserPersonas:     parsed.UserPersonas,
		UserStories:      parsed.UserStories,
		AcceptanceCriteria: parsed.AcceptanceCriteria,
		OutOfScope:       parsed.OutOfScope,
		Assumptions:      parsed.Assumptions,
		Dependencies:     parsed.Dependencies,
		SuccessMetrics:   parsed.SuccessMetrics,
	}

	title := parsed.Title
	if title == "" {
		title = "Generated PRD"
	}

	return &GeneratedPRDResult{
		Title:   title,
		Content: content,
	}, nil
}

// RefinePRD refines an existing PRD based on feedback
func (s *AIService) RefinePRD(ctx context.Context, content *models.PRDContent, feedback string) (*models.PRDContent, error) {
	if s.provider == nil {
		return nil, errors.New("LLM provider not configured")
	}

	// Marshal current content to JSON
	currentJSON, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	prompt := `You are refining an existing PRD based on feedback.

Current PRD:
` + string(currentJSON) + `

Feedback to address:
` + feedback + `

Please refine the PRD to address the feedback while maintaining the same JSON structure.
Return the complete refined PRD as valid JSON.`

	request := contract.GenerateRequest{
		SystemPrompt: prompts.PRDGeneratorSystemPrompt,
		Prompt:       prompt,
		MaxTokens:    4096,
		Temperature:  0.7,
	}

	response, err := s.provider.GenerateText(ctx, request)
	if err != nil {
		return nil, err
	}

	var refined models.PRDContent
	if err := json.Unmarshal([]byte(response.Text), &refined); err != nil {
		return nil, errors.New("failed to parse refined PRD: " + err.Error())
	}

	return &refined, nil
}

// GenerateUserStories generates user stories from PRD content
func (s *AIService) GenerateUserStories(ctx context.Context, content *models.PRDContent) ([]models.UserStory, error) {
	if s.provider == nil {
		return nil, errors.New("LLM provider not configured")
	}

	// Build the prompt
	tmpl, err := template.New("stories").Parse(prompts.UserStoryGeneratorPrompt)
	if err != nil {
		return nil, err
	}

	var promptBuf bytes.Buffer
	err = tmpl.Execute(&promptBuf, map[string]interface{}{
		"Overview":         content.Overview,
		"ProblemStatement": content.ProblemStatement,
		"Personas":         content.UserPersonas,
	})
	if err != nil {
		return nil, err
	}

	request := contract.GenerateRequest{
		Prompt:      promptBuf.String(),
		MaxTokens:   4096,
		Temperature: 0.7,
	}

	response, err := s.provider.GenerateText(ctx, request)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var result struct {
		UserStories []struct {
			ID                 string   `json:"id"`
			AsA                string   `json:"as_a"`
			IWant              string   `json:"i_want"`
			SoThat             string   `json:"so_that"`
			AcceptanceCriteria []string `json:"acceptance_criteria"`
			Priority           string   `json:"priority"`
			EstimatePoints     int      `json:"estimate_points"`
			Notes              string   `json:"notes"`
		} `json:"user_stories"`
	}

	if err := json.Unmarshal([]byte(response.Text), &result); err != nil {
		return nil, errors.New("failed to parse user stories: " + err.Error())
	}

	// Convert to domain model
	stories := make([]models.UserStory, len(result.UserStories))
	for i, s := range result.UserStories {
		stories[i] = models.UserStory{
			ID:                 s.ID,
			AsA:                s.AsA,
			IWant:              s.IWant,
			SoThat:             s.SoThat,
			AcceptanceCriteria: s.AcceptanceCriteria,
			Priority:           s.Priority,
			EstimatePoints:     s.EstimatePoints,
		}
	}

	return stories, nil
}

// GenerateAcceptanceCriteria generates acceptance criteria for a user story
func (s *AIService) GenerateAcceptanceCriteria(ctx context.Context, story *models.UserStory) ([]string, error) {
	if s.provider == nil {
		return nil, errors.New("LLM provider not configured")
	}

	tmpl, err := template.New("criteria").Parse(prompts.AcceptanceCriteriaPrompt)
	if err != nil {
		return nil, err
	}

	var promptBuf bytes.Buffer
	err = tmpl.Execute(&promptBuf, map[string]interface{}{
		"AsA":    story.AsA,
		"IWant":  story.IWant,
		"SoThat": story.SoThat,
	})
	if err != nil {
		return nil, err
	}

	request := contract.GenerateRequest{
		Prompt:      promptBuf.String(),
		MaxTokens:   2048,
		Temperature: 0.7,
	}

	response, err := s.provider.GenerateText(ctx, request)
	if err != nil {
		return nil, err
	}

	var result struct {
		AcceptanceCriteria []struct {
			Description string `json:"description"`
		} `json:"acceptance_criteria"`
	}

	if err := json.Unmarshal([]byte(response.Text), &result); err != nil {
		return nil, errors.New("failed to parse acceptance criteria: " + err.Error())
	}

	criteria := make([]string, len(result.AcceptanceCriteria))
	for i, c := range result.AcceptanceCriteria {
		criteria[i] = c.Description
	}

	return criteria, nil
}

// SummarizeMeetingInput holds input for meeting summarization
type SummarizeMeetingInput struct {
	Notes        string
	MeetingType  string
	Participants []string
}

// MeetingSummary represents a summarized meeting
type MeetingSummary struct {
	Summary         string            `json:"summary"`
	KeyPoints       []string          `json:"key_points"`
	Decisions       []MeetingDecision `json:"decisions"`
	ActionItems     []ActionItem      `json:"action_items"`
	OpenQuestions   []string          `json:"open_questions"`
	ProductInsights *ProductInsights  `json:"product_insights,omitempty"`
}

// MeetingDecision represents a decision made in a meeting
type MeetingDecision struct {
	Decision  string `json:"decision"`
	Context   string `json:"context"`
	DecidedBy string `json:"decided_by,omitempty"`
}

// ActionItem represents an action item from a meeting
type ActionItem struct {
	Action   string `json:"action"`
	Owner    string `json:"owner,omitempty"`
	DueDate  string `json:"due_date,omitempty"`
	Priority string `json:"priority"`
}

// ProductInsights represents product-related insights from a meeting
type ProductInsights struct {
	UserFeedback       []string `json:"user_feedback,omitempty"`
	FeatureRequests    []string `json:"feature_requests,omitempty"`
	ProblemsIdentified []string `json:"problems_identified,omitempty"`
}

// SummarizeMeeting summarizes meeting notes
func (s *AIService) SummarizeMeeting(ctx context.Context, input SummarizeMeetingInput) (*MeetingSummary, error) {
	if s.provider == nil {
		return nil, errors.New("LLM provider not configured")
	}

	tmpl, err := template.New("summarize").Parse(prompts.MeetingSummarizerPrompt)
	if err != nil {
		return nil, err
	}

	var promptBuf bytes.Buffer
	err = tmpl.Execute(&promptBuf, map[string]interface{}{
		"Notes":        input.Notes,
		"MeetingType":  input.MeetingType,
		"Participants": input.Participants,
	})
	if err != nil {
		return nil, err
	}

	request := contract.GenerateRequest{
		Prompt:      promptBuf.String(),
		MaxTokens:   4096,
		Temperature: 0.5,
	}

	response, err := s.provider.GenerateText(ctx, request)
	if err != nil {
		return nil, err
	}

	var summary MeetingSummary
	if err := json.Unmarshal([]byte(response.Text), &summary); err != nil {
		return nil, errors.New("failed to parse meeting summary: " + err.Error())
	}

	return &summary, nil
}
