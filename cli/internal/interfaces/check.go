package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newCheckCmd() *cobra.Command {
	check := &cobra.Command{
		Use:   "check",
		Short: "Check (gate verdict) operations",
	}

	record := &cobra.Command{
		Use:   "record",
		Short: "Record a gate verdict (deterministic, no free-text-only verdicts)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			verdict, _ := cmd.Flags().GetString("verdict")
			runID, _ := cmd.Flags().GetString("run-id")
			judge, _ := cmd.Flags().GetString("judge")
			judgeModel, _ := cmd.Flags().GetString("judge-model")
			proofLinksRaw, _ := cmd.Flags().GetString("proof-links")
			mode, _ := cmd.Flags().GetString("mode")
			return runCheckRecord(cmd, verdict, runID, judge, judgeModel, proofLinksRaw, mode)
		},
	}
	record.Flags().String("verdict", "", "APPROVED|APPROVE_WITH_REQUESTS|REQUEST_CHANGES")
	record.Flags().String("run-id", "", "ulid of the run being gated")
	record.Flags().String("judge", "", "independent|same-session")
	record.Flags().String("judge-model", "", "identifier of the model that produced the verdict")
	record.Flags().String("proof-links", "[]", `JSON array: [{"command":"...","output_ref":"...","artifact_path":"..."}]`)
	record.Flags().String("mode", domain.CheckModeGate, "gate|full — which check playbook mode produced the verdict")

	check.AddCommand(record)
	return check
}

func runCheckRecord(cmd *cobra.Command, verdict, runID, judge, judgeModel, proofLinksRaw, mode string) error {
	var proofLinks []domain.ProofLink
	if err := json.Unmarshal([]byte(proofLinksRaw), &proofLinks); err != nil {
		return newUserError("invalid_proof_links", fmt.Sprintf("check record: --proof-links is not valid JSON: %v", err))
	}
	if !domain.IsValidCheckMode(mode) {
		return newUserError("invalid_check_mode", fmt.Sprintf("check record: --mode must be one of gate, full, got %q", mode))
	}

	if !infrastructure.Exists(resolveDBPath()) {
		return missingDBError("check record")
	}
	db, err := infrastructure.Open(resolveDBPath())
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("check record: %v", err))
	}
	defer db.Close()

	id, err := application.RecordCheck(db, runID, verdict, judge, judgeModel, proofLinks, mode)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_not_writable", fmt.Sprintf("check record: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"id": id, "verdict": verdict})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "check %s recorded (%s)\n", id, verdict)
	return nil
}
