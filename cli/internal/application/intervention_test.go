package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestCreateIntervention(t *testing.T) {
	db, changesetDir := freshDB(t)
	checkID := seedCheck(t, db, changesetDir)

	id, path, err := CreateIntervention(db, changesetDir, checkID, "human override: acceptable risk")
	if err != nil {
		t.Fatalf("CreateIntervention: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "interventions", id, "intervention")
}

func TestCreateInterventionUnknownVerdictID(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := CreateIntervention(db, changesetDir, "01HZZZZZZZZZZZZZZZZZZZZZZZ", "human override")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_verdict_id" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_verdict_id}", err)
	}
	if got := countRows(t, db, "interventions"); got != 0 {
		t.Fatalf("interventions rows = %d, want 0", got)
	}
}
