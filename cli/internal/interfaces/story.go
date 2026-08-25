package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newStoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "story",
		Short: "Record a new story (phase); slug = phase slug",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, _ := cmd.Flags().GetString("slug")
			goal, _ := cmd.Flags().GetString("goal")
			dependsOn, _ := cmd.Flags().GetString("depends-on")
			return runStory(cmd, slug, goal, dependsOn)
		},
	}
	cmd.Flags().String("slug", "", "phase slug")
	cmd.Flags().String("goal", "", "phase goal")
	cmd.Flags().String("depends-on", "", "slug of a prerequisite phase (optional)")
	return cmd
}

func runStory(cmd *cobra.Command, slug, goal, dependsOn string) error {
	if !infrastructure.Exists(resolveDBPath()) {
		return missingDBError("story")
	}
	db, err := infrastructure.Open(resolveDBPath())
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("story: %v", err))
	}
	defer db.Close()

	id, err := application.CreateStory(db, slug, goal, dependsOn)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_not_writable", fmt.Sprintf("story: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
			"id":     id,
			"slug":   slug,
			"status": domain.StoryPlanned,
		})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "story %s (%s) created\n", id, slug)
	return nil
}
