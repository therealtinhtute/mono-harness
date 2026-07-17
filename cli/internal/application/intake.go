package application

import (
	"database/sql"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// CreateIntake validates and records a new intake entity (CONTRACT.md
// `intake`), changeset-first.
func CreateIntake(db *sql.DB, changesetDir, typ, summary, lane string) (id, path string, err error) {
	entity := domain.Intake{Type: typ, Summary: summary, Lane: lane}
	if err := entity.Validate(); err != nil {
		return "", "", err
	}

	at := time.Now().UTC().Format(time.RFC3339)
	id = ulid.Make().String()
	path, _, err = AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{
			Op:     "create",
			Entity: "intake",
			ID:     id,
			Fields: map[string]any{
				"type":       typ,
				"summary":    summary,
				"lane":       lane,
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
