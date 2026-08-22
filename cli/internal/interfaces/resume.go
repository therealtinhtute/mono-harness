package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

var resumeFacts string

func newResumeCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume",
		Short: "Snapshot of current position, drift, and readiness",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResume(cmd, version)
		},
	}
	cmd.Flags().StringVar(&resumeFacts, "facts", "", "git/WIP facts JSON — renders the full Vietnamese Recap text (mutually exclusive with --json)")
	return cmd
}

// runResume implements CONTRACT.md's `resume`: a missing db is a valid
// "no-harness" response, not a db_unreadable error (that's reserved for a
// db that exists but can't be opened/read).
func runResume(cmd *cobra.Command, version string) error {
	if jsonOutput && resumeFacts != "" {
		return newUserError("invalid_arguments", "resume: --facts and --json are mutually exclusive")
	}

	db, err := infrastructure.OpenReadOnly(resolveDBPath())
	if infrastructure.IsDatabaseNotFound(err) {
		return emitResume(cmd, application.ResumeView{Drift: []application.DriftFinding{}, Readiness: "no-harness"})
	}
	if err != nil {
		return mapReadOnlyOpenError("resume", err)
	}
	defer db.Close()

	view, err := application.Resume(db.Raw(), version)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("resume: %v", err))
	}
	return emitResume(cmd, view)
}

func emitResume(cmd *cobra.Command, view application.ResumeView) error {
	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(view)
	}
	if resumeFacts != "" {
		var facts application.RecapFacts
		if err := json.Unmarshal([]byte(resumeFacts), &facts); err != nil {
			return newUserError("facts_malformed", fmt.Sprintf("resume: --facts is not valid JSON: %v", err))
		}
		recap, err := application.RenderRecap(view, facts)
		if err != nil {
			if ve, ok := err.(*domain.ValidationError); ok {
				return mapValidationError(ve)
			}
			return newSystemError("recap_failed", err.Error())
		}
		fmt.Fprint(cmd.OutOrStdout(), recap)
		return nil
	}
	phase := "none"
	if view.Position.CurrentPhase != nil {
		phase = *view.Position.CurrentPhase
	}
	fmt.Fprintf(cmd.OutOrStdout(), "phase=%s readiness=%s drift=%d\n", phase, view.Readiness, len(view.Drift))
	return nil
}
