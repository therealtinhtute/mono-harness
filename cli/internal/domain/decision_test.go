package domain

import "testing"

func TestDecisionValidate(t *testing.T) {
	cases := []struct {
		name     string
		decision Decision
		wantCode string
	}{
		{"valid, minimal", Decision{Decision: "used JSON array batching", Rationale: "matches proof_links precedent"}, ""},
		{"valid, with phase and task", Decision{Decision: "d", Rationale: "r", Phase: "p1", Task: "wave 1 task 2"}, ""},
		{"missing decision", Decision{Decision: "", Rationale: "r"}, "missing_required_field"},
		{"blank decision", Decision{Decision: "   ", Rationale: "r"}, "missing_required_field"},
		{"missing rationale", Decision{Decision: "d", Rationale: ""}, "missing_required_field"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertValidationCode(t, tc.decision.Validate(), tc.wantCode)
		})
	}
}
