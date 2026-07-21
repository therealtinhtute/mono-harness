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

	id, path, err := CreateRun(db, changesetDir, "write-boundary", ".kit/runs/work/x.md", "")
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	lines, err := infrastructure.ReadChangeset(path)
	if err != nil {
		t.Fatalf("read changeset: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("changeset %s has %d lines, want 2 (create run + update meta)", path, len(lines))
	}
	if lines[0].ID != id || lines[0].Entity != "run" || lines[0].Op != "create" {
		t.Fatalf("changeset line 0 = %+v, want id=%s entity=run op=create", lines[0], id)
	}
	if lines[1].Entity != "meta" || lines[1].Op != "update" || lines[1].Fields["latest_run_id"] != id {
		t.Fatalf("changeset line 1 = %+v, want entity=meta op=update fields.latest_run_id=%s", lines[1], id)
	}

	var storySlug string
	if err := db.QueryRow(`SELECT story_slug FROM runs WHERE id = ?`, id).Scan(&storySlug); err != nil {
		t.Fatalf("query run: %v", err)
	}
	if storySlug != "write-boundary" {
		t.Fatalf("story_slug = %q, want write-boundary", storySlug)
	}

	var latestRunID string
	if err := db.QueryRow(`SELECT latest_run_id FROM meta LIMIT 1`).Scan(&latestRunID); err != nil {
		t.Fatalf("query meta: %v", err)
	}
	if latestRunID != id {
		t.Fatalf("meta.latest_run_id = %q, want %q", latestRunID, id)
	}
}

func TestRunCreateUnknownStory(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := CreateRun(db, changesetDir, "does-not-exist", ".kit/runs/work/x.md", "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_story" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_story}", err)
	}
	if got := countRows(t, db, "runs"); got != 0 {
		t.Fatalf("runs rows = %d, want 0", got)
	}
}

func TestRunCreateMissingArtifactPath(t *testing.T) {
	db, changesetDir := freshDB(t)
	if _, _, err := CreateStory(db, changesetDir, "write-boundary", "add run create", ""); err != nil {
		t.Fatalf("CreateStory (prereq): %v", err)
	}

	_, _, err := CreateRun(db, changesetDir, "write-boundary", "", "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "missing_required_field" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: missing_required_field}", err)
	}
}
