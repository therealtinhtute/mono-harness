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

// contextEligibleStages names the stages that unconditionally receive a
// stage-shaped context packet (R4). check is handled separately
// (checkContextEligible) since only its durable gate/full modes qualify —
// review/bounded stay response-only with zero packet cost. brainstorm/to-plan
// don't consult prior lifecycle state before writing (their playbooks call
// resume/query phases only as post-write verification, not as input), and
// git/interview own no harness entity (D7), so none of those four get a
// packet either.
var contextEligibleStages = map[string]bool{"watzup": true, "work": true, "handoff": true}

// checkContextEligible reports whether check's resolved preflight mode
// qualifies for a context packet (R6, docs/audit/sdlc-token-cache-audit.md):
// gate/full/auto/"" all resolve to domain.PreflightModeDurable per
// domain.preflightModes' check entry, exactly the modes whose playbook
// step (check.md step 1) reads lifecycle position at all — review/bounded
// resolve to reduced and stay packet-free, unchanged from before R6.
func checkContextEligible(stage string, resolvedMode string) bool {
	return stage == "check" && resolvedMode == domain.PreflightModeDurable
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
	dbStatus, docsStatus, hasInProgressPhase, db := observePreflightState(version)
	if db != nil {
		defer db.Close()
	}
	requestedMode = resolvePreflightRequestedMode(stage, requestedMode, dbStatus, hasInProgressPhase)
	playbook := preflightPlaybooks[stage]
	if docsStatus == application.PreflightDocsReady && playbook != "" && !infrastructure.Exists(playbook) {
		docsStatus = application.PreflightDocsMissing
	}
	if docsStatus != application.PreflightDocsReady {
		playbook = ""
	}
	view, err := application.Preflight(stage, requestedMode, dbStatus, docsStatus, playbook, version, hasCommittedPlans())
	if err != nil {
		if ve, ok := err.(*domain.ValidationError); ok {
			return mapValidationError(ve)
		}
		return newSystemError("preflight_failed", err.Error())
	}
	if db != nil && (contextEligibleStages[stage] || checkContextEligible(stage, view.Mode)) {
		pkg, cerr := application.BuildContextPacket(db.Raw(), stage, version)
		if cerr != nil {
			return newSystemError("preflight_failed", fmt.Sprintf("preflight: build context: %v", cerr))
		}
		view.Context = pkg
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
// It returns the opened handle (nil when the DB is missing/unreadable) so
// runPreflight can reuse it to build the context packet — one shared
// directory-inode lock and one SQLite handle for the complete preflight
// command lifetime (CONTRACT.md's own read-only-command lock invariant),
// rather than opening the DB a second time.
func observePreflightState(version string) (dbStatus, docsStatus string, hasInProgressPhase bool, db *infrastructure.ReadOnlyDB) {
	docsStatus = application.PreflightDocsReady
	if !infrastructure.Exists(preflightDocsPath) {
		docsStatus = application.PreflightDocsMissing
	}
	opened, err := infrastructure.OpenReadOnly(resolveDBPath())
	if infrastructure.IsDatabaseNotFound(err) {
		return application.PreflightDBMissing, docsStatus, false, nil
	}
	if err != nil {
		return application.PreflightDBUnreadable, docsStatus, false, nil
	}

	var docsVersion sql.NullString
	if err := opened.Raw().QueryRow(`SELECT docs_version FROM meta LIMIT 1`).Scan(&docsVersion); err != nil {
		opened.Close()
		return application.PreflightDBUnreadable, docsStatus, false, nil
	}
	if docsStatus == application.PreflightDocsReady && docsVersion.Valid && docsVersion.String != "" && docsVersion.String != "dev" && version != "dev" && docsVersion.String != version {
		docsStatus = application.PreflightDocsStale
	}

	var inProgressCount int
	if err := opened.Raw().QueryRow(`SELECT COUNT(*) FROM stories WHERE status = ?`, domain.StoryInProgress).Scan(&inProgressCount); err == nil {
		hasInProgressPhase = inProgressCount > 0
	}
	return application.PreflightDBReady, docsStatus, hasInProgressPhase, opened
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
