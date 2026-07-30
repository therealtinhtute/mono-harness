package infrastructure

import (
	"bytes"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
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

type enumApplySnapshot struct {
	stories              int
	runs                 int
	checks               int
	storyGoal            string
	storyStatus          string
	checkVerdict         string
	schemaVersion        int
	currentPhase         sql.NullString
	entryPhase           sql.NullString
	latestRunID          sql.NullString
	latestCheckID        sql.NullString
	lastAppliedChangeset sql.NullString
	docsVersion          sql.NullString
}

func takeEnumApplySnapshot(t *testing.T, db *sql.DB, storyID, checkID string) enumApplySnapshot {
	t.Helper()

	snapshot := enumApplySnapshot{
		stories: countRows(t, db, "stories"),
		runs:    countRows(t, db, "runs"),
		checks:  countRows(t, db, "checks"),
	}
	if err := db.QueryRow(`SELECT goal, status FROM stories WHERE id = ?`, storyID).Scan(&snapshot.storyGoal, &snapshot.storyStatus); err != nil {
		t.Fatalf("query story fields: %v", err)
	}
	if err := db.QueryRow(`SELECT verdict FROM checks WHERE id = ?`, checkID).Scan(&snapshot.checkVerdict); err != nil {
		t.Fatalf("query check verdict: %v", err)
	}
	if err := db.QueryRow(`
		SELECT schema_version, current_phase, entry_phase, latest_run_id, latest_check_id,
			last_applied_changeset, docs_version
		FROM meta
		LIMIT 1
	`).Scan(
		&snapshot.schemaVersion,
		&snapshot.currentPhase,
		&snapshot.entryPhase,
		&snapshot.latestRunID,
		&snapshot.latestCheckID,
		&snapshot.lastAppliedChangeset,
		&snapshot.docsVersion,
	); err != nil {
		t.Fatalf("query meta snapshot: %v", err)
	}
	return snapshot
}

func seedEnumApplyFixture(t *testing.T, db *sql.DB, dir string) (storyID, runID, checkID string) {
	t.Helper()

	storyID = ulid.Make().String()
	runID = ulid.Make().String()
	checkID = ulid.Make().String()
	at := "2026-07-28T00:00:00Z"
	path, err := WriteChangeset(dir, []ChangesetLine{
		{Op: "create", Entity: "story", ID: storyID, Fields: map[string]any{
			"slug": "enum-apply", "goal": "validate persisted enums", "status": domain.StoryPlanned, "created_at": at,
		}, At: at},
		{Op: "create", Entity: "run", ID: runID, Fields: map[string]any{
			"story_slug": "enum-apply", "trace_ids": []any{}, "artifact_path": "", "created_at": at,
		}, At: at},
		{Op: "create", Entity: "check", ID: checkID, Fields: map[string]any{
			"run_id": runID, "verdict": domain.VerdictApproved, "proof_links": []any{}, "created_at": at,
		}, At: at},
	})
	if err != nil {
		t.Fatalf("WriteChangeset(seed): %v", err)
	}
	if _, _, err := ApplyChangeset(db, path); err != nil {
		t.Fatalf("ApplyChangeset(seed): %v", err)
	}
	return storyID, runID, checkID
}

func TestChangesetRejectsInvalidLifecycleEnumsWithoutMutation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		entity string
		table  string
		op     string
		column string
	}{
		{name: "story create", entity: "story", table: "stories", op: "create", column: "status"},
		{name: "story update", entity: "story", table: "stories", op: "update", column: "status"},
		{name: "check create", entity: "check", table: "checks", op: "create", column: "verdict"},
		{name: "check update", entity: "check", table: "checks", op: "update", column: "verdict"},
		{name: "check create judge", entity: "check", table: "checks", op: "create", column: "judge"},
		{name: "check update judge", entity: "check", table: "checks", op: "update", column: "judge"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := freshDB(t)
			dir := t.TempDir()
			storyID, runID, checkID := seedEnumApplyFixture(t, db, dir)

			line := ChangesetLine{Op: tc.op, Entity: tc.entity, At: "2026-07-28T00:05:00Z"}
			switch tc.entity {
			case "story":
				line.ID = storyID
				line.Fields = map[string]any{"status": "bogus"}
				if tc.op == "create" {
					line.ID = ulid.Make().String()
					line.Fields["slug"] = "invalid-story-enum"
					line.Fields["goal"] = "must not persist"
					line.Fields["created_at"] = line.At
				}
			case "check":
				line.ID = checkID
				line.Fields = map[string]any{tc.column: "bogus"}
				if tc.column != "verdict" {
					line.Fields["verdict"] = domain.VerdictApproved
				}
				if tc.op == "create" {
					line.ID = ulid.Make().String()
					line.Fields["run_id"] = runID
					line.Fields["proof_links"] = []any{}
					line.Fields["created_at"] = line.At
				}
			}

			path, err := WriteChangeset(dir, []ChangesetLine{
				{Op: "update", Entity: "story", ID: storyID, Fields: map[string]any{"goal": "must not persist"}, At: line.At},
				{Op: "update", Entity: "meta", Fields: map[string]any{"current_phase": "enum-apply"}, At: line.At},
				line,
			})
			if err != nil {
				t.Fatalf("WriteChangeset(invalid): %v", err)
			}
			before := takeEnumApplySnapshot(t, db, storyID, checkID)
			if changesetULID(path) <= before.lastAppliedChangeset.String {
				t.Fatalf("invalid changeset %s is not above fence %s", changesetULID(path), before.lastAppliedChangeset.String)
			}

			applied, skipped, err := ApplyChangeset(db, path)
			wantError := "changeset_malformed: invalid " + tc.table + "." + tc.column + " value \"bogus\""
			if err == nil || err.Error() != wantError {
				t.Fatalf("ApplyChangeset error = %v, want %q", err, wantError)
			}
			if applied != 0 || skipped != 0 {
				t.Fatalf("ApplyChangeset = (applied=%d, skipped=%d), want (0, 0)", applied, skipped)
			}
			after := takeEnumApplySnapshot(t, db, storyID, checkID)
			if before != after {
				t.Fatalf("database changed after rejected enum:\nbefore = %+v\nafter  = %+v", before, after)
			}

			pending, appliedCount, fence, err := ChangesetStatus(db, dir)
			if err != nil {
				t.Fatalf("ChangesetStatus: %v", err)
			}
			if len(pending) != 1 || pending[0] != filepath.Base(path) {
				t.Fatalf("pending = %v, want [%s]", pending, filepath.Base(path))
			}
			if appliedCount != 1 || fence != before.lastAppliedChangeset.String {
				t.Fatalf("status = (applied=%d, fence=%q), want (1, %q)", appliedCount, fence, before.lastAppliedChangeset.String)
			}
		})
	}
}

