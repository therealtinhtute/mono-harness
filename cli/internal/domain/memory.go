package domain

// Memory scope enum (CONTRACT.md `memory add`, P5,
// docs/plans/active/durable-memory.md R5): "plan" ties an entry to one
// initiative via PlanID; "global" is repo-wide and carries no PlanID.
const (
	MemoryScopePlan   = "plan"
	MemoryScopeGlobal = "global"
)

var memoryScopes = map[string]bool{
	MemoryScopePlan:   true,
	MemoryScopeGlobal: true,
}

func IsValidMemoryScope(scope string) bool {
	return memoryScopes[scope]
}

// Memory is one docs/memory/{id}.md entry — the markdown-first source of
// truth a `memories` index row is derived from (R1). Type is free text
// (the caller's own classification, not a closed enum); Scope and PlanID
// together enforce R5's plan-vs-global distinction.
type Memory struct {
	Type    string
	Scope   string
	PlanID  string
	Summary string
}

func (m Memory) Validate() error {
	if m.Type == "" {
		return &ValidationError{Code: "missing_required_field", Message: "memory: type is required"}
	}
	if m.Summary == "" {
		return &ValidationError{Code: "missing_required_field", Message: "memory: summary is required"}
	}
	if !IsValidMemoryScope(m.Scope) {
		return &ValidationError{Code: "invalid_scope", Message: "memory: invalid scope " + m.Scope + " (want plan|global)"}
	}
	if m.Scope == MemoryScopePlan && m.PlanID == "" {
		return &ValidationError{Code: "missing_required_field", Message: "memory: scope=plan requires --plan-id"}
	}
	if m.Scope == MemoryScopeGlobal && m.PlanID != "" {
		return &ValidationError{Code: "invalid_scope", Message: "memory: scope=global must not set --plan-id"}
	}
	return nil
}
