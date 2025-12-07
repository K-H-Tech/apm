package prompts

// PRDGeneratorSystemPrompt is the system prompt for PRD generation
const PRDGeneratorSystemPrompt = `You are an expert Product Manager with 15+ years of experience creating comprehensive Product Requirements Documents (PRDs).

Your task is to generate a well-structured PRD from the provided notes, ideas, or meeting transcripts.

Follow these guidelines:
1. Extract and clearly articulate the core problem being solved
2. Identify target user personas with their goals and pain points
3. Create clear, testable user stories in "As a [persona], I want [capability], so that [benefit]" format
4. Define SMART success metrics (Specific, Measurable, Achievable, Relevant, Time-bound)
5. List explicit assumptions and dependencies
6. Clearly define what is OUT of scope
7. Be concise but thorough - every section should add value

Quality standards:
- User stories must have clear acceptance criteria
- Metrics must be quantifiable
- Scope must be clearly bounded
- Language should be clear and unambiguous`

// PRDGeneratorUserPrompt is the template for the user prompt
const PRDGeneratorUserPrompt = `Based on the following notes/ideas, generate a complete Product Requirements Document:

---
INPUT NOTES:
{{.Notes}}
---

{{if .TemplateGuidance}}
Follow this template structure:
{{.TemplateGuidance}}
{{end}}

{{if .AdditionalContext}}
Additional context:
{{.AdditionalContext}}
{{end}}

Generate a comprehensive PRD with the following sections:
1. Title - A clear, descriptive title
2. Overview - High-level summary (2-3 sentences)
3. Problem Statement - What problem are we solving and why now?
4. Goals - 3-5 measurable goals
5. User Personas - 1-3 target user personas with goals and pain points
6. User Stories - 3-7 user stories with acceptance criteria
7. Success Metrics - 3-5 quantifiable metrics with targets
8. Out of Scope - What we're NOT doing
9. Assumptions - Key assumptions being made
10. Dependencies - External dependencies

Respond in valid JSON format.`

// PRDRefinementPrompt is the prompt for refining PRD content
const PRDRefinementPrompt = `You are reviewing and improving a Product Requirements Document based on feedback.

Current PRD content:
{{.CurrentContent}}

Feedback to address:
{{.Feedback}}

{{if .SpecificSection}}
Focus specifically on the "{{.SpecificSection}}" section.
{{end}}

Please:
1. Address all feedback points
2. Maintain consistency with unchanged sections
3. Improve clarity where needed
4. Ensure all acceptance criteria are testable
5. Keep the same JSON structure

Respond with the complete updated PRD in JSON format.`

// ConsistencyCheckPrompt is the prompt for checking PRD consistency
const ConsistencyCheckPrompt = `Analyze this Product Requirements Document for consistency and quality issues:

{{.PRDContent}}

Check for:
1. Alignment between goals and success metrics
2. User stories that support stated goals
3. Acceptance criteria that are testable
4. Scope clarity (in-scope vs out-of-scope)
5. Completeness of all required sections
6. Clarity and unambiguity of language
7. Contradictions between sections
8. Missing dependencies or assumptions

For each issue found, provide:
- Section name
- Issue description
- Severity (error, warning, info)
- Suggested fix

Respond in JSON format with structure:
{
  "is_consistent": boolean,
  "issues": [
    {
      "section": "string",
      "issue": "string",
      "severity": "error|warning|info",
      "suggestion": "string"
    }
  ],
  "suggestions": ["string"]
}`

// PRDOutputSchema defines the expected JSON output structure
var PRDOutputSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"title": map[string]interface{}{
			"type":        "string",
			"description": "Clear, descriptive title for the PRD",
		},
		"overview": map[string]interface{}{
			"type":        "string",
			"description": "High-level summary in 2-3 sentences",
		},
		"problem_statement": map[string]interface{}{
			"type":        "string",
			"description": "The problem being solved and why now",
		},
		"goals": map[string]interface{}{
			"type":        "array",
			"items":       map[string]interface{}{"type": "string"},
			"description": "3-5 measurable goals",
		},
		"user_personas": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":        map[string]interface{}{"type": "string"},
					"description": map[string]interface{}{"type": "string"},
					"goals":       map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"pain_points": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
				},
			},
			"description": "1-3 target user personas",
		},
		"user_stories": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"as_a":                map[string]interface{}{"type": "string"},
					"i_want":              map[string]interface{}{"type": "string"},
					"so_that":             map[string]interface{}{"type": "string"},
					"acceptance_criteria": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"priority":            map[string]interface{}{"type": "string", "enum": []string{"high", "medium", "low"}},
				},
			},
			"description": "3-7 user stories with acceptance criteria",
		},
		"success_metrics": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":       map[string]interface{}{"type": "string"},
					"definition": map[string]interface{}{"type": "string"},
					"target":     map[string]interface{}{"type": "string"},
				},
			},
			"description": "3-5 quantifiable success metrics",
		},
		"out_of_scope": map[string]interface{}{
			"type":        "array",
			"items":       map[string]interface{}{"type": "string"},
			"description": "What is explicitly NOT included",
		},
		"assumptions": map[string]interface{}{
			"type":        "array",
			"items":       map[string]interface{}{"type": "string"},
			"description": "Key assumptions being made",
		},
		"dependencies": map[string]interface{}{
			"type":        "array",
			"items":       map[string]interface{}{"type": "string"},
			"description": "External dependencies",
		},
	},
	"required": []string{"title", "overview", "problem_statement", "goals", "user_stories", "success_metrics"},
}
