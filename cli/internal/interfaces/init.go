package interfaces

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/embedded"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

const dbPath = ".kit/harness.db"
const kitDir = ".kit"

func newInitCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create the zharness database and scaffold .kit/docs if missing",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")
			refreshDocs, _ := cmd.Flags().GetBool("refresh-docs")
			return runInit(cmd, force, refreshDocs, version)
		},
	}
	cmd.Flags().Bool("force", false, "reinitialize an empty db if one already exists")
	cmd.Flags().Bool("refresh-docs", false, "rewrite .kit/docs from the embedded doc set and re-stamp docs_version, even if docs already exist")
	return cmd
}

// runInit implements CONTRACT.md's `init`: safe-by-default (an existing db
// is left untouched and reported as "exists"; `--force` wipes and recreates
// it) plus doc scaffolding. Doc scaffolding runs independently of db status
// per the idempotency matrix in cli-embed-scaffold-CONTEXT.md — a missing
// {kitDir}/docs is always filled in, an existing one is left untouched unless
// `--refresh-docs` forces a canonical overwrite.
func runInit(cmd *cobra.Command, force, refreshDocs bool, version string) error {
	if err := os.MkdirAll(kitDir, 0o755); err != nil {
		return newSystemError("db_not_writable", fmt.Sprintf("init: %v", err))
	}

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

	scaffold, err := application.ScaffoldDocs(db, changesetDir, ".", kitDir, embedded.FS, version, refreshDocs)
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
	if scaffold.DocsWritten {
		fmt.Fprintf(cmd.OutOrStdout(), "scaffolded %s/docs (docs_version=%s)\n", kitDir, scaffold.DocsVersion)
	}
	if scaffold.AgentsShimWritten {
		fmt.Fprintln(cmd.OutOrStdout(), "wrote AGENTS.md shim at repo root")
	} else if scaffold.AgentsShimNoticePath != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "AGENTS.md already exists at repo root; canonical shim content is at %s\n", scaffold.AgentsShimNoticePath)
	}
	if scaffold.GitignoreUpdated {
		fmt.Fprintln(cmd.OutOrStdout(), "updated .gitignore")
	}
	return nil
}
