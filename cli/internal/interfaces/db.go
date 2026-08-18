package interfaces

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// dbRebuildActivePlanGlob mirrors preflightActivePlanGlob (preflight.go) —
// duplicated rather than shared so this wave's db.go change doesn't touch
// preflight.go's own const.
const dbStatusActivePlanGlob = "docs/plans/active/*.md"

func newDBCmd() *cobra.Command {
	db := &cobra.Command{
		Use:   "db",
		Short: "Database and changeset operations",
	}

	changeset := &cobra.Command{
		Use:   "changeset",
		Short: "Changeset apply/status operations",
	}

	changeset.AddCommand(&cobra.Command{
		Use:   "apply <path>",
		Short: "Apply a .jsonl changeset file to the db (idempotent)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDBChangesetApply(cmd, args[0])
		},
	})

	changeset.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show pending/applied changeset status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDBChangesetStatus(cmd)
		},
	})

	db.AddCommand(changeset)

	rebuild := &cobra.Command{
		Use:   "rebuild",
		Short: "Delete the database and rebuild it from committed changesets alone",
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
		Short: "Report schema version, fence, per-table row counts, and true pending changesets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDBStatus(cmd)
		},
	}
	db.AddCommand(statusCmd)

	return db
}

func runDBChangesetApply(cmd *cobra.Command, path string) error {
	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "db changeset apply: no db at "+dbPath+"; run `zharness init` first")
	}

	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("db changeset apply: %v", err))
	}
	defer db.Close()

	applied, skipped, err := application.ApplyChangesetForRecovery(db, changesetDir, path)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		var outOfOrder *infrastructure.ErrOutOfOrder
		if errors.As(err, &outOfOrder) {
			return newUserError("changeset_out_of_order", fmt.Sprintf("db changeset apply: %v", outOfOrder))
		}
		return newUserError("changeset_malformed", fmt.Sprintf("db changeset apply: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
			"applied":                 applied,
			"skipped_already_applied": skipped,
		})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "applied=%d skipped_already_applied=%d\n", applied, skipped)
	return nil
}

func runDBChangesetStatus(cmd *cobra.Command) error {
	db, err := infrastructure.OpenReadOnly(dbPath)
	if infrastructure.IsDatabaseNotFound(err) {
		return newSystemError("db_unreadable", "db changeset status: no db at "+dbPath+"; run `zharness init` first")
	}
	if err != nil {
		return mapReadOnlyOpenError("db changeset status", err)
	}
	defer db.Close()

	pending, appliedCount, lastApplied, unverifiedBelowFence, err := infrastructure.ChangesetStatus(db.Raw(), changesetDir)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("db changeset status: %v", err))
	}
	if pending == nil {
		pending = []string{}
	}
	if unverifiedBelowFence == nil {
		unverifiedBelowFence = []string{}
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
			"pending":                pending,
			"applied_count":          appliedCount,
			"last_applied":           lastApplied,
			"unverified_below_fence": unverifiedBelowFence,
		})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "pending=%v applied_count=%d last_applied=%q unverified_below_fence=%v\n",
		pending, appliedCount, lastApplied, unverifiedBelowFence)
	return nil
}

// runDBRebuild deletes harness.db and its WAL/SHM sidecars, then
// re-migrates and reconstructs every table from committed plan markdown
// under docs/plans/{active,completed}/*.md alone — no read of changesetDir
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

	for _, path := range []string{dbPath, dbPath + "-wal", dbPath + "-shm"} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return newSystemError("db_not_writable", fmt.Sprintf("db rebuild: %v", err))
		}
	}

	db, err := infrastructure.Open(dbPath)
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
		})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "rebuilt %s (schema_version=%d, stories=%d, intakes=%d, runs=%d, checks=%d, handoffs=%d, traces=%d, decisions=%d)\n",
		dbPath, schemaVersion, result.Stories, result.Intakes, result.Runs, result.Checks, result.Handoffs, result.Traces, result.Decisions)
	return nil
}

func runDBStatus(cmd *cobra.Command) error {
	db, err := infrastructure.OpenReadOnly(dbPath)
	if infrastructure.IsDatabaseNotFound(err) {
		return newSystemError("db_unreadable", "db status: no db at "+dbPath+"; run `zharness init` first")
	}
	if err != nil {
		return mapReadOnlyOpenError("db status", err)
	}
	defer db.Close()

	view, err := application.QueryDBStatus(db.Raw(), changesetDir, filepath.Join(docsDir, "playbooks"), dbStatusActivePlanGlob)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("db status: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(view)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "schema_version=%d fence=%q rows=%v pending=%v unverified_below_fence=%v\n",
		view.SchemaVersion, view.Fence, view.Rows, view.Pending, view.UnverifiedBelowFence)
	return nil
}
