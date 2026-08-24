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

// mapAuditError renders an application.Audit failure (R25). Audit content
// failures — an unreadable docs tree, a failing git subprocess — are not
// database problems, so they carry their own code instead of db_unreadable,
// and the "audit:" prefix is applied exactly once, here.
func mapAuditError(err error) *cliError {
	return newSystemError("audit_failed", fmt.Sprintf("audit: %v", err))
}

func runAudit(cmd *cobra.Command, version string) error {
	db, err := infrastructure.OpenReadOnly(resolveDBPath())
	if infrastructure.IsDatabaseNotFound(err) {
		return newSystemError("db_unreadable", "audit: no db at "+resolveDBPath()+"; run `zharness init` first")
	}
	if err != nil {
		return mapReadOnlyOpenError("audit", err)
	}
	defer db.Close()

	report, err := application.Audit(db.Raw(), version, ".")
	if err != nil {
		return mapAuditError(err)
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(report)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "pointer_drift=%d contract_violations=%d unlinked_proofs=%d\n",
		len(report.PointerDrift), len(report.ContractViolations), len(report.UnlinkedProofs))
	return nil
}
