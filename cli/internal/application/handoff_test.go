package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestHandoffRecord(t *testing.T) {
	db := freshDB(t)
	runID := seedRun(t, db)
	checkID := seedCheck(t, db)

	id, err := RecordHandoff(db, runID, checkID, "", []string{"finish continuity phase"}, false)
	if err != nil {
		t.Fatalf("RecordHandoff: %v", err)
	}
	assertRowExists(t, db, "handoffs", id)

	var gotRunID, gotCheckID string
	if err := db.QueryRow(`SELECT run_id, check_id FROM handoffs WHERE id = ?`, id).Scan(&gotRunID, &gotCheckID); err != nil {
		t.Fatalf("query handoff: %v", err)
	}
	if gotRunID != runID || gotCheckID != checkID {
		t.Fatalf("run_id=%q check_id=%q, want %q/%q", gotRunID, gotCheckID, runID, checkID)
	}
}

func TestHandoffRecordClosesCleanPhase(t *testing.T) {
	db := freshDB(t)
	runID := createLifecycleRun(t, db, "cli-domain")
	checkID, err := RecordCheck(db, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{{Command: "true", OutputRef: "pass"}})
	if err != nil {
		t.Fatalf("RecordCheck() error = %v", err)
	}

	if _, err := RecordHandoff(db, runID, checkID, "", nil, true); err != nil {
		t.Fatalf("RecordHandoff(close) error = %v", err)
	}
	if got := queryStoryStatus(t, db, "cli-domain"); got != domain.StoryDone {
		t.Fatalf("story status = %q, want done", got)
	}
}

func TestHandoffRecordRejectsDirtyPhaseClose(t *testing.T) {
	db := freshDB(t)
	runID := createLifecycleRun(t, db, "cli-domain")
	checkID, err := RecordCheck(db, runID, domain.VerdictRequestChanges, domain.JudgeIndependent, "test-model", nil)
	if err != nil {
		t.Fatalf("RecordCheck() error = %v", err)
	}

	before := takeLifecycleSnapshot(t, db, "cli-domain")
	id, err := RecordHandoff(db, runID, checkID, "", nil, true)
	assertLifecycleValidationError(t, err, "check_not_clean", "handoff record: cannot close a phase with REQUEST_CHANGES")
	if id != "" {
		t.Fatalf("rejected RecordHandoff returned id=%q, want empty value", id)
	}
	assertLifecycleUnchanged(t, before, takeLifecycleSnapshot(t, db, "cli-domain"))
}

func TestHandoffRecordNoAnchors(t *testing.T) {
	db := freshDB(t)

	id, err := RecordHandoff(db, "", "", "", nil, false)
	if err != nil {
		t.Fatalf("RecordHandoff: %v", err)
	}
	assertRowExists(t, db, "handoffs", id)
}

func TestHandoffRecordEmptyOpenItem(t *testing.T) {
	db := freshDB(t)

	_, err := RecordHandoff(db, "", "", "", []string{""}, false)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "invalid_open_items" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: invalid_open_items}", err)
	}
}

func TestHandoffRecordUnknownRunID(t *testing.T) {
	db := freshDB(t)

	_, err := RecordHandoff(db, "01HZZZZZZZZZZZZZZZZZZZZZZZ", "", "", nil, false)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_run_id" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_run_id}", err)
	}
	if got := countRows(t, db, "handoffs"); got != 0 {
		t.Fatalf("handoffs rows = %d, want 0", got)
	}
}

func TestHandoffRecordUnknownCheckID(t *testing.T) {
	db := freshDB(t)

	_, err := RecordHandoff(db, "", "01HZZZZZZZZZZZZZZZZZZZZZZZ", "", nil, false)
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
	db := freshDB(t)
	runID := seedRun(t, db)
	checkID := seedCheck(t, db)

	id, err := RecordHandoff(db, runID, checkID, "start p2-complete-the-index wave 1", []string{"owner decision pending"}, false)
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
	db := freshDB(t)
	runID := seedRun(t, db)
	checkID := seedCheck(t, db)

	if _, err := RecordHandoff(db, runID, checkID, "", nil, false); err != nil {
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

// TestHandoffRecordWritesPlanProgressEntry proves P3 wave 2 for handoff
// record: the entry lands in `## Progress` as an event-log line (handoff
// id, run, check, next action, open items), not as a rewrite of the
// snapshot-style `## Current State and Next Action` section.
func TestHandoffRecordWritesPlanProgressEntry(t *testing.T) {
	chdirFixture(t)
	planPath := writeActivePlanFixture(t, "demo")
	db := freshDB(t)
	runID := seedRun(t, db)
	checkID := seedCheck(t, db)

	id, err := RecordHandoff(db, runID, checkID, "start next phase", []string{"owner decision pending"}, false)
	if err != nil {
		t.Fatalf("RecordHandoff: %v", err)
	}

	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "handoff: `"+id+"`") {
		t.Fatalf("plan Progress missing handoff id:\n%s", content)
	}
	if !strings.Contains(content, "start next phase") || !strings.Contains(content, "owner decision pending") {
		t.Fatalf("plan Progress missing next action/open items:\n%s", content)
	}
	if !strings.Contains(content, "## Current State and Next Action\n- active_phase: none") {
		t.Fatalf("plan Current State section unexpectedly touched:\n%s", content)
	}
}

// TestHandoffRecordMalformedPlanBlocksDBWrite is handoff record's version
// of the atomicity proof: a missing `## Progress` section must fail before
// the DB write, using the same pre-check pattern check_record's deferred
// id/at minting requires.
func TestHandoffRecordMalformedPlanBlocksDBWrite(t *testing.T) {
	chdirFixture(t)
	planPath := writeActivePlanFixture(t, "demo")
	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	corrupted := strings.Replace(string(data), "## Progress", "## Renamed", 1)
	if err := os.WriteFile(planPath, []byte(corrupted), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	db := freshDB(t)
	runID := seedRun(t, db)
	checkID := seedCheck(t, db)

	_, err = RecordHandoff(db, runID, checkID, "", nil, false)
	if err == nil {
		t.Fatal("RecordHandoff = nil error, want a plan-section-not-found failure")
	}
	if got := countRows(t, db, "handoffs"); got != 0 {
		t.Fatalf("handoffs rows = %d, want 0 — DB write must not proceed when the plan can't be written to", got)
	}
	after, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("ReadFile after failed RecordHandoff: %v", err)
	}
	if string(after) != corrupted {
		t.Fatal("plan file changed despite the failed write")
	}
}

// TestHandoffRecordReadOnlyPlanBlocksDBWrite is R8's forced-failure proof: a
// markdown write that fails at the filesystem level (not just a malformed
// section) must still leave zero DB rows behind it
// (docs/plans/active/harness-markdown-truth.md).
func TestHandoffRecordReadOnlyPlanBlocksDBWrite(t *testing.T) {
	chdirFixture(t)
	planPath := writeActivePlanFixture(t, "demo")
	makeDirReadOnly(t, filepath.Dir(planPath))

	db := freshDB(t)
	runID := seedRun(t, db)
	checkID := seedCheck(t, db)

	_, err := RecordHandoff(db, runID, checkID, "", nil, false)
	if err == nil {
		t.Fatal("RecordHandoff = nil error, want a read-only plan write failure")
	}
	if got := countRows(t, db, "handoffs"); got != 0 {
		t.Fatalf("handoffs rows = %d, want 0 — DB write must not proceed when the plan write fails", got)
	}
}
