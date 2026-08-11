package interfaces

import (
	"encoding/json"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// TestTraceAddTaskGranularityRoundTrip is the CLI-level round trip for G1
// (docs/audit/workflow-harness-ceremony-audit.md): --task/--task-status
// reach the database and come back out of `query traces`, matching
// docs/playbooks/work.md's per-task Progress entries.
func TestTraceAddTaskGranularityRoundTrip(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}

	out, err := runDBCommand(t, "trace", "add", "--wave", "1", "--summary", "task done",
		"--task", "wave 1 task 2", "--task-status", domain.TaskStatusDone, "--json")
	if err != nil {
		t.Fatalf("trace add: %v (output=%s)", err, out)
	}

	out, err = runDBCommand(t, "query", "traces", "--json")
	if err != nil {
		t.Fatalf("query traces: %v (output=%s)", err, out)
	}
	var traces []struct {
		Wave       int     `json:"wave"`
		Summary    string  `json:"summary"`
		Task       *string `json:"task"`
		TaskStatus *string `json:"task_status"`
	}
	if err := json.Unmarshal([]byte(out), &traces); err != nil {
		t.Fatalf("decode query traces output %q: %v", out, err)
	}
	if len(traces) != 1 {
		t.Fatalf("query traces = %+v, want 1 row", traces)
	}
	if traces[0].Task == nil || *traces[0].Task != "wave 1 task 2" {
		t.Fatalf("traces[0].Task = %v, want %q", traces[0].Task, "wave 1 task 2")
	}
	if traces[0].TaskStatus == nil || *traces[0].TaskStatus != domain.TaskStatusDone {
		t.Fatalf("traces[0].TaskStatus = %v, want %q", traces[0].TaskStatus, domain.TaskStatusDone)
	}
}

func TestTraceAddRejectsInvalidTaskStatus(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}

	_, err := runDBCommand(t, "trace", "add", "--wave", "1", "--summary", "s", "--task-status", "BOGUS", "--json")
	ce, ok := err.(*cliError)
	if !ok || ce.Code != "invalid_task_status" {
		t.Fatalf("trace add with bogus task-status = %v, want invalid_task_status cliError", err)
	}
}

// TestTraceAddBatchAndQueryRoundTrip is the CLI-level round trip for R5
// (docs/audit/sdlc-token-cache-audit.md): --tasks reaches the database as
// one row per element and comes back out of `query traces` in input order.
func TestTraceAddBatchAndQueryRoundTrip(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}

	out, err := runDBCommand(t, "trace", "add", "--wave", "1", "--tasks",
		`[{"task":"task A","task_status":"DONE","summary":"did A"},{"task":"task B","task_status":"DONE_WITH_CONCERNS","summary":"did B"}]`,
		"--json")
	if err != nil {
		t.Fatalf("trace add --tasks: %v (output=%s)", err, out)
	}
	var addResult []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(out), &addResult); err != nil {
		t.Fatalf("decode trace add --tasks output %q: %v", out, err)
	}
	if len(addResult) != 2 || addResult[0].ID == "" || addResult[1].ID == "" || addResult[0].ID == addResult[1].ID {
		t.Fatalf("trace add --tasks ids = %+v, want 2 distinct non-empty ids", addResult)
	}

	out, err = runDBCommand(t, "query", "traces", "--json")
	if err != nil {
		t.Fatalf("query traces: %v (output=%s)", err, out)
	}
	var traces []struct {
		Wave       int     `json:"wave"`
		Summary    string  `json:"summary"`
		Task       *string `json:"task"`
		TaskStatus *string `json:"task_status"`
	}
	if err := json.Unmarshal([]byte(out), &traces); err != nil {
		t.Fatalf("decode query traces output %q: %v", out, err)
	}
	if len(traces) != 2 {
		t.Fatalf("query traces = %+v, want 2 rows", traces)
	}
	if traces[0].Task == nil || *traces[0].Task != "task A" || traces[0].TaskStatus == nil || *traces[0].TaskStatus != "DONE" {
		t.Fatalf("query traces[0] = %+v, want task A/DONE", traces[0])
	}
	if traces[1].Task == nil || *traces[1].Task != "task B" || traces[1].TaskStatus == nil || *traces[1].TaskStatus != "DONE_WITH_CONCERNS" {
		t.Fatalf("query traces[1] = %+v, want task B/DONE_WITH_CONCERNS", traces[1])
	}
}

func TestTraceAddBatchRejectsEmptyBatch(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}

	_, err := runDBCommand(t, "trace", "add", "--wave", "1", "--tasks", "[]", "--json")
	ce, ok := err.(*cliError)
	if !ok || ce.Code != "empty_tasks" {
		t.Fatalf("trace add --tasks [] = %v, want empty_tasks cliError", err)
	}
}

func TestTraceAddBatchRejectsInvalidJSON(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}

	_, err := runDBCommand(t, "trace", "add", "--wave", "1", "--tasks", "{not json", "--json")
	ce, ok := err.(*cliError)
	if !ok || ce.Code != "invalid_tasks" {
		t.Fatalf("trace add --tasks with malformed JSON = %v, want invalid_tasks cliError", err)
	}
}

// TestTraceAddBatchRejectsMixedFlags proves --tasks and the single-entry
// flags cannot be combined — the plan's "mutually exclusive with the
// single-task flags" requirement (R5).
func TestTraceAddBatchRejectsMixedFlags(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}

	_, err := runDBCommand(t, "trace", "add", "--wave", "1", "--tasks",
		`[{"task":"task A","task_status":"DONE","summary":"did A"}]`,
		"--task", "task A", "--json")
	ce, ok := err.(*cliError)
	if !ok || ce.Code != "invalid_arguments" {
		t.Fatalf("trace add --tasks with --task = %v, want invalid_arguments cliError", err)
	}
}
