package application

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func TestMigrateLayoutDryRunThenApply(t *testing.T) {
	root := t.TempDir()
	wantRunID, wantCheckID := seedLegacyLayout(t, root)
	legacyDB, err := infrastructure.Open(filepath.Join(root, ".kit", "harness.db"))
	if err != nil {
		t.Fatalf("Open(legacy) error = %v", err)
	}
	if _, err := legacyDB.Exec(`INSERT INTO traces (id, run_id, wave, summary, created_at) VALUES (?, ?, ?, ?, ?)`, "01JYYYYYYYYYYYYYYYYYYYYYYY", wantRunID, 9, "legacy-only trace", "2026-07-26T00:00:00Z"); err != nil {
		legacyDB.Close()
		t.Fatalf("insert legacy-only trace: %v", err)
	}
	if err := legacyDB.Close(); err != nil {
		t.Fatalf("close legacy db: %v", err)
	}

	dry, err := MigrateLayout(root, ".kit/harness.db", "harness.db", ".kit/changesets", ".kit", fixtureDocsFS, "0.6.0", true)
	if err != nil {
		t.Fatalf("MigrateLayout(dry-run) error = %v", err)
	}
	if dry.Status != "dry-run" || !dry.Parity || dry.Replayed == 0 || dry.Backfilled == 0 {
		t.Fatalf("dry-run result = %+v", dry)
	}
	if _, err := os.Stat(filepath.Join(root, "harness.db")); !os.IsNotExist(err) {
		t.Fatal("dry-run created root harness.db")
	}
	if _, err := os.Stat(filepath.Join(root, ".kit", "harness.db")); err != nil {
		t.Fatalf("dry-run removed legacy db: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".kit", "cache")); !os.IsNotExist(err) {
		t.Fatal("dry-run created repository cache state")
	}

	result, err := MigrateLayout(root, ".kit/harness.db", "harness.db", ".kit/changesets", ".kit", fixtureDocsFS, "0.6.0", false)
	if err != nil {
		t.Fatalf("MigrateLayout() error = %v", err)
	}
	if result.Status != "migrated" || !result.Parity || !result.DocsWritten {
		t.Fatalf("migration result = %+v", result)
	}
	if _, err := os.Stat(filepath.Join(root, ".kit", "harness.db")); !os.IsNotExist(err) {
		t.Fatal("legacy database still exists after migration")
	}
	if _, err := os.Stat(filepath.Join(root, "harness.db")); err != nil {
		t.Fatalf("root database missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "WORKFLOW.md")); err != nil {
		t.Fatalf("root WORKFLOW.md missing: %v", err)
	}

	db, err := infrastructure.OpenReadOnly(filepath.Join(root, "harness.db"))
	if err != nil {
		t.Fatalf("OpenReadOnly(root db) error = %v", err)
	}
	defer db.Close()
	state, err := QueryState(db.Raw())
	if err != nil {
		t.Fatalf("QueryState() error = %v", err)
	}
	if state.LatestRunID == nil || *state.LatestRunID != wantRunID || state.LatestCheckID == nil || *state.LatestCheckID != wantCheckID {
		t.Fatalf("migrated pointers = %+v, want run=%s check=%s", state, wantRunID, wantCheckID)
	}
	var managedCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM managed_docs`).Scan(&managedCount); err != nil {
		t.Fatalf("count managed_docs: %v", err)
	}
	if managedCount != 2 {
		t.Fatalf("managed docs count = %d, want 2", managedCount)
	}
}

func TestMigrateLayoutDocsConflictRollsBackActivation(t *testing.T) {
	root := t.TempDir()
	seedLegacyLayout(t, root)
	workflowPath := filepath.Join(root, "docs", "WORKFLOW.md")
	if err := os.MkdirAll(filepath.Dir(workflowPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	local := []byte("# Consumer-owned workflow\n")
	if err := os.WriteFile(workflowPath, local, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := MigrateLayout(root, ".kit/harness.db", "harness.db", ".kit/changesets", ".kit", fixtureDocsFS, "0.6.0", false)
	var conflict *ManagedDocsConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("MigrateLayout() error = %v, want ManagedDocsConflictError", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".kit", "harness.db")); err != nil {
		t.Fatalf("legacy db missing after conflict: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "harness.db")); !os.IsNotExist(err) {
		t.Fatal("root db activated despite docs conflict")
	}
	got, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read local workflow: %v", err)
	}
	if string(got) != string(local) {
		t.Fatalf("local workflow changed: %q", got)
	}
	if _, err := os.Stat(filepath.Join(root, ".kit", "conflicts", "docs", "WORKFLOW.md.upstream")); err != nil {
		t.Fatalf("staged upstream conflict missing: %v", err)
	}
}

func seedLegacyLayout(t *testing.T, root string) (runID, checkID string) {
	t.Helper()
	legacyPath := filepath.Join(root, ".kit", "harness.db")
	changesets := filepath.Join(root, ".kit", "changesets")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	db, err := infrastructure.Open(legacyPath)
	if err != nil {
		t.Fatalf("Open(legacy) error = %v", err)
	}
	if _, _, err := infrastructure.Migrate(db); err != nil {
		db.Close()
		t.Fatalf("Migrate(legacy) error = %v", err)
	}
	if _, _, err := CreateStory(db, changesets, "layout", "migrate layout", ""); err != nil {
		db.Close()
		t.Fatalf("CreateStory() error = %v", err)
	}
	artifactPath := filepath.Join(root, "run.md")
	if err := os.WriteFile(artifactPath, []byte("run"), 0o644); err != nil {
		db.Close()
		t.Fatalf("write run artifact: %v", err)
	}
	runID, _, err = CreateRun(db, changesets, "layout", artifactPath, "")
	if err != nil {
		db.Close()
		t.Fatalf("CreateRun() error = %v", err)
	}
	checkID, _, err = RecordCheck(db, changesets, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{{Command: "true", ArtifactPath: artifactPath}})
	if err != nil {
		db.Close()
		t.Fatalf("RecordCheck() error = %v", err)
	}
	setMeta(t, db, changesets, map[string]any{"current_phase": "layout", "entry_phase": "layout"})
	if err := db.Close(); err != nil {
		t.Fatalf("close legacy db: %v", err)
	}
	return runID, checkID
}
