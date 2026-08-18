package interfaces

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/embedded"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newInitCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create the root zharness database and safely scaffold managed docs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")
			refreshDocs, _ := cmd.Flags().GetBool("refresh-docs")
			forceDocs, _ := cmd.Flags().GetBool("force-docs")
			return runInit(cmd, force, refreshDocs, forceDocs, version)
		},
	}
	cmd.Flags().Bool("force", false, "reinitialize the root database if one already exists")
	cmd.Flags().Bool("refresh-docs", false, "safely refresh managed root docs from the embedded doc set")
	cmd.Flags().Bool("force-docs", false, "overwrite locally changed managed docs during refresh")
	return cmd
}

// runInit creates or migrates the one root database, then projects managed
// workflow docs with hash-based conflict protection. A legacy .kit database
// must use the explicit layout migrator so init cannot silently fork state.
func runInit(cmd *cobra.Command, force, refreshDocs, forceDocs bool, version string) error {
	if !infrastructure.Exists(dbPath) && infrastructure.Exists(legacyDBPath) {
		return newUserError("layout_migration_required", "init: legacy database found at "+legacyDBPath+"; run `zharness migrate layout --to v2`")
	}
	if err := os.MkdirAll(kitDir, 0o755); err != nil {
		return newSystemError("db_not_writable", fmt.Sprintf("init: %v", err))
	}

	status := "created"
	if infrastructure.Exists(dbPath) {
		if !force {
			status = "exists"
		} else {
			for _, path := range []string{dbPath, dbPath + "-wal", dbPath + "-shm"} {
				if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
					return newSystemError("db_not_writable", fmt.Sprintf("init: %v", err))
				}
			}
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

	scaffold, err := application.ScaffoldDocs(db, ".", kitDir, embedded.FS, version, refreshDocs, forceDocs)
	if err != nil {
		var validation *domain.ValidationError
		if errors.As(err, &validation) {
			return mapValidationError(validation)
		}
		var conflict *application.ManagedDocsConflictError
		if errors.As(err, &conflict) {
			return newUserError("docs_conflict", fmt.Sprintf("init: %v; inspect %s", conflict, conflictDir))
		}
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
	if scaffold.DocsWritten {
		fmt.Fprintf(cmd.OutOrStdout(), "scaffolded managed %s (docs_version=%s)\n", docsDir, scaffold.DocsVersion)
	}
	if scaffold.AgentsShimWritten {
		fmt.Fprintln(cmd.OutOrStdout(), "updated AGENTS.md managed block")
	}
	if scaffold.GitignoreUpdated {
		fmt.Fprintln(cmd.OutOrStdout(), "updated .gitignore")
	}
	return nil
}
