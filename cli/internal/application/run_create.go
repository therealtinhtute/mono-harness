package application

import (
	"database/sql"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// CreateRun validates and records a new run entity (CONTRACT.md `run
// create`), changeset-first, and atomically points meta.latest_run_id at
// it in the same changeset/tx — the two-line semantics work.md's playbook
// previously hand-authored, now owned by the CLI. unknown_story is
// DB-lookup-dependent, so it's enforced here rather than in
// domain.Run.Validate(). Full-mode only: simple mode has no story to
// reference (runs.story_slug is a NOT NULL FK) and must keep skipping DB
// registration entirely.
func CreateRun(db *sql.DB, changesetDir, storySlug, artifactPath, planID string) (id, path string, err error) {
	entity := domain.Run{StorySlug: storySlug, ArtifactPath: artifactPath}
	if err := entity.Validate(); err != nil {
		return "", "", err
	}
	if planID != "" {
		if _, err := ulid.ParseStrict(planID); err != nil {
			return "", "", &domain.ValidationError{Code: "invalid_plan_id", Message: "run create: plan_id must be a valid ULID"}
		}
	}

	storyID, storyStatus, exists, err := storyByExactSlug(db, storySlug)
	if err != nil {
		return "", "", err
	}
	if !exists {
		return "", "", &domain.ValidationError{Code: "unknown_story", Message: "run create: story slug " + storySlug + " not found"}
	}
	if storyStatus != domain.StoryPlanned && storyStatus != domain.StoryInProgress {
		return "", "", &domain.ValidationError{Code: "story_not_runnable", Message: "run create: story must be planned or in-progress"}
	}

	id, path, _, err = AppendNewEntityAndApply(db, changesetDir, func(id string) []infrastructure.ChangesetLine {
		at := orderedChangesetTime(id)
		fields := map[string]any{
			"story_slug":    storySlug,
			"artifact_path": artifactPath,
			"created_at":    at,
		}
		if planID != "" {
			fields["plan_id"] = planID
		}
		return []infrastructure.ChangesetLine{
			{Op: "create", Entity: "run", ID: id, Fields: fields, At: at},
			{Op: "update", Entity: "story", ID: storyID, Fields: map[string]any{"status": domain.StoryInProgress}, At: at},
			{Op: "update", Entity: "meta", ID: "meta", Fields: map[string]any{"latest_run_id": id, "current_phase": storySlug}, At: at},
		}
	})
	if err != nil {
		return "", "", err
	}
	return id, path, nil
}
