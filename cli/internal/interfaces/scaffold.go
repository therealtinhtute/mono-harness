package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/embedded"
)

func newScaffoldCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scaffold <run|check|handoff|spec>",
		Short: "Emit an artifact skeleton (frontmatter + section headers) for the agent to fill",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, _ := cmd.Flags().GetString("path")
			return runScaffold(cmd, args[0], path)
		},
	}
	cmd.Flags().String("path", "", "destination file path for the skeleton")
	return cmd
}

func runScaffold(cmd *cobra.Command, kind, path string) error {
	data, err := application.ScaffoldArtifact(embedded.Templates, kind, path)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("scaffold_failed", err.Error())
	}
	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"kind": kind, "path": path, "bytes": len(data)})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "scaffolded %s → %s\n", kind, path)
	return nil
}
