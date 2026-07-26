package interfaces

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

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

	applied, skipped, err := infrastructure.ApplyChangeset(db, path)
	if err != nil {
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
	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "db changeset status: no db at "+dbPath+"; run `zharness init` first")
	}

	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("db changeset status: %v", err))
	}
	defer db.Close()

	pending, appliedCount, lastApplied, err := infrastructure.ChangesetStatus(db, changesetDir)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("db changeset status: %v", err))
	}
	if pending == nil {
		pending = []string{}
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
			"pending":       pending,
			"applied_count": appliedCount,
			"last_applied":  lastApplied,
		})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "pending=%v applied_count=%d last_applied=%q\n", pending, appliedCount, lastApplied)
	return nil
}
