package application

import (
	"database/sql"
	"fmt"
	"strings"
)

// contextTraceTail is the window policy cap on a phase's trace history
// inside a context packet (P4 wave 2) — bounds packet size across phase
// counts (the F1 quadratic-read finding, docs/audit/workflow-harness-
// ceremony-audit.md) instead of returning every trace a long-running
// initiative has accumulated.
const contextTraceTail = 30

// ContextPacket is preflight's stage-shaped memory (R4): one call in place
// of the resume/query-phases/query-traces menu a stage would otherwise
// assemble by hand. Position/latest IDs/drift/readiness reuse Resume's own
// derivation, so the packet and `resume --json` can never disagree about
// lifecycle position. Phases and Traces are populated only for the stages
// that reference them (BuildContextPacket), so a packet never carries a
// field its own stage's playbook has no use for.
type ContextPacket struct {
	Position        Position        `json:"position"`
	LatestRunID     *string         `json:"latest_run_id"`
	LatestCheckID   *string         `json:"latest_check_id"`
	LatestHandoffID *string         `json:"latest_handoff_id"`
	Drift           []DriftFinding  `json:"drift"`
	Readiness       string          `json:"readiness"`
	Phases          []PhaseView     `json:"phases,omitempty"`
	Traces          []TraceView     `json:"traces,omitempty"`
	Memories        []MemoryListView `json:"memories,omitempty"`
	Omitted         []OmittedField  `json:"omitted,omitempty"`
}

// OmittedField declares a field a context packet bounded rather than
// returning in full, and how to fetch what was left out (R5 — "any
// bounded packet declares what it omitted and how to fetch it").
type OmittedField struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
	Fetch  string `json:"fetch"`
}

// contextPhasesStages names the stages whose playbook reads the phases
// list (`query phases`) today — work.md and handoff.md's "Load state"
// steps, and check.md's gate/full mode (R6, docs/audit/sdlc-token-cache-
// audit.md — check.md step 1's separate `zharness resume --json` call is
// replaced by this same packet). watzup.md does not call `query phases`,
// so its packet omits Phases entirely rather than carrying a field it
// never references.
var contextPhasesStages = map[string]bool{"work": true, "handoff": true, "check": true}

// BuildContextPacket assembles the stage-shaped packet for stage (R4).
// Callers gate which stages receive a packet at all, and for which modes
// (interfaces/preflight.go's contextEligibleStages/checkContextEligible) —
// check's response-only review/bounded modes still perform zero durable
// reads or writes beyond what their own playbook already does, so only
// its durable gate/full modes reach this function (R6 supersedes this
// initiative's own earlier NG2, which kept check's reads entirely
// separate; docs/audit/workflow-harness-ceremony-audit.md).
func BuildContextPacket(db *sql.DB, stage, version string) (*ContextPacket, error) {
	resumeView, err := Resume(db, version)
	if err != nil {
		return nil, fmt.Errorf("build context packet: %w", err)
	}

	pkg := &ContextPacket{
		Position:        resumeView.Position,
		LatestRunID:     resumeView.LatestRunID,
		LatestCheckID:   resumeView.LatestCheckID,
		LatestHandoffID: resumeView.LatestHandoffID,
		Drift:           resumeView.Drift,
		Readiness:       resumeView.Readiness,
	}

	if contextPhasesStages[stage] {
		phases, err := QueryPhases(db)
		if err != nil {
			return nil, fmt.Errorf("build context packet: %w", err)
		}
		pkg.Phases = phases
	}

	if resumeView.Position.CurrentPhase != nil {
		phaseSlug := *resumeView.Position.CurrentPhase
		total, err := countTracesForPhase(db, phaseSlug)
		if err != nil {
			return nil, fmt.Errorf("build context packet: %w", err)
		}
		traces, err := QueryTracesByPhase(db, phaseSlug, contextTraceTail)
		if err != nil {
			return nil, fmt.Errorf("build context packet: %w", err)
		}
		pkg.Traces = traces
		if total > contextTraceTail {
			pkg.Omitted = append(pkg.Omitted, OmittedField{
				Field:  "traces",
				Reason: fmt.Sprintf("phase %q has %d traces, capped at %d", phaseSlug, total, contextTraceTail),
				Fetch:  fmt.Sprintf("zharness query traces --phase %s --tail 0 --json", phaseSlug),
			})
		}
	}

	// Memories: include all active memories for agent context (R2 — superseded
	// entries are excluded by default, matching MemoryQuery's default).
	if err := populateContextMemories(db, pkg); err != nil {
		return nil, fmt.Errorf("build context packet: memories: %w", err)
	}

	return pkg, nil
}

func populateContextMemories(db *sql.DB, pkg *ContextPacket) error {
	// No memories table yet (pre-migration) — degrade, return empty.
	var tableCount int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='memories'`).Scan(&tableCount); err != nil {
		return err
	}
	if tableCount == 0 {
		return nil
	}
	rows, err := db.Query(`SELECT id, path, type, scope, plan_id, created_at, superseded_by, superseded_at FROM memories WHERE superseded_by IS NULL ORDER BY created_at DESC, id DESC`)
	if err != nil {
		if err.Error() != "" && containsNoSuchColumn(err.Error()) {
			return populateContextMemoriesLegacy(db, pkg)
		}
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var v MemoryListView
		var planIDCol sql.NullString
		var supBy sql.NullString
		var supAt sql.NullString
		if err := rows.Scan(&v.ID, &v.Path, &v.Type, &v.Scope, &planIDCol, &v.CreatedAt, &supBy, &supAt); err != nil {
			return err
		}
		v.PlanID = nullableString(planIDCol)
		if supBy.Valid && supBy.String != "" {
			v.SupersededBy = &supBy.String
		}
		if supAt.Valid && supAt.String != "" {
			v.SupersededAt = &supAt.String
		}
		pkg.Memories = append(pkg.Memories, v)
	}
	return rows.Err()
}

func populateContextMemoriesLegacy(db *sql.DB, pkg *ContextPacket) error {
	rows, err := db.Query(`SELECT id, path, type, scope, plan_id, created_at FROM memories ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var v MemoryListView
		var planIDCol sql.NullString
		if err := rows.Scan(&v.ID, &v.Path, &v.Type, &v.Scope, &planIDCol, &v.CreatedAt); err != nil {
			return err
		}
		v.PlanID = nullableString(planIDCol)
		pkg.Memories = append(pkg.Memories, v)
	}
	return rows.Err()
}

func containsNoSuchColumn(msg string) bool {
	return strings.Contains(msg, "no such column")
}
