package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newDecisionCmd() *cobra.Command {
	decision := &cobra.Command{
		Use:   "decision",
		Short: "Decision operations",
	}

	add := &cobra.Command{
		Use:   "add",
		Short: "Record one or more decisions in one transaction",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			decisionsRaw, _ := cmd.Flags().GetString("decisions")
			runID, _ := cmd.Flags().GetString("run-id")
			return runDecisionAdd(cmd, decisionsRaw, runID)
		},
	}
	add.Flags().String("decisions", "[]", `JSON array: [{"decision":"...","rationale":"...","phase":"slug","task":"..."}]`)
	add.Flags().String("run-id", "", "ulid of the associated run (optional, shared across the batch)")

	decision.AddCommand(add)
	return decision
}

func runDecisionAdd(cmd *cobra.Command, decisionsRaw, runID string) error {
	var decisions []domain.Decision
	if err := json.Unmarshal([]byte(decisionsRaw), &decisions); err != nil {
		return newUserError("invalid_decisions", fmt.Sprintf("decision add: --decisions is not valid JSON: %v", err))
	}

	if !infrastructure.Exists(resolveDBPath()) {
		return newSystemError("db_unreadable", "decision add: no db at "+resolveDBPath()+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(resolveDBPath())
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("decision add: %v", err))
	}
	defer db.Close()

	ids, err := application.RecordDecisions(db, runID, decisions)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_not_writable", fmt.Sprintf("decision add: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"ids": ids})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "decisions %v created\n", ids)
	return nil
}
