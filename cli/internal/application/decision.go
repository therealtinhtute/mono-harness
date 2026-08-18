package application

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// RecordDecisions validates and records a batch of decision entities in one
// changeset (CONTRACT.md `decision add`) — the compressed-index counterpart
// of a plan's `## Decisions` markdown section
// (docs/audit/workflow-harness-ceremony-audit.md, D1/D2/G1). Batching
// exists so a wave surfacing several decisions costs one call, not one per
// decision. runID is optional and shared across the whole batch, matching
// `trace add --run-id`'s pattern; unknown_run_id/unknown_phase are
// DB-lookup-dependent, so — like trace's unknown_run_id — they're enforced
// here rather than in domain.Decision.Validate().
func RecordDecisions(db *sql.DB, changesetDir, runID string, decisions []domain.Decision) (ids []string, path string, err error) {
	if len(decisions) == 0 {
		return nil, "", &domain.ValidationError{Code: "empty_decisions", Message: "decision add: at least one decision is required"}
	}
	for _, d := range decisions {
		if err := d.Validate(); err != nil {
			return nil, "", err
		}
	}

	if runID != "" {
		exists, err := runExists(db, runID)
		if err != nil {
			return nil, "", err
		}
		if !exists {
			return nil, "", &domain.ValidationError{Code: "unknown_run_id", Message: "decision add: run_id " + runID + " not found"}
		}
	}
	for _, d := range decisions {
		if d.Phase == "" {
			continue
		}
		exists, err := phaseExists(db, d.Phase)
		if err != nil {
			return nil, "", err
		}
		if !exists {
			return nil, "", &domain.ValidationError{Code: "unknown_phase", Message: "decision add: phase " + d.Phase + " not found"}
		}
	}

	at := time.Now().UTC().Format(time.RFC3339)

	entryLines := make([]string, len(decisions))
	for i, d := range decisions {
		entryLines[i] = formatDecisionEntry(at, d)
	}
	writePlan, err := preparePlanAppend(db, "Decisions", strings.Join(entryLines, "\n"))
	if err != nil {
		return nil, "", err
	}

	// Markdown is the write target: writePlan runs before the DB write, so
	// a failed markdown write leaves zero DB rows behind it (R8,
	// docs/plans/active/harness-markdown-truth.md).
	if err := writePlan(); err != nil {
		return nil, "", fmt.Errorf("plan write failed: %w", err)
	}

	lines := make([]infrastructure.ChangesetLine, 0, len(decisions))
	ids = make([]string, 0, len(decisions))
	for _, d := range decisions {
		id := ulid.Make().String()
		fields := map[string]any{
			"decision": d.Decision, "rationale": d.Rationale, "created_at": at,
		}
		if runID != "" {
			fields["run_id"] = runID
		}
		if d.Phase != "" {
			fields["phase"] = d.Phase
		}
		if d.Task != "" {
			fields["task"] = d.Task
		}
		lines = append(lines, infrastructure.ChangesetLine{Op: "create", Entity: "decision", ID: id, Fields: fields, At: at})
		ids = append(ids, id)
	}

	path, _, err = AppendAndApply(db, changesetDir, lines)
	if err != nil {
		return nil, "", fmt.Errorf("decisions %v: plan markdown recorded, but db write failed: %w", ids, err)
	}
	return ids, path, nil
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
