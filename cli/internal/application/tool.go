package application

import (
	"database/sql"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// CreateTool validates and records a new tool entity (CONTRACT.md `tool`),
// changeset-first.
func CreateTool(db *sql.DB, changesetDir, name, purpose string) (id, path string, err error) {
	entity := domain.Tool{Name: name, Purpose: purpose}
	if err := entity.Validate(); err != nil {
		return "", "", err
	}

	at := time.Now().UTC().Format(time.RFC3339)
	id = ulid.Make().String()
	path, _, err = AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{
			Op:     "create",
			Entity: "tool",
			ID:     id,
			Fields: map[string]any{
				"name":       name,
				"purpose":    purpose,
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
