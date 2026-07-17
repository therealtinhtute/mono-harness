package application

import (
	"database/sql"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// CreateBacklog validates and records a new backlog entity (CONTRACT.md
// `backlog`), changeset-first.
func CreateBacklog(db *sql.DB, changesetDir, summary, priority string) (id, path string, err error) {
	var priorityPtr *string
	if priority != "" {
		priorityPtr = &priority
	}
	entity := domain.Backlog{Summary: summary, Priority: priorityPtr}
	if err := entity.Validate(); err != nil {
		return "", "", err
	}

	at := time.Now().UTC().Format(time.RFC3339)
	id = ulid.Make().String()
	fields := map[string]any{
		"summary":    summary,
		"created_at": at,
	}
	if priority != "" {
		fields["priority"] = priority
	}
	path, _, err = AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "create", Entity: "backlog", ID: id, Fields: fields, At: at},
	})
	if err != nil {
		return "", "", err
	}
	return id, path, nil
}
