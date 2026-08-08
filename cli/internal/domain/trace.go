package domain

// Task status enum (CONTRACT.md `trace add`, migration 0008_trace_task_granularity):
// docs/playbooks/work.md's Status Routing table — the only four values a
// task execution status may hold in `## Progress`.
const (
	TaskStatusDone             = "DONE"
	TaskStatusDoneWithConcerns = "DONE_WITH_CONCERNS"
	TaskStatusNeedsContext     = "NEEDS_CONTEXT"
	TaskStatusBlocked          = "BLOCKED"
)

var taskStatuses = map[string]bool{
	TaskStatusDone:             true,
	TaskStatusDoneWithConcerns: true,
	TaskStatusNeedsContext:     true,
	TaskStatusBlocked:          true,
}

func IsValidTaskStatus(status string) bool {
	return taskStatuses[status]
}

// Trace is one wave- or task-level entry: the compressed-index counterpart
// of a plan's `## Progress` markdown entries. Task and TaskStatus are both
// optional — a wave-level trace (work.md step 9, fired once per wave) omits
// them; a task-level trace (step 7, one per attempted task) sets both.
type Trace struct {
	ID         string
	RunID      *string
	Wave       int
	Summary    string
	Task       *string
	TaskStatus *string
	CreatedAt  string
}

func (t Trace) Validate() error {
	if t.Summary == "" {
		return &ValidationError{Code: "missing_required_field", Message: "trace: summary is required"}
	}
	if t.Wave < 0 {
		// Not CONTRACT.md-documented (only unknown_run_id is listed for
		// `trace`): a negative --wave is a cobra int-parse-shaped mistake,
		// not a contract-classified failure.
		return &ValidationError{Message: "trace: wave must be >= 0"}
	}
	if t.TaskStatus != nil && *t.TaskStatus != "" && !IsValidTaskStatus(*t.TaskStatus) {
		return &ValidationError{Code: "invalid_task_status", Message: "trace: invalid task_status " + *t.TaskStatus}
	}
	return nil
}
