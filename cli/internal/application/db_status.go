package application

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// contextCostStages are the six spine playbooks whose preconditions read
// the active plan today (docs/audit/workflow-harness-ceremony-audit.md,
// F1). `git` and `interview` are excluded: neither owns a durable plan
// section, so neither reads the plan file.
var contextCostStages = []string{"watzup", "brainstorm", "to-plan", "work", "check", "handoff"}

// StageContextCost is one stage's estimated read cost under today's
// full-plan-read path.
type StageContextCost struct {
	PlaybookBytes        int `json:"playbook_bytes"`
	EstimatedTokensToday int `json:"estimated_tokens_today"`
}

// ContextCostEstimate answers G4 (docs/audit/workflow-harness-ceremony-audit.md):
// the harness has no way to see its own context cost. It reports, per
// stage, what reading the current active plan in full would cost today —
// the read path this initiative's later phases replace with an index read.
type ContextCostEstimate struct {
	ActivePlanPath  string                      `json:"active_plan_path,omitempty"`
	ActivePlanBytes int                         `json:"active_plan_bytes"`
	Stages          map[string]StageContextCost `json:"stages"`
	Note            string                      `json:"note"`
}

// DBStatusView is the `db status` payload: schema position, per-table row
// counts, and the context-cost estimate.
type DBStatusView struct {
	SchemaVersion int                 `json:"schema_version"`
	Rows          map[string]int      `json:"rows"`
	ContextCost   ContextCostEstimate `json:"context_cost_estimate"`
}

// QueryDBStatus assembles the `db status` view. playbooksDir and
// activePlanGlob are passed in rather than hardcoded so this stays
// path-agnostic like the rest of the application layer; interfaces/db.go
// supplies the real repository-relative values.
func QueryDBStatus(db *sql.DB, playbooksDir, activePlanGlob string) (DBStatusView, error) {
	var schemaVersion int
	if err := db.QueryRow(`SELECT schema_version FROM meta LIMIT 1`).Scan(&schemaVersion); err != nil {
		return DBStatusView{}, fmt.Errorf("db status: read meta: %w", err)
	}

	tableNames, err := lifecycleTableNames(db)
	if err != nil {
		return DBStatusView{}, err
	}
	rows := make(map[string]int, len(tableNames))
	for _, table := range tableNames {
		var n int
		// table comes only from sqlite_master, never from external input.
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil { //nolint:gosec
			return DBStatusView{}, fmt.Errorf("db status: count %s: %w", table, err)
		}
		rows[table] = n
	}

	cost, err := estimateContextCost(db, playbooksDir, activePlanGlob)
	if err != nil {
		return DBStatusView{}, err
	}

	return DBStatusView{
		SchemaVersion: schemaVersion,
		Rows:          rows,
		ContextCost:   cost,
	}, nil
}

