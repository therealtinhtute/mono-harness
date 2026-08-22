package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newPlanCmd() *cobra.Command {
	plan := &cobra.Command{
		Use:   "plan",
		Short: "Active-plan lifecycle operations",
	}

	complete := &cobra.Command{
		Use:   "complete",
		Short: "Move the active plan to docs/plans/completed/ with status: completed",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPlanComplete(cmd)
		},
	}

	abandon := &cobra.Command{
		Use:   "abandon",
		Short: "Move the active plan to docs/plans/completed/ with status: abandoned",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			reason, _ := cmd.Flags().GetString("reason")
			return runPlanAbandon(cmd, reason)
		},
	}
	abandon.Flags().String("reason", "", "why this plan will never ship (required)")

	plan.AddCommand(complete)
	plan.AddCommand(abandon)
	return plan
}

func runPlanComplete(cmd *cobra.Command) error {
	if !infrastructure.Exists(resolveDBPath()) {
		return newSystemError("db_unreadable", "plan complete: no db at "+resolveDBPath()+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(resolveDBPath())
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("plan complete: %v", err))
	}
	defer db.Close()

	dest, stop, err := application.PlanComplete(db)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("plan_unwritable", fmt.Sprintf("plan complete: %v", err))
	}
	if stop != nil {
		return mapStop(stop)
	}
	return emitPlanTransition(cmd, dest)
}

func runPlanAbandon(cmd *cobra.Command, reason string) error {
	dest, stop, err := application.PlanAbandon(reason)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("plan_unwritable", fmt.Sprintf("plan abandon: %v", err))
	}
	if stop != nil {
		return mapStop(stop)
	}
	return emitPlanTransition(cmd, dest)
}

func emitPlanTransition(cmd *cobra.Command, dest string) error {
	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"path": dest})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "plan moved to %s\n", dest)
	return nil
}
