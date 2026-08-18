package application

import (
	"os"
	"path/filepath"
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

// TestCreateTracesBatchAndQueryRoundTrip is the round trip for R5
// (docs/audit/sdlc-token-cache-audit.md): a batch of task-level entries
// lands as separate, individually queryable trace rows in one call.
func TestCreateTracesBatchAndQueryRoundTrip(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	ids, path, err := CreateTraces(db, changesetDir, 1, runID, []domain.TraceTask{
		{Task: "task A", TaskStatus: domain.TaskStatusDone, Summary: "did A"},
		{Task: "task B", TaskStatus: domain.TaskStatusDoneWithConcerns, Summary: "did B, minor concern"},
	})
	if err != nil {
		t.Fatalf("CreateTraces: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("CreateTraces ids = %v, want 2", ids)
	}
	if ids[0] == ids[1] {
		t.Fatalf("CreateTraces minted duplicate ids: %v", ids)
	}
	if path == "" {
		t.Fatal("CreateTraces path is empty, want a written changeset")
	}
	if got := countRows(t, db, "traces"); got != 2 {
		t.Fatalf("traces rows = %d, want 2", got)
	}

	var task, taskStatus string
	if err := db.QueryRow(`SELECT task, task_status FROM traces WHERE id = ?`, ids[0]).Scan(&task, &taskStatus); err != nil {
		t.Fatalf("query traces[0]: %v", err)
	}
	if task != "task A" || taskStatus != domain.TaskStatusDone {
		t.Fatalf("traces[0] (task, task_status) = (%q, %q), want (%q, %q)", task, taskStatus, "task A", domain.TaskStatusDone)
	}
	if err := db.QueryRow(`SELECT task, task_status FROM traces WHERE id = ?`, ids[1]).Scan(&task, &taskStatus); err != nil {
		t.Fatalf("query traces[1]: %v", err)
	}
	if task != "task B" || taskStatus != domain.TaskStatusDoneWithConcerns {
		t.Fatalf("traces[1] (task, task_status) = (%q, %q), want (%q, %q)", task, taskStatus, "task B", domain.TaskStatusDoneWithConcerns)
	}
}

func TestCreateTracesEmptyBatch(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := CreateTraces(db, changesetDir, 1, "", nil)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "empty_tasks" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: empty_tasks}", err)
	}
	if got := countRows(t, db, "traces"); got != 0 {
		t.Fatalf("traces rows = %d, want 0", got)
	}
}

func TestCreateTracesTooManyTasks(t *testing.T) {
	db, changesetDir := freshDB(t)

	tasks := make([]domain.TraceTask, 21)
	for i := range tasks {
		tasks[i] = domain.TraceTask{Task: "t", TaskStatus: domain.TaskStatusDone, Summary: "s"}
	}
	_, _, err := CreateTraces(db, changesetDir, 1, "", tasks)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "too_many_tasks" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: too_many_tasks}", err)
	}
	if got := countRows(t, db, "traces"); got != 0 {
		t.Fatalf("traces rows = %d, want 0", got)
	}
}

// TestCreateTracesBatchIsAtomicOnValidationFailure proves a batch with one
// invalid element writes nothing — not a partial batch — matching
// RecordDecisions's precedent (decision_test.go).
func TestCreateTracesBatchIsAtomicOnValidationFailure(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := CreateTraces(db, changesetDir, 1, "", []domain.TraceTask{
		{Task: "task A", TaskStatus: domain.TaskStatusDone, Summary: "did A"},
		{Task: "task B", TaskStatus: "BOGUS", Summary: "did B"},
	})
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "invalid_task_status" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: invalid_task_status}", err)
	}
	if got := countRows(t, db, "traces"); got != 0 {
		t.Fatalf("traces rows = %d, want 0 — the valid first element must not be written alone", got)
	}
}

func TestCreateTracesUnknownRunID(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := CreateTraces(db, changesetDir, 1, "01HZZZZZZZZZZZZZZZZZZZZZZZ", []domain.TraceTask{
		{Task: "task A", TaskStatus: domain.TaskStatusDone, Summary: "did A"},
	})
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_run_id" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_run_id}", err)
	}
	if got := countRows(t, db, "traces"); got != 0 {
		t.Fatalf("traces rows = %d, want 0", got)
	}
}

// TestCreateTracesWritesPlanProgressEntryPerElement proves the whole batch
// lands as separate, individually-formatted lines in the plan's
// `## Progress` section, in one operation (P3 wave 1, R5).
func TestCreateTracesWritesPlanProgressEntryPerElement(t *testing.T) {
	chdirFixture(t)
	planPath := writeActivePlanFixture(t, "demo")
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	ids, _, err := CreateTraces(db, changesetDir, 1, runID, []domain.TraceTask{
		{Task: "task A", TaskStatus: domain.TaskStatusDone, Summary: "did A"},
		{Task: "task B", TaskStatus: domain.TaskStatusDoneWithConcerns, Summary: "did B, minor concern"},
	})
	if err != nil {
		t.Fatalf("CreateTraces: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("CreateTraces ids = %v, want 2", ids)
	}

	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "task A") || !strings.Contains(content, "did A") {
		t.Fatalf("plan Progress missing first entry:\n%s", content)
	}
	if !strings.Contains(content, "task B") || !strings.Contains(content, "did B, minor concern") {
		t.Fatalf("plan Progress missing second entry:\n%s", content)
	}
	if !strings.Contains(content, "## Decisions\n<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->\n- none") {
		t.Fatalf("plan Decisions section corrupted:\n%s", content)
	}
}

// TestCreateTracesAtomicityIncludesPlanWrite extends the atomicity proof to
// the plan-write path: a malformed plan (Progress section missing) must
// block the DB write too, not just leave the plan corrupted.
func TestCreateTracesAtomicityIncludesPlanWrite(t *testing.T) {
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
	_, _, err = CreateTraces(db, changesetDir, 1, "", []domain.TraceTask{
		{Task: "task A", TaskStatus: domain.TaskStatusDone, Summary: "did A"},
	})
	if err == nil {
		t.Fatal("CreateTraces = nil error, want a plan-section-not-found failure")
	}
	if got := countRows(t, db, "traces"); got != 0 {
		t.Fatalf("traces rows = %d, want 0 — DB write must not proceed when the plan can't be written to", got)
	}
	after, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("ReadFile after failed CreateTraces: %v", err)
	}
	if string(after) != corrupted {
		t.Fatalf("plan file changed despite the failed write:\nbefore=%s\nafter=%s", corrupted, string(after))
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

// TestCreateTraceReadOnlyPlanBlocksDBWrite is R8's forced-failure proof: a
// markdown write that fails at the filesystem level (not just a malformed
// section) must still leave zero DB rows behind it
// (docs/plans/active/harness-markdown-truth.md).
func TestCreateTraceReadOnlyPlanBlocksDBWrite(t *testing.T) {
	chdirFixture(t)
	planPath := writeActivePlanFixture(t, "demo")
	makeDirReadOnly(t, filepath.Dir(planPath))

	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	_, _, err := CreateTrace(db, changesetDir, 1, "wave 1 done", runID, "", "")
	if err == nil {
		t.Fatal("CreateTrace = nil error, want a read-only plan write failure")
	}
	if got := countRows(t, db, "traces"); got != 0 {
		t.Fatalf("traces rows = %d, want 0 — DB write must not proceed when the plan write fails", got)
	}
}
