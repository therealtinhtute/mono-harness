package infrastructure

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func freshDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "harness.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, _, err := Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func countRows(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func TestChangesetIdempotent(t *testing.T) {
	db := freshDB(t)
	dir := t.TempDir()

	path, err := WriteChangeset(dir, []ChangesetLine{
		{
			Op:     "create",
			Entity: "story",
			ID:     "01HZZZZZZZZZZZZZZZZZZZZZZ0",
			Fields: map[string]any{
				"slug":       "cli-core",
				"goal":       "working zharness core",
				"status":     "in_progress",
				"created_at": "2026-07-17T12:00:00Z",
			},
			At: "2026-07-17T12:00:00Z",
		},
	})
	if err != nil {
		t.Fatalf("WriteChangeset: %v", err)
	}

	applied, skipped, err := ApplyChangeset(db, path)
	if err != nil {
		t.Fatalf("ApplyChangeset (first): %v", err)
	}
	if applied != 1 || skipped != 0 {
		t.Fatalf("first apply = (applied=%d, skipped=%d), want (1, 0)", applied, skipped)
	}
	if got := countRows(t, db, "stories"); got != 1 {
		t.Fatalf("stories rows after first apply = %d, want 1", got)
	}

	// Re-applying the same file must be a no-op: skipped via the
	// file-level fence, no duplicate row, no error (ULID == fence).
	applied, skipped, err = ApplyChangeset(db, path)
	if err != nil {
		t.Fatalf("ApplyChangeset (second): %v", err)
	}
	if applied != 0 || skipped != 1 {
		t.Fatalf("second apply = (applied=%d, skipped=%d), want (0, 1)", applied, skipped)
	}
	if got := countRows(t, db, "stories"); got != 1 {
		t.Fatalf("stories rows after re-apply = %d, want 1 (duplicate row created)", got)
	}
}

func TestChangesetRejectsUnknownFields(t *testing.T) {
	db := freshDB(t)
	dir := t.TempDir()

	// An unknown field key would, without an allowlist, be spliced
	// directly into SQL identifier position (e.g. `status=(select ...)`).
	// Confirm it's rejected before it ever reaches query text, on both
	// applyCreate and applyUpdate.
	injectKey := `status='x'); ATTACH DATABASE '/tmp/x' AS z; --`

	createPath, err := WriteChangeset(dir, []ChangesetLine{
		{
			Op:     "create",
			Entity: "story",
			ID:     "01HZZZZZZZZZZZZZZZZZZZZZZ1",
			Fields: map[string]any{
				"slug":    "cli-core",
				"goal":    "working zharness core",
				injectKey: "in_progress",
			},
			At: "2026-07-17T12:00:00Z",
		},
	})
	if err != nil {
		t.Fatalf("WriteChangeset (create): %v", err)
	}
	if _, _, err := ApplyChangeset(db, createPath); err == nil {
		t.Fatal("ApplyChangeset (create with unknown field) = nil error, want changeset_malformed")
	}
	if got := countRows(t, db, "stories"); got != 0 {
		t.Fatalf("stories rows after rejected create = %d, want 0", got)
	}

	// Seed a legitimate row, then attempt an update carrying the same
	// unknown field.
	seedPath, err := WriteChangeset(dir, []ChangesetLine{
		{
			Op:     "create",
			Entity: "story",
			ID:     "01HZZZZZZZZZZZZZZZZZZZZZZ2",
			Fields: map[string]any{
				"slug":       "cli-core",
				"goal":       "working zharness core",
				"status":     "planned",
				"created_at": "2026-07-17T12:00:00Z",
			},
			At: "2026-07-17T12:00:00Z",
		},
	})
	if err != nil {
		t.Fatalf("WriteChangeset (seed): %v", err)
	}
	if _, _, err := ApplyChangeset(db, seedPath); err != nil {
		t.Fatalf("ApplyChangeset (seed): %v", err)
	}

	updatePath, err := WriteChangeset(dir, []ChangesetLine{
		{
			Op:     "update",
			Entity: "story",
			ID:     "01HZZZZZZZZZZZZZZZZZZZZZZ2",
			Fields: map[string]any{injectKey: "in_progress"},
			At:     "2026-07-17T12:05:00Z",
		},
	})
	if err != nil {
		t.Fatalf("WriteChangeset (update): %v", err)
	}
	if _, _, err := ApplyChangeset(db, updatePath); err == nil {
		t.Fatal("ApplyChangeset (update with unknown field) = nil error, want changeset_malformed")
	}

	var status string
	if err := db.QueryRow(`SELECT status FROM stories WHERE id = ?`, "01HZZZZZZZZZZZZZZZZZZZZZZ2").Scan(&status); err != nil {
		t.Fatalf("query story status: %v", err)
	}
	if status != "planned" {
		t.Fatalf("story status after rejected update = %q, want unchanged %q", status, "planned")
	}
}

