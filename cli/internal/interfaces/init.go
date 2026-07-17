package interfaces

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

const dbPath = ".kit/harness.db"

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create the zharness database if it doesn't exist",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")
			return runInit(cmd, force)
		},
	}
	cmd.Flags().Bool("force", false, "reinitialize an empty db if one already exists")
	return cmd
}

// runInit implements CONTRACT.md's `init`: safe-by-default (an existing
// db is left untouched and reported as "exists"); `--force` wipes and
// recreates it.
func runInit(cmd *cobra.Command, force bool) error {
	status := "created"
	if infrastructure.Exists(dbPath) {
		if !force {
			status = "exists"
		} else if err := os.Remove(dbPath); err != nil {
			return newSystemError("db_not_writable", fmt.Sprintf("init: %v", err))
		}
	}

	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_not_writable", fmt.Sprintf("init: %v", err))
	}
	defer db.Close()

	_, schemaVersion, err := infrastructure.Migrate(db)
	if err != nil {
		return newSystemError("db_not_writable", fmt.Sprintf("init: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
			"status":         status,
			"db_path":        dbPath,
			"schema_version": schemaVersion,
		})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s %s (schema_version=%d)\n", status, dbPath, schemaVersion)
	return nil
}
