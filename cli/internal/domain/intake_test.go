package domain

import "testing"

func TestIntakeValidate(t *testing.T) {
	cases := []struct {
		name     string
		intake   Intake
		wantCode string // "" means Validate() must return nil
	}{
		{"valid new-spec tiny", Intake{Type: IntakeNewSpec, Summary: "add x", Lane: LaneTiny}, ""},
		{"valid spec-slice normal", Intake{Type: IntakeSpecSlice, Summary: "slice y", Lane: LaneNormal}, ""},
		{"valid harness-improvement high-risk", Intake{Type: IntakeHarnessImprovement, Summary: "z", Lane: LaneHighRisk}, ""},
		{"invalid type", Intake{Type: "bogus", Summary: "add x", Lane: LaneTiny}, "invalid_type"},
		{"empty type", Intake{Type: "", Summary: "add x", Lane: LaneTiny}, "invalid_type"},
		{"missing summary", Intake{Type: IntakeNewSpec, Summary: "", Lane: LaneTiny}, "missing_required_field"},
		{"invalid lane", Intake{Type: IntakeNewSpec, Summary: "add x", Lane: "low"}, "invalid_lane"},
		{"underscore alias not accepted", Intake{Type: "new_spec", Summary: "add x", Lane: LaneTiny}, "invalid_type"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.intake.Validate()
			assertValidationCode(t, err, tc.wantCode)
		})
	}
}

// assertValidationCode asserts err is nil (wantCode == "") or a
// *ValidationError with exactly wantCode.
func assertValidationCode(t *testing.T, err error, wantCode string) {
	t.Helper()
	if wantCode == "" {
		if err != nil {
			t.Fatalf("Validate() = %v, want nil", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("Validate() = nil, want ValidationError code %q", wantCode)
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("Validate() error type = %T, want *ValidationError", err)
	}
	if ve.Code != wantCode {
		t.Fatalf("Validate() code = %q, want %q", ve.Code, wantCode)
	}
}
