package application

import (
	"database/sql"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// RecordCheck validates and records a new check (gate verdict) entity
// (CONTRACT.md `check record`), changeset-first, and atomically points
// meta.latest_check_id at it in the same changeset/tx — the hand-authored
// meta changeset check.md's playbook previously required is now owned by
// the CLI. unknown_run_id is DB-lookup-dependent, so it's enforced here
// rather than in domain.Check.Validate() (invalid_verdict and
// empty_proof_links are already covered there).
func RecordCheck(db *sql.DB, changesetDir, runID, verdict string, proofLinks []domain.ProofLink) (id, path string, err error) {
	entity := domain.Check{RunID: runID, Verdict: verdict, ProofLinks: proofLinks}
	if err := entity.Validate(); err != nil {
		return "", "", err
	}

	var storyID, storyStatus, latestRunID string
	err = db.QueryRow(`
		SELECT stories.id, stories.status,
			(
				SELECT latest.id
				FROM runs AS latest
				WHERE latest.story_slug = stories.slug
				ORDER BY latest.created_at DESC, latest.id DESC
				LIMIT 1
			)
		FROM runs
		JOIN stories ON stories.slug = runs.story_slug
		WHERE runs.id = ?
	`, runID).Scan(&storyID, &storyStatus, &latestRunID)
	if err == sql.ErrNoRows {
		return "", "", &domain.ValidationError{Code: "unknown_run_id", Message: "check record: run_id " + runID + " not found"}
	}
	if err != nil {
		return "", "", err
	}
	if storyStatus != domain.StoryInProgress {
		return "", "", &domain.ValidationError{Code: "story_not_checkable", Message: "check record: story must be in-progress"}
	}
	if latestRunID != runID {
		return "", "", &domain.ValidationError{Code: "run_not_latest", Message: "check record: run_id is not the latest run for its story"}
	}

	proofLinksAny := make([]any, len(proofLinks))
	for i, pl := range proofLinks {
		proofLinksAny[i] = map[string]any{
			"command":       pl.Command,
			"output_ref":    pl.OutputRef,
			"artifact_path": pl.ArtifactPath,
		}
	}
	id, path, _, err = AppendNewEntityAndApply(db, changesetDir, func(id string) []infrastructure.ChangesetLine {
		at := orderedChangesetTime(id)
		lines := []infrastructure.ChangesetLine{
			{
				Op:     "create",
				Entity: "check",
				ID:     id,
				Fields: map[string]any{
					"run_id":      runID,
					"verdict":     verdict,
					"proof_links": proofLinksAny,
					"created_at":  at,
				},
				At: at,
			},
		}
		if verdict != domain.VerdictRequestChanges {
			lines = append(lines, infrastructure.ChangesetLine{Op: "update", Entity: "story", ID: storyID, Fields: map[string]any{"status": domain.StoryChecked}, At: at})
		}
		return append(lines, infrastructure.ChangesetLine{Op: "update", Entity: "meta", ID: "meta", Fields: map[string]any{"latest_check_id": id}, At: at})
	})
	if err != nil {
		return "", "", err
	}
	return id, path, nil
}
