package application

import (
	"database/sql"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestCreateBacklog(t *testing.T) {
	db, changesetDir := freshDB(t)

	id, path, err := CreateBacklog(db, changesetDir, "improve trace scoring", domain.LaneTiny)
	if err != nil {
		t.Fatalf("CreateBacklog: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "backlog", id, "backlog")

	var priority sql.NullString
	if err := db.QueryRow(`SELECT priority FROM backlog WHERE id = ?`, id).Scan(&priority); err != nil {
		t.Fatalf("query priority: %v", err)
	}
	if !priority.Valid || priority.String != domain.LaneTiny {
		t.Fatalf("priority = %+v, want %q", priority, domain.LaneTiny)
	}
}

func TestCreateBacklogNoPriority(t *testing.T) {
	db, changesetDir := freshDB(t)

	id, _, err := CreateBacklog(db, changesetDir, "improve trace scoring", "")
	if err != nil {
		t.Fatalf("CreateBacklog: %v", err)
	}
	var priority sql.NullString
	if err := db.QueryRow(`SELECT priority FROM backlog WHERE id = ?`, id).Scan(&priority); err != nil {
		t.Fatalf("query priority: %v", err)
	}
	if priority.Valid {
		t.Fatalf("priority = %q, want NULL", priority.String)
	}
}

func TestCreateBacklogInvalidPriority(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := CreateBacklog(db, changesetDir, "improve trace scoring", "urgent")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "invalid_lane" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: invalid_lane}", err)
	}
}
