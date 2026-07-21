package application

import (
	"database/sql"
	"time"

	"github.com/oklog/ulid/v2"

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

	exists, err := runExists(db, runID)
	if err != nil {
		return "", "", err
	}
	if !exists {
		return "", "", &domain.ValidationError{Code: "unknown_run_id", Message: "check record: run_id " + runID + " not found"}
	}

	at := time.Now().UTC().Format(time.RFC3339)
	id = ulid.Make().String()
	proofLinksAny := make([]any, len(proofLinks))
	for i, pl := range proofLinks {
		proofLinksAny[i] = map[string]any{
			"command":       pl.Command,
			"output_ref":    pl.OutputRef,
			"artifact_path": pl.ArtifactPath,
		}
	}
	path, _, err = AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
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
		{Op: "update", Entity: "meta", ID: "meta", Fields: map[string]any{"latest_check_id": id}, At: at},
	})
	if err != nil {
		return "", "", err
	}
	return id, path, nil
}
