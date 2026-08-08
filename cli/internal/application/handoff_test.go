package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestHandoffRecord(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)
	checkID := seedCheck(t, db, changesetDir)

	id, path, err := RecordHandoff(db, changesetDir, runID, checkID, "", []string{"finish continuity phase"}, false)
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

	if _, _, err := RecordHandoff(db, changesetDir, runID, checkID, "", nil, true); err != nil {
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
	id, path, err := RecordHandoff(db, changesetDir, runID, checkID, "", nil, true)
	assertLifecycleValidationError(t, err, "check_not_clean", "handoff record: cannot close a phase with REQUEST_CHANGES")
	if id != "" || path != "" {
		t.Fatalf("rejected RecordHandoff returned id=%q path=%q, want empty values", id, path)
	}
	assertLifecycleUnchanged(t, before, takeLifecycleSnapshot(t, db, changesetDir, "cli-domain"))
}

func TestHandoffRecordNoAnchors(t *testing.T) {
	db, changesetDir := freshDB(t)

	id, path, err := RecordHandoff(db, changesetDir, "", "", "", nil, false)
	if err != nil {
		t.Fatalf("RecordHandoff: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "handoffs", id, "handoff")
}

func TestHandoffRecordEmptyOpenItem(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := RecordHandoff(db, changesetDir, "", "", "", []string{""}, false)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "invalid_open_items" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: invalid_open_items}", err)
	}
}

func TestHandoffRecordUnknownRunID(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := RecordHandoff(db, changesetDir, "01HZZZZZZZZZZZZZZZZZZZZZZZ", "", "", nil, false)
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

	_, _, err := RecordHandoff(db, changesetDir, "", "01HZZZZZZZZZZZZZZZZZZZZZZZ", "", nil, false)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_check_id" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_check_id}", err)
	}
	if got := countRows(t, db, "handoffs"); got != 0 {
		t.Fatalf("handoffs rows = %d, want 0", got)
	}
}

// TestHandoffRecordNextActionRoundTripsThroughResumeAndQuery is the round
// trip the plan calls for (docs/plans/active/harness-memory-ceremony-convergence.md,
// P2 wave 2): --next-action persists into handoffs.anchors with no
// migration (anchors is already free-form JSON), resume's latest_handoff_id
// points at the row, and query handoff --latest reads exact_next_action
// back out.
func TestHandoffRecordNextActionRoundTripsThroughResumeAndQuery(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)
	checkID := seedCheck(t, db, changesetDir)

	id, _, err := RecordHandoff(db, changesetDir, runID, checkID, "start p2-complete-the-index wave 1", []string{"owner decision pending"}, false)
	if err != nil {
		t.Fatalf("RecordHandoff: %v", err)
	}

	resumeView, err := Resume(db, "dev")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if resumeView.LatestHandoffID == nil || *resumeView.LatestHandoffID != id {
		t.Fatalf("Resume().LatestHandoffID = %v, want %q", resumeView.LatestHandoffID, id)
	}

	handoffView, ok, err := QueryLatestHandoff(db)
	if err != nil {
		t.Fatalf("QueryLatestHandoff: %v", err)
	}
	if !ok {
		t.Fatal("QueryLatestHandoff ok = false, want true")
	}
	if handoffView.ID != id {
		t.Fatalf("QueryLatestHandoff.ID = %q, want %q", handoffView.ID, id)
	}
	if handoffView.NextAction == nil || *handoffView.NextAction != "start p2-complete-the-index wave 1" {
		t.Fatalf("QueryLatestHandoff.NextAction = %v, want %q", handoffView.NextAction, "start p2-complete-the-index wave 1")
	}
	if len(handoffView.OpenItems) != 1 || handoffView.OpenItems[0] != "owner decision pending" {
		t.Fatalf("QueryLatestHandoff.OpenItems = %v, want [owner decision pending]", handoffView.OpenItems)
	}
}

// TestHandoffRecordNextActionOptional proves the field is additive: a
// handoff with no --next-action stores none, and the round trip reports
// exact_next_action as nil rather than an empty string.
func TestHandoffRecordNextActionOptional(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)
	checkID := seedCheck(t, db, changesetDir)

	if _, _, err := RecordHandoff(db, changesetDir, runID, checkID, "", nil, false); err != nil {
		t.Fatalf("RecordHandoff: %v", err)
	}

	handoffView, ok, err := QueryLatestHandoff(db)
	if err != nil {
		t.Fatalf("QueryLatestHandoff: %v", err)
	}
	if !ok {
		t.Fatal("QueryLatestHandoff ok = false, want true")
	}
	if handoffView.NextAction != nil {
		t.Fatalf("QueryLatestHandoff.NextAction = %v, want nil", handoffView.NextAction)
	}
}
