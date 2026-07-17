package interfaces

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import [path]",
		Short: "Parse legacy .kit/ state into changesets + DB rows (defaults to .kit/)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			legacyDir := ".kit/"
			if len(args) == 1 {
				legacyDir = args[0]
			}
			return runImport(cmd, legacyDir)
		},
	}
}

func runImport(cmd *cobra.Command, legacyDir string) error {
	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "import: no db at "+dbPath+"; run `zharness init` first")
	}

	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("import: %v", err))
	}
	defer db.Close()

	result, err := application.Import(db, legacyDir, changesetDir)
	if err != nil {
		var unmapped *application.ErrLegacyFieldUnmapped
		if errors.As(err, &unmapped) {
			return newUserError("legacy_field_unmapped", fmt.Sprintf("import: %v", unmapped))
		}
		return newSystemError("import_failed", fmt.Sprintf("import: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "imported=%d skipped=%d changesets_written=%v\n",
		result.Imported, result.Skipped, result.ChangesetsWritten)
	return nil
}
