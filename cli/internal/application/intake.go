package application

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// CreateIntake validates and records a new intake entity (CONTRACT.md
// `intake`). planID is optional and, when supplied, must be the same plan
// ULID passed to `run create --plan-id` for that initiative's runs — the
// join check record uses to resolve a run's lane (G2,
// docs/audit/workflow-harness-ceremony-audit.md/V2). It carries no FK,
// matching runs.plan_id's own precedent (not part of lifecycle-link
// validation).
func CreateIntake(db *sql.DB, typ, summary, lane, planPath, planID string) (id string, err error) {
	entity := domain.Intake{Type: typ, Summary: summary, Lane: lane, PlanPath: planPath, PlanID: planID}
	if err := entity.Validate(); err != nil {
		return "", err
	}
	if planID != "" {
		if _, err := ulid.ParseStrict(planID); err != nil {
			return "", &domain.ValidationError{Code: "invalid_plan_id", Message: "intake: plan_id must be a valid ULID"}
		}
	}

	at := time.Now().UTC().Format(time.RFC3339)
	id = ulid.Make().String()
	var planPathArg, planIDArg any
	if planPath != "" {
		planPathArg = planPath
	}
	if planID != "" {
		planIDArg = planID
	}
	if _, err := db.Exec(
		`INSERT INTO intakes (id, type, summary, lane, plan_path, plan_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, typ, summary, lane, planPathArg, planIDArg, at,
	); err != nil {
		return "", fmt.Errorf("insert intake: %w", err)
	}
	return id, nil
}
