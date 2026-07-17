package domain

import "testing"

func TestCheckValidate(t *testing.T) {
	proof := []ProofLink{{Command: "go test ./...", OutputRef: "run-log#L10", ArtifactPath: ".kit/runs/work/x.md"}}

	cases := []struct {
		name     string
		check    Check
		wantCode string
	}{
		{"valid approved with proof", Check{RunID: "r1", Verdict: VerdictApproved, ProofLinks: proof}, ""},
		{"valid approve-with-requests with proof", Check{RunID: "r1", Verdict: VerdictApproveWithRequests, ProofLinks: proof}, ""},
		{"valid request-changes without proof", Check{RunID: "r1", Verdict: VerdictRequestChanges}, ""},
		{"missing run_id", Check{RunID: "", Verdict: VerdictApproved, ProofLinks: proof}, "missing_required_field"},
		{"invalid verdict", Check{RunID: "r1", Verdict: "MAYBE", ProofLinks: proof}, "invalid_verdict"},
		{"approved without proof", Check{RunID: "r1", Verdict: VerdictApproved}, "empty_proof_links"},
		{"approve-with-requests without proof", Check{RunID: "r1", Verdict: VerdictApproveWithRequests}, "empty_proof_links"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.check.Validate()
			assertValidationCode(t, err, tc.wantCode)
		})
	}
}
