package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newBacklogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backlog",
		Short: "Record a backlog item (general-purpose, ad hoc)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			summary, _ := cmd.Flags().GetString("summary")
			priority, _ := cmd.Flags().GetString("priority")
			return runBacklog(cmd, summary, priority)
		},
	}
	cmd.Flags().String("summary", "", "one-line summary")
	cmd.Flags().String("priority", "", "tiny|normal|high-risk (optional)")
	return cmd
}

func runBacklog(cmd *cobra.Command, summary, priority string) error {
	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "backlog: no db at "+dbPath+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("backlog: %v", err))
	}
	defer db.Close()

	id, _, err := application.CreateBacklog(db, changesetDir, summary, priority)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_not_writable", fmt.Sprintf("backlog: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"id": id})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "backlog %s created\n", id)
	return nil
}
