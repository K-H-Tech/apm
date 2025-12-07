package prompts

// UserStoryGeneratorPrompt is the prompt for generating user stories
const UserStoryGeneratorPrompt = `You are an expert Product Manager creating user stories from a PRD.

Given this PRD overview and context:

Overview: {{.Overview}}
Problem Statement: {{.ProblemStatement}}

User Personas:
{{range .Personas}}
- {{.Name}}: {{.Description}}
  Goals: {{range .Goals}}{{.}}, {{end}}
  Pain Points: {{range .PainPoints}}{{.}}, {{end}}
{{end}}

Generate comprehensive user stories following these guidelines:

1. Use the format: "As a [persona], I want [capability], so that [benefit]"
2. Each story should:
   - Be independent and deliverable
   - Provide clear value to the user
   - Be testable
3. Include 3-5 specific, testable acceptance criteria per story
4. Suggest story point estimates using Fibonacci (1, 2, 3, 5, 8, 13)
5. Assign priority (high, medium, low) based on user value and dependencies

Generate 5-10 user stories that together deliver the PRD's goals.

Respond in JSON format with this structure:
{
  "user_stories": [
    {
      "id": "US-001",
      "as_a": "persona name",
      "i_want": "capability",
      "so_that": "benefit",
      "acceptance_criteria": ["criterion 1", "criterion 2", ...],
      "priority": "high|medium|low",
      "estimate_points": 3,
      "notes": "optional implementation notes"
    }
  ]
}`

// AcceptanceCriteriaPrompt is the prompt for generating acceptance criteria
const AcceptanceCriteriaPrompt = `You are writing acceptance criteria for a user story.

User Story:
"As a {{.AsA}}, I want {{.IWant}}, so that {{.SoThat}}"

{{if .Context}}
Additional Context:
{{.Context}}
{{end}}

Generate comprehensive acceptance criteria following these guidelines:

1. Use Given-When-Then format where appropriate
2. Each criterion must be:
   - Specific and unambiguous
   - Testable (can verify pass/fail)
   - Complete (covers the scenario fully)
3. Include:
   - Happy path scenarios
   - Edge cases
   - Error handling
   - Validation rules (if applicable)
   - Performance requirements (if applicable)
4. Aim for 5-8 criteria

Respond in JSON format:
{
  "acceptance_criteria": [
    {
      "id": "AC-001",
      "description": "Full criterion description",
      "type": "functional|edge_case|validation|performance|error_handling",
      "given_when_then": {
        "given": "precondition",
        "when": "action",
        "then": "expected result"
      }
    }
  ]
}`

// UserStoryRefinementPrompt is the prompt for refining user stories
const UserStoryRefinementPrompt = `Review and improve this user story:

Current Story:
As a {{.AsA}}, I want {{.IWant}}, so that {{.SoThat}}

Current Acceptance Criteria:
{{range $i, $c := .AcceptanceCriteria}}
{{$i}}. {{$c}}
{{end}}

{{if .Feedback}}
Feedback to address:
{{.Feedback}}
{{end}}

Please:
1. Improve clarity and specificity of the story
2. Ensure the "so that" clearly articulates user value
3. Make acceptance criteria more testable
4. Add any missing edge cases
5. Suggest appropriate story points if not set

Respond in JSON format with the improved story.`

// StoryOutputSchema defines the expected JSON output structure for user stories
var StoryOutputSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"user_stories": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Story identifier (e.g., US-001)",
					},
					"as_a": map[string]interface{}{
						"type":        "string",
						"description": "The user persona",
					},
					"i_want": map[string]interface{}{
						"type":        "string",
						"description": "The desired capability",
					},
					"so_that": map[string]interface{}{
						"type":        "string",
						"description": "The benefit/value",
					},
					"acceptance_criteria": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Testable acceptance criteria",
					},
					"priority": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"high", "medium", "low"},
						"description": "Story priority",
					},
					"estimate_points": map[string]interface{}{
						"type":        "integer",
						"enum":        []int{1, 2, 3, 5, 8, 13, 21},
						"description": "Fibonacci story points",
					},
					"notes": map[string]interface{}{
						"type":        "string",
						"description": "Optional implementation notes",
					},
				},
				"required": []string{"as_a", "i_want", "so_that", "acceptance_criteria"},
			},
		},
	},
	"required": []string{"user_stories"},
}

// AcceptanceCriteriaOutputSchema defines the expected JSON output for acceptance criteria
var AcceptanceCriteriaOutputSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"acceptance_criteria": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type": "string",
					},
					"description": map[string]interface{}{
						"type": "string",
					},
					"type": map[string]interface{}{
						"type": "string",
						"enum": []string{"functional", "edge_case", "validation", "performance", "error_handling"},
					},
					"given_when_then": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"given": map[string]interface{}{"type": "string"},
							"when":  map[string]interface{}{"type": "string"},
							"then":  map[string]interface{}{"type": "string"},
						},
					},
				},
				"required": []string{"description"},
			},
		},
	},
	"required": []string{"acceptance_criteria"},
}
