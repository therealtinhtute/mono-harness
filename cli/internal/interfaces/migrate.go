package interfaces

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/embedded"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newMigrateCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Apply schema migrations or migrate repository layout",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrate(cmd)
		},
	}

	var target string
	var dryRun bool
	layout := &cobra.Command{
		Use:   "layout",
		Short: "Replay legacy .kit state into the root database layout",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if target != "v2" {
				return newUserError("invalid_layout", "migrate layout: --to must be v2")
			}
			return runMigrateLayout(cmd, version, dryRun)
		},
	}
	layout.Flags().StringVar(&target, "to", "", "target layout version")
	layout.Flags().BoolVar(&dryRun, "dry-run", false, "verify replay parity without changing repository files")
	cmd.AddCommand(layout)
	return cmd
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

func runMigrateLayout(cmd *cobra.Command, version string, dryRun bool) error {
	result, err := application.MigrateLayout(".", legacyDBPath, dbPath, changesetDir, kitDir, embedded.FS, version, dryRun)
	if err != nil {
		var conflict *application.ManagedDocsConflictError
		if errors.As(err, &conflict) {
			return newUserError("docs_conflict", fmt.Sprintf("migrate layout: %v; inspect %s", conflict, conflictDir))
		}
		return newSystemError("layout_migration_failed", err.Error())
	}
	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "status=%s parity=%v replayed=%d backfilled=%d source=%s target=%s docs_written=%v\n", result.Status, result.Parity, result.Replayed, result.Backfilled, result.SourceDB, result.TargetDB, result.DocsWritten)
	return nil
}
