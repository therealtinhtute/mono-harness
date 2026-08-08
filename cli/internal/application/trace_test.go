package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestCreateTrace(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	id, path, err := CreateTrace(db, changesetDir, 1, "wave 1 done", runID, "", "")
	if err != nil {
		t.Fatalf("CreateTrace: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "traces", id, "trace")
}

func TestCreateTraceNoRunID(t *testing.T) {
	db, changesetDir := freshDB(t)

	id, path, err := CreateTrace(db, changesetDir, 2, "standalone trace", "", "", "")
	if err != nil {
		t.Fatalf("CreateTrace: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "traces", id, "trace")
}

func TestCreateTraceUnknownRunID(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := CreateTrace(db, changesetDir, 1, "wave 1 done", "01HZZZZZZZZZZZZZZZZZZZZZZZ", "", "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_run_id" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_run_id}", err)
	}
	if got := countRows(t, db, "traces"); got != 0 {
		t.Fatalf("traces rows = %d, want 0", got)
	}
}

// TestCreateTraceTaskGranularity is the round trip for G1
// (docs/audit/workflow-harness-ceremony-audit.md): a task-level trace
// records task and task_status, matching docs/playbooks/work.md's
// per-task Progress entries rather than only wave-level summaries.
func TestCreateTraceTaskGranularity(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	id, path, err := CreateTrace(db, changesetDir, 1, "task done", runID, "wave 1 task 2", domain.TaskStatusDone)
	if err != nil {
		t.Fatalf("CreateTrace: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "traces", id, "trace")

	var task, taskStatus string
	if err := db.QueryRow(`SELECT task, task_status FROM traces WHERE id = ?`, id).Scan(&task, &taskStatus); err != nil {
		t.Fatalf("query task-level trace: %v", err)
	}
	if task != "wave 1 task 2" || taskStatus != domain.TaskStatusDone {
		t.Fatalf("trace (task, task_status) = (%q, %q), want (%q, %q)", task, taskStatus, "wave 1 task 2", domain.TaskStatusDone)
	}
}

func TestCreateTraceInvalidTaskStatus(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	_, _, err := CreateTrace(db, changesetDir, 1, "task done", runID, "task", "BOGUS")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "invalid_task_status" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: invalid_task_status}", err)
	}
	if got := countRows(t, db, "traces"); got != 0 {
		t.Fatalf("traces rows = %d, want 0", got)
	}
}
