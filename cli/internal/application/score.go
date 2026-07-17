package application

import (
	"database/sql"
	"fmt"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// TraceScore mirrors CONTRACT.md's locked `score-trace --json` shape.
type TraceScore struct {
	Tier    string   `json:"tier"`
	Reasons []string `json:"reasons"`
}

// ContextScore mirrors CONTRACT.md's locked `score-context --json` shape.
type ContextScore struct {
	Score   int      `json:"score"`
	Reasons []string `json:"reasons"`
}

type traceRow struct {
	ID        string
	RunID     *string
	Wave      int
	Summary   string
	CreatedAt string
}

func loadTrace(db *sql.DB, id string) (traceRow, bool, error) {
	var tr traceRow
	err := db.QueryRow(`SELECT id, run_id, wave, summary, created_at FROM traces WHERE id = ?`, id).
		Scan(&tr.ID, &tr.RunID, &tr.Wave, &tr.Summary, &tr.CreatedAt)
	if err == sql.ErrNoRows {
		return traceRow{}, false, nil
	}
	if err != nil {
		return traceRow{}, false, fmt.Errorf("query trace %q: %w", id, err)
	}
	return tr, true, nil
}

func countTracesForRun(db *sql.DB, runID string) (int, error) {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM traces WHERE run_id = ?`, runID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count traces for run %q: %w", runID, err)
	}
	return n, nil
}

// ScoreTrace adapts upstream TRACE_SPEC.md's Minimal/Standard/Detailed tier
// system to zharness's actual trace shape (id, run_id, wave, summary,
// created_at). Upstream's tiers key off fields this schema does not carry
// (task_summary, intake_id, story_id, agent, actions_taken, files_read,
// files_changed, decisions_made, errors, outcome, duration_seconds,
// token_estimate, harness_friction, notes) — this is a deliberate,
// documented deviation (see .kit/implementation-notes.md) rather than a
// literal port:
//   - minimal: default tier
//   - standard: summary >= 10 chars (upstream's own Minimal rule, the one
//     field-for-field survivor) AND linked to a run (the one linkage field
//     this schema has, proxying upstream's intake_id/story_id linkage check)
//   - detailed: standard AND summary >= 40 chars AND more than one trace
//     recorded against the same run (proxying upstream's richer
//     decisions_made/errors/harness_friction evidence with the closest
//     available structural signal: multi-wave tracking)
func ScoreTrace(db *sql.DB, id string) (TraceScore, error) {
	tr, exists, err := loadTrace(db, id)
	if err != nil {
		return TraceScore{}, err
	}
	if !exists {
		return TraceScore{}, &domain.ValidationError{Code: "unknown_trace_id", Message: "score-trace: trace " + id + " not found"}
	}

	reasons := []string{}
	hasSummary := len(tr.Summary) >= 10
	if hasSummary {
		reasons = append(reasons, "summary is present and >= 10 chars")
	} else {
		reasons = append(reasons, "summary missing or under 10 chars")
	}

	linked := tr.RunID != nil
	if linked {
		reasons = append(reasons, "linked to run "+*tr.RunID)
	} else {
		reasons = append(reasons, "not linked to any run (run_id is null)")
	}

	tier := "minimal"
	if hasSummary && linked {
		tier = "standard"
	}

	if tier == "standard" && len(tr.Summary) >= 40 {
		n, err := countTracesForRun(db, *tr.RunID)
		if err != nil {
			return TraceScore{}, err
		}
		if n > 1 {
			tier = "detailed"
			reasons = append(reasons, fmt.Sprintf("summary >= 40 chars and %d traces recorded for this run", n))
		}
	}

	return TraceScore{Tier: tier, Reasons: reasons}, nil
}

// ScoreContext is CONTRACT.md's reserved `score-context` command (not
// consumed by any skill this initiative). Upstream's context-read scoring
// compares actual files_read against expected context — a field this
// schema does not capture (same gap ScoreTrace documents). Rather than
// fabricate a files-read comparison, this returns a coarse, honest proxy
// score (0-2) using only the signals the schema actually has, and says so
// in its own reasons.
func ScoreContext(db *sql.DB, id string) (ContextScore, error) {
	tr, exists, err := loadTrace(db, id)
	if err != nil {
		return ContextScore{}, err
	}
	if !exists {
		return ContextScore{}, &domain.ValidationError{Code: "unknown_trace_id", Message: "score-context: trace " + id + " not found"}
	}

	reasons := []string{}
	score := 0
	if len(tr.Summary) >= 10 {
		score++
		reasons = append(reasons, "summary is present and >= 10 chars")
	}
	if tr.RunID != nil {
		score++
		reasons = append(reasons, "linked to run "+*tr.RunID)
	}
	reasons = append(reasons, "zharness's trace schema does not capture files_read; score reflects structural linkage only, not verified context reads")

	return ContextScore{Score: score, Reasons: reasons}, nil
}
