package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query <view>",
		Short: "Read-only views: state, phases, artifacts, check",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			phaseFilter, _ := cmd.Flags().GetString("phase")
			latest, _ := cmd.Flags().GetBool("latest")
			return runQuery(cmd, args[0], phaseFilter, latest)
		},
	}
	cmd.Flags().String("phase", "", "filter the artifacts view by phase slug")
	cmd.Flags().Bool("latest", false, "return the most recent verdict (check view)")
	return cmd
}

func runQuery(cmd *cobra.Command, view, phaseFilter string, latest bool) error {
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
