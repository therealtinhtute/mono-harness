package domain

// Run corresponds to one `work` execution log.
type Run struct {
	ID           string
	StorySlug    string
	PlanID       *string // known gap: unpopulated until phase-plan-template.md carries spec_id
	TraceIDs     []string
	ArtifactPath string
	CreatedAt    string
}

func (r Run) Validate() error {
	if r.StorySlug == "" {
		return &ValidationError{Code: "missing_required_field", Message: "run: story_slug is required"}
	}
	return nil
}
