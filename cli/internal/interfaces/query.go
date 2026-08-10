package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query <view>",
		Short: "Read-only views: state, phases, artifacts, check, checks, traces, decisions, handoff, plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			phaseFilter, _ := cmd.Flags().GetString("phase")
			latest, _ := cmd.Flags().GetBool("latest")
			runID, _ := cmd.Flags().GetString("run-id")
			tail, _ := cmd.Flags().GetInt("tail")
			section, _ := cmd.Flags().GetString("section")
			return runQuery(cmd, args[0], phaseFilter, latest, runID, tail, section)
		},
	}
	cmd.Flags().String("phase", "", "filter the artifacts/decisions/traces/checks view by phase slug, or select the phase block for the plan view")
	cmd.Flags().Bool("latest", false, "return the most recent verdict (check view)")
	cmd.Flags().String("run-id", "", "filter the traces view to one run")
	cmd.Flags().Int("tail", 0, "limit the traces/decisions/checks view to the N most recent entries (0 = unbounded)")
	cmd.Flags().String("section", "", "plan view: current-state or phase (phase requires --phase)")
	return cmd
}

func runQuery(cmd *cobra.Command, view, phaseFilter string, latest bool, runID string, tail int, section string) error {
	if view == "plan" {
		v, err := application.QueryPlanSection(section, phaseFilter)
		if err != nil {
			if ve, ok := err.(*domain.ValidationError); ok {
				return mapValidationError(ve)
			}
			return newSystemError("plan_unreadable", fmt.Sprintf("query plan: %v", err))
		}
		return emitQueryResult(cmd, v, fmt.Sprintf("%+v", v))
	}

	db, err := infrastructure.OpenReadOnly(dbPath)
	if infrastructure.IsDatabaseNotFound(err) {
		return newSystemError("db_unreadable", "query: no db at "+dbPath+"; run `zharness init` first")
	}
	if err != nil {
		return mapReadOnlyOpenError("query", err)
	}
	defer db.Close()
	raw := db.Raw()

	switch view {
	case "state":
		v, err := application.QueryState(raw)
		if err != nil {
			return newSystemError("db_unreadable", fmt.Sprintf("query state: %v", err))
		}
		return emitQueryResult(cmd, v, fmt.Sprintf("%+v", v))

	case "phases":
		v, err := application.QueryPhases(raw)
		if err != nil {
			return newSystemError("db_unreadable", fmt.Sprintf("query phases: %v", err))
		}
		return emitQueryResult(cmd, v, fmt.Sprintf("%+v", v))

	case "artifacts":
		v, err := application.QueryArtifacts(raw, phaseFilter)
		if err != nil {
			return newSystemError("db_unreadable", fmt.Sprintf("query artifacts: %v", err))
		}
		return emitQueryResult(cmd, v, fmt.Sprintf("%+v", v))

	case "check":
		if !latest {
			return newUserError("unknown_view", "query check: only --latest is supported")
		}
		v, ok, err := application.QueryLatestCheck(raw)
		if err != nil {
			return newSystemError("db_unreadable", fmt.Sprintf("query check --latest: %v", err))
		}
		if !ok {
			return newUserError("no_check_found", "query check --latest: no check rows found")
		}
		return emitQueryResult(cmd, v, fmt.Sprintf("%+v", v))

	case "checks":
		v, err := application.QueryChecks(raw, phaseFilter, tail)
		if err != nil {
			return newSystemError("db_unreadable", fmt.Sprintf("query checks: %v", err))
		}
		return emitQueryResult(cmd, v, fmt.Sprintf("%+v", v))

	case "traces":
		var v []application.TraceView
		var err error
		if phaseFilter != "" {
			v, err = application.QueryTracesByPhase(raw, phaseFilter, tail)
		} else {
			v, err = application.QueryTraces(raw, runID, tail)
		}
		if err != nil {
			return newSystemError("db_unreadable", fmt.Sprintf("query traces: %v", err))
		}
		return emitQueryResult(cmd, v, fmt.Sprintf("%+v", v))

	case "decisions":
		v, err := application.QueryDecisions(raw, phaseFilter, tail)
		if err != nil {
			return newSystemError("db_unreadable", fmt.Sprintf("query decisions: %v", err))
		}
		return emitQueryResult(cmd, v, fmt.Sprintf("%+v", v))

	case "handoff":
		if !latest {
			return newUserError("unknown_view", "query handoff: only --latest is supported")
		}
		v, ok, err := application.QueryLatestHandoff(raw)
		if err != nil {
			return newSystemError("db_unreadable", fmt.Sprintf("query handoff --latest: %v", err))
		}
		if !ok {
			return newUserError("no_handoff_found", "query handoff --latest: no handoff rows found")
		}
		return emitQueryResult(cmd, v, fmt.Sprintf("%+v", v))

	default:
		return newUserError("unknown_view", fmt.Sprintf("query: unknown view %q", view))
	}
}

func emitQueryResult(cmd *cobra.Command, v any, plain string) error {
	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(v)
	}
	fmt.Fprintln(cmd.OutOrStdout(), plain)
	return nil
}
