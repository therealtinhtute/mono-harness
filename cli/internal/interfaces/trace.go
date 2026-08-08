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
			return runTraceAdd(cmd, wave, summary, runID, task, taskStatus)
		},
	}
	add.Flags().Int("wave", 0, "wave number")
	add.Flags().String("summary", "", "one-line summary")
	add.Flags().String("run-id", "", "ulid of the associated run (optional)")
	add.Flags().String("task", "", "task name, for a task-level trace entry (optional)")
	add.Flags().String("task-status", "", "DONE|DONE_WITH_CONCERNS|NEEDS_CONTEXT|BLOCKED (optional; requires --task context, but not enforced together)")

	trace.AddCommand(add)
	return trace
}

func runTraceAdd(cmd *cobra.Command, wave int, summary, runID, task, taskStatus string) error {
	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "trace add: no db at "+dbPath+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("trace add: %v", err))
	}
	defer db.Close()

	id, _, err := application.CreateTrace(db, changesetDir, wave, summary, runID, task, taskStatus)
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
