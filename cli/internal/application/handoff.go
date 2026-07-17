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
func RecordHandoff(db *sql.DB, changesetDir, runID, checkID string, openItems []string) (id, path string, err error) {
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
	path, _, err = AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "create", Entity: "handoff", ID: id, Fields: fields, At: at},
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
