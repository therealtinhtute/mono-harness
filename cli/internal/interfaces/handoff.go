package interfaces

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func newHandoffCmd() *cobra.Command {
	handoff := &cobra.Command{
		Use:   "handoff",
		Short: "Handoff (close-out record) operations",
	}

	record := &cobra.Command{
		Use:   "record",
		Short: "Record a handoff (anchors: latest run/check IDs, open items)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runID, _ := cmd.Flags().GetString("run-id")
			checkID, _ := cmd.Flags().GetString("check-id")
			openItemsRaw, _ := cmd.Flags().GetString("open-items")
			return runHandoffRecord(cmd, runID, checkID, openItemsRaw)
		},
	}
	record.Flags().String("run-id", "", "ulid of the latest run (optional)")
	record.Flags().String("check-id", "", "ulid of the latest check (optional)")
	record.Flags().String("open-items", "[]", `JSON array of strings: ["open item", ...]`)

	handoff.AddCommand(record)
	return handoff
}

func runHandoffRecord(cmd *cobra.Command, runID, checkID, openItemsRaw string) error {
	var openItems []string
	if err := json.Unmarshal([]byte(openItemsRaw), &openItems); err != nil {
		return newUserError("invalid_open_items", fmt.Sprintf("handoff record: --open-items is not valid JSON: %v", err))
	}

	if !infrastructure.Exists(dbPath) {
		return newSystemError("db_unreadable", "handoff record: no db at "+dbPath+"; run `zharness init` first")
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		return newSystemError("db_unreadable", fmt.Sprintf("handoff record: %v", err))
	}
	defer db.Close()

	id, _, err := application.RecordHandoff(db, changesetDir, runID, checkID, openItems)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("db_not_writable", fmt.Sprintf("handoff record: %v", err))
	}

	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"id": id})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "handoff %s recorded\n", id)
	return nil
}
