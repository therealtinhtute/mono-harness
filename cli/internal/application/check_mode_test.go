package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestCheckRecordFullModeOnCheckedStorySucceeds(t *testing.T) {
	db := freshDB(t)
	slug := "full-on-checked"
	runID := createLifecycleRun(t, db, slug)
	if _, err := RecordCheck(db, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{{Command: "true", OutputRef: "ok"}}, domain.CheckModeGate); err != nil {
		t.Fatalf("gate RecordCheck: %v", err)
	}
	if got := queryStoryStatus(t, db, slug); got != domain.StoryChecked {
		t.Fatalf("story status = %q, want checked after gate", got)
	}

	id, err := RecordCheck(db, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{{Command: "true", OutputRef: "ok"}}, domain.CheckModeFull)
	if err != nil {
		t.Fatalf("full RecordCheck on checked story: %v", err)
	}
	if id == "" {
		t.Fatalf("full RecordCheck returned empty id")
	}
	if got := queryStoryStatus(t, db, slug); got != domain.StoryChecked {
		t.Fatalf("story status = %q, want still checked after full record", got)
	}
	var mode string
	if err := db.QueryRow(`SELECT mode FROM checks WHERE id = ?`, id).Scan(&mode); err != nil {
		t.Fatalf("query check mode: %v", err)
	}
	if mode != domain.CheckModeFull {
		t.Fatalf("check mode = %q, want full", mode)
	}
}

func TestCheckRecordRequestChangesFullOnCheckedStoryReopens(t *testing.T) {
	db := freshDB(t)
	slug := "reopen-on-request-changes"
	runID := createLifecycleRun(t, db, slug)
	if _, err := RecordCheck(db, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{{Command: "true", OutputRef: "ok"}}, domain.CheckModeGate); err != nil {
		t.Fatalf("gate RecordCheck: %v", err)
	}

	if _, err := RecordCheck(db, runID, domain.VerdictRequestChanges, domain.JudgeIndependent, "test-model", nil, domain.CheckModeFull); err != nil {
		t.Fatalf("request_changes full RecordCheck on checked story: %v", err)
	}
	if got := queryStoryStatus(t, db, slug); got != domain.StoryInProgress {
		t.Fatalf("story status = %q, want reopened in-progress after request_changes full record", got)
	}
	if n := countRows(t, db, "checks"); n != 2 {
		t.Fatalf("checks row count = %d, want 2", n)
	}
}

func TestCheckRecordGateModeOnCheckedStoryStillRefused(t *testing.T) {
	db := freshDB(t)
	slug := "gate-on-checked-refused"
	runID := createLifecycleRun(t, db, slug)
	recordCleanLifecycleCheck(t, db, runID)

	before := takeLifecycleSnapshot(t, db, slug)
	id, err := RecordCheck(db, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{{Command: "true", OutputRef: "ok"}}, domain.CheckModeGate)
	assertLifecycleValidationError(t, err, "story_not_checkable", "check record: story must be in-progress")
	if id != "" {
		t.Fatalf("rejected gate RecordCheck returned id=%q, want empty value", id)
	}
	assertLifecycleUnchanged(t, before, takeLifecycleSnapshot(t, db, slug))
}

func TestCheckRecordFullModeOnDoneStoryStillRefused(t *testing.T) {
	db := freshDB(t)
	slug := "full-on-done-refused"
	runID := createLifecycleRun(t, db, slug)
	checkID := recordCleanLifecycleCheck(t, db, runID)
	if _, err := RecordHandoff(db, runID, checkID, "", nil, true); err != nil {
		t.Fatalf("RecordHandoff: %v", err)
	}

	before := takeLifecycleSnapshot(t, db, slug)
	id, err := RecordCheck(db, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{{Command: "true", OutputRef: "ok"}}, domain.CheckModeFull)
	assertLifecycleValidationError(t, err, "story_not_checkable", "check record: story must be in-progress")
	if id != "" {
		t.Fatalf("rejected full RecordCheck on done story returned id=%q, want empty value", id)
	}
	assertLifecycleUnchanged(t, before, takeLifecycleSnapshot(t, db, slug))
}

func TestCheckRecordRejectsInvalidMode(t *testing.T) {
	db := freshDB(t)
	slug := "invalid-mode"
	runID := createLifecycleRun(t, db, slug)

	before := takeLifecycleSnapshot(t, db, slug)
	id, err := RecordCheck(db, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{{Command: "true", OutputRef: "ok"}}, "banana")
	assertLifecycleValidationError(t, err, "invalid_check_mode", "check: mode must be one of gate, full")
	if id != "" {
		t.Fatalf("rejected invalid-mode RecordCheck returned id=%q, want empty value", id)
	}
	assertLifecycleUnchanged(t, before, takeLifecycleSnapshot(t, db, slug))
}

func TestCheckRecordGateDefaultPersistsModeColumn(t *testing.T) {
	db := freshDB(t)
	slug := "default-mode-persisted"
	runID := createLifecycleRun(t, db, slug)

	id, err := RecordCheck(db, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{{Command: "true", OutputRef: "ok"}}, "")
	if err != nil {
		t.Fatalf("empty-mode RecordCheck: %v", err)
	}
	var mode string
	if err := db.QueryRow(`SELECT mode FROM checks WHERE id = ?`, id).Scan(&mode); err != nil {
		t.Fatalf("query check mode: %v", err)
	}
	if mode != domain.CheckModeGate {
		t.Fatalf("check mode = %q, want gate for empty input", mode)
	}
}
