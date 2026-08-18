package application

import (
	"database/sql"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

const validRunPlanID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"

type runCreateSnapshot struct {
	runRows              int
	storyStatus          string
	schemaVersion        int
	currentPhase         sql.NullString
	entryPhase           sql.NullString
	latestRunID          sql.NullString
	latestCheckID        sql.NullString
	lastAppliedChangeset sql.NullString
	docsVersion          sql.NullString
}

func takeRunCreateSnapshot(t *testing.T, db *sql.DB, storySlug string) runCreateSnapshot {
	t.Helper()

	snapshot := runCreateSnapshot{
		runRows:     countRows(t, db, "runs"),
		storyStatus: queryStoryStatus(t, db, storySlug),
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
		t.Fatalf("query run-create meta snapshot: %v", err)
	}
	return snapshot
}

func TestRunCreate(t *testing.T) {
	db := freshDB(t)
	if _, err := CreateStory(db, "write-boundary", "add run create", ""); err != nil {
		t.Fatalf("CreateStory (prereq): %v", err)
	}

	id, err := CreateRun(db, "write-boundary", "", validRunPlanID)
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	var storySlug, artifactPath, planID string
	if err := db.QueryRow(`SELECT story_slug, artifact_path, plan_id FROM runs WHERE id = ?`, id).Scan(&storySlug, &artifactPath, &planID); err != nil {
		t.Fatalf("query run: %v", err)
	}
	if storySlug != "write-boundary" || artifactPath != "" || planID != validRunPlanID {
		t.Fatalf("run = (%q, %q, %q)", storySlug, artifactPath, planID)
	}
	if got := queryStoryStatus(t, db, "write-boundary"); got != domain.StoryInProgress {
		t.Fatalf("story status = %q, want in-progress", got)
	}
	var latestRunID, currentPhase string
	if err := db.QueryRow(`SELECT latest_run_id, current_phase FROM meta`).Scan(&latestRunID, &currentPhase); err != nil {
		t.Fatalf("query meta: %v", err)
	}
	if latestRunID != id || currentPhase != "write-boundary" {
		t.Fatalf("meta = (latest_run_id=%q, current_phase=%q), want (%q, write-boundary)", latestRunID, currentPhase, id)
	}
}

func TestRunCreateAcceptsEmptyPlanID(t *testing.T) {
	db := freshDB(t)
	if _, err := CreateStory(db, "empty-plan-id", "allow a run without a plan ID", ""); err != nil {
		t.Fatalf("CreateStory: %v", err)
	}

	id, err := CreateRun(db, "empty-plan-id", "", "")
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	if id == "" {
		t.Fatal("CreateRun returned empty id, want persisted run")
	}
	var planID sql.NullString
	if err := db.QueryRow(`SELECT plan_id FROM runs WHERE id = ?`, id).Scan(&planID); err != nil {
		t.Fatalf("query run plan_id: %v", err)
	}
	if planID.Valid {
		t.Fatalf("plan_id = %q, want NULL for empty input", planID.String)
	}
	if got := queryStoryStatus(t, db, "empty-plan-id"); got != domain.StoryInProgress {
		t.Fatalf("story status = %q, want in-progress", got)
	}
}

func TestRunCreateRejectsInvalidPlanIDWithoutWrites(t *testing.T) {
	for _, tc := range []struct {
		name   string
		planID string
	}{
		{name: "malformed", planID: "not-a-ulid"},
		{name: "overflow", planID: "ZZZZZZZZZZZZZZZZZZZZZZZZZZ"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := freshDB(t)
			storySlug := "invalid-plan-id-" + tc.name
			if _, err := CreateStory(db, storySlug, "reject invalid plan IDs", ""); err != nil {
				t.Fatalf("CreateStory: %v", err)
			}

			before := takeRunCreateSnapshot(t, db, storySlug)
			id, err := CreateRun(db, storySlug, "", tc.planID)
			assertLifecycleValidationError(t, err, "invalid_plan_id", "run create: plan_id must be a valid ULID")
			if id != "" {
				t.Fatalf("rejected CreateRun returned id=%q, want empty", id)
			}
			after := takeRunCreateSnapshot(t, db, storySlug)
			if before != after {
				t.Fatalf("run-create state changed after invalid plan ID:\nbefore = %+v\nafter  = %+v", before, after)
			}
		})
	}
}

func TestRunCreateUnknownStory(t *testing.T) {
	db := freshDB(t)

	_, err := CreateRun(db, "does-not-exist", "", "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_story" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_story}", err)
	}
	if got := countRows(t, db, "runs"); got != 0 {
		t.Fatalf("runs rows = %d, want 0", got)
	}
}
