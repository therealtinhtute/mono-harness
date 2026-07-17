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

	report, err := Audit(db, root)
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
	if report.EntropyScore != 5 {
		t.Fatalf("entropy_score = %d, want 5 (one contract_violation * weight 5)", report.EntropyScore)
	}
}

// TestAuditUnlinkedProofFixture proves audit's own new check (unlinked
// proof links) fires and moves the entropy score, per T2's verification:
// "staled-pointer fixture changes the score and lists the finding."
func TestAuditUnlinkedProofFixture(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()

	baseline, err := Audit(db, root)
	if err != nil {
		t.Fatalf("Audit (baseline): %v", err)
	}
	// root is an empty temp dir with no SPEC.md, so Validate's own doc-walk
	// always reports its one baseline missing_key finding (see
	// TestValidateMissingSpec) — that's the floor, not zero.
	if len(baseline.UnlinkedProofs) != 0 || baseline.EntropyScore != 5 {
		t.Fatalf("baseline = %+v, want zero unlinked_proofs and score 5 (one contract_violation)", baseline)
	}

	runID := seedRun(t, db, changesetDir)
	proofLinks := []domain.ProofLink{
		{Command: "go test ./...", OutputRef: "PASS", ArtifactPath: filepath.Join(root, "missing-report.md")},
	}
	if _, _, err := RecordCheck(db, changesetDir, runID, "APPROVED", proofLinks); err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}

	after, err := Audit(db, root)
	if err != nil {
		t.Fatalf("Audit (after): %v", err)
	}
	if len(after.UnlinkedProofs) != 1 {
		t.Fatalf("unlinked_proofs = %v, want exactly one (missing artifact_path)", after.UnlinkedProofs)
	}
	if after.EntropyScore != 13 {
		t.Fatalf("entropy_score = %d, want 13 (one contract_violation*5 + one unlinked_proof*8)", after.EntropyScore)
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

	first, err := Audit(db, root)
	if err != nil {
		t.Fatalf("Audit (first): %v", err)
	}
	second, err := Audit(db, root)
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

func TestProposeFromAuditFindings(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	runID := seedRun(t, db, changesetDir)
	proofLinks := []domain.ProofLink{
		{Command: "go test ./...", OutputRef: "PASS", ArtifactPath: filepath.Join(root, "missing-report.md")},
	}
	if _, _, err := RecordCheck(db, changesetDir, runID, "APPROVED", proofLinks); err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}

	report, err := Propose(db, root)
	if err != nil {
		t.Fatalf("Propose: %v", err)
	}
	found := false
	for _, p := range report.Proposals {
		if p.Pattern == "unlinked_proofs" {
			found = true
		}
	}
	if !found {
		t.Fatalf("proposals = %v, want one with pattern unlinked_proofs", report.Proposals)
	}
}

// TestProposeCleanState uses the chain-valid fixture, which the existing
// TestValidateChainValidFixture proves produces exactly one
// (not_yet_implemented) Validate finding — the known, already-documented
// SPEC->PLAN gap, not a new violation. A truly zero-finding project isn't
// currently achievable (that known gap always fires), so "clean" here
// means exactly that one proposal and nothing else.
func TestProposeCleanState(t *testing.T) {
	db, changesetDir := freshDB(t)
	seedRun(t, db, changesetDir)

	report, err := Propose(db, filepath.Join("..", "..", "testdata", "chain-valid"))
	if err != nil {
		t.Fatalf("Propose: %v", err)
	}
	if len(report.Proposals) != 1 || report.Proposals[0].Pattern != "contract_violations" {
		t.Fatalf("proposals = %v, want exactly one with pattern contract_violations (the known not_yet_implemented gap)", report.Proposals)
	}
}
