package application

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// freshDB opens a fresh, migrated db plus an adjacent (not-yet-existing)
// changeset dir, both scoped to a per-test temp dir.
func freshDB(t *testing.T) (db *sql.DB, changesetDir string) {
	t.Helper()
	root := t.TempDir()
	db, err := infrastructure.Open(filepath.Join(root, "harness.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, _, err := infrastructure.Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, filepath.Join(root, ".kit", "changesets")
}

func countRows(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

// assertChangesetBeforeRow proves the changeset-first invariant for a
// single-line create: the returned path names a changeset file whose sole
// line already carries id/entity, and only after that does the db row for
// id show up in table.
func assertChangesetBeforeRow(t *testing.T, db *sql.DB, path, table, id, entity string) {
	t.Helper()
	if path == "" {
		t.Fatal("changeset path is empty, want a written .jsonl file")
	}
	lines, err := infrastructure.ReadChangeset(path)
	if err != nil {
		t.Fatalf("ReadChangeset(%s): %v", path, err)
	}
	if len(lines) != 1 {
		t.Fatalf("changeset %s has %d lines, want 1", path, len(lines))
	}
	if lines[0].ID != id || lines[0].Entity != entity || lines[0].Op != "create" {
		t.Fatalf("changeset line = %+v, want id=%s entity=%s op=create", lines[0], id, entity)
	}

	var found string
	if err := db.QueryRow("SELECT id FROM "+table+" WHERE id = ?", id).Scan(&found); err != nil {
		t.Fatalf("row for %s not found in %s after apply: %v", id, table, err)
	}
}

// seedRun writes a story + run row directly (bypassing domain validation
// and the Create* application functions under test) so FK-dependent
// fixtures (checks.run_id, traces.run_id) have something real to point
// at. FKs are enforced (PRAGMA foreign_keys=ON, see store.go), so a bare
// literal ID would fail at apply time.
func seedRun(t *testing.T, db *sql.DB, changesetDir string) (runID string) {
	t.Helper()
	at := "2026-07-17T12:00:00Z"
	storyID := ulid.Make().String()
	if _, _, err := AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "create", Entity: "story", ID: storyID, Fields: map[string]any{
			"slug": "cli-domain", "goal": "ported domain commands work", "status": "planned", "created_at": at,
		}, At: at},
	}); err != nil {
		t.Fatalf("seed story: %v", err)
	}

	runID = ulid.Make().String()
	if _, _, err := AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "create", Entity: "run", ID: runID, Fields: map[string]any{
			"story_slug": "cli-domain", "artifact_path": ".kit/runs/work/x.md", "created_at": at,
		}, At: at},
	}); err != nil {
		t.Fatalf("seed run: %v", err)
	}
	return runID
}

// seedCheck extends seedRun with a checks row, for intervention fixtures.
func seedCheck(t *testing.T, db *sql.DB, changesetDir string) (checkID string) {
	t.Helper()
	runID := seedRun(t, db, changesetDir)
	at := "2026-07-17T12:05:00Z"
	checkID = ulid.Make().String()
	if _, _, err := AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "create", Entity: "check", ID: checkID, Fields: map[string]any{
			"run_id": runID, "verdict": "APPROVED", "created_at": at,
		}, At: at},
	}); err != nil {
		t.Fatalf("seed check: %v", err)
	}
	return checkID
}
