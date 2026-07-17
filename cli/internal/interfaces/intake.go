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
			return runIntake(cmd, typ, summary, lane)
		},
	}
	cmd.Flags().String("type", "", "new-spec|spec-slice|change-request|new-initiative|maintenance|harness-improvement")
	cmd.Flags().String("summary", "", "one-line summary")
	cmd.Flags().String("lane", "", "tiny|normal|high-risk")
	return cmd
}

func runIntake(cmd *cobra.Command, typ, summary, lane string) error {
	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "intake: no db at "+dbPath+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("intake: %v", err))
	}
	defer db.Close()

	id, _, err := application.CreateIntake(db, changesetDir, typ, summary, lane)
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
