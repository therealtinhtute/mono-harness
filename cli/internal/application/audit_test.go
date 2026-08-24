package application

import (
	"encoding/json"
	"os"
	"os/exec"
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

// initGitRepoFixture turns dir into a deterministic git repository and
// returns a runner for further read/write git commands inside it.
func initGitRepoFixture(t *testing.T, dir string) func(...string) string {
	t.Helper()
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
			"GIT_OPTIONAL_LOCKS=0")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "test")
	return run
}

// TestAuditUnpinnedDocStaysSilent pins nothing: a document citing source
// paths without a pin declaration produces no finding and no error (R5), so
// audit output on an unpinned tree is unchanged by this mechanism.
func TestAuditUnpinnedDocStaysSilent(t *testing.T) {
	repoRoot := t.TempDir()
	docs := filepath.Join(repoRoot, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	body := "# Note\n\nSee `cli/internal/application/audit.go:33` for the entry point.\n"
	if err := os.WriteFile(filepath.Join(docs, "Note.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write unpinned doc: %v", err)
	}

	db := freshDB(t)
	seedRun(t, db)
	report, err := Audit(db, "dev", repoRoot)
	if err != nil {
		t.Fatalf("Audit with unpinned doc: %v", err)
	}
	if len(report.ContractViolations) != 0 {
		t.Fatalf("contract_violations = %v, want none for an unpinned document (R5)", report.ContractViolations)
	}
}

// TestAuditPinScopingExemptDirectories pins stale-looking files in each of
// the three exempt directories with a deliberately bogus pin SHA. If scoping
// ever regressed to reading subdirectory documents, measurement would run and
// the bogus SHA would surface as an error — so no findings AND no error is
// the proof that exclusion happens before any git call (R4).
func TestAuditPinScopingExemptDirectories(t *testing.T) {
	repoRoot := t.TempDir()
	bogusPin := "<!-- zharness:pin deadbeefdeadbeef -->\n"
	for _, sub := range []string{"plans", "decisions", "audit"} {
		dir := filepath.Join(repoRoot, "docs", sub)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir docs/%s: %v", sub, err)
		}
		name := filepath.Join(dir, "record.md")
		body := bogusPin + "\nCites `cli/internal/application/audit.go:1`.\n"
		if err := os.WriteFile(name, []byte(body), 0o644); err != nil {
			t.Fatalf("write docs/%s/record.md: %v", sub, err)
		}
	}

	db := freshDB(t)
	seedRun(t, db)
	report, err := Audit(db, "dev", repoRoot)
	if err != nil {
		t.Fatalf("Audit with pinned exempt-directory docs: %v", err)
	}
	if len(report.ContractViolations) != 0 {
		t.Fatalf("contract_violations = %v, want none: exempt directories are never measured (R4)", report.ContractViolations)
	}
}

