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
	add.Flags().Bool("force", false, "bypass dedup gate when similar entries exist (use with similar_memory)")

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
			includeSuperseded, _ := cmd.Flags().GetBool("include-superseded")
			if keywords != "" {
				return runMemoryQueryRanked(cmd, keywords, memType, scope, planID, includeSuperseded)
			}
			return runMemoryQuery(cmd, memType, scope, planID, includeSuperseded)
		},
	}
	query.Flags().String("type", "", "memory type filter (required unless --keywords is set)")
	query.Flags().String("scope", "", "plan|global (optional)")
	query.Flags().String("plan-id", "", "initiative plan ulid (optional)")
	query.Flags().String("keywords", "", "rank results by keyword match against type+body instead of exact filtering (optional; --type becomes optional too)")
	query.Flags().Bool("include-superseded", false, "include superseded entries (default: exclude)")

	supersede := &cobra.Command{
		Use:   "supersede",
		Short: "Mark one memory entry as superseded by another",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			oldID, _ := cmd.Flags().GetString("old-id")
			newID, _ := cmd.Flags().GetString("new-id")
			return runMemorySupersede(cmd, oldID, newID)
		},
	}
	supersede.Flags().String("old-id", "", "memory entry ulid to supersede (required)")
	supersede.Flags().String("new-id", "", "memory entry ulid that supersedes the old one (required)")

	memory.AddCommand(add, get, query, supersede)
	return memory
}

func runMemoryAdd(cmd *cobra.Command, memType, scope, planID, summary string) error {
	if !infrastructure.Exists(resolveDBPath()) {
		return missingDBError("memory add")
	}
	db, err := infrastructure.Open(resolveDBPath())
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("memory add: %v", err))
	}
	defer db.Close()

	// Pre-insert dedup gate (R3): rank new summary against existing entries;
	// best score >=4 folded-token matches refuses with similar_memory unless --force.
	force, _ := cmd.Flags().GetBool("force")
	if !force && strings.TrimSpace(summary) != "" {
		ranked, rerr := application.MemoryQueryRanked(db, summary, "", "", "")
		if rerr == nil && len(ranked) > 0 && ranked[0].Score >= 4 {
			var ids []string
			for _, r := range ranked {
				if r.Score >= 4 {
					ids = append(ids, r.ID)
				}
				if len(ids) >= 5 {
					break
				}
			}
			ve := &domain.ValidationError{Code: "similar_memory", Message: fmt.Sprintf("memory add: similar entries exist: %s (use --force to bypass)", strings.Join(ids, ", "))}
			return mapValidationError(ve)
		}
	}

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
		return missingDBError("memory get")
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

func runMemoryQuery(cmd *cobra.Command, memType, scope, planID string, includeSuperseded bool) error {
	db, err := infrastructure.OpenReadOnly(resolveDBPath())
	if infrastructure.IsDatabaseNotFound(err) {
		return missingDBError("memory query")
	}
	if err != nil {
		return mapReadOnlyOpenError("memory query", err)
	}
	defer db.Close()

	var views []application.MemoryListView
	if includeSuperseded {
		views, err = application.MemoryQueryWithIncludeSuperseded(db.Raw(), memType, scope, planID, true)
	} else {
		views, err = application.MemoryQuery(db.Raw(), memType, scope, planID)
	}
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

func runMemoryQueryRanked(cmd *cobra.Command, keywords, memType, scope, planID string, includeSuperseded bool) error {
	db, err := infrastructure.OpenReadOnly(resolveDBPath())
	if infrastructure.IsDatabaseNotFound(err) {
		return missingDBError("memory query")
	}
	if err != nil {
		return mapReadOnlyOpenError("memory query", err)
	}
	defer db.Close()

	var views []application.MemoryScoredView
	if includeSuperseded {
		views, err = application.MemoryQueryRankedWithIncludeSuperseded(db.Raw(), keywords, memType, scope, planID, true)
	} else {
		views, err = application.MemoryQueryRanked(db.Raw(), keywords, memType, scope, planID)
	}
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

func runMemorySupersede(cmd *cobra.Command, oldID, newID string) error {
	if oldID == "" {
		return newUserError("missing_required_field", "memory supersede: --old-id is required")
	}
	if newID == "" {
		return newUserError("missing_required_field", "memory supersede: --new-id is required")
	}
	if !infrastructure.Exists(resolveDBPath()) {
		return missingDBError("memory supersede")
	}
	db, err := infrastructure.Open(resolveDBPath())
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("memory supersede: %v", err))
	}
	defer db.Close()

	if err := application.SupersedeMemory(db, oldID, newID); err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_not_writable", fmt.Sprintf("memory supersede: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"old_id": oldID, "new_id": newID, "superseded": true})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "memory %s superseded by %s\n", oldID, newID)
	return nil
}
