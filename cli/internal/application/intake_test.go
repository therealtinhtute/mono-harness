package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestCreateIntake(t *testing.T) {
	db := freshDB(t)

	planPath := "docs/plans/active/harness.md"
	id, err := CreateIntake(db, domain.IntakeNewSpec, "add zharness domain commands", domain.LaneNormal, planPath, "")
	if err != nil {
		t.Fatalf("CreateIntake: %v", err)
	}
	assertRowExists(t, db, "intakes", id)
	if got := countRows(t, db, "intakes"); got != 1 {
		t.Fatalf("intakes rows = %d, want 1", got)
	}
	var gotPlanPath string
	if err := db.QueryRow(`SELECT plan_path FROM intakes WHERE id = ?`, id).Scan(&gotPlanPath); err != nil {
		t.Fatalf("query plan_path: %v", err)
	}
	if gotPlanPath != planPath {
		t.Fatalf("plan_path = %q, want %q", gotPlanPath, planPath)
	}
}

func TestCreateIntakeInvalidType(t *testing.T) {
	db := freshDB(t)

	_, err := CreateIntake(db, "not-a-type", "summary", domain.LaneNormal, "", "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "invalid_type" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: invalid_type}", err)
	}
	if got := countRows(t, db, "intakes"); got != 0 {
		t.Fatalf("intakes rows after rejected create = %d, want 0", got)
	}
}

func TestCreateIntakeInvalidLane(t *testing.T) {
	db := freshDB(t)

	_, err := CreateIntake(db, domain.IntakeNewSpec, "summary", "urgent", "", "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "invalid_lane" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: invalid_lane}", err)
	}
}