func TestChangesetAcceptsLifecycleEnums(t *testing.T) {
	db := freshDB(t)
	dir := t.TempDir()
	at := "2026-07-28T01:00:00Z"
	statuses := []string{domain.StoryPlanned, domain.StoryInProgress, domain.StoryChecked, domain.StoryDone}
	verdicts := []string{domain.VerdictApproved, domain.VerdictApproveWithRequests, domain.VerdictRequestChanges}
	judges := []string{domain.JudgeIndependent, domain.JudgeSameSession}

	storyIDs := make([]string, len(statuses))
	createLines := make([]ChangesetLine, 0, len(statuses)+len(verdicts)+1)
	for i, status := range statuses {
		storyIDs[i] = ulid.Make().String()
		createLines = append(createLines, ChangesetLine{Op: "create", Entity: "story", ID: storyIDs[i], Fields: map[string]any{
			"slug": "valid-status-" + status, "goal": "accept valid status", "status": status, "created_at": at,
		}, At: at})
	}
	runID := ulid.Make().String()
	createLines = append(createLines, ChangesetLine{Op: "create", Entity: "run", ID: runID, Fields: map[string]any{
		"story_slug": "valid-status-planned", "trace_ids": []any{}, "artifact_path": "", "created_at": at,
	}, At: at})
	checkIDs := make([]string, len(verdicts))
	for i, verdict := range verdicts {
		checkIDs[i] = ulid.Make().String()
		createLines = append(createLines, ChangesetLine{Op: "create", Entity: "check", ID: checkIDs[i], Fields: map[string]any{
			"run_id": runID, "verdict": verdict, "judge": judges[i%len(judges)], "judge_model": "test-model", "proof_links": []any{}, "created_at": at,
		}, At: at})
	}

	createPath, err := WriteChangeset(dir, createLines)
	if err != nil {
		t.Fatalf("WriteChangeset(create enums): %v", err)
	}
	if applied, _, err := ApplyChangeset(db, createPath); err != nil || applied != len(createLines) {
		t.Fatalf("ApplyChangeset(create enums) = (applied=%d, err=%v), want (%d, nil)", applied, err, len(createLines))
	}

	updateLines := make([]ChangesetLine, 0, len(statuses)+len(verdicts))
	for _, status := range statuses {
		updateLines = append(updateLines, ChangesetLine{Op: "update", Entity: "story", ID: storyIDs[0], Fields: map[string]any{"status": status}, At: at})
	}
	for _, verdict := range verdicts {
		updateLines = append(updateLines, ChangesetLine{Op: "update", Entity: "check", ID: checkIDs[0], Fields: map[string]any{"verdict": verdict}, At: at})
	}
	for _, judge := range judges {
		updateLines = append(updateLines, ChangesetLine{Op: "update", Entity: "check", ID: checkIDs[0], Fields: map[string]any{"judge": judge}, At: at})
	}
	updatePath, err := WriteChangeset(dir, updateLines)
	if err != nil {
		t.Fatalf("WriteChangeset(update enums): %v", err)
	}
	if applied, _, err := ApplyChangeset(db, updatePath); err != nil || applied != len(updateLines) {
		t.Fatalf("ApplyChangeset(update enums) = (applied=%d, err=%v), want (%d, nil)", applied, err, len(updateLines))
	}

	var status, verdict, judge string
	if err := db.QueryRow(`SELECT status FROM stories WHERE id = ?`, storyIDs[0]).Scan(&status); err != nil {
		t.Fatalf("query final story status: %v", err)
	}
	if err := db.QueryRow(`SELECT verdict, judge FROM checks WHERE id = ?`, checkIDs[0]).Scan(&verdict, &judge); err != nil {
		t.Fatalf("query final check verdict: %v", err)
	}
	if status != domain.StoryDone || verdict != domain.VerdictRequestChanges || judge != domain.JudgeSameSession {
		t.Fatalf("final enums = (%q, %q, %q), want (%q, %q, %q)", status, verdict, judge, domain.StoryDone, domain.VerdictRequestChanges, domain.JudgeSameSession)
	}
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
				"status":     domain.StoryInProgress,
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
				injectKey: domain.StoryInProgress,
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
			Fields: map[string]any{injectKey: domain.StoryInProgress},
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

func TestNextChangesetIDAdvancesPastCanonicalFloorWithoutSleeping(t *testing.T) {
	dir := t.TempDir()
	entropy := bytes.NewReader(bytes.Repeat([]byte{0xff}, 10))
	existing, err := ulid.New(ulid.MaxTime()-1, entropy)
	if err != nil {
		t.Fatalf("construct future ULID: %v", err)
	}
	if _, err := WriteChangesetWithID(dir, existing.String(), []ChangesetLine{{Op: "update", Entity: "meta", Fields: map[string]any{}, At: "2026-07-27T00:00:00Z"}}); err != nil {
		t.Fatalf("WriteChangesetWithID: %v", err)
	}

	next, err := NextChangesetID(dir, "01ARZ3NDEKTSV4RRFFQ69G5FAV")
	if err != nil {
		t.Fatalf("NextChangesetID: %v", err)
	}
	if next <= existing.String() {
		t.Fatalf("next id = %s, want above existing floor %s", next, existing)
	}
	parsed, err := ulid.ParseStrict(next)
	if err != nil || parsed.String() != next {
		t.Fatalf("next id = %q, want canonical ULID (err=%v)", next, err)
	}
}

func TestNextChangesetIDReportsULIDOverflow(t *testing.T) {
	var maximum ulid.ULID
	for i := range maximum {
		maximum[i] = 0xff
	}
	_, err := NextChangesetID(t.TempDir(), maximum.String())
	var overflow *ChangesetIDOverflowError
	if !errors.As(err, &overflow) {
		t.Fatalf("NextChangesetID error = %v, want ChangesetIDOverflowError", err)
	}
}

func TestApplyChangesetAcquiresWriteReservationBeforeFenceRead(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "harness.db")
	first, err := Open(path)
	if err != nil {
		t.Fatalf("Open first: %v", err)
	}
	defer first.Close()
	if _, _, err := Migrate(first); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	second, err := Open(path)
	if err != nil {
		t.Fatalf("Open second: %v", err)
	}
	defer second.Close()

	changeset, err := WriteChangeset(t.TempDir(), []ChangesetLine{{Op: "update", Entity: "meta", Fields: map[string]any{"docs_version": "blocked"}, At: "2026-07-27T00:00:00Z"}})
	if err != nil {
		t.Fatalf("WriteChangeset: %v", err)
	}
	holder, err := first.Begin()
	if err != nil {
		t.Fatalf("Begin holder: %v", err)
	}
	if _, err := holder.Exec(`UPDATE meta SET last_applied_changeset = last_applied_changeset`); err != nil {
		holder.Rollback()
		t.Fatalf("reserve writer: %v", err)
	}

	_, _, err = ApplyChangeset(second, changeset)
	if err == nil || !strings.Contains(err.Error(), "lock apply tx") {
		holder.Rollback()
		t.Fatalf("concurrent ApplyChangeset error = %v, want failure before fence read", err)
	}
	var fence sql.NullString
	if err := holder.QueryRow(`SELECT last_applied_changeset FROM meta`).Scan(&fence); err != nil {
		holder.Rollback()
		t.Fatalf("read held fence: %v", err)
	}
	if fence.Valid {
		holder.Rollback()
		t.Fatalf("fence changed while writer reservation held: %q", fence.String)
	}
	if err := holder.Rollback(); err != nil {
		t.Fatalf("rollback holder: %v", err)
	}
	if _, _, err := ApplyChangeset(second, changeset); err != nil {
		t.Fatalf("ApplyChangeset after release: %v", err)
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
				"status": domain.StoryInProgress,
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
	if status1 != domain.StoryInProgress || status2 != status1 {
		t.Fatalf("story status: db1=%q db2=%q, want both %q", status1, status2, domain.StoryInProgress)
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
