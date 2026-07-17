package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestHandoffRecord(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)
	checkID := seedCheck(t, db, changesetDir)

	id, path, err := RecordHandoff(db, changesetDir, runID, checkID, []string{"finish continuity phase"})
	if err != nil {
		t.Fatalf("RecordHandoff: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "handoffs", id, "handoff")

	var gotRunID, gotCheckID string
	if err := db.QueryRow(`SELECT run_id, check_id FROM handoffs WHERE id = ?`, id).Scan(&gotRunID, &gotCheckID); err != nil {
		t.Fatalf("query handoff: %v", err)
	}
	if gotRunID != runID || gotCheckID != checkID {
		t.Fatalf("run_id=%q check_id=%q, want %q/%q", gotRunID, gotCheckID, runID, checkID)
	}
}

func TestHandoffRecordNoAnchors(t *testing.T) {
	db, changesetDir := freshDB(t)

	id, path, err := RecordHandoff(db, changesetDir, "", "", nil)
	if err != nil {
		t.Fatalf("RecordHandoff: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "handoffs", id, "handoff")
}

func TestHandoffRecordEmptyOpenItem(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := RecordHandoff(db, changesetDir, "", "", []string{""})
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "invalid_open_items" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: invalid_open_items}", err)
	}
}

func TestHandoffRecordUnknownRunID(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := RecordHandoff(db, changesetDir, "01HZZZZZZZZZZZZZZZZZZZZZZZ", "", nil)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_run_id" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_run_id}", err)
	}
	if got := countRows(t, db, "handoffs"); got != 0 {
		t.Fatalf("handoffs rows = %d, want 0", got)
	}
}

func TestHandoffRecordUnknownCheckID(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := RecordHandoff(db, changesetDir, "", "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_check_id" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_check_id}", err)
	}
	if got := countRows(t, db, "handoffs"); got != 0 {
		t.Fatalf("handoffs rows = %d, want 0", got)
	}
}
