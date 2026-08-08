package application

import (
	"database/sql"
	"fmt"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// RecordCheck validates and records a new check (gate verdict) entity
// (CONTRACT.md `check record`), changeset-first, and atomically points
// meta.latest_check_id at it in the same changeset/tx — the hand-authored
// meta changeset check.md's playbook previously required is now owned by
// the CLI. unknown_run_id is DB-lookup-dependent, so it's enforced here
// rather than in domain.Check.Validate() (invalid_verdict, invalid_judge,
// and empty_proof_links are already covered there).
func RecordCheck(db *sql.DB, changesetDir, runID, verdict, judge, judgeModel string, proofLinks []domain.ProofLink) (id, path string, err error) {
	entity := domain.Check{RunID: runID, Verdict: verdict, Judge: judge, JudgeModel: judgeModel, ProofLinks: proofLinks}
	if err := entity.Validate(); err != nil {
		return "", "", err
	}

	var storyID, storySlug, storyStatus, latestRunID string
	err = db.QueryRow(`
		SELECT stories.id, stories.slug, stories.status,
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
	`, runID).Scan(&storyID, &storySlug, &storyStatus, &latestRunID)
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

	lane, laneResolved, err := resolveLaneForRun(db, runID)
	if err != nil {
		return "", "", err
	}
	if laneResolved && lane == domain.LaneHighRisk && judge != domain.JudgeIndependent {
		return "", "", &domain.ValidationError{Code: "independent_judge_required", Message: "check record: lane is high-risk, --judge must be independent"}
	}

	// AppendNewEntityAndApply mints id/at internally (from the changeset
	// ULID, for clock-precision "latest check" ordering — see its own
	// doc comment), so the Validation entry's exact text isn't knowable
	// until after. Pre-validate the section is writable now, with
	// placeholder content, so the common failure mode (missing section)
	// still fails before the DB write — same "index and markdown cannot
	// diverge" guarantee trace/decision get from computing their entry
	// text upfront, just reached by checking writability instead.
	if err := planSectionWritable("Validation"); err != nil {
		return "", "", err
	}

	proofLinksAny := make([]any, len(proofLinks))
	for i, pl := range proofLinks {
		proofLinksAny[i] = map[string]any{
			"command":       pl.Command,
			"output_ref":    pl.OutputRef,
			"artifact_path": pl.ArtifactPath,
		}
	}
	var entryAt string
	id, path, _, err = AppendNewEntityAndApply(db, changesetDir, func(id string) []infrastructure.ChangesetLine {
		at := orderedChangesetTime(id)
		entryAt = at
		lines := []infrastructure.ChangesetLine{
			{
				Op:     "create",
				Entity: "check",
				ID:     id,
				Fields: map[string]any{
					"run_id":      runID,
					"verdict":     verdict,
					"judge":       judge,
					"judge_model": judgeModel,
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

	writePlan, err := preparePlanAppend("Validation", formatCheckValidationEntry(entryAt, id, storySlug, runID, verdict, judge, judgeModel, proofLinks))
	if err != nil {
		// The pre-check above already proved the section exists; a failure
		// here is a race (the plan changed mid-command) or an I/O error,
		// not the ordinary missing-section case.
		return id, path, fmt.Errorf("check %s recorded, but plan markdown update failed: %w", id, err)
	}
	if err := writePlan(); err != nil {
		return id, path, fmt.Errorf("check %s recorded, but plan markdown update failed: %w", id, err)
	}
	return id, path, nil
}

// planSectionWritable checks whether the active plan (if exactly one
// exists) currently has the named section, without needing final entry
// content — for callers (check record) whose entry text isn't computable
// until after the DB write that mints its id/timestamp.
func planSectionWritable(section string) error {
	_, err := preparePlanAppend(section, "")
	return err
}

// formatCheckValidationEntry renders a `## Validation` line using only the
// fields check record actually receives (P3, "one writer") — proof links
// become nested sub-bullets, mirroring the nested-bullet convention real
// hand-authored Validation entries already use
// (docs/plans/completed/eval-layer.md).
func formatCheckValidationEntry(at, id, phase, runID, verdict, judge, judgeModel string, proofLinks []domain.ProofLink) string {
	line := fmt.Sprintf("- `%s` — check. verdict: `%s`. check: `%s`. run: `%s`.", at, verdict, id, runID)
	if phase != "" {
		line += fmt.Sprintf(" phase: `%s`.", phase)
	}
	line += fmt.Sprintf(" judge: `%s` (%s).", judge, judgeModel)
	for _, pl := range proofLinks {
		sub := fmt.Sprintf("\n  - `%s`", pl.Command)
		if pl.OutputRef != "" {
			sub += fmt.Sprintf(" → %s", pl.OutputRef)
		}
		line += sub
	}
	return line
}

// resolveLaneForRun joins runs.plan_id -> intakes.plan_id to find the
// initiative lane a run belongs to (G2, docs/audit/workflow-harness-
// ceremony-audit.md/V2). Both sides are optional (run_create --plan-id,
// intake --plan-id), so an unresolvable lane is not an error: the gate
// this enables simply does not apply. Only a resolved high-risk lane
// blocks anything — every other lane, and every run with no plan_id
// trail, behaves exactly as before this feature existed.
func resolveLaneForRun(db *sql.DB, runID string) (lane string, ok bool, err error) {
	var planID sql.NullString
	if err := db.QueryRow(`SELECT plan_id FROM runs WHERE id = ?`, runID).Scan(&planID); err != nil {
		if err == sql.ErrNoRows {
			return "", false, nil
		}
		return "", false, err
	}
	if !planID.Valid || planID.String == "" {
		return "", false, nil
	}

	var laneCol sql.NullString
	err = db.QueryRow(`SELECT lane FROM intakes WHERE plan_id = ? ORDER BY created_at DESC LIMIT 1`, planID.String).Scan(&laneCol)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if !laneCol.Valid || laneCol.String == "" {
		return "", false, nil
	}
	return laneCol.String, true, nil
}
