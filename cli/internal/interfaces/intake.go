package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newIntakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "intake",
		Short: "Record a new intake (SPEC lock trigger)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			typ, _ := cmd.Flags().GetString("type")
			summary, _ := cmd.Flags().GetString("summary")
			lane, _ := cmd.Flags().GetString("lane")
			planPath, _ := cmd.Flags().GetString("plan-path")
			planID, _ := cmd.Flags().GetString("plan-id")
			return runIntake(cmd, typ, summary, lane, planPath, planID)
		},
	}
	cmd.Flags().String("type", "", "new-spec|spec-slice|change-request|new-initiative|maintenance|harness-improvement")
	cmd.Flags().String("summary", "", "one-line summary")
	cmd.Flags().String("lane", "", "tiny|normal|high-risk")
	cmd.Flags().String("plan-path", "", "repository-relative path to the initiative's evolving plan (optional)")
	cmd.Flags().String("plan-id", "", "the plan's own ULID, same value passed to `run create --plan-id` (optional; enables lane-aware check gating)")
	return cmd
}

func runIntake(cmd *cobra.Command, typ, summary, lane, planPath, planID string) error {
	if !infrastructure.Exists(resolveDBPath()) {
		return newSystemError("db_unreadable", "intake: no db at "+resolveDBPath()+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(resolveDBPath())
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("intake: %v", err))
	}
	defer db.Close()

	id, err := application.CreateIntake(db, typ, summary, lane, planPath, planID)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_not_writable", fmt.Sprintf("intake: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"id": id})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "intake %s created\n", id)
	return nil
}
