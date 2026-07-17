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
	cmd := &cobra.Command{
		Use:   "decision",
		Short: "Record a decision (general-purpose, ad hoc)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			summary, _ := cmd.Flags().GetString("summary")
			rationale, _ := cmd.Flags().GetString("rationale")
			rejected, _ := cmd.Flags().GetString("rejected")
			return runDecision(cmd, summary, rationale, rejected)
		},
	}
	cmd.Flags().String("summary", "", "one-line summary")
	cmd.Flags().String("rationale", "", "why this decision was made")
	cmd.Flags().String("rejected", "", "rejected alternatives, if any")
	return cmd
}

func runDecision(cmd *cobra.Command, summary, rationale, rejected string) error {
	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "decision: no db at "+dbPath+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("decision: %v", err))
	}
	defer db.Close()

	id, _, err := application.CreateDecision(db, changesetDir, summary, rationale, rejected)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_not_writable", fmt.Sprintf("decision: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"id": id})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "decision %s created\n", id)
	return nil
}
