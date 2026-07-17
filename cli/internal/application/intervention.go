package application

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func checkExists(db *sql.DB, id string) (bool, error) {
	var found string
	err := db.QueryRow(`SELECT id FROM checks WHERE id = ?`, id).Scan(&found)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("query check %q: %w", id, err)
	}
	return true, nil
}

// CreateIntervention validates and records a new intervention entity
// (CONTRACT.md `intervention`), changeset-first. unknown_verdict_id is
// DB-lookup-dependent, so it's enforced here rather than in
// domain.Intervention.Validate().
func CreateIntervention(db *sql.DB, changesetDir, verdictID, reason string) (id, path string, err error) {
	entity := domain.Intervention{VerdictID: verdictID, Reason: reason}
	if err := entity.Validate(); err != nil {
		return "", "", err
	}

	exists, err := checkExists(db, verdictID)
	if err != nil {
		return "", "", err
	}
	if !exists {
		return "", "", &domain.ValidationError{Code: "unknown_verdict_id", Message: "intervention: verdict_id " + verdictID + " not found"}
	}

	at := time.Now().UTC().Format(time.RFC3339)
	id = ulid.Make().String()
	path, _, err = AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{
			Op:     "create",
			Entity: "intervention",
			ID:     id,
			Fields: map[string]any{
				"verdict_id": verdictID,
				"reason":     reason,
				"created_at": at,
			},
			At: at,
		},
	})
	if err != nil {
		return "", "", err
	}
	return id, path, nil
}
