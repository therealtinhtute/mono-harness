package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newMemoryCmd() *cobra.Command {
	memory := &cobra.Command{
		Use:   "memory",
		Short: "Durable cross-session memory operations",
	}

	add := &cobra.Command{
		Use:   "add",
		Short: "Record a memory entry under docs/memory/",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			memType, _ := cmd.Flags().GetString("type")
			scope, _ := cmd.Flags().GetString("scope")
			planID, _ := cmd.Flags().GetString("plan-id")
			summary, _ := cmd.Flags().GetString("summary")
			return runMemoryAdd(cmd, memType, scope, planID, summary)
		},
	}
	add.Flags().String("type", "", "memory type (free text, required)")
	add.Flags().String("scope", "", "plan|global (required)")
	add.Flags().String("plan-id", "", "initiative plan ulid (required when --scope=plan, disallowed when --scope=global)")
	add.Flags().String("summary", "", "entry body (required)")

	get := &cobra.Command{
		Use:   "get",
		Short: "Read one memory entry by id",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, _ := cmd.Flags().GetString("id")
			return runMemoryGet(cmd, id)
		},
	}
	get.Flags().String("id", "", "memory entry ulid (required)")

	query := &cobra.Command{
		Use:   "query",
		Short: "List memory entries by type, optionally filtered by scope/plan-id, or ranked by --keywords",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			memType, _ := cmd.Flags().GetString("type")
			scope, _ := cmd.Flags().GetString("scope")
			planID, _ := cmd.Flags().GetString("plan-id")
			keywords, _ := cmd.Flags().GetString("keywords")
			if keywords != "" {
				return runMemoryQueryRanked(cmd, keywords, memType, scope, planID)
			}
			return runMemoryQuery(cmd, memType, scope, planID)
		},
	}
	query.Flags().String("type", "", "memory type filter (required unless --keywords is set)")
	query.Flags().String("scope", "", "plan|global (optional)")
	query.Flags().String("plan-id", "", "initiative plan ulid (optional)")
	query.Flags().String("keywords", "", "rank results by keyword match against type+body instead of exact filtering (optional; --type becomes optional too)")

	memory.AddCommand(add, get, query)
	return memory
}

func runMemoryAdd(cmd *cobra.Command, memType, scope, planID, summary string) error {
	if !infrastructure.Exists(resolveDBPath()) {
		return newSystemError("db_unreadable", "memory add: no db at "+resolveDBPath()+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(resolveDBPath())
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("memory add: %v", err))
	}
	defer db.Close()

	id, err := application.CreateMemory(db, memType, scope, planID, summary)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_not_writable", fmt.Sprintf("memory add: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"id": id})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "memory %s created\n", id)
	return nil
}

func runMemoryGet(cmd *cobra.Command, id string) error {
	if id == "" {
		return newUserError("missing_required_field", "memory get: --id is required")
	}

	db, err := infrastructure.OpenReadOnly(resolveDBPath())
	if infrastructure.IsDatabaseNotFound(err) {
		return newSystemError("db_unreadable", "memory get: no db at "+resolveDBPath()+"; run `zharness init` first")
	}
	if err != nil {
		return mapReadOnlyOpenError("memory get", err)
	}
	defer db.Close()

	view, err := application.MemoryGet(db.Raw(), id)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_unreadable", fmt.Sprintf("memory get: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(view)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%+v\n", view)
	return nil
}

func runMemoryQuery(cmd *cobra.Command, memType, scope, planID string) error {
	db, err := infrastructure.OpenReadOnly(resolveDBPath())
	if infrastructure.IsDatabaseNotFound(err) {
		return newSystemError("db_unreadable", "memory query: no db at "+resolveDBPath()+"; run `zharness init` first")
	}
	if err != nil {
		return mapReadOnlyOpenError("memory query", err)
	}
	defer db.Close()

	views, err := application.MemoryQuery(db.Raw(), memType, scope, planID)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_unreadable", fmt.Sprintf("memory query: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(views)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%+v\n", views)
	return nil
}

func runMemoryQueryRanked(cmd *cobra.Command, keywords, memType, scope, planID string) error {
	db, err := infrastructure.OpenReadOnly(resolveDBPath())
	if infrastructure.IsDatabaseNotFound(err) {
		return newSystemError("db_unreadable", "memory query: no db at "+resolveDBPath()+"; run `zharness init` first")
	}
	if err != nil {
		return mapReadOnlyOpenError("memory query", err)
	}
	defer db.Close()

	views, err := application.MemoryQueryRanked(db.Raw(), keywords, memType, scope, planID)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_unreadable", fmt.Sprintf("memory query: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(views)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%+v\n", views)
	return nil
}
