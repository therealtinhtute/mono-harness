package interfaces

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// kitRoot is the .kit/-equivalent directory validate walks. Per
// CONTRACT.md, `validate` takes no path argument — it always assumes
// .kit/ relative to cwd, same convention as dbPath/changesetDir.
const kitRoot = ".kit"

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Walk SPEC->PLAN->RUN->CHECK->HANDOFF cross-links and report findings",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd)
		},
	}
}

// runValidate implements CONTRACT.md's `validate`: a missing db only
// skips the freshness-vs-DB checks (mirrors resume's "no db is a valid
// state"), it never blocks the doc-to-doc cross-link walk.
func runValidate(cmd *cobra.Command) error {
	var db *sql.DB
	if infrastructure.Exists(dbPath) {
		var err error
		db, err = infrastructure.Open(dbPath)
		if err != nil {
			return newSystemError("db_unreadable", fmt.Sprintf("validate: %v", err))
		}
		defer db.Close()
	}

	result, err := application.Validate(db, kitRoot)
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

	// CONTRACT.md: "exits 1 with non-empty findings on any violation" —
	// the body above is the actual response, so a plain cliError (which
	// would print a second {"error": ...} envelope) doesn't fit here.
	if !result.Valid {
		return newSilentExit(1)
	}
	return nil
}
