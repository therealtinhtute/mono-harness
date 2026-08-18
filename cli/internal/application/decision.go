package application

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// RecordDecisions validates and records a batch of decision entities in one
// transaction (CONTRACT.md `decision add`) — the compressed-index
// counterpart of a plan's `## Decisions` markdown section
// (docs/audit/workflow-harness-ceremony-audit.md, D1/D2/G1). Batching
// exists so a wave surfacing several decisions costs one call, not one per
// decision. runID is optional and shared across the whole batch, matching
// `trace add --run-id`'s pattern; unknown_run_id/unknown_phase are
// DB-lookup-dependent, so — like trace's unknown_run_id — they're enforced
// here rather than in domain.Decision.Validate().
func RecordDecisions(db *sql.DB, runID string, decisions []domain.Decision) (ids []string, err error) {
	if len(decisions) == 0 {
		return nil, &domain.ValidationError{Code: "empty_decisions", Message: "decision add: at least one decision is required"}
	}
	for _, d := range decisions {
		if err := d.Validate(); err != nil {
			return nil, err
		}
	}

	if runID != "" {
		exists, err := runExists(db, runID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, &domain.ValidationError{Code: "unknown_run_id", Message: "decision add: run_id " + runID + " not found"}
		}
	}
	for _, d := range decisions {
		if d.Phase == "" {
			continue
		}
		exists, err := phaseExists(db, d.Phase)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, &domain.ValidationError{Code: "unknown_phase", Message: "decision add: phase " + d.Phase + " not found"}
		}
	}

	at := time.Now().UTC().Format(time.RFC3339)

	entryLines := make([]string, len(decisions))
	for i, d := range decisions {
		entryLines[i] = formatDecisionEntry(at, d)
	}
	writePlan, err := preparePlanAppend(db, "Decisions", strings.Join(entryLines, "\n"))
	if err != nil {
		return nil, err
	}

	// Markdown is the write target: writePlan runs before the DB write, so
	// a failed markdown write leaves zero DB rows behind it (R8,
	// docs/plans/active/harness-markdown-truth.md).
	if err := writePlan(); err != nil {
		return nil, fmt.Errorf("plan write failed: %w", err)
	}

	ids = make([]string, 0, len(decisions))
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("decisions: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck
	for _, d := range decisions {
		id := ulid.Make().String()
		var runIDArg, phaseArg, taskArg any
		if runID != "" {
			runIDArg = runID
		}
		if d.Phase != "" {
			phaseArg = d.Phase
		}
		if d.Task != "" {
			taskArg = d.Task
		}
		if _, err := tx.Exec(
			`INSERT INTO decisions (id, run_id, phase, task, decision, rationale, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			id, runIDArg, phaseArg, taskArg, d.Decision, d.Rationale, at,
		); err != nil {
			return nil, fmt.Errorf("decisions %v: plan markdown recorded, but db write failed: %w", ids, err)
		}
		ids = append(ids, id)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("decisions %v: plan markdown recorded, but db commit failed: %w", ids, err)
	}
	return ids, nil
}

// formatDecisionEntry renders one `## Decisions` line using only the
// fields decision add actually receives (P3, "one writer") — not the
// richer Discovered/Rationale/Result narrative structure a hand-authored
// decision uses, and not an auto-numbered "D{n}" title, which would
// require scanning existing entries for the highest number and risk
// colliding with a hand-authored one.
func formatDecisionEntry(at string, d domain.Decision) string {
	line := fmt.Sprintf("- `%s` — %s", at, d.Decision)
	if d.Phase != "" {
		line += fmt.Sprintf(" (phase: `%s`)", d.Phase)
	}
	if d.Task != "" {
		line += fmt.Sprintf(", task: %s", d.Task)
	}
	line += fmt.Sprintf(". rationale: %s.", d.Rationale)
	return line
}

func phaseExists(db *sql.DB, slug string) (bool, error) {
	var found string
	err := db.QueryRow(`SELECT slug FROM stories WHERE slug = ?`, slug).Scan(&found)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
