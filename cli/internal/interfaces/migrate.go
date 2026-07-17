package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Apply all pending versioned migrations to the current schema",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrate(cmd)
		},
	}
}

func runMigrate(cmd *cobra.Command) error {
	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "migrate: no db at "+dbPath+"; run `zharness init` first")
	}

	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("migrate: %v", err))
	}
	defer db.Close()

	applied, schemaVersion, err := infrastructure.Migrate(db)
	if err != nil {
		return newSystemError("migration_conflict", fmt.Sprintf("migrate: %v", err))
	}
	if applied == nil {
		applied = []string{}
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
			"applied":        applied,
			"schema_version": schemaVersion,
		})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "applied=%v schema_version=%d\n", applied, schemaVersion)
	return nil
}
