package domain

import "testing"

func TestInterventionValidate(t *testing.T) {
	cases := []struct {
		name         string
		intervention Intervention
		wantCode     string
	}{
		{"valid", Intervention{VerdictID: "01J0000000000000000000CHEK", Reason: "human override: acceptable risk"}, ""},
		{"missing verdict_id", Intervention{VerdictID: "", Reason: "human override"}, "missing_required_field"},
		{"missing reason", Intervention{VerdictID: "01J0000000000000000000CHEK", Reason: ""}, "missing_required_field"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.intervention.Validate()
			assertValidationCode(t, err, tc.wantCode)
		})
	}
}
