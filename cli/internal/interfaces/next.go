package interfaces

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newNextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "next [argument]",
		Short: "Resolve work.md's mode + active-phase routing (auto/simple/full[/phase <slug>])",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNext(cmd, strings.Join(args, " "))
		},
	}
}

// runNext mirrors resume.go's "missing db is a valid state" handling: a
// harness that hasn't been initialized yet just means every roadmap phase
// counts as not-yet-done, not a db_unreadable error.
func runNext(cmd *cobra.Command, argument string) error {
	conn, err := infrastructure.OpenReadOnly(dbPath)
	if infrastructure.IsDatabaseNotFound(err) {
		view, err := application.Next(nil, argument)
		if err != nil {
			if ve, ok := err.(*domain.ValidationError); ok {
				return mapValidationError(ve)
			}
			return newSystemError("next_failed", err.Error())
		}
		return emitNext(cmd, view)
	}
	if err != nil {
		return mapReadOnlyOpenError("next", err)
	}
	defer conn.Close()

	view, err := application.Next(conn.Raw(), argument)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("next_failed", err.Error())
	}
	return emitNext(cmd, view)
}

func emitNext(cmd *cobra.Command, view application.NextView) error {
	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(view)
	}
	phase := "none"
	if view.ActivePhase != nil {
		phase = *view.ActivePhase
	}
	stop := "none"
	if view.Stop != nil {
		stop = view.Stop.Code
	}
	fmt.Fprintf(cmd.OutOrStdout(), "mode=%s active_phase=%s stop=%s\n", view.Mode, phase, stop)
	return nil
}
