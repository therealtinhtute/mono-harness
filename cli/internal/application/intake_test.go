package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestCreateIntake(t *testing.T) {
	db, changesetDir := freshDB(t)

	planPath := "docs/plans/active/harness.md"
	id, path, err := CreateIntake(db, changesetDir, domain.IntakeNewSpec, "add zharness domain commands", domain.LaneNormal, planPath, "")
	if err != nil {
		t.Fatalf("CreateIntake: %v", err)
	}
	assertChangesetBeforeRow(t, db, path, "intakes", id, "intake")
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
	db, changesetDir := freshDB(t)

	_, _, err := CreateIntake(db, changesetDir, "not-a-type", "summary", domain.LaneNormal, "", "")
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

	_, _, err := CreateIntake(db, changesetDir, domain.IntakeNewSpec, "summary", "urgent", "", "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "invalid_lane" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: invalid_lane}", err)
	}
}
