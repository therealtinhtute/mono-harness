package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newAuditCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "audit",
		Short: "Report pointer drift and lifecycle contract violations",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAudit(cmd, version)
		},
	}
}

func runAudit(cmd *cobra.Command, version string) error {
	db, err := infrastructure.OpenReadOnly(dbPath)
	if infrastructure.IsDatabaseNotFound(err) {
		return newSystemError("db_unreadable", "audit: no db at "+dbPath+"; run `zharness init` first")
	}
	if err != nil {
		return mapReadOnlyOpenError("audit", err)
	}
	defer db.Close()

	report, err := application.Audit(db.Raw(), version, ".")
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("audit: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(report)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "pointer_drift=%d contract_violations=%d unlinked_proofs=%d\n",
		len(report.PointerDrift), len(report.ContractViolations), len(report.UnlinkedProofs))
	return nil
}
