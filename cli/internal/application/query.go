package application

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/therealtinhtute/skills/cli/internal/domain"
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

// HandoffView is the `query handoff --latest` view: the most recent
// handoff's anchors, flattened out of its JSON column — the read half of
// the round trip for `handoff record --next-action` (D1,
// docs/audit/workflow-harness-ceremony-audit.md).
type HandoffView struct {
	ID         string   `json:"id"`
	RunID      *string  `json:"run_id"`
	CheckID    *string  `json:"check_id"`
	OpenItems  []string `json:"open_items"`
	NextAction *string  `json:"exact_next_action"`
	CreatedAt  string   `json:"created_at"`
}

// QueryLatestHandoff reads the most recently created handoff row and
// unmarshals its anchors JSON column back into typed fields.
func QueryLatestHandoff(db *sql.DB) (HandoffView, bool, error) {
	var v HandoffView
	var runID, checkID sql.NullString
	var anchorsRaw string
	err := db.QueryRow(`
		SELECT id, run_id, check_id, anchors, created_at
		FROM handoffs
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`).Scan(&v.ID, &runID, &checkID, &anchorsRaw, &v.CreatedAt)
	if err == sql.ErrNoRows {
		return v, false, nil
	}
	if err != nil {
		return v, false, fmt.Errorf("query latest handoff: %w", err)
	}
	v.RunID = nullableString(runID)
	v.CheckID = nullableString(checkID)

	var anchors domain.HandoffAnchors
	if err := json.Unmarshal([]byte(anchorsRaw), &anchors); err != nil {
		return v, false, fmt.Errorf("query latest handoff: decode anchors: %w", err)
	}
	v.OpenItems = anchors.OpenItems
	if v.OpenItems == nil {
		v.OpenItems = []string{}
	}
	v.NextAction = anchors.NextAction
	return v, true, nil
}

// TraceView is one row of the `query traces` view: a wave- or task-level
// compressed summary, the compact index counterpart of a `## Progress`
// entry. Task and TaskStatus are nil for a wave-level trace.
type TraceView struct {
	ID         string  `json:"id"`
	RunID      *string `json:"run_id"`
	Wave       int     `json:"wave"`
	Summary    string  `json:"summary"`
	Task       *string `json:"task"`
	TaskStatus *string `json:"task_status"`
	CreatedAt  string  `json:"created_at"`
}

// QueryTraces reads trace rows in chronological order, optionally filtered
// to one run (`--run-id`) and/or capped to the most recent N (`--tail`, 0
// means unbounded). Filtering is applied in SQL so --tail counts from the
// filtered set, not the whole table.
func QueryTraces(db *sql.DB, runID string, tail int) ([]TraceView, error) {
	q := `SELECT id, run_id, wave, summary, task, task_status, created_at FROM traces`
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
		var runIDCol, taskCol, taskStatusCol sql.NullString
		if err := rows.Scan(&v.ID, &runIDCol, &v.Wave, &v.Summary, &taskCol, &taskStatusCol, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan trace row: %w", err)
		}
		v.RunID = nullableString(runIDCol)
		v.Task = nullableString(taskCol)
		v.TaskStatus = nullableString(taskStatusCol)
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

// QueryTracesByPhase reads trace rows for one phase, joining through
// runs.story_slug, in chronological order and capped to the most recent N
// (tail, 0 = unbounded) — the phase-scoped counterpart of QueryTraces'
// run-scoped filter, for callers (preflight's context packet) that need a
// phase's trace history rather than one run's.
func QueryTracesByPhase(db *sql.DB, phaseSlug string, tail int) ([]TraceView, error) {
	q := `
		SELECT traces.id, traces.run_id, traces.wave, traces.summary, traces.task, traces.task_status, traces.created_at
		FROM traces
		JOIN runs ON runs.id = traces.run_id
		WHERE runs.story_slug = ?
		ORDER BY traces.created_at DESC, traces.id DESC`
	args := []any{phaseSlug}
	if tail > 0 {
		q += ` LIMIT ?`
		args = append(args, tail)
	}

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("query traces by phase: %w", err)
	}
	defer rows.Close()

	views := []TraceView{}
	for rows.Next() {
		var v TraceView
		var runIDCol, taskCol, taskStatusCol sql.NullString
		if err := rows.Scan(&v.ID, &runIDCol, &v.Wave, &v.Summary, &taskCol, &taskStatusCol, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan trace row: %w", err)
		}
		v.RunID = nullableString(runIDCol)
		v.Task = nullableString(taskCol)
		v.TaskStatus = nullableString(taskStatusCol)
		views = append(views, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query traces by phase: %w", err)
	}

	// DESC fetched the most recent `tail`; restore chronological order.
	for i, j := 0, len(views)-1; i < j; i, j = i+1, j-1 {
		views[i], views[j] = views[j], views[i]
	}
	return views, nil
}

// countTracesForPhase reports how many trace rows exist for a phase,
// regardless of any tail cap — BuildContextPacket uses this to decide
// whether QueryTracesByPhase's windowed result needs an Omitted entry.
func countTracesForPhase(db *sql.DB, phaseSlug string) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM traces
		JOIN runs ON runs.id = traces.run_id
		WHERE runs.story_slug = ?`, phaseSlug).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count traces by phase: %w", err)
	}
	return count, nil
}

// DecisionView is one row of the `query decisions` view: the compressed-
// index counterpart of a `## Decisions` markdown entry.
type DecisionView struct {
	ID        string  `json:"id"`
	RunID     *string `json:"run_id"`
	Phase     *string `json:"phase"`
	Task      *string `json:"task"`
	Decision  string  `json:"decision"`
	Rationale string  `json:"rationale"`
	CreatedAt string  `json:"created_at"`
}

// QueryDecisions reads decision rows in chronological order, optionally
// filtered to one phase (`--phase`, matching `query artifacts`'s filter
// convention) and/or capped to the most recent N (`--tail`, 0 = unbounded).
func QueryDecisions(db *sql.DB, phase string, tail int) ([]DecisionView, error) {
	q := `SELECT id, run_id, phase, task, decision, rationale, created_at FROM decisions`
	args := []any{}
	if phase != "" {
		q += ` WHERE phase = ?`
		args = append(args, phase)
	}
	q += ` ORDER BY created_at DESC, id DESC`
	if tail > 0 {
		q += ` LIMIT ?`
		args = append(args, tail)
	}

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("query decisions: %w", err)
	}
	defer rows.Close()

	views := []DecisionView{}
	for rows.Next() {
		var v DecisionView
		var runID, phaseCol, task sql.NullString
		if err := rows.Scan(&v.ID, &runID, &phaseCol, &task, &v.Decision, &v.Rationale, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan decision row: %w", err)
		}
		v.RunID = nullableString(runID)
		v.Phase = nullableString(phaseCol)
		v.Task = nullableString(task)
		views = append(views, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query decisions: %w", err)
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
