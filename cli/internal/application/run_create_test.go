package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func TestRunCreate(t *testing.T) {
	db, changesetDir := freshDB(t)
	if _, _, err := CreateStory(db, changesetDir, "write-boundary", "add run create", ""); err != nil {
		t.Fatalf("CreateStory (prereq): %v", err)
	}

	id, path, err := CreateRun(db, changesetDir, "write-boundary", "", "01JPLANPLANPLANPLANPLANPLAN")
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
	if storySlug != "write-boundary" || artifactPath != "" || planID != "01JPLANPLANPLANPLANPLANPLAN" {
		t.Fatalf("run = (%q, %q, %q)", storySlug, artifactPath, planID)
	}
	if got := queryStoryStatus(t, db, "write-boundary"); got != domain.StoryInProgress {
		t.Fatalf("story status = %q, want in-progress", got)
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
