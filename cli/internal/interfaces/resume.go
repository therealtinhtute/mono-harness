package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newResumeCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "resume",
		Short: "Snapshot of current position, drift, and readiness",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResume(cmd, version)
		},
	}
}

// runResume implements CONTRACT.md's `resume`: a missing db is a valid
// "no-harness" response, not a db_unreadable error (that's reserved for a
// db that exists but can't be opened/read).
func runResume(cmd *cobra.Command, version string) error {
	if !infrastructure.Exists(dbPath) {
		return emitResume(cmd, application.ResumeView{Drift: []application.DriftFinding{}, Readiness: "no-harness"})
	}

	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("resume: %v", err))
	}
	defer db.Close()

	view, err := application.Resume(db, version)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("resume: %v", err))
	}
	return emitResume(cmd, view)
}

func emitResume(cmd *cobra.Command, view application.ResumeView) error {
	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(view)
	}
	phase := "none"
	if view.Position.CurrentPhase != nil {
		phase = *view.Position.CurrentPhase
	}
	fmt.Fprintf(cmd.OutOrStdout(), "phase=%s readiness=%s drift=%d\n", phase, view.Readiness, len(view.Drift))
	return nil
}
