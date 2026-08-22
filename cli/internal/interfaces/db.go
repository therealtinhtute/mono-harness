package interfaces

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// dbRebuildActivePlanGlob mirrors preflightActivePlanGlob (preflight.go) —
// duplicated rather than shared so this wave's db.go change doesn't touch
// preflight.go's own const.
const dbStatusActivePlanGlob = "docs/plans/active/*.md"

func newDBCmd() *cobra.Command {
	db := &cobra.Command{
		Use:   "db",
		Short: "Database operations",
	}

	rebuild := &cobra.Command{
		Use:   "rebuild",
		Short: "Delete the database and rebuild it from committed plan markdown alone",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, _ := cmd.Flags().GetBool("yes")
			return runDBRebuild(cmd, yes)
		},
	}
	rebuild.Flags().Bool("yes", false, "confirm deleting harness.db and its WAL/SHM sidecars before rebuilding")
	db.AddCommand(rebuild)

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Report schema version and per-table row counts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDBStatus(cmd)
		},
	}
	db.AddCommand(statusCmd)

	return db
}

// runDBRebuild deletes harness.db and its WAL/SHM sidecars, then
// re-migrates and reconstructs every table from committed plan markdown
// under docs/plans/{active,completed}/*.md alone
// (P3, docs/plans/active/harness-markdown-truth.md: markdown is the
// source of truth, the db is a rebuildable derived index). It touches
// nothing under docs/ — unlike `init`, this is a database-only operation,
// so it never re-scaffolds managed docs as a side effect. Requires --yes:
// rebuilding is destructive to any DB-only state that never made it into
// markdown (NG4).
func runDBRebuild(cmd *cobra.Command, yes bool) error {
	if !yes {
		return newUserError("confirmation_required",
			"db rebuild: pass --yes to confirm deleting harness.db (and its -wal/-shm sidecars) and rebuilding it from committed plan markdown alone")
	}

	for _, path := range []string{resolveDBPath(), resolveDBPath() + "-wal", resolveDBPath() + "-shm"} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return newSystemError("db_not_writable", fmt.Sprintf("db rebuild: %v", err))
		}
	}

	db, err := infrastructure.Open(resolveDBPath())
	if err != nil {
		return newSystemError("db_not_writable", fmt.Sprintf("db rebuild: %v", err))
	}
	defer db.Close()

	_, schemaVersion, err := infrastructure.Migrate(db)
	if err != nil {
		return newSystemError("db_not_writable", fmt.Sprintf("db rebuild: %v", err))
	}

	result, err := application.RebuildFromMarkdown(db)
	if err != nil {
		return newSystemError("markdown_rebuild_failed", fmt.Sprintf("db rebuild: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
			"status":         "rebuilt",
			"schema_version": schemaVersion,
			"stories":        result.Stories,
			"intakes":        result.Intakes,
			"runs":           result.Runs,
			"checks":         result.Checks,
			"handoffs":       result.Handoffs,
			"traces":         result.Traces,
			"decisions":      result.Decisions,
			"memories":       result.Memories,
		})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "rebuilt %s (schema_version=%d, stories=%d, intakes=%d, runs=%d, checks=%d, handoffs=%d, traces=%d, decisions=%d, memories=%d)\n",
		resolveDBPath(), schemaVersion, result.Stories, result.Intakes, result.Runs, result.Checks, result.Handoffs, result.Traces, result.Decisions, result.Memories)
	return nil
}

func runDBStatus(cmd *cobra.Command) error {
	db, err := infrastructure.OpenReadOnly(resolveDBPath())
	if infrastructure.IsDatabaseNotFound(err) {
		return newSystemError("db_unreadable", "db status: no db at "+resolveDBPath()+"; run `zharness init` first")
	}
	if err != nil {
		return mapReadOnlyOpenError("db status", err)
	}
	defer db.Close()

	view, err := application.QueryDBStatus(db.Raw(), filepath.Join(docsDir, "playbooks"), dbStatusActivePlanGlob)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("db status: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(view)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "schema_version=%d rows=%v\n", view.SchemaVersion, view.Rows)
	return nil
}
