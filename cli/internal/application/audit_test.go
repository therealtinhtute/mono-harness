package application

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestAuditCleanState(t *testing.T) {
	db, changesetDir := freshDB(t)
	seedRun(t, db, changesetDir)

	report, err := Audit(db, "dev")
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(report.PointerDrift) != 0 || len(report.ContractViolations) != 0 || len(report.UnlinkedProofs) != 0 {
		t.Fatalf("report = %+v, want no findings", report)
	}
}

func TestAuditSurfacesInvalidLifecycleEnums(t *testing.T) {
	db, changesetDir := freshDB(t)
	ids := createValidRetainedLifecycle(t, db, changesetDir)
	if _, err := db.Exec(`UPDATE stories SET status = 'bogus' WHERE id = ?`, ids["story"]); err != nil {
		t.Fatalf("corrupt story status: %v", err)
	}
	if _, err := db.Exec(`UPDATE checks SET verdict = 'bogus' WHERE id = ?`, ids["check"]); err != nil {
		t.Fatalf("corrupt check verdict: %v", err)
	}

	report, err := Audit(db, "dev")
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(report.ContractViolations) != 2 {
		t.Fatalf("contract_violations = %v, want story status and check verdict findings", report.ContractViolations)
	}
	if report.ContractViolations[0].Link != "DB->STORY" || report.ContractViolations[1].Link != "DB->CHECK" {
		t.Fatalf("contract_violations = %v, want deterministic story then check findings", report.ContractViolations)
	}
	if len(report.PointerDrift) != 1 || report.PointerDrift[0].Type != "invalid_status" {
		t.Fatalf("pointer_drift = %v, want one invalid_status finding", report.PointerDrift)
	}
}

func TestAuditIgnoresLegacyProofArtifactPaths(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := createLifecycleRun(t, db, changesetDir, "cli-domain")
	proofLinks := []domain.ProofLink{
		{Command: "go test ./...", OutputRef: "PASS"},
		{Command: "go vet ./...", OutputRef: "PASS", ArtifactPath: filepath.Join(t.TempDir(), "missing-report.md")},
	}
	if _, _, err := RecordCheck(db, changesetDir, runID, domain.VerdictApproved, proofLinks); err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}

	report, err := Audit(db, "dev")
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(report.ContractViolations) != 0 {
		t.Fatalf("contract_violations = %v, want none", report.ContractViolations)
	}
	if len(report.UnlinkedProofs) != 0 {
		t.Fatalf("unlinked_proofs = %v, want none for optional legacy artifact paths", report.UnlinkedProofs)
	}
}

func TestAuditDeterministic(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := createLifecycleRun(t, db, changesetDir, "cli-domain")
	if _, _, err := RecordCheck(db, changesetDir, runID, domain.VerdictApproved, []domain.ProofLink{{Command: "go test ./..."}}); err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}

	first, err := Audit(db, "dev")
	if err != nil {
		t.Fatalf("Audit (first): %v", err)
	}
	second, err := Audit(db, "dev")
	if err != nil {
		t.Fatalf("Audit (second): %v", err)
	}

	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal first: %v", err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("marshal second: %v", err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("audit output not deterministic:\nfirst:  %s\nsecond: %s", firstJSON, secondJSON)
	}
}
