package domain

type Tool struct {
	ID        string
	Name      string
	Purpose   string
	CreatedAt string
}

func (t Tool) Validate() error {
	if t.Name == "" {
		return &ValidationError{Code: "missing_required_field", Message: "tool: name is required"}
	}
	if t.Purpose == "" {
		return &ValidationError{Code: "missing_required_field", Message: "tool: purpose is required"}
	}
	return nil
}
