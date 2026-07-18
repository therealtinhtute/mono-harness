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
		Short: "Report pointer drift, contract violations, unlinked proofs, and an entropy score",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAudit(cmd, version)
		},
	}
}

func runAudit(cmd *cobra.Command, version string) error {
	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "audit: no db at "+dbPath+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("audit: %v", err))
	}
	defer db.Close()

	report, err := application.Audit(db, kitRoot, version)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("audit: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(report)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "entropy_score=%d pointer_drift=%d contract_violations=%d unlinked_proofs=%d\n",
		report.EntropyScore, len(report.PointerDrift), len(report.ContractViolations), len(report.UnlinkedProofs))
	return nil
}

func newProposeCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "propose",
		Short: "Reserved: suggest improvements from observed audit patterns (documented only, not adopted into any skill)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPropose(cmd, version)
		},
	}
}

func runPropose(cmd *cobra.Command, version string) error {
	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "propose: no db at "+dbPath+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("propose: %v", err))
	}
	defer db.Close()

	report, err := application.Propose(db, kitRoot, version)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("propose: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(report)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "proposals=%d\n", len(report.Proposals))
	return nil
}
