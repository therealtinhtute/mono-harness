package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newTraceCmd() *cobra.Command {
	trace := &cobra.Command{
		Use:   "trace",
		Short: "Trace operations",
	}

	add := &cobra.Command{
		Use:   "add",
		Short: "Record a trace entry (fires at each wave completion, or per task for finer-grained history)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			wave, _ := cmd.Flags().GetInt("wave")
			summary, _ := cmd.Flags().GetString("summary")
			runID, _ := cmd.Flags().GetString("run-id")
			task, _ := cmd.Flags().GetString("task")
			taskStatus, _ := cmd.Flags().GetString("task-status")
			tasksRaw, _ := cmd.Flags().GetString("tasks")
			return runTraceAdd(cmd, wave, summary, runID, task, taskStatus, tasksRaw)
		},
	}
	add.Flags().Int("wave", 0, "wave number")
	add.Flags().String("summary", "", "one-line summary (single-entry form)")
	add.Flags().String("run-id", "", "ulid of the associated run (optional, shared across a batch)")
	add.Flags().String("task", "", "task name, for a task-level trace entry (single-entry form; optional)")
	add.Flags().String("task-status", "", "DONE|DONE_WITH_CONCERNS|NEEDS_CONTEXT|BLOCKED (single-entry form; optional)")
	add.Flags().String("tasks", "", `JSON array for a batched flush, 1-20 entries: [{"task":"...","task_status":"DONE|DONE_WITH_CONCERNS|NEEDS_CONTEXT|BLOCKED","summary":"..."}, ...]; mutually exclusive with --task/--task-status/--summary`)

	trace.AddCommand(add)
	return trace
}

func runTraceAdd(cmd *cobra.Command, wave int, summary, runID, task, taskStatus, tasksRaw string) error {
	if tasksRaw != "" {
		return runTraceAddBatch(cmd, wave, runID, task, taskStatus, summary, tasksRaw)
	}

	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "trace add: no db at "+dbPath+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("trace add: %v", err))
	}
	defer db.Close()

	id, err := application.CreateTrace(db, wave, summary, runID, task, taskStatus)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_not_writable", fmt.Sprintf("trace add: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"id": id})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "trace %s created\n", id)
	return nil
}

func runTraceAddBatch(cmd *cobra.Command, wave int, runID, task, taskStatus, summary, tasksRaw string) error {
	if task != "" || taskStatus != "" || summary != "" {
		return newUserError("invalid_arguments", "trace add: --tasks is mutually exclusive with --task/--task-status/--summary")
	}
	var tasks []domain.TraceTask
	if err := json.Unmarshal([]byte(tasksRaw), &tasks); err != nil {
		return newUserError("invalid_tasks", fmt.Sprintf("trace add: --tasks is not valid JSON: %v", err))
	}

	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "trace add: no db at "+dbPath+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("trace add: %v", err))
	}
	defer db.Close()

	ids, err := application.CreateTraces(db, wave, runID, tasks)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_not_writable", fmt.Sprintf("trace add: %v", err))
	}

	if jsonOutput {
		results := make([]map[string]string, len(ids))
		for i, id := range ids {
			results[i] = map[string]string{"id": id}
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(results)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "traces %v created\n", ids)
	return nil
}