func TestChangesetReplay(t *testing.T) {
	dir := t.TempDir()
	storyID := "01J0000000000000000000STRY"
	runID := "01J0000000000000000000RUNN"

	changesets := [][]ChangesetLine{
		{{
			Op:     "create",
			Entity: "story",
			ID:     storyID,
			Fields: map[string]any{
				"slug":       "cli-core",
				"goal":       "working zharness core",
				"status":     "planned",
				"created_at": "2026-07-17T12:00:00Z",
			},
			At: "2026-07-17T12:00:00Z",
		}},
		{{
			Op:     "update",
			Entity: "story",
			ID:     storyID,
			Fields: map[string]any{
				"status": "in_progress",
			},
			At: "2026-07-17T12:05:00Z",
		}},
		{{
			Op:     "create",
			Entity: "run",
			ID:     runID,
			Fields: map[string]any{
				"story_slug":    "cli-core",
				"trace_ids":     []any{"t1", "t2"},
				"artifact_path": ".kit/runs/work/x.md",
				"created_at":    "2026-07-17T12:10:00Z",
			},
			At: "2026-07-17T12:10:00Z",
		}},
	}

	// Simulate the normal incremental flow: write + apply each changeset
	// against db1 as it's produced.
	db1 := freshDB(t)
	for _, lines := range changesets {
		path, err := WriteChangeset(dir, lines)
		if err != nil {
			t.Fatalf("WriteChangeset: %v", err)
		}
		if _, _, err := ApplyChangeset(db1, path); err != nil {
			t.Fatalf("ApplyChangeset (incremental): %v", err)
		}
	}

	// Full replay: a second, freshly migrated db rebuilt purely from the
	// changeset directory, in ULID order, from empty.
	db2 := freshDB(t)
	totalApplied, err := Replay(db2, dir)
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if totalApplied != 3 {
		t.Fatalf("Replay totalApplied = %d, want 3", totalApplied)
	}

	var status1, status2 string
	if err := db1.QueryRow("SELECT status FROM stories WHERE id = ?", storyID).Scan(&status1); err != nil {
		t.Fatalf("query db1 story status: %v", err)
	}
	if err := db2.QueryRow("SELECT status FROM stories WHERE id = ?", storyID).Scan(&status2); err != nil {
		t.Fatalf("query db2 story status: %v", err)
	}
	if status1 != "in_progress" || status2 != status1 {
		t.Fatalf("story status: db1=%q db2=%q, want both %q", status1, status2, "in_progress")
	}

	var traceIDs1, traceIDs2, artifact1, artifact2 string
	if err := db1.QueryRow("SELECT trace_ids, artifact_path FROM runs WHERE id = ?", runID).Scan(&traceIDs1, &artifact1); err != nil {
		t.Fatalf("query db1 run: %v", err)
	}
	if err := db2.QueryRow("SELECT trace_ids, artifact_path FROM runs WHERE id = ?", runID).Scan(&traceIDs2, &artifact2); err != nil {
		t.Fatalf("query db2 run: %v", err)
	}
	if traceIDs1 != traceIDs2 || artifact1 != artifact2 {
		t.Fatalf("run mismatch: db1=(%q,%q) db2=(%q,%q)", traceIDs1, artifact1, traceIDs2, artifact2)
	}

	if got := countRows(t, db1, "stories"); got != 1 {
		t.Fatalf("db1 stories rows = %d, want 1", got)
	}
	if got := countRows(t, db2, "stories"); got != 1 {
		t.Fatalf("db2 stories rows = %d, want 1", got)
	}
}
