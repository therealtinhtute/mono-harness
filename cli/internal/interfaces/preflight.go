package interfaces

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

const (
	preflightDocsPath = "docs"
	preflightSpecPath = ".kit/planning/SPEC.md"
)

var preflightPlaybooks = map[string]string{
	"brainstorm": "docs/playbooks/brainstorm.md",
	"to-plan":    "docs/playbooks/to-plan.md",
	"work":       "docs/playbooks/work.md",
	"check":      "docs/playbooks/check.md",
	"handoff":    "docs/playbooks/handoff.md",
	"watzup":     "docs/playbooks/watzup.md",
}

func newPreflightCmd(version string) *cobra.Command {
	var mode string
	cmd := &cobra.Command{
		Use:   "preflight <stage>",
		Short: "Resolve workflow stage readiness without mutating harness state",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPreflight(cmd, args[0], mode, version)
		},
	}
	cmd.Flags().StringVar(&mode, "mode", "", "stage mode (for example: explore, lock, simple, full, review)")
	return cmd
}

func runPreflight(cmd *cobra.Command, stage, requestedMode, version string) error {
	stage = strings.ToLower(strings.TrimSpace(stage))
	requestedMode = resolvePreflightRequestedMode(stage, requestedMode)
	dbStatus, docsStatus := observePreflightState(version)
	playbook := preflightPlaybooks[stage]
	if docsStatus == application.PreflightDocsReady && playbook != "" && !infrastructure.Exists(playbook) {
		docsStatus = application.PreflightDocsMissing
	}
	if docsStatus != application.PreflightDocsReady {
		playbook = ""
	}
	view, err := application.Preflight(stage, requestedMode, dbStatus, docsStatus, playbook)
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("preflight_failed", err.Error())
	}
	return emitPreflight(cmd, view)
}

func resolvePreflightRequestedMode(stage, requestedMode string) string {
	requestedMode = strings.ToLower(strings.TrimSpace(requestedMode))
	if stage == "work" && (requestedMode == "" || requestedMode == "auto") {
		if infrastructure.Exists(preflightSpecPath) {
			return "full"
		}
		return "simple"
	}
	return requestedMode
}

func observePreflightState(version string) (dbStatus, docsStatus string) {
	docsStatus = application.PreflightDocsReady
	if !infrastructure.Exists(preflightDocsPath) {
		docsStatus = application.PreflightDocsMissing
	}
	if !infrastructure.Exists(dbPath) {
		return application.PreflightDBMissing, docsStatus
	}

	db, err := infrastructure.OpenReadOnly(dbPath)
	if err != nil {
		return application.PreflightDBUnreadable, docsStatus
	}
	defer db.Close()

	var docsVersion sql.NullString
	if err := db.QueryRow(`SELECT docs_version FROM meta LIMIT 1`).Scan(&docsVersion); err != nil {
		return application.PreflightDBUnreadable, docsStatus
	}
	if docsStatus == application.PreflightDocsReady && docsVersion.Valid && docsVersion.String != "" && docsVersion.String != "dev" && version != "dev" && docsVersion.String != version {
		docsStatus = application.PreflightDocsStale
	}
	return application.PreflightDBReady, docsStatus
}

func emitPreflight(cmd *cobra.Command, view application.PreflightView) error {
	if jsonOutput {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(view)
	}
	stop := "none"
	if view.Stop != nil {
		stop = view.Stop.Code
	}
	fmt.Fprintf(cmd.OutOrStdout(), "stage=%s mode=%s db=%s docs=%s readiness=%s stop=%s playbook=%s\n", view.Stage, view.Mode, view.DB, view.Docs, view.Readiness, stop, view.Playbook)
	return nil
}
