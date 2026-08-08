package application

import (
	"os"
	"strings"
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

// TestCreateTraceWritesPlanProgressEntry is P3 wave 2's core proof for
// trace add: the same call that creates the DB row also appends a
// `## Progress` line to the single active plan, in one operation.
func TestCreateTraceWritesPlanProgressEntry(t *testing.T) {
	chdirFixture(t)
	planPath := writeActivePlanFixture(t, "demo")
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	id, _, err := CreateTrace(db, changesetDir, 1, "wave 1 done", runID, "task A", domain.TaskStatusDone)
	if err != nil {
		t.Fatalf("CreateTrace: %v", err)
	}
	if id == "" {
		t.Fatal("CreateTrace returned empty id")
	}

	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", planPath, err)
	}
	content := string(data)
	if !strings.Contains(content, "## Decisions\n<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->\n- none") {
		t.Fatalf("plan Decisions section corrupted:\n%s", content)
	}
	if !strings.Contains(content, "task A") || !strings.Contains(content, runID) || !strings.Contains(content, "wave 1") {
		t.Fatalf("plan Progress section missing expected trace fields:\n%s", content)
	}
	if !strings.Contains(content, "## Progress\n<!--") {
		t.Fatalf("plan Progress heading/comment corrupted:\n%s", content)
	}
}

// TestCreateTraceNoActivePlanSkipsMarkdownWrite proves the feature is
// additive: with zero active plans (the ordinary case for most of this
// package's other tests, and for bounded/simple work), the DB write still
// succeeds and nothing is written to disk.
func TestCreateTraceNoActivePlanSkipsMarkdownWrite(t *testing.T) {
	chdirFixture(t)
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	id, _, err := CreateTrace(db, changesetDir, 1, "wave 1 done", runID, "", "")
	if err != nil {
		t.Fatalf("CreateTrace: %v", err)
	}
	if id == "" {
		t.Fatal("CreateTrace returned empty id")
	}
	if _, err := os.Stat("docs"); err == nil {
		t.Fatal("docs/ unexpectedly created with no active plan")
	}
}

// TestCreateTraceMalformedPlanBlocksDBWrite is the atomicity proof risk
// R-A/wave 2 requires: when the active plan is missing the Progress
// section entirely, the whole operation fails BEFORE the changeset/DB
// write — the common failure mode has zero side effects, so index and
// markdown cannot diverge from it.
func TestCreateTraceMalformedPlanBlocksDBWrite(t *testing.T) {
	chdirFixture(t)
	planPath := writeActivePlanFixture(t, "demo")
	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	corrupted := strings.Replace(string(data), "## Progress", "## Renamed Somehow", 1)
	if err := os.WriteFile(planPath, []byte(corrupted), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	_, _, err = CreateTrace(db, changesetDir, 1, "wave 1 done", runID, "", "")
	if err == nil {
		t.Fatal("CreateTrace = nil error, want a plan-section-not-found failure")
	}
	if got := countRows(t, db, "traces"); got != 0 {
		t.Fatalf("traces rows = %d, want 0 — DB write must not proceed when the plan can't be written to", got)
	}
	after, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("ReadFile after failed CreateTrace: %v", err)
	}
	if string(after) != corrupted {
		t.Fatalf("plan file changed despite the failed write:\nbefore=%s\nafter=%s", corrupted, string(after))
	}
}
