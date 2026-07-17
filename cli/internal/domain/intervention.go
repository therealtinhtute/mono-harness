package domain

// Intervention records a human override of a check verdict (CONTRACT.md
// `intervention`; documented escalation path, not auto-invoked).
type Intervention struct {
	ID        string
	VerdictID string // FK checks.id
	Reason    string
	CreatedAt string
}

func (iv Intervention) Validate() error {
	if iv.VerdictID == "" {
		return &ValidationError{Code: "missing_required_field", Message: "intervention: verdict_id is required"}
	}
	if iv.Reason == "" {
		return &ValidationError{Code: "missing_required_field", Message: "intervention: reason is required"}
	}
	return nil
}
