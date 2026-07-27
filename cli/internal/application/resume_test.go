package application

import (
	"database/sql"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// setMeta applies a meta update changeset line, mirroring the shape
// applyMetaLine expects (no id, entity "meta", op "update").
func setMeta(t *testing.T, db *sql.DB, changesetDir string, fields map[string]any) {
	t.Helper()
	if _, _, err := AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "update", Entity: "meta", Fields: fields, At: "2026-07-17T12:10:00Z"},
	}); err != nil {
		t.Fatalf("setMeta: %v", err)
	}
}

func TestResumeCleanState(t *testing.T) {
	db, changesetDir := freshDB(t)
	seedRun(t, db, changesetDir) // seeds the "cli-domain" story; artifact_path is fake so latest_run_id is deliberately left unset here

	if _, err := db.Exec(`UPDATE stories SET status = ? WHERE slug = ?`, domain.StoryDone, "cli-domain"); err != nil {
		t.Fatalf("seed story status: %v", err)
	}
	setMeta(t, db, changesetDir, map[string]any{"current_phase": "cli-domain"})

	view, err := Resume(db, "dev")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if view.Readiness != "clean" {
		t.Fatalf("readiness = %q, want clean (drift=%v)", view.Readiness, view.Drift)
	}
	if len(view.Drift) != 0 {
		t.Fatalf("drift = %v, want none", view.Drift)
	}
	if view.Position.CurrentPhase == nil || *view.Position.CurrentPhase != "cli-domain" {
		t.Fatalf("current_phase = %v, want cli-domain", view.Position.CurrentPhase)
	}
}

func TestResumeInProgress(t *testing.T) {
	db, changesetDir := freshDB(t)
	seedRun(t, db, changesetDir)
	setMeta(t, db, changesetDir, map[string]any{"current_phase": "cli-domain"})

	view, err := Resume(db, "dev")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if view.Readiness != "in-progress" {
		t.Fatalf("readiness = %q, want in-progress", view.Readiness)
	}
}

func TestResumeInvalidCurrentStoryStatusDrift(t *testing.T) {
	db, changesetDir := freshDB(t)
	seedRun(t, db, changesetDir)
	setMeta(t, db, changesetDir, map[string]any{"current_phase": "cli-domain"})
	if _, err := db.Exec(`UPDATE stories SET status = 'bogus' WHERE slug = 'cli-domain'`); err != nil {
		t.Fatalf("corrupt story status: %v", err)
	}

	view, err := Resume(db, "dev")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if view.Readiness != "drifted" {
		t.Fatalf("readiness = %q, want drifted", view.Readiness)
	}
	if view.Position.Status == nil || *view.Position.Status != "bogus" {
		t.Fatalf("position status = %v, want bogus", view.Position.Status)
	}
	if len(view.Drift) != 1 || view.Drift[0].Type != "invalid_status" || view.Drift[0].Recovery == "" {
		t.Fatalf("drift = %v, want one recoverable invalid_status finding", view.Drift)
	}
}

// TestResumeUnknownPhaseDrift exercises Resume's defensive unknown_phase
// branch. meta.current_phase carries its own FK to stories(slug), so the
// CLI's own changeset-first writes can never produce this state (there's no
// delete command in the 19-command surface either) — it only models a
// pointer gone stale via a manual db edit, restore, or future migration.
// Constructing the fixture requires bypassing FK enforcement directly.
func TestResumeUnknownPhaseDrift(t *testing.T) {
	db, _ := freshDB(t)
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA foreign_keys=OFF;`); err != nil {
		t.Fatalf("disable foreign_keys: %v", err)
	}
	if _, err := db.Exec(`UPDATE meta SET current_phase = ?`, "no-such-phase"); err != nil {
		t.Fatalf("seed stale current_phase: %v", err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		t.Fatalf("re-enable foreign_keys: %v", err)
	}

	view, err := Resume(db, "dev")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if view.Readiness != "drifted" {
		t.Fatalf("readiness = %q, want drifted", view.Readiness)
	}
	if len(view.Drift) != 1 || view.Drift[0].Type != "unknown_phase" {
		t.Fatalf("drift = %v, want one unknown_phase finding", view.Drift)
	}
}

func TestResumeIgnoresMissingLegacyRunArtifact(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir) // seedRun's legacy artifact_path (.kit/runs/work/x.md) never exists on disk
	setMeta(t, db, changesetDir, map[string]any{"latest_run_id": runID})

	view, err := Resume(db, "dev")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if view.Readiness != "clean" {
		t.Fatalf("readiness = %q, want clean", view.Readiness)
	}
	if len(view.Drift) != 0 {
		t.Fatalf("drift = %v, want none", view.Drift)
	}
}

// TestResumeOutOfOrderStaledPointer is the staled-pointer fixture named by
// cli-domain-PLAN.md's T3 verification step: meta.latest_check_id points at
// a check belonging to an older run than meta.latest_run_id, asserting a
// non-empty `recovery` per finding and that out_of_order is present.
func TestResumeOutOfOrderStaledPointer(t *testing.T) {
	db, changesetDir := freshDB(t)

	staleCheckID := seedCheck(t, db, changesetDir) // belongs to its own freshly seeded run
	newRunID := seedRun(t, db, changesetDir)       // a second, newer run

	setMeta(t, db, changesetDir, map[string]any{
		"latest_run_id":   newRunID,
		"latest_check_id": staleCheckID,
	})

	view, err := Resume(db, "dev")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if view.Readiness != "drifted" {
		t.Fatalf("readiness = %q, want drifted", view.Readiness)
	}
	found := false
	for _, d := range view.Drift {
		if d.Recovery == "" {
			t.Fatalf("drift finding %+v has empty recovery, want non-empty", d)
		}
		if d.Type == "out_of_order" {
			found = true
		}
	}
	if !found {
		t.Fatalf("drift = %v, want an out_of_order finding", view.Drift)
	}
}

func TestResumeSurfacesLatestHandoffID(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	handoffID := ulid.Make().String()
	if _, _, err := AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "create", Entity: "handoff", ID: handoffID, Fields: map[string]any{
			"run_id": runID, "anchors": "{}", "created_at": "2026-07-17T12:20:00Z",
		}, At: "2026-07-17T12:20:00Z"},
	}); err != nil {
		t.Fatalf("seed handoff: %v", err)
	}

	view, err := Resume(db, "dev")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if view.LatestHandoffID == nil || *view.LatestHandoffID != handoffID {
		t.Fatalf("latest_handoff_id = %v, want %s", view.LatestHandoffID, handoffID)
	}
}
