package application

import (
	"database/sql"
	"fmt"
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
	Position        Position       `json:"position"`
	LatestRunID     *string        `json:"latest_run_id"`
	LatestCheckID   *string        `json:"latest_check_id"`
	LatestHandoffID *string        `json:"latest_handoff_id"`
	Drift           []DriftFinding `json:"drift"`
	Readiness       string         `json:"readiness"`
	Phases          []PhaseView    `json:"phases,omitempty"`
	Traces          []TraceView    `json:"traces,omitempty"`
	Omitted         []OmittedField `json:"omitted,omitempty"`
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
// steps. watzup.md does not call `query phases`, so its packet omits
// Phases entirely rather than carrying a field it never references.
var contextPhasesStages = map[string]bool{"work": true, "handoff": true}

// BuildContextPacket assembles the stage-shaped packet for stage (R4).
// Callers gate which stages receive a packet at all (interfaces/preflight.go's
// contextEligibleStages) — check.md keeps its own separate resume/query
// calls by design (NG2: its full-plan read is audit, not ceremony this
// initiative optimizes), so check never reaches this function.
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

	return pkg, nil
}
