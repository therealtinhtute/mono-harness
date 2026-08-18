package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newRunCmd() *cobra.Command {
	run := &cobra.Command{
		Use:   "run",
		Short: "Run operations",
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "Register a work run (full-mode only) and point meta.latest_run_id at it",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, _ := cmd.Flags().GetString("slug")
			artifactPath, _ := cmd.Flags().GetString("artifact-path")
			planID, _ := cmd.Flags().GetString("plan-id")
			return runRunCreate(cmd, slug, artifactPath, planID)
		},
	}
	create.Flags().String("slug", "", "story (phase) slug this run executes")
	create.Flags().String("artifact-path", "", "deprecated optional path to a legacy run markdown artifact")
	create.Flags().String("plan-id", "", "ulid of the phase PLAN this run executes (optional)")

	run.AddCommand(create)
	return run
}

func runRunCreate(cmd *cobra.Command, slug, artifactPath, planID string) error {
	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "run create: no db at "+dbPath+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("run create: %v", err))
	}
	defer db.Close()

	id, err := application.CreateRun(db, slug, artifactPath, planID)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_not_writable", fmt.Sprintf("run create: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"id": id})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "run %s created\n", id)
	return nil
}
