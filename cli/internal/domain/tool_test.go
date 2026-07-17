package domain

import "testing"

func TestToolValidate(t *testing.T) {
	cases := []struct {
		name     string
		tool     Tool
		wantCode string
	}{
		{"valid", Tool{Name: "gh", Purpose: "GitHub CLI for release publishing"}, ""},
		{"missing name", Tool{Name: "", Purpose: "GitHub CLI"}, "missing_required_field"},
		{"missing purpose", Tool{Name: "gh", Purpose: ""}, "missing_required_field"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.tool.Validate()
			assertValidationCode(t, err, tc.wantCode)
		})
	}
}
