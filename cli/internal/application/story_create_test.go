package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestCreateStory(t *testing.T) {
	db, changesetDir := freshDB(t)

	id, path, err := CreateStory(db, changesetDir, "cli-domain", "ported domain commands work", "")
	if err != nil {
		t.Fatalf("CreateStory: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "stories", id, "story")

	var status string
	if err := db.QueryRow(`SELECT status FROM stories WHERE id = ?`, id).Scan(&status); err != nil {
		t.Fatalf("query status: %v", err)
	}
	if status != domain.StoryPlanned {
		t.Fatalf("status = %q, want %q", status, domain.StoryPlanned)
	}
}

func TestCreateStoryWithDependsOn(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := CreateStory(db, changesetDir, "cli-core", "working zharness core", "")
	if err != nil {
		t.Fatalf("CreateStory (prereq): %v", err)
	}

	id, _, err := CreateStory(db, changesetDir, "cli-domain", "ported domain commands work", "cli-core")
	if err != nil {
		t.Fatalf("CreateStory (dependent): %v", err)
	}

	var dependsOn string
	if err := db.QueryRow(`SELECT depends_on FROM stories WHERE id = ?`, id).Scan(&dependsOn); err != nil {
		t.Fatalf("query depends_on: %v", err)
	}
	if dependsOn != "cli-core" {
		t.Fatalf("depends_on = %q, want cli-core", dependsOn)
	}
}

func TestCreateStoryDuplicateSlug(t *testing.T) {
	db, changesetDir := freshDB(t)

	if _, _, err := CreateStory(db, changesetDir, "cli-domain", "first goal", ""); err != nil {
		t.Fatalf("CreateStory (first): %v", err)
	}

	_, _, err := CreateStory(db, changesetDir, "cli-domain", "second goal", "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "duplicate_slug" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: duplicate_slug}", err)
	}
	if got := countRows(t, db, "stories"); got != 1 {
		t.Fatalf("stories rows = %d, want 1 (duplicate not written)", got)
	}
}

func TestCreateStoryUnknownDependency(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := CreateStory(db, changesetDir, "cli-domain", "ported domain commands work", "does-not-exist")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_dependency" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_dependency}", err)
	}
	if got := countRows(t, db, "stories"); got != 0 {
		t.Fatalf("stories rows = %d, want 0", got)
	}
}
