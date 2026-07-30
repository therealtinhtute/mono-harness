package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestHandoffRecord(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)
	checkID := seedCheck(t, db, changesetDir)

	id, path, err := RecordHandoff(db, changesetDir, runID, checkID, []string{"finish continuity phase"}, false)
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

func TestHandoffRecordClosesCleanPhase(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := createLifecycleRun(t, db, changesetDir, "cli-domain")
	checkID, _, err := RecordCheck(db, changesetDir, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{{Command: "go test ./...", OutputRef: "pass"}})
	if err != nil {
		t.Fatalf("RecordCheck() error = %v", err)
	}

	if _, _, err := RecordHandoff(db, changesetDir, runID, checkID, nil, true); err != nil {
		t.Fatalf("RecordHandoff(close) error = %v", err)
	}
	if got := queryStoryStatus(t, db, "cli-domain"); got != domain.StoryDone {
		t.Fatalf("story status = %q, want done", got)
	}
}

func TestHandoffRecordRejectsDirtyPhaseClose(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := createLifecycleRun(t, db, changesetDir, "cli-domain")
	checkID, _, err := RecordCheck(db, changesetDir, runID, domain.VerdictRequestChanges, domain.JudgeIndependent, "test-model", nil)
	if err != nil {
		t.Fatalf("RecordCheck() error = %v", err)
	}

	before := takeLifecycleSnapshot(t, db, changesetDir, "cli-domain")
	id, path, err := RecordHandoff(db, changesetDir, runID, checkID, nil, true)
	assertLifecycleValidationError(t, err, "check_not_clean", "handoff record: cannot close a phase with REQUEST_CHANGES")
	if id != "" || path != "" {
		t.Fatalf("rejected RecordHandoff returned id=%q path=%q, want empty values", id, path)
	}
	assertLifecycleUnchanged(t, before, takeLifecycleSnapshot(t, db, changesetDir, "cli-domain"))
}

func TestHandoffRecordNoAnchors(t *testing.T) {
	db, changesetDir := freshDB(t)

	id, path, err := RecordHandoff(db, changesetDir, "", "", nil, false)
	if err != nil {
		t.Fatalf("RecordHandoff: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "handoffs", id, "handoff")
}

func TestHandoffRecordEmptyOpenItem(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := RecordHandoff(db, changesetDir, "", "", []string{""}, false)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "invalid_open_items" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: invalid_open_items}", err)
	}
}

func TestHandoffRecordUnknownRunID(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := RecordHandoff(db, changesetDir, "01HZZZZZZZZZZZZZZZZZZZZZZZ", "", nil, false)
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

	_, _, err := RecordHandoff(db, changesetDir, "", "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil, false)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_check_id" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_check_id}", err)
	}
	if got := countRows(t, db, "handoffs"); got != 0 {
		t.Fatalf("handoffs rows = %d, want 0", got)
	}
}
