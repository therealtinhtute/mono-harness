package interfaces

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
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

func listPreflightChangesets(t *testing.T) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(".kit", "changesets", "*.jsonl"))
	if err != nil {
		t.Fatalf("Glob() changesets error = %v", err)
	}
	sort.Strings(matches)
	return matches
}

func listBoundedLifecycleMarkdown(t *testing.T) []string {
	t.Helper()
	var matches []string
	handoffPath := filepath.Join(".kit", "HANDOFF.md")
	if _, err := os.Stat(handoffPath); err == nil {
		matches = append(matches, handoffPath)
	} else if !os.IsNotExist(err) {
		t.Fatalf("Stat() handoff error = %v", err)
	}

	roots := []string{
		filepath.Join(".kit", "planning"),
		filepath.Join(".kit", "runs"),
		filepath.Join(".kit", "reports"),
		filepath.Join("docs", "plans", "active"),
		filepath.Join("docs", "plans", "completed"),
	}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() && filepath.Ext(path) == ".md" {
				matches = append(matches, path)
			}
			return nil
		})
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatalf("WalkDir(%q) error = %v", root, err)
		}
	}
	sort.Strings(matches)
	return matches
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

func TestPreflightCommandWorkAutoRequiresHarnessForActivePlanOnly(t *testing.T) {
	t.Chdir(t.TempDir())
	planPath := filepath.Join("docs", "plans", "active", "initiative.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(planPath, []byte("# active plan\n"), 0o644); err != nil {
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

func TestPreflightCommandWorkAutoUsesReducedWithoutActivePlan(t *testing.T) {
	t.Chdir(t.TempDir())

	out, err := runPreflightCommand(t, "dev", "preflight", "work", "--mode", "auto", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Mode != "reduced" || view.Stop != nil {
		t.Fatalf("view = %+v, want reduced route without an active plan", view)
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
	workPlaybook := filepath.Join(preflightDocsPath, "playbooks", "work.md")
	if err := os.WriteFile(workPlaybook, []byte("# work\n"), 0o644); err != nil {
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
	beforeDB, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("ReadFile() db error = %v", err)
	}
	beforeHash := sha256.Sum256(beforeDB)
	beforeChangesets := listPreflightChangesets(t)
	if beforeMarkdown := listBoundedLifecycleMarkdown(t); len(beforeMarkdown) != 0 {
		t.Fatalf("bounded lifecycle markdown fixture = %v, want none", beforeMarkdown)
	}

	out, err := runPreflightCommand(t, "dev", "preflight", "work", "--mode", "bounded", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Stage != "work" || view.Mode != "reduced" || view.Readiness != application.PreflightReady || view.DB != application.PreflightDBReady || view.Docs != application.PreflightDocsReady {
		t.Fatalf("view = %+v, want reduced/ready work route", view)
	}
	if view.Stop != nil {
		t.Fatalf("view stop = %+v, want none", view.Stop)
	}
	if view.Playbook != workPlaybook {
		t.Fatalf("view playbook = %q, want %q", view.Playbook, workPlaybook)
	}

	afterDB, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("ReadFile() db after preflight error = %v", err)
	}
	afterHash := sha256.Sum256(afterDB)
	if beforeHash != afterHash {
		t.Fatalf("preflight mutated closed db: before=%x after=%x", beforeHash, afterHash)
	}
	afterChangesets := listPreflightChangesets(t)
	if len(beforeChangesets) != len(afterChangesets) {
		t.Fatalf("preflight changeset count = %d, want %d", len(afterChangesets), len(beforeChangesets))
	}
	if !slices.Equal(beforeChangesets, afterChangesets) {
		t.Fatalf("preflight changesets = %v, want %v", afterChangesets, beforeChangesets)
	}
	if afterMarkdown := listBoundedLifecycleMarkdown(t); len(afterMarkdown) != 0 {
		t.Fatalf("preflight created bounded lifecycle markdown = %v", afterMarkdown)
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
