package application

import (
	"strings"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestRecordCheckRejectsApprovedVerdictWithFailingProof(t *testing.T) {
	db := freshDB(t)
	runID := createLifecycleRun(t, db, "cli-domain")

	id, err := RecordCheck(db, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{
		{Command: "false", OutputRef: "should not be accepted"},
	}, "")
	if id != "" {
		t.Fatalf("RecordCheck returned id=%q, want empty on verification failure", id)
	}
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "proof_verification_failed" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: proof_verification_failed}", err)
	}
	if !strings.Contains(ve.Message, "false") {
		t.Fatalf("error message = %q, want it to name the failing command", ve.Message)
	}
	if got := countRows(t, db, "checks"); got != 0 {
		t.Fatalf("checks rows = %d, want 0 — a verification failure must not record the check", got)
	}
}

func TestRecordCheckRejectsApproveWithRequestsVerdictWithFailingProof(t *testing.T) {
	db := freshDB(t)
	runID := createLifecycleRun(t, db, "cli-domain")

	_, err := RecordCheck(db, runID, domain.VerdictApproveWithRequests, domain.JudgeIndependent, "test-model", []domain.ProofLink{
		{Command: "false", OutputRef: "should not be accepted"},
	}, "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "proof_verification_failed" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: proof_verification_failed}", err)
	}
}

func TestRecordCheckAcceptsApprovedVerdictWithPassingProof(t *testing.T) {
	db := freshDB(t)
	runID := createLifecycleRun(t, db, "cli-domain")

	id, err := RecordCheck(db, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{
		{Command: "true", OutputRef: "ok"},
	}, "")
	if err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}
	if id == "" {
		t.Fatal("RecordCheck returned empty id")
	}
}

// TestRecordCheckDoesNotVerifyRequestChangesProof proves the verdict-
// conditional design: REQUEST_CHANGES commonly cites a failing command as
// the evidence of the problem it's reporting, so requiring exit 0 there
// would reject exactly the proof a REQUEST_CHANGES verdict needs to carry.
func TestRecordCheckDoesNotVerifyRequestChangesProof(t *testing.T) {
	db := freshDB(t)
	runID := createLifecycleRun(t, db, "cli-domain")

	id, err := RecordCheck(db, runID, domain.VerdictRequestChanges, domain.JudgeIndependent, "test-model", []domain.ProofLink{
		{Command: "false", OutputRef: "reproduces the bug"},
	}, "")
	if err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}
	if id == "" {
		t.Fatal("RecordCheck returned empty id")
	}
}

func TestVerifyProofLinksCapturesOutputTailOnFailure(t *testing.T) {
	err := verifyProofLinks([]domain.ProofLink{{Command: "echo boom && false"}})
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "proof_verification_failed" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: proof_verification_failed}", err)
	}
	if !strings.Contains(ve.Message, "boom") {
		t.Fatalf("error message = %q, want it to include the command's captured output", ve.Message)
	}
}

func TestVerifyProofLinksPassesForEveryPassingCommand(t *testing.T) {
	err := verifyProofLinks([]domain.ProofLink{
		{Command: "true"},
		{Command: "exit 0"},
	})
	if err != nil {
		t.Fatalf("verifyProofLinks: %v", err)
	}
}

func TestVerifyProofLinksStopsAtTheFirstFailingCommand(t *testing.T) {
	err := verifyProofLinks([]domain.ProofLink{
		{Command: "true"},
		{Command: "false"},
	})
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "proof_verification_failed" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: proof_verification_failed}", err)
	}
}