// TestDocCitationPatternExtraction locks the citation shape (R24): only
// repository-relative path:line tokens with at least one directory segment
// are citations. Bare filename tokens, globs, versions, and bare times are
// prose — a bare token joined against the repository root would resolve to a
// non-existent path and be misreported as missing.
func TestDocCitationPatternExtraction(t *testing.T) {
	body := "`cli/internal/application/plan_write.go:36` states the rule. " +
		"Also (`interfaces/trace.go:65`) and `docs/README.md:15`.\n" +
		"Bare `trace.go:65`, globs like `docs/plans/active/*.md`, versions like `v0.11.0`, " +
		"and times like 12:30 never match.\n"
	got := docCitationPattern.FindAllString(body, -1)
	want := []string{
		"cli/internal/application/plan_write.go:36",
		"interfaces/trace.go:65",
		"docs/README.md:15",
	}
	if len(got) != len(want) {
		t.Fatalf("citations = %q, want exactly %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("citation[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	if bare := docCitationPattern.FindAllString("see trace.go:65 for details", -1); len(bare) != 0 {
		t.Fatalf("bare filename tokens extracted as citations: %q (R24)", bare)
	}
}

// TestScanPinnedDocsExtractsPinAndCitations checks pin parsing plus unique,
// sorted citation extraction from one eligible top-level document.
func TestScanPinnedDocsExtractsPinAndCitations(t *testing.T) {
	repoRoot := t.TempDir()
	docs := filepath.Join(repoRoot, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	body := "# Guide\n\n<!-- zharness:pin 1a2b3c4d5e6f7a8b9c0d -->\n\n" +
		"Cites `pkg/b.go:2` then `pkg/a.go:1` then `pkg/b.go:2` again.\n"
	if err := os.WriteFile(filepath.Join(docs, "Guide.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write pinned doc: %v", err)
	}

	pins, err := scanPinnedDocs(repoRoot)
	if err != nil {
		t.Fatalf("scanPinnedDocs: %v", err)
	}
	if len(pins) != 1 {
		t.Fatalf("pins = %+v, want one", pins)
	}
	doc := pins[0]
	if doc.Name != "Guide.md" || doc.Pin != "1a2b3c4d5e6f7a8b9c0d" {
		t.Fatalf("doc = %+v, want Guide.md pinned at 1a2b3c4d5e6f7a8b9c0d", doc)
	}
	want := []string{"pkg/a.go:1", "pkg/b.go:2"}
	if len(doc.Citations) != len(want) || doc.Citations[0] != want[0] || doc.Citations[1] != want[1] {
		t.Fatalf("citations = %q, want unique sorted %q", doc.Citations, want)
	}
}

// TestAuditPinDriftReportsMovedCitations builds a real git history: commit a
// cited file, pin a document to that commit, move the file, and expect one
// info-severity finding naming the document, the moved citation, and the line
// delta. Re-pinning to HEAD clears it.
func TestAuditPinDriftReportsMovedCitations(t *testing.T) {
	repoRoot := t.TempDir()
	git := initGitRepoFixture(t, repoRoot)

	if err := os.MkdirAll(filepath.Join(repoRoot, "pkg"), 0o755); err != nil {
		t.Fatalf("mkdir pkg: %v", err)
	}
	source := "one\ntwo\nthree\n"
	if err := os.WriteFile(filepath.Join(repoRoot, "pkg", "thing.go"), []byte(source), 0o644); err != nil {
		t.Fatalf("write thing.go: %v", err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "base")
	baseSha := git("rev-parse", "HEAD")

	docs := filepath.Join(repoRoot, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	guidePath := filepath.Join(docs, "Guide.md")
	guide := "# Guide\n\n<!-- zharness:pin " + baseSha + " -->\n\nSee `pkg/thing.go:1`.\n"
	if err := os.WriteFile(guidePath, []byte(guide), 0o644); err != nil {
		t.Fatalf("write Guide.md: %v", err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "doc")

	if err := os.WriteFile(filepath.Join(repoRoot, "pkg", "thing.go"), []byte("one\nfour\nfive\nsix\n"), 0o644); err != nil {
		t.Fatalf("move thing.go forward: %v", err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "drift")

	db := freshDB(t)
	seedRun(t, db)
	report, err := Audit(db, "dev", repoRoot)
	if err != nil {
		t.Fatalf("Audit with drifted pin: %v", err)
	}
	if len(report.ContractViolations) != 1 {
		t.Fatalf("contract_violations = %+v, want exactly one pin-drift finding", report.ContractViolations)
	}
	finding := report.ContractViolations[0]
	if finding.Identifier != "authored_doc_pin_drift" || finding.Severity != "info" {
		t.Fatalf("finding = %+v, want authored_doc_pin_drift info — distinct identifier and severity from authored_docs_missing warning (R6)", finding)
	}
	if !strings.Contains(finding.Detail, "Guide.md") || !strings.Contains(finding.Detail, "pkg/thing.go:1") {
		t.Fatalf("detail = %q, want document name and moved citation", finding.Detail)
	}
	if !strings.Contains(finding.Detail, "+3/-2") { // one,two,three -> one,four,five,six
		t.Fatalf("detail = %q, want the +3/-2 line delta since the pin", finding.Detail)
	}
	if !strings.Contains(finding.Detail, "freshness signal, not a correctness judgment") {
		t.Fatalf("detail = %q, want freshness-only wording (R9)", finding.Detail)
	}

	headSha := git("rev-parse", "HEAD")
	pinnedToHead := strings.Replace(guide, baseSha, headSha[:len(baseSha)], 1)
	if err := os.WriteFile(guidePath, []byte(pinnedToHead), 0o644); err != nil {
		t.Fatalf("repin Guide.md to HEAD: %v", err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "repin to head")
	report, err = Audit(db, "dev", repoRoot)
	if err != nil {
		t.Fatalf("Audit repinned to HEAD: %v", err)
	}
	if len(report.ContractViolations) != 0 {
		t.Fatalf("contract_violations = %+v, want none once pinned to HEAD", report.ContractViolations)
	}

	statusOut := git("status", "--porcelain")
	if statusOut != "" {
		t.Fatalf("git status after audits = %q, want clean — audit performs no repin and writes nothing (R7)", statusOut)
	}
}

// TestAuditPinDriftMissingPathDistinct cites one live-but-moved path and one
// path that no longer exists; the finding must list the missing citation in
// its own section rather than folding it into the moved set.
func TestAuditPinDriftMissingPathDistinct(t *testing.T) {
	repoRoot := t.TempDir()
	git := initGitRepoFixture(t, repoRoot)

	if err := os.MkdirAll(filepath.Join(repoRoot, "pkg"), 0o755); err != nil {
		t.Fatalf("mkdir pkg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "pkg", "thing.go"), []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatalf("write thing.go: %v", err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "base")
	baseSha := git("rev-parse", "HEAD")

	docs := filepath.Join(repoRoot, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	guide := "# Guide\n\n<!-- zharness:pin " + baseSha + " -->\n\nSee `pkg/thing.go:1` and `pkg/gone.go:1`.\n"
	if err := os.WriteFile(filepath.Join(docs, "Guide.md"), []byte(guide), 0o644); err != nil {
		t.Fatalf("write Guide.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repoRoot, "pkg", "thing.go"), []byte("one\ntwo\nthree\n"), 0o644); err != nil {
		t.Fatalf("move thing.go forward: %v", err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "drift")

	db := freshDB(t)
	seedRun(t, db)
	report, err := Audit(db, "dev", repoRoot)
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(report.ContractViolations) != 1 {
		t.Fatalf("contract_violations = %+v, want one finding covering both citations", report.ContractViolations)
	}
	detail := report.ContractViolations[0].Detail
	movedIdx := strings.Index(detail, "Moved citations:")
	missingIdx := strings.Index(detail, "no longer exists")
	if movedIdx < 0 || missingIdx < 0 {
		t.Fatalf("detail = %q, want distinct moved and missing sections", detail)
	}
	movedSection := detail[movedIdx:missingIdx]
	if !strings.Contains(movedSection, "pkg/thing.go:1") || strings.Contains(movedSection, "pkg/gone.go:1") {
		t.Fatalf("moved section = %q, want gone.go excluded from it", movedSection)
	}
	if !strings.Contains(detail[missingIdx:], "pkg/gone.go:1") {
		t.Fatalf("detail = %q, want gone.go listed as missing", detail)
	}
}

// TestAuditUnresolvablePinDegradesToWarning pins one document validly (its
// cited source moves afterwards) and a second document to a bogus SHA whose
// cited path exists on disk — the review probe's exact failure mode. Audit
// must stay exit-0 clean: no error, the bogus pin yields exactly one
// authored_doc_pin_invalid warning naming document and pin, and the
// validly-pinned neighbor still yields its authored_doc_pin_drift finding
// (R23: degrade, never fail, never mask).
func TestAuditUnresolvablePinDegradesToWarning(t *testing.T) {
	repoRoot := t.TempDir()
	git := initGitRepoFixture(t, repoRoot)

	if err := os.MkdirAll(filepath.Join(repoRoot, "pkg"), 0o755); err != nil {
		t.Fatalf("mkdir pkg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "pkg", "thing.go"), []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatalf("write thing.go: %v", err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "base")
	baseSha := git("rev-parse", "HEAD")

	docs := filepath.Join(repoRoot, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	valid := "# Guide\n\n<!-- zharness:pin " + baseSha + " -->\n\nSee `pkg/thing.go:1`.\n"
	if err := os.WriteFile(filepath.Join(docs, "Guide.md"), []byte(valid), 0o644); err != nil {
		t.Fatalf("write Guide.md: %v", err)
	}
	bogus := "# Broken\n\n<!-- zharness:pin deadbeefdeadbeef -->\n\nSee `pkg/thing.go:1`.\n"
	if err := os.WriteFile(filepath.Join(docs, "Broken.md"), []byte(bogus), 0o644); err != nil {
		t.Fatalf("write Broken.md: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repoRoot, "pkg", "thing.go"), []byte("one\ntwo\nthree\n"), 0o644); err != nil {
		t.Fatalf("move thing.go forward: %v", err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "drift")

	db := freshDB(t)
	seedRun(t, db)
	report, err := Audit(db, "dev", repoRoot)
	if err != nil {
		t.Fatalf("Audit with an unresolvable pin among valid ones: %v (R23: never fail)", err)
	}

	var invalid, drift *AuditFinding
	for i := range report.ContractViolations {
		switch report.ContractViolations[i].Identifier {
		case "authored_doc_pin_invalid":
			invalid = &report.ContractViolations[i]
		case "authored_doc_pin_drift":
			drift = &report.ContractViolations[i]
		}
	}
	if invalid == nil {
		t.Fatalf("contract_violations = %+v, want an authored_doc_pin_invalid warning for the bogus pin (R23)", report.ContractViolations)
	}
	if invalid.Severity != "warning" {
		t.Fatalf("severity = %q, want warning", invalid.Severity)
	}
	if !strings.Contains(invalid.Detail, "Broken.md") || !strings.Contains(invalid.Detail, "deadbeefdeadbeef") {
		t.Fatalf("detail = %q, want document name and unresolvable pin value", invalid.Detail)
	}
	if !strings.Contains(invalid.Detail, "No drift was measured") {
		t.Fatalf("detail = %q, want explicit statement that measurement was skipped", invalid.Detail)
	}
	if drift == nil {
		t.Fatalf("contract_violations = %+v, want the validly-pinned neighbor's drift finding preserved (R23: never mask)", report.ContractViolations)
	}
	if len(report.ContractViolations) != 2 {
		t.Fatalf("contract_violations = %+v, want exactly the two pin findings", report.ContractViolations)
	}
}

// TestMeasureCitationBinaryFileCountsOnePair proves the binary-numstat
// fallback is deterministic (R25): a binary file modified after the pin moves
// the citation and contributes exactly one added and one removed line pair.
func TestMeasureCitationBinaryFileCountsOnePair(t *testing.T) {
	repoRoot := t.TempDir()
	git := initGitRepoFixture(t, repoRoot)

	if err := os.MkdirAll(filepath.Join(repoRoot, "pkg"), 0o755); err != nil {
		t.Fatalf("mkdir pkg: %v", err)
	}
	binaryV1 := []byte{0x00, 0x01, 0x02, 0xff}
	if err := os.WriteFile(filepath.Join(repoRoot, "pkg", "asset.bin"), binaryV1, 0o644); err != nil {
		t.Fatalf("write asset.bin: %v", err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "base")
	baseSha := git("rev-parse", "HEAD")

	binaryV2 := []byte{0x00, 0x01, 0x02, 0xfe, 0xfd}
	if err := os.WriteFile(filepath.Join(repoRoot, "pkg", "asset.bin"), binaryV2, 0o644); err != nil {
		t.Fatalf("rewrite asset.bin: %v", err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "drift")

	drift, err := measureCitation(repoRoot, baseSha, "pkg/asset.bin:1")
	if err != nil {
		t.Fatalf("measureCitation: %v", err)
	}
	if !drift.Moved {
		t.Fatalf("drift = %+v, want Moved for a binary file changed after the pin", drift)
	}
	if drift.LinesAdded != 1 || drift.LinesRemoved != 1 {
		t.Fatalf("lines = +%d/-%d, want the fixed +1/-1 sentinel pair per binary entry (R25)", drift.LinesAdded, drift.LinesRemoved)
	}
}

// TestAuditFailsWhenDocsIsNotADirectory forces a real application.Audit error
// path: `docs` exists as a regular file, so scanPinnedDocs cannot read it as a
// directory and Audit returns an error rather than a finding.
func TestAuditFailsWhenDocsIsNotADirectory(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoRoot, "docs"), []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("write docs file: %v", err)
	}
	db := freshDB(t)
	seedRun(t, db)
	if _, err := Audit(db, "dev", repoRoot); err == nil {
		t.Fatal("Audit with docs-as-file succeeded, want an error from the unreadable docs tree")
	}
}

// TestAuditReportsUnansweredArchitectureFormAsNotDocs scaffolds the R15
// question form into an otherwise docs-less repository and proves both
// behaviors at once: the form is reported unanswered (its own info finding,
// R6) AND it does not satisfy the authored-docs guard's precondition (R15:
// never counted as documentation), so authored_docs_missing still fires.
// Answering the form — replacing the marker with real prose — clears both.
func TestAuditReportsUnansweredArchitectureFormAsNotDocs(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoRoot, "AGENTS.md"), []byte("managed\n"), 0o644); err != nil {
		t.Fatalf("write managed root doc: %v", err)
	}
	docs := filepath.Join(repoRoot, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	form := strings.ReplaceAll(architectureQuestionFormBody, "~", "`")
	if err := os.WriteFile(filepath.Join(docs, "ARCHITECTURE.md"), []byte(form), 0o644); err != nil {
		t.Fatalf("write scaffolded question form: %v", err)
	}

	db := freshDB(t)
	seedRun(t, db)
	report, err := Audit(db, "dev", repoRoot)
	if err != nil {
		t.Fatalf("Audit with the unanswered question form: %v", err)
	}
	if len(report.ContractViolations) != 2 {
		t.Fatalf("contract_violations = %+v, want authored_docs_missing plus architecture_elicitation_unanswered", report.ContractViolations)
	}
	var missing, unanswered *AuditFinding
	for i := range report.ContractViolations {
		switch report.ContractViolations[i].Identifier {
		case "authored_docs_missing":
			missing = &report.ContractViolations[i]
		case "architecture_elicitation_unanswered":
			unanswered = &report.ContractViolations[i]
		}
	}
	if missing == nil || missing.Severity != "warning" {
		t.Fatalf("missing = %+v, want the authored_docs_missing warning — the form is not documentation (R15)", missing)
	}
	if unanswered == nil || unanswered.Severity != "info" {
		t.Fatalf("unanswered = %+v, want the architecture_elicitation_unanswered info finding", unanswered)
	}
	lower := strings.ToLower(unanswered.Detail)
	if strings.Contains(lower, "correct") || strings.Contains(lower, "accurate") {
		t.Fatalf("detail = %q, want no correctness claim about answers (R9)", unanswered.Detail)
	}

	answered := strings.Replace(form,
		"<!-- zharness:unanswered -- `zharness init` scaffolded this form because\ndocs/ARCHITECTURE.md was absent. Answer the five questions below in your own\nwords, then delete this comment; while it remains, `zharness audit` reports\nthis file as an unanswered form rather than documentation. -->",
		"The harness is a CLI that keeps durable agent workflow state in git-tracked\nmarkdown and derives a database from it.", 1)
	if answered == form {
		t.Fatal("marker replacement did not apply; fixture stale")
	}
	if err := os.WriteFile(filepath.Join(docs, "ARCHITECTURE.md"), []byte(answered), 0o644); err != nil {
		t.Fatalf("answer the form: %v", err)
	}
	report, err = Audit(db, "dev", repoRoot)
	if err != nil {
		t.Fatalf("Audit with an answered ARCHITECTURE.md: %v", err)
	}
	if len(report.ContractViolations) != 0 {
		t.Fatalf("contract_violations = %+v, want none once the form is answered", report.ContractViolations)
	}
}
