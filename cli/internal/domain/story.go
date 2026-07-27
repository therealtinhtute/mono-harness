package domain

import "fmt"

// Story status enum (STATE.md).
const (
	StoryPlanned    = "planned"
	StoryInProgress = "in-progress"
	StoryChecked    = "checked"
	StoryDone       = "done"
)

var storyStatuses = map[string]bool{
	StoryPlanned:    true,
	StoryInProgress: true,
	StoryChecked:    true,
	StoryDone:       true,
}

func IsValidStoryStatus(status string) bool {
	return storyStatuses[status]
}

// Story carries phase semantics — see SCHEMA.md's table-count note (no
// separate `phases` table; story slug = phase slug).
type Story struct {
	ID        string
	Slug      string
	Goal      string
	Status    string
	DependsOn *string
	CreatedAt string
}

func (s Story) Validate() error {
	if s.Slug == "" {
		return &ValidationError{Code: "missing_required_field", Message: "story: slug is required"}
	}
	if s.Goal == "" {
		return &ValidationError{Code: "missing_required_field", Message: "story: goal is required"}
	}
	if !IsValidStoryStatus(s.Status) {
		// Not a CONTRACT.md-documented `story` error: status is always set
		// internally (creation always starts "planned"), never a direct
		// user-supplied argument, so this guards an internal invariant only.
		return &ValidationError{Message: fmt.Sprintf("story: invalid status %q", s.Status)}
	}
	return nil
}
