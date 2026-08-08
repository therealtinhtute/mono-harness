package application

import (
	"database/sql"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func TestPendingChangesetBlocksOrdinaryMutationUntilRecovery(t *testing.T) {
	db, dir := freshDB(t)
	storyID := ulid.Make().String()
	initialPath, _, err := AppendAndApply(db, dir, []infrastructure.ChangesetLine{{
		Op: "create", Entity: "story", ID: storyID,
		Fields: map[string]any{"slug": "recovery", "goal": "before recovery", "status": domain.StoryPlanned, "created_at": "2026-07-27T00:00:00Z"},
		At:     "2026-07-27T00:00:00Z",
	}})
	if err != nil {
		t.Fatalf("seed story: %v", err)
	}
	_, _, fence, _, err := infrastructure.ChangesetStatus(db, dir)
	if err != nil {
		t.Fatalf("ChangesetStatus before pending: %v", err)
	}
	pendingPath, err := infrastructure.WriteChangesetAbove(dir, fence, []infrastructure.ChangesetLine{{
		Op: "update", Entity: "story", ID: storyID,
		Fields: map[string]any{"goal": "recovered"},
		At:     "2026-07-27T00:01:00Z",
	}})
	if err != nil {
		t.Fatalf("write pending changeset: %v", err)
	}

	beforeFiles, err := infrastructure.ListChangesets(dir)
	if err != nil {
		t.Fatalf("ListChangesets before blocked mutation: %v", err)
	}
	before := recoverySnapshotForTest(t, db, storyID)
	_, _, err = CreateIntake(db, dir, "maintenance", "blocked while recovery is pending", "tiny", "", "")
	validation, ok := err.(*domain.ValidationError)
	if !ok || validation.Code != "changeset_recovery_required" || validation.Message == "" {
		t.Fatalf("CreateIntake error = %T %v, want changeset_recovery_required", err, err)
	}
	afterFiles, err := infrastructure.ListChangesets(dir)
	if err != nil {
		t.Fatalf("ListChangesets after blocked mutation: %v", err)
	}
	if !reflect.DeepEqual(afterFiles, beforeFiles) {
		t.Fatalf("blocked mutation changed files: before=%v after=%v", beforeFiles, afterFiles)
	}
	if after := recoverySnapshotForTest(t, db, storyID); after != before {
		t.Fatalf("blocked mutation changed database: before=%+v after=%+v", before, after)
	}

	if _, _, err := ApplyChangesetForRecovery(db, dir, pendingPath); err != nil {
		t.Fatalf("apply earliest pending changeset: %v", err)
	}
	if _, _, err := CreateIntake(db, dir, "maintenance", "continues after recovery", "tiny", "", ""); err != nil {
		t.Fatalf("CreateIntake after recovery: %v", err)
	}
	live := recoverySnapshotForTest(t, db, storyID)
	if live.goal != "recovered" || live.intakes != 1 {
		t.Fatalf("live state after recovery = %+v", live)
	}

	replayDB := freshReplayDBForTest(t)
	if _, err := infrastructure.Replay(replayDB, dir); err != nil {
		t.Fatalf("Replay: %v", err)
	}
	replayed := recoverySnapshotForTest(t, replayDB, storyID)
	if replayed != live {
		t.Fatalf("replay state differs: live=%+v replay=%+v (initial=%s pending=%s)", live, replayed, filepath.Base(initialPath), filepath.Base(pendingPath))
	}
}

func TestDirectApplyRejectsExternalSameBasename(t *testing.T) {
	db, dir := freshDB(t)
	storyID := ulid.Make().String()
	if _, _, err := AppendAndApply(db, dir, []infrastructure.ChangesetLine{{
		Op: "create", Entity: "story", ID: storyID,
		Fields: map[string]any{"slug": "recovery-identity", "goal": "initial", "status": domain.StoryPlanned, "created_at": "2026-07-27T00:00:00Z"},
		At:     "2026-07-27T00:00:00Z",
	}}); err != nil {
		t.Fatalf("seed story: %v", err)
	}
	trackedPath, err := infrastructure.WriteChangeset(dir, []infrastructure.ChangesetLine{{
		Op: "update", Entity: "story", ID: storyID,
		Fields: map[string]any{"goal": "A"},
		At:     "2026-07-27T00:01:00Z",
	}})
	if err != nil {
		t.Fatalf("write tracked pending changeset: %v", err)
	}
	trackedID := strings.TrimSuffix(filepath.Base(trackedPath), ".changeset.jsonl")
	impostorPath, err := infrastructure.WriteChangesetWithID(t.TempDir(), trackedID, []infrastructure.ChangesetLine{{
		Op: "update", Entity: "story", ID: storyID,
		Fields: map[string]any{"goal": "B"},
		At:     "2026-07-27T00:01:00Z",
	}})
	if err != nil {
		t.Fatalf("write external impostor: %v", err)
	}

	pendingBefore, appliedBefore, fenceBefore, _, err := infrastructure.ChangesetStatus(db, dir)
	if err != nil {
		t.Fatalf("ChangesetStatus before impostor: %v", err)
	}
	stateBefore := recoverySnapshotForTest(t, db, storyID)
	_, _, err = ApplyChangesetForRecovery(db, dir, impostorPath)
	validation, ok := err.(*domain.ValidationError)
	if !ok || validation.Code != "changeset_recovery_required" {
		t.Fatalf("apply impostor error = %T %v, want changeset_recovery_required", err, err)
	}
	if stateAfter := recoverySnapshotForTest(t, db, storyID); stateAfter != stateBefore {
		t.Fatalf("impostor changed database: before=%+v after=%+v", stateBefore, stateAfter)
	}
	pendingAfter, appliedAfter, fenceAfter, _, err := infrastructure.ChangesetStatus(db, dir)
	if err != nil {
		t.Fatalf("ChangesetStatus after impostor: %v", err)
	}
	if !reflect.DeepEqual(pendingAfter, pendingBefore) || appliedAfter != appliedBefore || fenceAfter != fenceBefore {
		t.Fatalf("impostor changed status: before=(%v, %d, %q) after=(%v, %d, %q)", pendingBefore, appliedBefore, fenceBefore, pendingAfter, appliedAfter, fenceAfter)
	}

	if _, _, err := ApplyChangesetForRecovery(db, dir, trackedPath); err != nil {
		t.Fatalf("apply tracked pending changeset: %v", err)
	}
	live := recoverySnapshotForTest(t, db, storyID)
	if live.goal != "A" {
		t.Fatalf("live goal = %q, want A", live.goal)
	}

	replayDB := freshReplayDBForTest(t)
	if _, err := infrastructure.Replay(replayDB, dir); err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if replayed := recoverySnapshotForTest(t, replayDB, storyID); replayed != live {
		t.Fatalf("replay state differs: live=%+v replay=%+v", live, replayed)
	}
}

func TestDirectApplyRejectsLaterPendingUntilEarliestApplied(t *testing.T) {
	db, dir := freshDB(t)
	storyID := ulid.Make().String()
	first, err := infrastructure.WriteChangeset(dir, []infrastructure.ChangesetLine{{
		Op: "create", Entity: "story", ID: storyID,
		Fields: map[string]any{"slug": "ordered-recovery", "goal": "first", "status": domain.StoryPlanned, "created_at": "2026-07-27T00:00:00Z"},
		At:     "2026-07-27T00:00:00Z",
	}})
	if err != nil {
		t.Fatalf("write first pending: %v", err)
	}
	second, err := infrastructure.WriteChangeset(dir, []infrastructure.ChangesetLine{{
		Op: "update", Entity: "story", ID: storyID,
		Fields: map[string]any{"goal": "second"},
		At:     "2026-07-27T00:01:00Z",
	}})
	if err != nil {
		t.Fatalf("write second pending: %v", err)
	}

	_, _, err = ApplyChangesetForRecovery(db, dir, second)
	validation, ok := err.(*domain.ValidationError)
	if !ok || validation.Code != "changeset_recovery_required" {
		t.Fatalf("apply second first error = %T %v, want changeset_recovery_required", err, err)
	}
	if got := countRows(t, db, "stories"); got != 0 {
		t.Fatalf("stories after rejected later apply = %d, want 0", got)
	}

	if _, _, err := ApplyChangesetForRecovery(db, dir, first); err != nil {
		t.Fatalf("apply first: %v", err)
	}
	if _, _, err := ApplyChangesetForRecovery(db, dir, second); err != nil {
		t.Fatalf("apply second: %v", err)
	}
	live := recoverySnapshotForTest(t, db, storyID)
	if live.goal != "second" {
		t.Fatalf("live goal = %q, want second", live.goal)
	}

	replayDB := freshReplayDBForTest(t)
	if _, err := infrastructure.Replay(replayDB, dir); err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if replayed := recoverySnapshotForTest(t, replayDB, storyID); replayed != live {
		t.Fatalf("replay state differs: live=%+v replay=%+v", live, replayed)
	}
}

type recoveryTestSnapshot struct {
	goal    string
	intakes int
	fence   sql.NullString
}

func recoverySnapshotForTest(t *testing.T, db *sql.DB, storyID string) recoveryTestSnapshot {
	t.Helper()
	var snapshot recoveryTestSnapshot
	if err := db.QueryRow(`SELECT goal FROM stories WHERE id = ?`, storyID).Scan(&snapshot.goal); err != nil {
		t.Fatalf("query story goal: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM intakes`).Scan(&snapshot.intakes); err != nil {
		t.Fatalf("count intakes: %v", err)
	}
	if err := db.QueryRow(`SELECT last_applied_changeset FROM meta`).Scan(&snapshot.fence); err != nil {
		t.Fatalf("query fence: %v", err)
	}
	return snapshot
}

func freshReplayDBForTest(t *testing.T) *sql.DB {
	t.Helper()
	db, err := infrastructure.Open(filepath.Join(t.TempDir(), "harness.db"))
	if err != nil {
		t.Fatalf("Open replay db: %v", err)
	}
	if _, _, err := infrastructure.Migrate(db); err != nil {
		db.Close()
		t.Fatalf("Migrate replay db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
