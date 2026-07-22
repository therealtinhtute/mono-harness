package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newScoreTraceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "score-trace <trace-id>",
		Short: "Score a trace's quality tier (minimal/standard/detailed)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScoreTrace(cmd, args[0])
		},
	}
}

func runScoreTrace(cmd *cobra.Command, id string) error {
	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "score-trace: no db at "+dbPath+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("score-trace: %v", err))
	}
	defer db.Close()

	score, err := application.ScoreTrace(db, id)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_unreadable", fmt.Sprintf("score-trace: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(score)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "tier=%s reasons=%v\n", score.Tier, score.Reasons)
	return nil
}
