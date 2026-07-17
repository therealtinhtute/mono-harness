package domain

import "testing"

func TestDecisionValidate(t *testing.T) {
	rejected := "considered X, rejected because Y"
	cases := []struct {
		name     string
		decision Decision
		wantCode string
	}{
		{"valid", Decision{Summary: "use SQLite", Rationale: "durable local storage"}, ""},
		{"valid with rejected", Decision{Summary: "use SQLite", Rationale: "durable local storage", Rejected: &rejected}, ""},
		{"missing summary", Decision{Summary: "", Rationale: "durable local storage"}, "missing_required_field"},
		{"missing rationale", Decision{Summary: "use SQLite", Rationale: ""}, "missing_required_field"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.decision.Validate()
			assertValidationCode(t, err, tc.wantCode)
		})
	}
}
