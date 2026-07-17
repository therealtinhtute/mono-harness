package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newToolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool",
		Short: "Record a tool/capability usage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			purpose, _ := cmd.Flags().GetString("purpose")
			return runTool(cmd, name, purpose)
		},
	}
	cmd.Flags().String("name", "", "tool name")
	cmd.Flags().String("purpose", "", "why this tool was used")
	return cmd
}

func runTool(cmd *cobra.Command, name, purpose string) error {
	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "tool: no db at "+dbPath+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("tool: %v", err))
	}
	defer db.Close()

	id, _, err := application.CreateTool(db, changesetDir, name, purpose)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_not_writable", fmt.Sprintf("tool: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"id": id})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "tool %s created\n", id)
	return nil
}
