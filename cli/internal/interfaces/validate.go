package interfaces

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the durable lifecycle graph",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd)
		},
	}
}

func runValidate(cmd *cobra.Command) error {
	var raw *sql.DB
	db, err := infrastructure.OpenReadOnly(resolveDBPath())
	if infrastructure.IsDatabaseNotFound(err) {
		db = nil
	} else if err != nil {
		return mapReadOnlyOpenError("validate", err)
	} else {
		defer db.Close()
		raw = db.Raw()
	}

	result, err := application.Validate(raw)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("validate: %v", err))
	}

	if jsonOutput {
		if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
			return err
		}
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "valid=%v findings=%d\n", result.Valid, len(result.Findings))
		for _, f := range result.Findings {
			fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s: %s\n", f.Link, f.Issue, f.Detail)
		}
	}

	if !result.Valid {
		return newSilentExit(1)
	}
	return nil
}
