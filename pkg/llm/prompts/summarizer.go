package prompts

// MeetingSummarizerPrompt is the prompt for summarizing meeting notes
const MeetingSummarizerPrompt = `You are an expert at extracting actionable insights from meeting notes.

Analyze these meeting notes and extract key information:

---
MEETING NOTES:
{{.Notes}}
---

{{if .MeetingType}}
Meeting Type: {{.MeetingType}}
{{end}}

{{if .Participants}}
Participants: {{.Participants}}
{{end}}

Please provide:

1. **Executive Summary** (2-3 sentences)
   - What was the main purpose/outcome of this meeting?

2. **Key Discussion Points** (3-7 bullet points)
   - Main topics discussed
   - Important insights shared

3. **Decisions Made**
   - List each decision with context
   - Note who made the decision if mentioned

4. **Action Items**
   - What needs to be done
   - Who is responsible (if mentioned)
   - Due date (if mentioned)
   - Priority level

5. **Open Questions**
   - Unresolved issues
   - Items needing follow-up
   - Questions for future discussion

6. **Product Insights** (if applicable)
   - User feedback mentioned
   - Feature requests
   - Problem areas identified

Respond in JSON format.`

// MeetingSummaryOutputSchema defines the expected JSON output structure
var MeetingSummaryOutputSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"summary": map[string]interface{}{
			"type":        "string",
			"description": "Executive summary in 2-3 sentences",
		},
		"key_points": map[string]interface{}{
			"type":        "array",
			"items":       map[string]interface{}{"type": "string"},
			"description": "Main discussion points",
		},
		"decisions": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"decision":    map[string]interface{}{"type": "string"},
					"context":     map[string]interface{}{"type": "string"},
					"decided_by":  map[string]interface{}{"type": "string"},
				},
			},
			"description": "Decisions made during the meeting",
		},
		"action_items": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action":   map[string]interface{}{"type": "string"},
					"owner":    map[string]interface{}{"type": "string"},
					"due_date": map[string]interface{}{"type": "string"},
					"priority": map[string]interface{}{"type": "string", "enum": []string{"high", "medium", "low"}},
				},
				"required": []string{"action"},
			},
			"description": "Action items with owners and due dates",
		},
		"open_questions": map[string]interface{}{
			"type":        "array",
			"items":       map[string]interface{}{"type": "string"},
			"description": "Unresolved questions and items needing follow-up",
		},
		"product_insights": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"user_feedback":    map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
				"feature_requests": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
				"problems_identified": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			},
			"description": "Product-related insights from the meeting",
		},
	},
	"required": []string{"summary", "key_points", "action_items"},
}

// DiscoveryInsightsPrompt is the prompt for generating insights from discovery items
const DiscoveryInsightsPrompt = `You are a product discovery expert analyzing research data to extract actionable insights.

Analyze this {{.ItemType}} content:

---
{{.Content}}
---

{{if .RelatedItems}}
Related Discovery Items:
{{range .RelatedItems}}
- {{.Type}}: {{.Title}}
  Key Findings: {{range .KeyFindings}}{{.}}, {{end}}
{{end}}
{{end}}

Please provide:

1. **Summary**
   - Concise summary of the key findings (2-3 sentences)

2. **Key Themes**
   - Recurring patterns or themes identified
   - Common user behaviors or needs

3. **Patterns**
   - Cross-cutting observations
   - Connections between data points

4. **Opportunities**
   - Potential product opportunities identified
   - Unmet user needs
   - Areas for improvement

5. **Suggested Actions**
   - Recommended next steps
   - Priority investigations
   - Validation experiments to run

Respond in JSON format.`

// DiscoveryInsightsOutputSchema defines the expected JSON output structure
var DiscoveryInsightsOutputSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"summary": map[string]interface{}{
			"type":        "string",
			"description": "Concise summary of key findings",
		},
		"key_themes": map[string]interface{}{
			"type":        "array",
			"items":       map[string]interface{}{"type": "string"},
			"description": "Recurring themes identified",
		},
		"patterns": map[string]interface{}{
			"type":        "array",
			"items":       map[string]interface{}{"type": "string"},
			"description": "Cross-cutting patterns and observations",
		},
		"opportunities": map[string]interface{}{
			"type":        "array",
			"items":       map[string]interface{}{"type": "string"},
			"description": "Product opportunities identified",
		},
		"suggested_actions": map[string]interface{}{
			"type":        "array",
			"items":       map[string]interface{}{"type": "string"},
			"description": "Recommended next steps",
		},
	},
	"required": []string{"summary", "key_themes", "opportunities", "suggested_actions"},
}

// NotesToPRDPrompt is the prompt for converting meeting notes directly to PRD sections
const NotesToPRDPrompt = `You are converting meeting notes into PRD-ready content.

Meeting Notes:
{{.Notes}}

{{if .ExistingPRD}}
This should be integrated with the existing PRD:
{{.ExistingPRD}}
{{end}}

Extract and structure the following from the notes:

1. **Problems Discussed** → Problem Statement
2. **User Mentions** → User Personas
3. **Feature Ideas** → User Stories
4. **Success Criteria Mentioned** → Success Metrics
5. **Concerns/Risks** → Assumptions or Dependencies
6. **Scope Discussions** → In/Out of Scope

For each extracted item, indicate your confidence level (high, medium, low).

Respond in JSON format with clearly labeled sections.`

// NotesToPRDOutputSchema defines the expected JSON output structure for notes-to-PRD conversion
var NotesToPRDOutputSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"problems_discussed": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"problem":    map[string]interface{}{"type": "string"},
					"confidence": map[string]interface{}{"type": "string", "enum": []string{"high", "medium", "low"}},
				},
				"required": []string{"problem", "confidence"},
			},
			"description": "Problems discussed that form the problem statement",
		},
		"user_mentions": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_type":  map[string]interface{}{"type": "string"},
					"context":    map[string]interface{}{"type": "string"},
					"confidence": map[string]interface{}{"type": "string", "enum": []string{"high", "medium", "low"}},
				},
				"required": []string{"user_type", "confidence"},
			},
			"description": "User types mentioned that map to personas",
		},
		"feature_ideas": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"feature":    map[string]interface{}{"type": "string"},
					"context":    map[string]interface{}{"type": "string"},
					"confidence": map[string]interface{}{"type": "string", "enum": []string{"high", "medium", "low"}},
				},
				"required": []string{"feature", "confidence"},
			},
			"description": "Feature ideas that map to user stories",
		},
		"success_criteria_mentioned": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"criterion":  map[string]interface{}{"type": "string"},
					"confidence": map[string]interface{}{"type": "string", "enum": []string{"high", "medium", "low"}},
				},
				"required": []string{"criterion", "confidence"},
			},
			"description": "Success criteria mentioned for metrics",
		},
		"concerns_risks": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"concern":    map[string]interface{}{"type": "string"},
					"type":       map[string]interface{}{"type": "string", "enum": []string{"assumption", "dependency", "risk"}},
					"confidence": map[string]interface{}{"type": "string", "enum": []string{"high", "medium", "low"}},
				},
				"required": []string{"concern", "confidence"},
			},
			"description": "Concerns and risks for assumptions/dependencies",
		},
		"scope_discussions": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"in_scope": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string"},
				},
				"out_of_scope": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string"},
				},
			},
			"description": "Scope discussions for in/out of scope items",
		},
	},
	"required": []string{"problems_discussed", "feature_ideas"},
}
