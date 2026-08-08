package application

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
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
// counts, true changeset coverage (see ChangesetStatus), and the context-
// cost estimate.
type DBStatusView struct {
	SchemaVersion        int                 `json:"schema_version"`
	Fence                string              `json:"fence"`
	Rows                 map[string]int      `json:"rows"`
	Pending              []string            `json:"pending"`
	UnverifiedBelowFence []string            `json:"unverified_below_fence"`
	ContextCost          ContextCostEstimate `json:"context_cost_estimate"`
}

// QueryDBStatus assembles the `db status` view. playbooksDir and
// activePlanGlob are passed in rather than hardcoded so this stays
// path-agnostic like the rest of the application layer; interfaces/db.go
// supplies the real repository-relative values.
func QueryDBStatus(db *sql.DB, changesetDir, playbooksDir, activePlanGlob string) (DBStatusView, error) {
	var schemaVersion int
	var fence sql.NullString
	if err := db.QueryRow(`SELECT schema_version, last_applied_changeset FROM meta LIMIT 1`).Scan(&schemaVersion, &fence); err != nil {
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

	pending, _, _, unverified, err := infrastructure.ChangesetStatus(db, changesetDir)
	if err != nil {
		return DBStatusView{}, fmt.Errorf("db status: changeset status: %w", err)
	}
	if pending == nil {
		pending = []string{}
	}
	if unverified == nil {
		unverified = []string{}
	}

	cost, err := estimateContextCost(playbooksDir, activePlanGlob)
	if err != nil {
		return DBStatusView{}, err
	}

	return DBStatusView{
		SchemaVersion:        schemaVersion,
		Fence:                fence.String,
		Rows:                 rows,
		Pending:              pending,
		UnverifiedBelowFence: unverified,
		ContextCost:          cost,
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

func estimateContextCost(playbooksDir, activePlanGlob string) (ContextCostEstimate, error) {
	cost := ContextCostEstimate{
		Stages: map[string]StageContextCost{},
		Note: "bytes/4 heuristic (docs/audit/workflow-harness-ceremony-audit.md's tokenizer " +
			"methodology), not an exact count. Reflects today's full-plan-read path; the " +
			"index read path this initiative adds is not yet reflected here.",
	}

	planPath, planBytes, err := activePlanSize(activePlanGlob)
	if err != nil {
		return cost, err
	}
	cost.ActivePlanPath = planPath
	cost.ActivePlanBytes = planBytes

	for _, stage := range contextCostStages {
		playbookBytes, err := fileSizeOrZero(filepath.Join(playbooksDir, stage+".md"))
		if err != nil {
			return cost, err
		}
		// Every stage's precondition step reads the whole active plan
		// today, on top of its own playbook (F1) — that sum is what P5
		// changes, so it is what "today" must report.
		cost.Stages[stage] = StageContextCost{
			PlaybookBytes:        playbookBytes,
			EstimatedTokensToday: (playbookBytes + planBytes) / 4,
		}
	}
	return cost, nil
}

// activePlanSize finds the first non-empty match for the active-plan glob,
// mirroring next.go's findActivePlans preference order, and returns its
// path and byte size. No match (or every match empty) returns ("", 0, nil).
func activePlanSize(glob string) (path string, size int, err error) {
	matches, err := filepath.Glob(glob)
	if err != nil {
		return "", 0, fmt.Errorf("db status: glob active plans: %w", err)
	}
	sort.Strings(matches)
	for _, m := range matches {
		info, statErr := os.Stat(m)
		if statErr != nil {
			return "", 0, fmt.Errorf("db status: stat %s: %w", m, statErr)
		}
		if info.Size() > 0 {
			return m, int(info.Size()), nil
		}
	}
	return "", 0, nil
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
