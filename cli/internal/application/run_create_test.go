package application

import (
	"database/sql"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
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
	changesets           int
}

func takeRunCreateSnapshot(t *testing.T, db *sql.DB, changesetDir, storySlug string) runCreateSnapshot {
	t.Helper()

	snapshot := runCreateSnapshot{
		runRows:     countRows(t, db, "runs"),
		storyStatus: queryStoryStatus(t, db, storySlug),
		changesets:  countLifecycleChangesets(t, changesetDir),
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
	db, changesetDir := freshDB(t)
	if _, _, err := CreateStory(db, changesetDir, "write-boundary", "add run create", ""); err != nil {
		t.Fatalf("CreateStory (prereq): %v", err)
	}

	id, path, err := CreateRun(db, changesetDir, "write-boundary", "", validRunPlanID)
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	lines, err := infrastructure.ReadChangeset(path)
	if err != nil {
		t.Fatalf("read changeset: %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("changeset %s has %d lines, want 3 (run + story status + meta)", path, len(lines))
	}
	if lines[0].ID != id || lines[0].Entity != "run" || lines[0].Op != "create" {
		t.Fatalf("changeset line 0 = %+v", lines[0])
	}
	if lines[1].Entity != "story" || lines[1].Fields["status"] != domain.StoryInProgress {
		t.Fatalf("changeset line 1 = %+v, want story in-progress", lines[1])
	}
	if lines[2].Entity != "meta" || lines[2].Fields["latest_run_id"] != id || lines[2].Fields["current_phase"] != "write-boundary" {
		t.Fatalf("changeset line 2 = %+v, want latest run/current phase", lines[2])
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
}

func TestRunCreateAcceptsEmptyPlanID(t *testing.T) {
	db, changesetDir := freshDB(t)
	if _, _, err := CreateStory(db, changesetDir, "empty-plan-id", "allow a run without a plan ID", ""); err != nil {
		t.Fatalf("CreateStory: %v", err)
	}

	id, path, err := CreateRun(db, changesetDir, "empty-plan-id", "", "")
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	if id == "" || path == "" {
		t.Fatalf("CreateRun returned id=%q path=%q, want persisted run", id, path)
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
			db, changesetDir := freshDB(t)
			storySlug := "invalid-plan-id-" + tc.name
			if _, _, err := CreateStory(db, changesetDir, storySlug, "reject invalid plan IDs", ""); err != nil {
				t.Fatalf("CreateStory: %v", err)
			}

			before := takeRunCreateSnapshot(t, db, changesetDir, storySlug)
			id, path, err := CreateRun(db, changesetDir, storySlug, "", tc.planID)
			assertLifecycleValidationError(t, err, "invalid_plan_id", "run create: plan_id must be a valid ULID")
			if id != "" || path != "" {
				t.Fatalf("rejected CreateRun returned id=%q path=%q, want empty values", id, path)
			}
			after := takeRunCreateSnapshot(t, db, changesetDir, storySlug)
			if before != after {
				t.Fatalf("run-create state changed after invalid plan ID:\nbefore = %+v\nafter  = %+v", before, after)
			}
		})
	}
}

func TestRunCreateUnknownStory(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := CreateRun(db, changesetDir, "does-not-exist", "", "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_story" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_story}", err)
	}
	if got := countRows(t, db, "runs"); got != 0 {
		t.Fatalf("runs rows = %d, want 0", got)
	}
}
