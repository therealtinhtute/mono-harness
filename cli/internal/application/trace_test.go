package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestCreateTrace(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	id, path, err := CreateTrace(db, changesetDir, 1, "wave 1 done", runID)
	if err != nil {
		t.Fatalf("CreateTrace: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "traces", id, "trace")
}

func TestCreateTraceNoRunID(t *testing.T) {
	db, changesetDir := freshDB(t)

	id, path, err := CreateTrace(db, changesetDir, 2, "standalone trace", "")
	if err != nil {
		t.Fatalf("CreateTrace: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "traces", id, "trace")
}

func TestCreateTraceUnknownRunID(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := CreateTrace(db, changesetDir, 1, "wave 1 done", "01HZZZZZZZZZZZZZZZZZZZZZZZ")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_run_id" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_run_id}", err)
	}
	if got := countRows(t, db, "traces"); got != 0 {
		t.Fatalf("traces rows = %d, want 0", got)
	}
}
