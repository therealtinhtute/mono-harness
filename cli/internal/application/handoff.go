package application

import (
	"database/sql"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// RecordHandoff validates and records a new handoff entity (CONTRACT.md
// `handoff record`), changeset-first. unknown_run_id/unknown_check_id are
// DB-lookup-dependent (only checked when the corresponding flag is given),
// so they're enforced here rather than in domain.Handoff.Validate() — same
// pattern as CreateTrace's optional --run-id.
func RecordHandoff(db *sql.DB, changesetDir, runID, checkID string, openItems []string, closePhase bool) (id, path string, err error) {
	entity := domain.Handoff{
		RunID:   optionalString(runID),
		CheckID: optionalString(checkID),
		Anchors: domain.HandoffAnchors{OpenItems: openItems},
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

	id, path, _, err = AppendNewEntityAndApply(db, changesetDir, func(id string) []infrastructure.ChangesetLine {
		at := orderedChangesetTime(id)
		anchors := map[string]any{"open_items": openItems}
		if runID != "" {
			anchors["latest_run_id"] = runID
		}
		if checkID != "" {
			anchors["latest_check_id"] = checkID
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
		return lines
	})
	if err != nil {
		return "", "", err
	}
	return id, path, nil
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
