package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestCreateTool(t *testing.T) {
	db, changesetDir := freshDB(t)

	id, path, err := CreateTool(db, changesetDir, "gh", "GitHub CLI for release publishing")
	if err != nil {
		t.Fatalf("CreateTool: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "tools", id, "tool")
}

func TestCreateToolMissingName(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := CreateTool(db, changesetDir, "", "GitHub CLI")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "missing_required_field" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: missing_required_field}", err)
	}
}
