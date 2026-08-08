package application

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func runExists(db *sql.DB, id string) (bool, error) {
	var found string
	err := db.QueryRow(`SELECT id FROM runs WHERE id = ?`, id).Scan(&found)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("query run %q: %w", id, err)
	}
	return true, nil
}

// CreateTrace validates and records a new trace entity (CONTRACT.md
// `trace add`), changeset-first. unknown_run_id is DB-lookup-dependent
// (only checked when --run-id is given), so it's enforced here rather
// than in domain.Trace.Validate(). task and taskStatus are both optional:
// a wave-level trace (work.md step 9) omits them; a task-level trace (step
// 7, addressing G1 — docs/audit/workflow-harness-ceremony-audit.md) sets
// both, so a mid-wave interruption still leaves a queryable index entry
// for every task actually completed.
func CreateTrace(db *sql.DB, changesetDir string, wave int, summary, runID, task, taskStatus string) (id, path string, err error) {
	var taskStatusPtr *string
	if taskStatus != "" {
		taskStatusPtr = &taskStatus
	}
	entity := domain.Trace{Wave: wave, Summary: summary, TaskStatus: taskStatusPtr}
	if err := entity.Validate(); err != nil {
		return "", "", err
	}

	if runID != "" {
		exists, err := runExists(db, runID)
		if err != nil {
			return "", "", err
		}
		if !exists {
			return "", "", &domain.ValidationError{Code: "unknown_run_id", Message: "trace: run_id " + runID + " not found"}
		}
	}

	at := time.Now().UTC().Format(time.RFC3339)
	id = ulid.Make().String()
	fields := map[string]any{
		"wave":       wave,
		"summary":    summary,
		"created_at": at,
	}
	if runID != "" {
		fields["run_id"] = runID
	}
	if task != "" {
		fields["task"] = task
	}
	if taskStatus != "" {
		fields["task_status"] = taskStatus
	}
	path, _, err = AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "create", Entity: "trace", ID: id, Fields: fields, At: at},
	})
	if err != nil {
		return "", "", err
	}
	return id, path, nil
}
