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
