package application

import (
	"database/sql"
	"fmt"
)

// StateView mirrors CONTRACT.md's locked `query state --json` shape.
type StateView struct {
	CurrentPhase  *string `json:"current_phase"`
	EntryPhase    *string `json:"entry_phase"`
	SchemaVersion int     `json:"schema_version"`
	LatestRunID   *string `json:"latest_run_id"`
	LatestCheckID *string `json:"latest_check_id"`
}

func QueryState(db *sql.DB) (StateView, error) {
	var v StateView
	var currentPhase, entryPhase, latestRunID, latestCheckID sql.NullString
	err := db.QueryRow(`SELECT current_phase, entry_phase, schema_version, latest_run_id, latest_check_id FROM meta LIMIT 1`).
		Scan(&currentPhase, &entryPhase, &v.SchemaVersion, &latestRunID, &latestCheckID)
	if err != nil {
		return v, fmt.Errorf("query state: %w", err)
	}
	v.CurrentPhase = nullableString(currentPhase)
	v.EntryPhase = nullableString(entryPhase)
	v.LatestRunID = nullableString(latestRunID)
	v.LatestCheckID = nullableString(latestCheckID)
	return v, nil
}

// PhaseView is one row of the (CONTRACT.md-undocumented) `query phases`
// view: a story with its dependency slug, if any.
type PhaseView struct {
	Slug      string  `json:"slug"`
	Goal      string  `json:"goal"`
	Status    string  `json:"status"`
	DependsOn *string `json:"depends_on"`
	CreatedAt string  `json:"created_at"`
}

func QueryPhases(db *sql.DB) ([]PhaseView, error) {
	rows, err := db.Query(`SELECT slug, goal, status, depends_on, created_at FROM stories ORDER BY created_at, slug`)
	if err != nil {
		return nil, fmt.Errorf("query phases: %w", err)
	}
	defer rows.Close()

	views := []PhaseView{}
	for rows.Next() {
		var v PhaseView
		var dependsOn sql.NullString
		if err := rows.Scan(&v.Slug, &v.Goal, &v.Status, &dependsOn, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan phase row: %w", err)
		}
		v.DependsOn = nullableString(dependsOn)
		views = append(views, v)
	}
	return views, rows.Err()
}

// ArtifactView is one row of the (CONTRACT.md-undocumented) `query
// artifacts` view: a run, optionally filtered by story slug (`--phase`).
type ArtifactView struct {
	ID           string `json:"id"`
	StorySlug    string `json:"story_slug"`
	ArtifactPath string `json:"artifact_path"`
	CreatedAt    string `json:"created_at"`
}

func QueryArtifacts(db *sql.DB, phase string) ([]ArtifactView, error) {
	q := `SELECT id, story_slug, artifact_path, created_at FROM runs`
	args := []any{}
	if phase != "" {
		q += ` WHERE story_slug = ?`
		args = append(args, phase)
	}
	q += ` ORDER BY created_at, id`

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("query artifacts: %w", err)
	}
	defer rows.Close()

	views := []ArtifactView{}
	for rows.Next() {
		var v ArtifactView
		if err := rows.Scan(&v.ID, &v.StorySlug, &v.ArtifactPath, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan artifact row: %w", err)
		}
		views = append(views, v)
	}
	return views, rows.Err()
}

// CheckView mirrors CONTRACT.md's locked `query check --latest --json`
// shape: the most recent check's verdict and the phase (story slug) its
// run belongs to. Judge and JudgeModel are nullable — checks recorded
// before migration 0006 carry neither.
type CheckView struct {
	ID         string  `json:"id"`
	Verdict    string  `json:"verdict"`
	Phase      string  `json:"phase"`
	Judge      *string `json:"judge"`
	JudgeModel *string `json:"judge_model"`
}

// QueryLatestCheck reads the most recently created check row, joined to
// its run's story_slug. Independent of import (which never creates
// checks rows itself) — any check row applied via `db changeset apply`
// is visible here.
func QueryLatestCheck(db *sql.DB) (CheckView, bool, error) {
	var v CheckView
	var judge, judgeModel sql.NullString
	err := db.QueryRow(`
		SELECT checks.id, checks.verdict, runs.story_slug, checks.judge, checks.judge_model
		FROM checks
		JOIN runs ON runs.id = checks.run_id
		ORDER BY checks.created_at DESC, checks.id DESC
		LIMIT 1
	`).Scan(&v.ID, &v.Verdict, &v.Phase, &judge, &judgeModel)
	if err == sql.ErrNoRows {
		return v, false, nil
	}
	if err != nil {
		return v, false, fmt.Errorf("query latest check: %w", err)
	}
	v.Judge = nullableString(judge)
	v.JudgeModel = nullableString(judgeModel)
	return v, true, nil
}

// TraceView is one row of the `query traces` view: a wave's compressed
// summary, the compact index counterpart of a `## Progress` entry.
type TraceView struct {
	ID        string  `json:"id"`
	RunID     *string `json:"run_id"`
	Wave      int     `json:"wave"`
	Summary   string  `json:"summary"`
	CreatedAt string  `json:"created_at"`
}

// QueryTraces reads trace rows in chronological order, optionally filtered
// to one run (`--run-id`) and/or capped to the most recent N (`--tail`, 0
// means unbounded). Filtering is applied in SQL so --tail counts from the
// filtered set, not the whole table.
func QueryTraces(db *sql.DB, runID string, tail int) ([]TraceView, error) {
	q := `SELECT id, run_id, wave, summary, created_at FROM traces`
	args := []any{}
	if runID != "" {
		q += ` WHERE run_id = ?`
		args = append(args, runID)
	}
	q += ` ORDER BY created_at DESC, id DESC`
	if tail > 0 {
		q += ` LIMIT ?`
		args = append(args, tail)
	}

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("query traces: %w", err)
	}
	defer rows.Close()

	views := []TraceView{}
	for rows.Next() {
		var v TraceView
		var runIDCol sql.NullString
		if err := rows.Scan(&v.ID, &runIDCol, &v.Wave, &v.Summary, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan trace row: %w", err)
		}
		v.RunID = nullableString(runIDCol)
		views = append(views, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query traces: %w", err)
	}

	// DESC fetched the most recent `tail`; restore chronological order.
	for i, j := 0, len(views)-1; i < j; i, j = i+1, j-1 {
		views[i], views[j] = views[j], views[i]
	}
	return views, nil
}

func nullableString(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}
