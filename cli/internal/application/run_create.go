package application

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// CreateRun validates and records a new run entity (CONTRACT.md `run
// create`), and atomically points meta.latest_run_id at it in the same
// transaction — the two-line semantics work.md's playbook previously
// hand-authored, now owned by the CLI. unknown_story is DB-lookup-dependent,
// so it's enforced here rather than in domain.Run.Validate(). Full-mode
// only: simple mode has no story to reference (runs.story_slug is a NOT
// NULL FK) and must keep skipping DB registration entirely.
func CreateRun(db *sql.DB, storySlug, artifactPath, planID string) (id string, err error) {
	entity := domain.Run{StorySlug: storySlug, ArtifactPath: artifactPath}
	if err := entity.Validate(); err != nil {
		return "", err
	}
	if planID != "" {
		if _, err := ulid.ParseStrict(planID); err != nil {
			return "", &domain.ValidationError{Code: "invalid_plan_id", Message: "run create: plan_id must be a valid ULID"}
		}
	}

	storyID, storyStatus, exists, err := storyByExactSlug(db, storySlug)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", &domain.ValidationError{Code: "unknown_story", Message: "run create: story slug " + storySlug + " not found"}
	}
	if storyStatus != domain.StoryPlanned && storyStatus != domain.StoryInProgress {
		return "", &domain.ValidationError{Code: "story_not_runnable", Message: "run create: story must be planned or in-progress"}
	}

	// Markdown is the write target for the story's phase-block status,
	// same as story create (R8, P3 wave 1,
	// docs/plans/active/harness-markdown-truth.md).
	writePlan, err := preparePlanPhaseStatus(db, storySlug, domain.StoryInProgress)
	if err != nil {
		return "", err
	}
	if err := writePlan(); err != nil {
		return "", fmt.Errorf("plan write failed: %w", err)
	}

	at := time.Now().UTC().Format(time.RFC3339)
	id = ulid.Make().String()
	var planIDArg any
	if planID != "" {
		planIDArg = planID
	}

	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("run create: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.Exec(
		`INSERT INTO runs (id, story_slug, plan_id, artifact_path, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, storySlug, planIDArg, artifactPath, at,
	); err != nil {
		return "", fmt.Errorf("run create: insert run: %w", err)
	}
	if _, err := tx.Exec(`UPDATE stories SET status = ? WHERE id = ?`, domain.StoryInProgress, storyID); err != nil {
		return "", fmt.Errorf("run create: update story: %w", err)
	}
	if _, err := tx.Exec(`UPDATE meta SET latest_run_id = ?, current_phase = ?`, id, storySlug); err != nil {
		return "", fmt.Errorf("run create: update meta: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("run create: commit: %w", err)
	}
	return id, nil
}
