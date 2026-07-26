package interfaces

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func runPreflightCommand(t *testing.T, version string, args ...string) (string, error) {
	t.Helper()
	jsonOutput = false
	cmd := NewRootCmd(version)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), err
}

func decodePreflight(t *testing.T, output string) application.PreflightView {
	t.Helper()
	var view application.PreflightView
	if err := json.Unmarshal([]byte(output), &view); err != nil {
		t.Fatalf("unmarshal preflight output %q: %v", output, err)
	}
	return view
}

func TestPreflightCommandReducedWithoutHarness(t *testing.T) {
	t.Chdir(t.TempDir())

	out, err := runPreflightCommand(t, "dev", "preflight", "work", "--mode", "simple", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Readiness != application.PreflightReduced || view.DB != application.PreflightDBMissing || view.Docs != application.PreflightDocsMissing {
		t.Fatalf("view = %+v", view)
	}
	if view.Playbook != "" {
		t.Fatalf("view playbook = %q, want empty while docs are missing", view.Playbook)
	}
}

func TestPreflightCommandTreatsMissingStagePlaybookAsMissingDocs(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(preflightDocsPath, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	out, err := runPreflightCommand(t, "dev", "preflight", "work", "--mode", "simple", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Docs != application.PreflightDocsMissing || view.Playbook != "" {
		t.Fatalf("view = %+v, want missing docs and no playbook", view)
	}
}

func TestPreflightCommandWorkAutoRequiresHarnessWhenSpecExists(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(filepath.Dir(preflightSpecPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(preflightSpecPath, []byte("# locked spec\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	out, err := runPreflightCommand(t, "dev", "preflight", "work", "--mode", "auto", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Mode != "durable" || view.Stop == nil || view.Stop.Code != "harness_required" {
		t.Fatalf("view = %+v, want durable harness_required stop", view)
	}
}

func TestPreflightCommandBlocksDurableWithoutHarness(t *testing.T) {
	t.Chdir(t.TempDir())

	out, err := runPreflightCommand(t, "dev", "preflight", "to-plan", "--mode", "full", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Stop == nil || view.Stop.Code != "harness_required" {
		t.Fatalf("view stop = %+v, want harness_required", view.Stop)
	}
}

func TestPreflightCommandReadsReadyStateWithoutMutation(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.MkdirAll(filepath.Join(preflightDocsPath, "playbooks"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(preflightDocsPath, "playbooks", "watzup.md"), []byte("# watzup\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, _, err := infrastructure.Migrate(db); err != nil {
		db.Close()
		t.Fatalf("Migrate() error = %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	before, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	out, err := runPreflightCommand(t, "dev", "preflight", "watzup", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Readiness != application.PreflightReady || view.DB != application.PreflightDBReady || view.Docs != application.PreflightDocsReady {
		t.Fatalf("view = %+v", view)
	}
	after, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("Stat() after preflight error = %v", err)
	}
	if before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		t.Fatalf("preflight mutated db metadata: before=%+v after=%+v", before, after)
	}
	if matches, err := filepath.Glob(filepath.Join(".kit", "changesets", "*.jsonl")); err != nil || len(matches) != 0 {
		t.Fatalf("preflight changesets = %v, err = %v", matches, err)
	}
}

func TestPreflightCommandReportsStaleDocs(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(filepath.Join(preflightDocsPath, "playbooks"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(preflightDocsPath, "playbooks", "check.md"), []byte("# check\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, _, err := infrastructure.Migrate(db); err != nil {
		db.Close()
		t.Fatalf("Migrate() error = %v", err)
	}
	if _, err := db.Exec(`UPDATE meta SET docs_version = '0.5.0'`); err != nil {
		db.Close()
		t.Fatalf("set docs version: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	out, err := runPreflightCommand(t, "0.6.0", "preflight", "check", "--mode", "full", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Docs != application.PreflightDocsStale || view.Stop == nil || view.Stop.Code != "stale_docs" {
		t.Fatalf("view = %+v", view)
	}
}
