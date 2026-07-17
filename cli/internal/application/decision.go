package application

import (
	"database/sql"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// CreateDecision validates and records a new decision entity (CONTRACT.md
// `decision`), changeset-first.
func CreateDecision(db *sql.DB, changesetDir, summary, rationale, rejected string) (id, path string, err error) {
	entity := domain.Decision{Summary: summary, Rationale: rationale}
	if err := entity.Validate(); err != nil {
		return "", "", err
	}

	at := time.Now().UTC().Format(time.RFC3339)
	id = ulid.Make().String()
	fields := map[string]any{
		"summary":    summary,
		"rationale":  rationale,
		"created_at": at,
	}
	if rejected != "" {
		fields["rejected"] = rejected
	}
	path, _, err = AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "create", Entity: "decision", ID: id, Fields: fields, At: at},
	})
	if err != nil {
		return "", "", err
	}
	return id, path, nil
}
