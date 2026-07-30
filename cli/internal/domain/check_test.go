package domain

import "testing"

func TestCheckValidate(t *testing.T) {
	proof := []ProofLink{{Command: "go test ./...", OutputRef: "run-log#L10", ArtifactPath: ".kit/runs/work/x.md"}}

	cases := []struct {
		name     string
		check    Check
		wantCode string
	}{
		{"valid approved with proof", Check{RunID: "r1", Verdict: VerdictApproved, Judge: JudgeIndependent, JudgeModel: "claude-opus-5", ProofLinks: proof}, ""},
		{"valid approve-with-requests with proof", Check{RunID: "r1", Verdict: VerdictApproveWithRequests, Judge: JudgeIndependent, JudgeModel: "claude-opus-5", ProofLinks: proof}, ""},
		{"valid request-changes without proof", Check{RunID: "r1", Verdict: VerdictRequestChanges, Judge: JudgeSameSession, JudgeModel: "claude-opus-5"}, ""},
		{"missing run_id", Check{RunID: "", Verdict: VerdictApproved, Judge: JudgeIndependent, JudgeModel: "claude-opus-5", ProofLinks: proof}, "missing_required_field"},
		{"invalid verdict", Check{RunID: "r1", Verdict: "MAYBE", Judge: JudgeIndependent, JudgeModel: "claude-opus-5", ProofLinks: proof}, "invalid_verdict"},
		{"approved without proof", Check{RunID: "r1", Verdict: VerdictApproved, Judge: JudgeIndependent, JudgeModel: "claude-opus-5"}, "empty_proof_links"},
		{"approve-with-requests without proof", Check{RunID: "r1", Verdict: VerdictApproveWithRequests, Judge: JudgeIndependent, JudgeModel: "claude-opus-5"}, "empty_proof_links"},
		{"invalid judge", Check{RunID: "r1", Verdict: VerdictApproved, Judge: "maybe", JudgeModel: "claude-opus-5", ProofLinks: proof}, "invalid_judge"},
		{"missing judge_model", Check{RunID: "r1", Verdict: VerdictApproved, Judge: JudgeIndependent, ProofLinks: proof}, "missing_required_field"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.check.Validate()
			assertValidationCode(t, err, tc.wantCode)
		})
	}
}
