package domain

type Decision struct {
	ID        string
	Summary   string
	Rationale string
	Rejected  *string
	CreatedAt string
}

func (d Decision) Validate() error {
	if d.Summary == "" {
		return &ValidationError{Code: "missing_required_field", Message: "decision: summary is required"}
	}
	if d.Rationale == "" {
		return &ValidationError{Code: "missing_required_field", Message: "decision: rationale is required"}
	}
	return nil
}
