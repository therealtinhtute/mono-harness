package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestCreateIntake(t *testing.T) {
	db, changesetDir := freshDB(t)

	id, path, err := CreateIntake(db, changesetDir, domain.IntakeNewSpec, "add zharness domain commands", domain.LaneNormal)
	if err != nil {
		t.Fatalf("CreateIntake: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "intakes", id, "intake")
	if got := countRows(t, db, "intakes"); got != 1 {
		t.Fatalf("intakes rows = %d, want 1", got)
	}
}

func TestCreateIntakeInvalidType(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := CreateIntake(db, changesetDir, "not-a-type", "summary", domain.LaneNormal)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "invalid_type" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: invalid_type}", err)
	}
	if got := countRows(t, db, "intakes"); got != 0 {
		t.Fatalf("intakes rows after rejected create = %d, want 0", got)
	}
}

func TestCreateIntakeInvalidLane(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := CreateIntake(db, changesetDir, domain.IntakeNewSpec, "summary", "urgent")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "invalid_lane" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: invalid_lane}", err)
	}
}
