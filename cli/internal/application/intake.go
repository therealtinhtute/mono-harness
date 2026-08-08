package application

import (
	"database/sql"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// CreateIntake validates and records a new intake entity (CONTRACT.md
// `intake`), changeset-first. planID is optional and, when supplied, must
// be the same plan ULID passed to `run create --plan-id` for that
// initiative's runs — the join check record uses to resolve a run's lane
// (G2, docs/audit/workflow-harness-ceremony-audit.md/V2). It carries no FK,
// matching runs.plan_id's own precedent (not part of lifecycle-link
// validation).
func CreateIntake(db *sql.DB, changesetDir, typ, summary, lane, planPath, planID string) (id, path string, err error) {
	entity := domain.Intake{Type: typ, Summary: summary, Lane: lane, PlanPath: planPath, PlanID: planID}
	if err := entity.Validate(); err != nil {
		return "", "", err
	}
	if planID != "" {
		if _, err := ulid.ParseStrict(planID); err != nil {
			return "", "", &domain.ValidationError{Code: "invalid_plan_id", Message: "intake: plan_id must be a valid ULID"}
		}
	}

	at := time.Now().UTC().Format(time.RFC3339)
	id = ulid.Make().String()
	fields := map[string]any{
		"type": typ, "summary": summary, "lane": lane, "created_at": at,
	}
	if planPath != "" {
		fields["plan_path"] = planPath
	}
	if planID != "" {
		fields["plan_id"] = planID
	}
	path, _, err = AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "create", Entity: "intake", ID: id, Fields: fields, At: at},
	})
	if err != nil {
		return "", "", err
	}
	return id, path, nil
}