// lifecycleTableNames introspects the schema instead of hardcoding a table
// list, so a later migration (e.g. this initiative's own `decisions`
// table) is picked up automatically. `meta` is excluded: its one row's
// relevant fields are already surfaced as schema_version and fence.
func lifecycleTableNames(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type = 'table' AND name != 'meta' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("db status: list tables: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("db status: scan table name: %w", err)
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

// contextCostSectionStages names the stages whose playbook now reads a
// bounded slice of the active plan rather than the whole file (R7,
// docs/audit/sdlc-token-cache-audit.md — the section-read path P3's wave 1
// shipped: watzup.md's `query plan --section current-state`, work.md's
// `query plan --section phase`, handoff.md's `query plan --section
// current-state`). brainstorm/to-plan/check are absent: their playbooks
// still read the plan in full (brainstorm/to-plan define it; check's
// full-plan alignment review is audit, not ceremony — see check.md's own
// Preconditions step 1), so their estimate keeps modeling a full-plan read.
var contextCostSectionStages = map[string]bool{"watzup": true, "work": true, "handoff": true}

func estimateContextCost(db *sql.DB, playbooksDir, activePlanGlob string) (ContextCostEstimate, error) {
	cost := ContextCostEstimate{
		Stages: map[string]StageContextCost{},
		Note: "bytes/4 heuristic (docs/audit/workflow-harness-ceremony-audit.md's tokenizer " +
			"methodology), not an exact count. watzup/work/handoff report their own documented " +
			"section-read path (R7, docs/audit/sdlc-token-cache-audit.md) — Outcome+Current State " +
			"for watzup, the selected phase's block for work, Current State for handoff — not a " +
			"full-plan read; brainstorm/to-plan/check still read the plan in full, matching their " +
			"own playbooks, so their estimate keeps modeling that.",
	}

	planPath, planContent, err := activePlanContent(activePlanGlob)
	if err != nil {
		return cost, err
	}
	cost.ActivePlanPath = planPath
	cost.ActivePlanBytes = len(planContent)

	currentPhase, err := currentPhaseSlugOrEmpty(db)
	if err != nil {
		return cost, err
	}

	for _, stage := range contextCostStages {
		playbookBytes, err := fileSizeOrZero(filepath.Join(playbooksDir, stage+".md"))
		if err != nil {
			return cost, err
		}
		planReadBytes := len(planContent)
		if contextCostSectionStages[stage] {
			planReadBytes = stagePlanReadBytes(stage, planContent, currentPhase)
		}
		cost.Stages[stage] = StageContextCost{
			PlaybookBytes:        playbookBytes,
			EstimatedTokensToday: (playbookBytes + planReadBytes) / 4,
		}
	}
	return cost, nil
}

// stagePlanReadBytes models exactly what watzup/work/handoff's own
// playbook text says it reads from the plan (see contextCostSectionStages).
// A section that can't be found degrades to the full plan, mirroring
// `query plan`'s own degraded-read fallback (QueryPlanSection) — a
// malformed plan costs what a full read always cost, not less.
func stagePlanReadBytes(stage, planContent, currentPhase string) int {
	switch stage {
	case "watzup":
		return sectionBytesOrFull(planContent, "Outcome") + sectionBytesOrFull(planContent, "Current State and Next Action")
	case "work":
		if currentPhase == "" {
			return 0
		}
		if body, ok := extractPlanPhaseBlock(planContent, currentPhase); ok {
			return len(body)
		}
		return len(planContent)
	case "handoff":
		return sectionBytesOrFull(planContent, "Current State and Next Action")
	default:
		return len(planContent)
	}
}

func sectionBytesOrFull(planContent, name string) int {
	if body, ok := extractPlanSection(planContent, name); ok {
		return len(body)
	}
	return len(planContent)
}

func currentPhaseSlugOrEmpty(db *sql.DB) (string, error) {
	var slug sql.NullString
	if err := db.QueryRow(`SELECT current_phase FROM meta LIMIT 1`).Scan(&slug); err != nil {
		return "", fmt.Errorf("db status: read current_phase: %w", err)
	}
	return slug.String, nil
}

// activePlanContent finds the first non-empty match for the active-plan
// glob, mirroring next.go's findActivePlans preference order, and returns
// its path and content. No match (or every match empty) returns ("", "", nil).
func activePlanContent(glob string) (path, content string, err error) {
	matches, err := filepath.Glob(glob)
	if err != nil {
		return "", "", fmt.Errorf("db status: glob active plans: %w", err)
	}
	sort.Strings(matches)
	for _, m := range matches {
		data, readErr := os.ReadFile(m)
		if readErr != nil {
			return "", "", fmt.Errorf("db status: read %s: %w", m, readErr)
		}
		if strings.TrimSpace(string(data)) != "" {
			return m, string(data), nil
		}
	}
	return "", "", nil
}

func fileSizeOrZero(path string) (int, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("db status: stat %s: %w", path, err)
	}
	return int(info.Size()), nil
}
