package application

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// RecordHandoff validates and records a new handoff entity (CONTRACT.md
// `handoff record`), changeset-first. unknown_run_id/unknown_check_id are
// DB-lookup-dependent (only checked when the corresponding flag is given),
// so they're enforced here rather than in domain.Handoff.Validate() — same
// pattern as CreateTrace's optional --run-id.
func RecordHandoff(db *sql.DB, changesetDir, runID, checkID, nextAction string, openItems []string, closePhase bool) (id, path string, err error) {
	entity := domain.Handoff{
		RunID:   optionalString(runID),
		CheckID: optionalString(checkID),
		Anchors: domain.HandoffAnchors{OpenItems: openItems, NextAction: optionalString(nextAction)},
	}
	if err := entity.Validate(); err != nil {
		return "", "", err
	}

	if runID != "" {
		exists, err := runExists(db, runID)
		if err != nil {
			return "", "", err
		}
		if !exists {
			return "", "", &domain.ValidationError{Code: "unknown_run_id", Message: "handoff record: run_id " + runID + " not found"}
		}
	}
	if checkID != "" {
		exists, err := checkExists(db, checkID)
		if err != nil {
			return "", "", err
		}
		if !exists {
			return "", "", &domain.ValidationError{Code: "unknown_check_id", Message: "handoff record: check_id " + checkID + " not found"}
		}
	}

	var closingStoryID string
	if closePhase {
		if runID == "" || checkID == "" {
			return "", "", &domain.ValidationError{Code: "missing_required_field", Message: "handoff record: --close-phase requires run_id and check_id"}
		}
		checkRunID, verdict, exists, err := checkForPhaseClose(db, checkID)
		if err != nil {
			return "", "", err
		}
		if !exists {
			return "", "", &domain.ValidationError{Code: "unknown_check_id", Message: "handoff record: check_id " + checkID + " not found"}
		}
		if checkRunID != runID {
			return "", "", &domain.ValidationError{Code: "check_run_mismatch", Message: "handoff record: check does not gate the supplied run"}
		}
		if verdict != domain.VerdictApproved && verdict != domain.VerdictApproveWithRequests {
			return "", "", &domain.ValidationError{Code: "check_not_clean", Message: "handoff record: cannot close a phase with REQUEST_CHANGES"}
		}

		var storyStatus, latestRunID string
		err = db.QueryRow(`
			SELECT stories.id, stories.status,
				(
					SELECT latest.id
					FROM runs AS latest
					WHERE latest.story_slug = stories.slug
					ORDER BY latest.created_at DESC, latest.id DESC
					LIMIT 1
				)
			FROM runs
			JOIN stories ON stories.slug = runs.story_slug
			WHERE runs.id = ?
		`, runID).Scan(&closingStoryID, &storyStatus, &latestRunID)
		if err == sql.ErrNoRows {
			return "", "", &domain.ValidationError{Code: "unknown_run_id", Message: "handoff record: run_id " + runID + " not found"}
		}
		if err != nil {
			return "", "", err
		}
		if latestRunID != runID {
			return "", "", &domain.ValidationError{Code: "run_not_latest", Message: "handoff record: run_id is not the latest run for its story"}
		}

		var latestCheckID string
		err = db.QueryRow(`
			SELECT id
			FROM checks
			WHERE run_id = ?
			ORDER BY created_at DESC, id DESC
			LIMIT 1
		`, runID).Scan(&latestCheckID)
		if err == sql.ErrNoRows {
			return "", "", &domain.ValidationError{Code: "check_not_latest", Message: "handoff record: check_id is not the latest check for its run"}
		}
		if err != nil {
			return "", "", err
		}
		if latestCheckID != checkID {
			return "", "", &domain.ValidationError{Code: "check_not_latest", Message: "handoff record: check_id is not the latest check for its run"}
		}
		if storyStatus != domain.StoryChecked {
			return "", "", &domain.ValidationError{Code: "phase_not_checked", Message: "handoff record: story must be checked before phase close"}
		}
	}

	// handoff record targets `## Progress`, not `## Current State and Next
	// Action` — Current State is a snapshot (last-write-wins), not an
	// append-only log, and the plan_section appender only knows how to
	// append after the last line of a section, so a handoff entry joins
	// Progress's event log alongside trace/decision/check entries instead.
	// id is minted from the changeset ULID up front (same value
	// AppendNewEntityAndApply used to mint inside its DB-write callback),
	// so the Progress entry text is knowable before anything is written.
	// Markdown is the write target: writePlan below runs before the DB
	// write, so a failed markdown write leaves zero DB rows behind it (R8,
	// docs/plans/active/harness-markdown-truth.md).
	id, err = prepareChangesetAppend(db, changesetDir)
	if err != nil {
		return "", "", err
	}
	at := orderedChangesetTime(id)

	writePlan, err := preparePlanAppend(db, "Progress", formatHandoffProgressEntry(at, id, runID, checkID, nextAction, openItems, closePhase))
	if err != nil {
		return "", "", err
	}
	if err := writePlan(); err != nil {
		return "", "", fmt.Errorf("plan write failed: %w", err)
	}

	anchors := map[string]any{"open_items": openItems}
	if runID != "" {
		anchors["latest_run_id"] = runID
	}
	if checkID != "" {
		anchors["latest_check_id"] = checkID
	}
	if nextAction != "" {
		anchors["exact_next_action"] = nextAction
	}
	fields := map[string]any{
		"anchors":    anchors,
		"created_at": at,
	}
	if runID != "" {
		fields["run_id"] = runID
	}
	if checkID != "" {
		fields["check_id"] = checkID
	}
	lines := []infrastructure.ChangesetLine{{Op: "create", Entity: "handoff", ID: id, Fields: fields, At: at}}
	if closePhase {
		lines = append(lines, infrastructure.ChangesetLine{Op: "update", Entity: "story", ID: closingStoryID, Fields: map[string]any{"status": domain.StoryDone}, At: at})
	}

	path, _, err = writeAndApplyPreparedChangeset(db, changesetDir, id, lines)
	if err != nil {
		return "", "", fmt.Errorf("handoff %s: plan markdown recorded, but db write failed: %w", id, err)
	}
	return id, path, nil
}

// formatHandoffProgressEntry renders a `## Progress` line for a handoff
// record — an event-log entry ("handoff recorded"), not a rewrite of the
// snapshot-style `## Current State and Next Action` section.
func formatHandoffProgressEntry(at, id, runID, checkID, nextAction string, openItems []string, closePhase bool) string {
	line := fmt.Sprintf("- `%s` — handoff recorded. handoff: `%s`.", at, id)
	if runID != "" {
		line += fmt.Sprintf(" run: `%s`.", runID)
	}
	if checkID != "" {
		line += fmt.Sprintf(" check: `%s`.", checkID)
	}
	if closePhase {
		line += " phase closed."
	}
	if nextAction != "" {
		line += fmt.Sprintf(" next action: %s.", nextAction)
	}
	if len(openItems) > 0 {
		line += fmt.Sprintf(" open items: %s.", strings.Join(openItems, "; "))
	}
	return line
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
