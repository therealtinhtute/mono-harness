package application

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestAuditCleanState(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	seedRun(t, db, changesetDir) // no meta pointer set, so Resume's drift checks have nothing to cross-check

	report, err := Audit(db, root, "dev")
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(report.PointerDrift) != 0 {
		t.Fatalf("pointer_drift = %v, want none", report.PointerDrift)
	}
	if len(report.UnlinkedProofs) != 0 {
		t.Fatalf("unlinked_proofs = %v, want none (no checks recorded)", report.UnlinkedProofs)
	}
	// root is an empty temp dir with no SPEC.md, so Validate's own doc-walk
	// still reports its one baseline missing_key finding (see
	// TestValidateMissingSpec) — composed here unchanged, not duplicated.
	if len(report.ContractViolations) != 1 || report.ContractViolations[0].Issue != "missing_key" {
		t.Fatalf("contract_violations = %v, want exactly one missing_key (composed from Validate)", report.ContractViolations)
	}
}

// TestAuditUnlinkedProofFixture proves audit's own new check (unlinked
// proof links) fires and lists the finding, per T2's verification:
// "staled-pointer fixture ... lists the finding."
func TestAuditUnlinkedProofFixture(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()

	baseline, err := Audit(db, root, "dev")
	if err != nil {
		t.Fatalf("Audit (baseline): %v", err)
	}
	// root is an empty temp dir with no SPEC.md, so Validate's own doc-walk
	// always reports its one baseline missing_key finding (see
	// TestValidateMissingSpec) — that's the floor, not zero.
	if len(baseline.UnlinkedProofs) != 0 {
		t.Fatalf("baseline = %+v, want zero unlinked_proofs", baseline)
	}

	runID := seedRun(t, db, changesetDir)
	proofLinks := []domain.ProofLink{
		{Command: "go test ./...", OutputRef: "PASS", ArtifactPath: filepath.Join(root, "missing-report.md")},
	}
	if _, _, err := RecordCheck(db, changesetDir, runID, "APPROVED", proofLinks); err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}

	after, err := Audit(db, root, "dev")
	if err != nil {
		t.Fatalf("Audit (after): %v", err)
	}
	if len(after.UnlinkedProofs) != 1 {
		t.Fatalf("unlinked_proofs = %v, want exactly one (missing artifact_path)", after.UnlinkedProofs)
	}
}

// TestAuditDeterministic proves identical input produces byte-identical
// JSON across repeated calls (T4's determinism requirement, checked here
// at the audit-composition level ahead of the full gate-flow fixture).
func TestAuditDeterministic(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	runID := seedRun(t, db, changesetDir)
	proofLinks := []domain.ProofLink{
		{Command: "go test ./...", OutputRef: "PASS", ArtifactPath: filepath.Join(root, "missing-report.md")},
	}
	if _, _, err := RecordCheck(db, changesetDir, runID, "APPROVED", proofLinks); err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}

	first, err := Audit(db, root, "dev")
	if err != nil {
		t.Fatalf("Audit (first): %v", err)
	}
	second, err := Audit(db, root, "dev")
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
