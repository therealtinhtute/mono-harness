package domain

import "testing"

func TestTraceValidate(t *testing.T) {
	runID := "01J0000000000000000000RUNN"
	cases := []struct {
		name        string
		trace       Trace
		wantCode    string
		wantAnyCode bool // true: expect a non-nil, uncoded error (Code == "")
	}{
		{"valid, no run id", Trace{Wave: 1, Summary: "wave 1 done"}, "", false},
		{"valid, with run id", Trace{RunID: &runID, Wave: 2, Summary: "wave 2 done"}, "", false},
		{"valid, wave zero", Trace{Wave: 0, Summary: "initial"}, "", false},
		{"missing summary", Trace{Wave: 1, Summary: ""}, "missing_required_field", false},
		{"negative wave", Trace{Wave: -1, Summary: "bad wave"}, "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.trace.Validate()
			if tc.wantAnyCode {
				if err == nil {
					t.Fatal("Validate() = nil, want a non-nil error")
				}
				if ve, ok := err.(*ValidationError); ok && ve.Code != "" {
					t.Fatalf("Validate() code = %q, want empty (uncoded internal invariant)", ve.Code)
				}
				return
			}
			assertValidationCode(t, err, tc.wantCode)
		})
	}
}
