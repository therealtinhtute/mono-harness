package application

import (
	"database/sql"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestCreateDecision(t *testing.T) {
	db, changesetDir := freshDB(t)

	id, path, err := CreateDecision(db, changesetDir, "use ULIDs for entity ids", "sortable + collision-free", "sequential ints")
	if err != nil {
		t.Fatalf("CreateDecision: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "decisions", id, "decision")

	var rejected sql.NullString
	if err := db.QueryRow(`SELECT rejected FROM decisions WHERE id = ?`, id).Scan(&rejected); err != nil {
		t.Fatalf("query rejected: %v", err)
	}
	if !rejected.Valid || rejected.String != "sequential ints" {
		t.Fatalf("rejected = %+v, want %q", rejected, "sequential ints")
	}
}

func TestCreateDecisionNoRejected(t *testing.T) {
	db, changesetDir := freshDB(t)

	id, _, err := CreateDecision(db, changesetDir, "use ULIDs for entity ids", "sortable + collision-free", "")
	if err != nil {
		t.Fatalf("CreateDecision: %v", err)
	}
	var rejected sql.NullString
	if err := db.QueryRow(`SELECT rejected FROM decisions WHERE id = ?`, id).Scan(&rejected); err != nil {
		t.Fatalf("query rejected: %v", err)
	}
	if rejected.Valid {
		t.Fatalf("rejected = %q, want NULL", rejected.String)
	}
}

func TestCreateDecisionMissingRationale(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := CreateDecision(db, changesetDir, "use ULIDs for entity ids", "", "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "missing_required_field" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: missing_required_field}", err)
	}
}
