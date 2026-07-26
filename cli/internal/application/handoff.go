package application

import (
	"database/sql"
	"time"

	"github.com/oklog/ulid/v2"

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
		if verdict == domain.VerdictRequestChanges {
			return "", "", &domain.ValidationError{Code: "check_not_clean", Message: "handoff record: cannot close a phase with REQUEST_CHANGES"}
		}
		closingStoryID, _, exists, err = storyForRun(db, runID)
		if err != nil {
			return "", "", err
		}
		if !exists {
			return "", "", &domain.ValidationError{Code: "unknown_run_id", Message: "handoff record: run_id " + runID + " not found"}
		}
	}

	at := time.Now().UTC().Format(time.RFC3339)
	id = ulid.Make().String()
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
	path, _, err = AppendAndApply(db, changesetDir, lines)
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
