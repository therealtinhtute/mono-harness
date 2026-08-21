package application

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestAuditCleanState(t *testing.T) {
	db := freshDB(t)
	seedRun(t, db)

	report, err := Audit(db, "dev", ".")
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(report.PointerDrift) != 0 || len(report.ContractViolations) != 0 || len(report.UnlinkedProofs) != 0 {
		t.Fatalf("report = %+v, want no findings", report)
	}
}

func TestAuditSurfacesInvalidLifecycleEnums(t *testing.T) {
	db := freshDB(t)
	ids := createValidRetainedLifecycle(t, db)
	if _, err := db.Exec(`UPDATE stories SET status = 'bogus' WHERE id = ?`, ids["story"]); err != nil {
		t.Fatalf("corrupt story status: %v", err)
	}
	if _, err := db.Exec(`UPDATE checks SET verdict = 'bogus' WHERE id = ?`, ids["check"]); err != nil {
		t.Fatalf("corrupt check verdict: %v", err)
	}

	report, err := Audit(db, "dev", ".")
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
	db := freshDB(t)
	runID := createLifecycleRun(t, db, "cli-domain")
	proofLinks := []domain.ProofLink{
		{Command: "true", OutputRef: "PASS"},
		{Command: "true", OutputRef: "PASS", ArtifactPath: filepath.Join(t.TempDir(), "missing-report.md")},
	}
	if _, err := RecordCheck(db, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", proofLinks); err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}

	report, err := Audit(db, "dev", ".")
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
	db := freshDB(t)
	runID := createLifecycleRun(t, db, "cli-domain")
	if _, err := RecordCheck(db, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{{Command: "true"}}); err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}

	first, err := Audit(db, "dev", ".")
	if err != nil {
		t.Fatalf("Audit (first): %v", err)
	}
	second, err := Audit(db, "dev", ".")
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

func TestAuditReportsMissingAuthoredDocs(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoRoot, "AGENTS.md"), []byte("managed\n"), 0o644); err != nil {
		t.Fatalf("write managed root doc: %v", err)
	}

	db := freshDB(t)
	seedRun(t, db)
	report, err := Audit(db, "dev", repoRoot)
	if err != nil {
		t.Fatalf("Audit without authored docs: %v", err)
	}
	if len(report.ContractViolations) != 1 {
		t.Fatalf("contract_violations = %v, want one authored-docs finding", report.ContractViolations)
	}
	finding := report.ContractViolations[0]
	if finding.Identifier != "authored_docs_missing" || finding.Severity != "warning" {
		t.Fatalf("finding = %+v, want authored_docs_missing warning", finding)
	}
	if !strings.Contains(finding.Detail, "presence") || strings.Contains(strings.ToLower(finding.Detail), "correctness") {
		t.Fatalf("finding detail = %q, want presence-only wording", finding.Detail)
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if len(fields) != 3 {
		t.Fatalf("audit top-level fields = %v, want exactly three arrays", fields)
	}
	for _, field := range []string{"pointer_drift", "contract_violations", "unlinked_proofs"} {
		if _, ok := fields[field]; !ok {
			t.Fatalf("audit top-level fields = %v, missing %q", fields, field)
		}
	}

	if err := os.MkdirAll(filepath.Join(repoRoot, "docs"), 0o755); err != nil {
		t.Fatalf("create docs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "docs", "README.md"), []byte("# Authored\n"), 0o644); err != nil {
		t.Fatalf("write authored doc: %v", err)
	}
	report, err = Audit(db, "dev", repoRoot)
	if err != nil {
		t.Fatalf("Audit with authored docs: %v", err)
	}
	if len(report.ContractViolations) != 0 {
		t.Fatalf("contract_violations = %v, want none after authored doc is restored", report.ContractViolations)
	}
}
