package domain

import "strings"

// Decision is one entry of the `decisions` table — the compressed-index
// counterpart of a plan's `## Decisions` markdown section. Phase and Task
// are optional: a decision need not be tied to a specific in-flight task
// (e.g. a decision made during to-plan, before any task exists).
type Decision struct {
	Decision  string `json:"decision"`
	Rationale string `json:"rationale"`
	Phase     string `json:"phase"`
	Task      string `json:"task"`
}

func (d Decision) Validate() error {
	if strings.TrimSpace(d.Decision) == "" {
		return &ValidationError{Code: "missing_required_field", Message: "decision: decision text is required"}
	}
	if strings.TrimSpace(d.Rationale) == "" {
		return &ValidationError{Code: "missing_required_field", Message: "decision: rationale is required"}
	}
	return nil
}
