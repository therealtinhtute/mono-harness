package domain

type Trace struct {
	ID        string
	RunID     *string
	Wave      int
	Summary   string
	CreatedAt string
}

func (t Trace) Validate() error {
	if t.Summary == "" {
		return &ValidationError{Code: "missing_required_field", Message: "trace: summary is required"}
	}
	if t.Wave < 0 {
		// Not CONTRACT.md-documented (only unknown_run_id is listed for
		// `trace`): a negative --wave is a cobra int-parse-shaped mistake,
		// not a contract-classified failure.
		return &ValidationError{Message: "trace: wave must be >= 0"}
	}
	return nil
}
