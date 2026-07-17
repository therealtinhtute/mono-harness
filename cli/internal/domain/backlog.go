package domain

import "fmt"

type Backlog struct {
	ID        string
	Summary   string
	Priority  *string // enum: tiny|normal|high-risk, see LaneTiny etc.
	CreatedAt string
}

func (b Backlog) Validate() error {
	if b.Summary == "" {
		return &ValidationError{Code: "missing_required_field", Message: "backlog: summary is required"}
	}
	if b.Priority != nil && !lanes[*b.Priority] {
		// Reuses `invalid_lane` (not separately documented for `backlog`):
		// --priority is the exact same tiny|normal|high-risk enum CONTRACT.md
		// names for `intake --lane`, so the existing code classifies it
		// accurately without inventing a new one.
		return &ValidationError{Code: "invalid_lane", Message: fmt.Sprintf("backlog: invalid priority %q", *b.Priority)}
	}
	return nil
}
