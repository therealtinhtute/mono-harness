package application

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

const maxTraceTasksPerBatch = 20

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

	writePlan, err := preparePlanAppend("Progress", formatTraceProgressEntry(at, wave, summary, runID, task, taskStatus))
	if err != nil {
		return "", "", err
	}

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

	if err := writePlan(); err != nil {
		return id, path, fmt.Errorf("trace %s recorded, but plan markdown update failed: %w", id, err)
	}
	return id, path, nil
}

// CreateTraces validates and records a batch of task-level trace entities in
// one changeset — the flush-once-per-wave counterpart of CreateTrace (R5,
// docs/audit/sdlc-token-cache-audit.md). All entries share wave and the
// optional run_id; each still mints its own id/timestamp and gets its own
// `## Progress` line, one per element, so a mid-wave interruption leaves the
// same queryable per-task granularity CreateTrace's single-call form
// already provides — batching removes the round trips, not the index rows.
func CreateTraces(db *sql.DB, changesetDir string, wave int, runID string, tasks []domain.TraceTask) (ids []string, path string, err error) {
	if len(tasks) == 0 {
		return nil, "", &domain.ValidationError{Code: "empty_tasks", Message: "trace add: at least one task is required"}
	}
	if len(tasks) > maxTraceTasksPerBatch {
		return nil, "", &domain.ValidationError{Code: "too_many_tasks", Message: fmt.Sprintf("trace add: at most %d tasks per batch", maxTraceTasksPerBatch)}
	}
	for _, t := range tasks {
		if err := t.Validate(); err != nil {
			return nil, "", err
		}
	}
	if wave < 0 {
		return nil, "", &domain.ValidationError{Message: "trace: wave must be >= 0"}
	}

	if runID != "" {
		exists, err := runExists(db, runID)
		if err != nil {
			return nil, "", err
		}
		if !exists {
			return nil, "", &domain.ValidationError{Code: "unknown_run_id", Message: "trace: run_id " + runID + " not found"}
		}
	}

	at := time.Now().UTC().Format(time.RFC3339)

	entryLines := make([]string, len(tasks))
	for i, t := range tasks {
		entryLines[i] = formatTraceProgressEntry(at, wave, t.Summary, runID, t.Task, t.TaskStatus)
	}
	writePlan, err := preparePlanAppend("Progress", strings.Join(entryLines, "\n"))
	if err != nil {
		return nil, "", err
	}

	lines := make([]infrastructure.ChangesetLine, 0, len(tasks))
	ids = make([]string, 0, len(tasks))
	for _, t := range tasks {
		id := ulid.Make().String()
		fields := map[string]any{
			"wave": wave, "summary": t.Summary, "task": t.Task, "task_status": t.TaskStatus, "created_at": at,
		}
		if runID != "" {
			fields["run_id"] = runID
		}
		lines = append(lines, infrastructure.ChangesetLine{Op: "create", Entity: "trace", ID: id, Fields: fields, At: at})
		ids = append(ids, id)
	}

	path, _, err = AppendAndApply(db, changesetDir, lines)
	if err != nil {
		return nil, "", err
	}

	if err := writePlan(); err != nil {
		return ids, path, fmt.Errorf("traces %v recorded, but plan markdown update failed: %w", ids, err)
	}
	return ids, path, nil
}

// formatTraceProgressEntry renders a `## Progress` line using only the
// fields trace add actually receives — it is not a substitute for the
// richer changed-surfaces/verification detail work.md's playbook asks an
// agent to write by hand, only an honest, always-current compressed line
// that cannot drift from the traces row it accompanies (P3, "one writer").
func formatTraceProgressEntry(at string, wave int, summary, runID, task, taskStatus string) string {
	line := fmt.Sprintf("- `%s` — wave %d", at, wave)
	if task != "" {
		line += fmt.Sprintf(", task %s", task)
	}
	line += "."
	if taskStatus != "" {
		line += fmt.Sprintf(" task_status: `%s`.", taskStatus)
	}
	if runID != "" {
		line += fmt.Sprintf(" run: `%s`.", runID)
	}
	line += fmt.Sprintf(" summary: %s.", summary)
	return line
}
