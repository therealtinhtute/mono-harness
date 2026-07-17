package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newInterventionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "intervention",
		Short: "Record a human override of a check verdict",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			verdictID, _ := cmd.Flags().GetString("verdict-id")
			reason, _ := cmd.Flags().GetString("reason")
			return runIntervention(cmd, verdictID, reason)
		},
	}
	cmd.Flags().String("verdict-id", "", "ulid of the check being overridden")
	cmd.Flags().String("reason", "", "why the override is acceptable")
	return cmd
}

func runIntervention(cmd *cobra.Command, verdictID, reason string) error {
	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "intervention: no db at "+dbPath+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("intervention: %v", err))
	}
	defer db.Close()

	id, _, err := application.CreateIntervention(db, changesetDir, verdictID, reason)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_not_writable", fmt.Sprintf("intervention: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"id": id})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "intervention %s created\n", id)
	return nil
}
