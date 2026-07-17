package domain

import "testing"

func TestBacklogValidate(t *testing.T) {
	tiny := LaneTiny
	bogus := "low"
	cases := []struct {
		name     string
		backlog  Backlog
		wantCode string
	}{
		{"valid, no priority", Backlog{Summary: "improve trace scoring"}, ""},
		{"valid, with priority", Backlog{Summary: "improve trace scoring", Priority: &tiny}, ""},
		{"missing summary", Backlog{Summary: ""}, "missing_required_field"},
		{"invalid priority", Backlog{Summary: "improve trace scoring", Priority: &bogus}, "invalid_lane"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.backlog.Validate()
			assertValidationCode(t, err, tc.wantCode)
		})
	}
}
