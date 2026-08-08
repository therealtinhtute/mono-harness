package interfaces

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

const (
	preflightDocsPath       = "docs"
	preflightActivePlanGlob = "docs/plans/active/*.md"
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
	dbStatus, docsStatus, hasInProgressPhase := observePreflightState(version)
	requestedMode = resolvePreflightRequestedMode(stage, requestedMode, dbStatus, hasInProgressPhase)
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

// resolvePreflightRequestedMode auto-resolves work's mode. No active plan
// file at all still means simple, unchanged. With a plan file present, a
// *readable* harness that shows no story actually in-progress means the
// plan isn't the work at hand right now, so this downgrades to simple
// instead of routing any small unrelated change through the full
// 62-operation ceremony path merely because some active plan exists
// (docs/audit/workflow-harness-ceremony-audit.md, F2/V2). A missing or
// unreadable harness can't answer that question, so an active plan file
// alone still resolves to full there, the same as before this fix —
// otherwise a real durable initiative with no db yet would silently
// downgrade to simple instead of surfacing `harness_required`.
func resolvePreflightRequestedMode(stage, requestedMode, dbStatus string, hasInProgressPhase bool) string {
	requestedMode = strings.ToLower(strings.TrimSpace(requestedMode))
	if stage == "work" && (requestedMode == "" || requestedMode == "auto") {
		if !hasNonEmptyActivePlan() {
			return "simple"
		}
		if dbStatus == application.PreflightDBReady && !hasInProgressPhase {
			return "simple"
		}
		return "full"
	}
	return requestedMode
}

func hasNonEmptyActivePlan() bool {
	matches, err := filepath.Glob(preflightActivePlanGlob)
	if err != nil {
		return false
	}
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err == nil && strings.TrimSpace(string(data)) != "" {
			return true
		}
	}
	return false
}

// observePreflightState reads db/docs readiness and, in the same handle,
// whether any story is in-progress — the signal resolvePreflightRequestedMode
// needs to tell live durable work apart from a merely-present plan file.
func observePreflightState(version string) (dbStatus, docsStatus string, hasInProgressPhase bool) {
	docsStatus = application.PreflightDocsReady
	if !infrastructure.Exists(preflightDocsPath) {
		docsStatus = application.PreflightDocsMissing
	}
	db, err := infrastructure.OpenReadOnly(dbPath)
	if infrastructure.IsDatabaseNotFound(err) {
		return application.PreflightDBMissing, docsStatus, false
	}
	if err != nil {
		return application.PreflightDBUnreadable, docsStatus, false
	}
	defer db.Close()

	var docsVersion sql.NullString
	if err := db.QueryRow(`SELECT docs_version FROM meta LIMIT 1`).Scan(&docsVersion); err != nil {
		return application.PreflightDBUnreadable, docsStatus, false
	}
	if docsStatus == application.PreflightDocsReady && docsVersion.Valid && docsVersion.String != "" && docsVersion.String != "dev" && version != "dev" && docsVersion.String != version {
		docsStatus = application.PreflightDocsStale
	}

	var inProgressCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM stories WHERE status = ?`, domain.StoryInProgress).Scan(&inProgressCount); err == nil {
		hasInProgressPhase = inProgressCount > 0
	}
	return application.PreflightDBReady, docsStatus, hasInProgressPhase
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
