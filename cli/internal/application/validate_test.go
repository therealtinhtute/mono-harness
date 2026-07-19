package application

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oklog/ulid/v2"
)

func TestValidateChainValidFixture(t *testing.T) {
	result, err := Validate(nil, filepath.Join("..", "..", "testdata", "chain-valid"))
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !result.Valid {
		t.Fatalf("valid = false (findings=%v), want true", result.Findings)
	}
	if len(result.Findings) != 1 || result.Findings[0].Issue != "not_yet_implemented" {
		t.Fatalf("findings = %v, want exactly one not_yet_implemented finding", result.Findings)
	}
}

func TestValidateChainBrokenFixture(t *testing.T) {
	result, err := Validate(nil, filepath.Join("..", "..", "testdata", "chain-broken"))
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.Valid {
		t.Fatalf("valid = true, want false (chain-broken has one broken cross-link)")
	}
	found := false
	for _, f := range result.Findings {
		if f.Link == "RUN->CHECK" && f.Issue == "broken_link" {
			found = true
		}
	}
	if !found {
		t.Fatalf("findings = %v, want a RUN->CHECK broken_link finding naming the break", result.Findings)
	}
}

func TestValidateMissingSpec(t *testing.T) {
	result, err := Validate(nil, t.TempDir())
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.Valid {
		t.Fatalf("valid = true, want false (no SPEC.md at all)")
	}
	if len(result.Findings) != 1 || result.Findings[0].Link != "SPEC->PLAN" || result.Findings[0].Issue != "missing_key" {
		t.Fatalf("findings = %v, want a single SPEC->PLAN missing_key finding", result.Findings)
	}
}

// TestValidateMalformedULID exercises the "ULID formats" leg from
// cli-domain-CONTEXT.md's Locked Decisions: a present-but-malformed id
// value is reported (as missing_key, since CONTRACT.md's issue enum has
// no dedicated slot for it) rather than silently accepted.
func TestValidateMalformedULID(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "runs", "work")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatalf("mkdir run dir: %v", err)
	}
	content := "---\nid: not-a-real-ulid\ntype: run\nphase: sample-phase\nlane: normal\nplan_id: 01KXQKT39GV8YK5QBBDCJ33A32\ntrace_ids: []\ncreated: 2026-01-01\nupdated: 2026-01-01\n---\n\n# COOK RUN\n"
	if err := os.WriteFile(filepath.Join(runDir, "20260101-1200-sample-phase.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write run: %v", err)
	}

	result, err := Validate(nil, root)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.Valid {
		t.Fatalf("valid = true, want false (malformed RUN id)")
	}
	found := false
	for _, f := range result.Findings {
		if f.Link == "PLAN->RUN" && f.Issue == "missing_key" && strings.Contains(f.Detail, "not a valid ULID") {
			found = true
		}
	}
	if !found {
		t.Fatalf("findings = %v, want a PLAN->RUN missing_key finding naming the malformed ULID", result.Findings)
	}
}

// TestValidateStalePointerAgainstDB exercises the "freshness vs DB" leg:
// a RUN file whose id has no matching row in a live db's runs table.
func TestValidateStalePointerAgainstDB(t *testing.T) {
	db, _ := freshDB(t)
	root := t.TempDir()

	planDir := filepath.Join(root, "planning", "phases", "sample-phase")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("mkdir plan dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "sample-phase-PLAN.md"), []byte("# Plan: sample-phase\n"), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}

	runDir := filepath.Join(root, "runs", "work")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatalf("mkdir run dir: %v", err)
	}
	runID := ulid.Make().String()
	content := fmt.Sprintf("---\nid: %s\ntype: run\nphase: sample-phase\nlane: normal\nplan_id: %s\ntrace_ids: []\ncreated: 2026-01-01\nupdated: 2026-01-01\n---\n\n# COOK RUN\n", runID, ulid.Make().String())
	if err := os.WriteFile(filepath.Join(runDir, "20260101-1200-sample-phase.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write run: %v", err)
	}

	result, err := Validate(db, root)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.Valid {
		t.Fatalf("valid = true, want false (run id has no matching db row)")
	}
	found := false
	for _, f := range result.Findings {
		if f.Link == "PLAN->RUN" && f.Issue == "stale_pointer" {
			found = true
		}
	}
	if !found {
		t.Fatalf("findings = %v, want a PLAN->RUN stale_pointer finding", result.Findings)
	}
}

// writeMinimalSpec writes a valid-shape planning/SPEC.md (real ULID id) so
// the SPEC->PLAN leg reports the expected not_yet_implemented finding
// instead of a blocking missing_key for "no SPEC at all" — isolating the
// RUN/CHECK-side assertions these mode-aware tests actually care about.
func writeMinimalSpec(t *testing.T, root string) {
	t.Helper()
	planningDir := filepath.Join(root, "planning")
	if err := os.MkdirAll(planningDir, 0o755); err != nil {
		t.Fatalf("mkdir planning dir: %v", err)
	}
	content := fmt.Sprintf("---\nid: %s\ntype: spec\nphase: none\nlane: tiny\ncreated: 2026-01-01\nupdated: 2026-01-01\n---\n\n# SPEC: fixture\n", ulid.Make().String())
	if err := os.WriteFile(filepath.Join(planningDir, "SPEC.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
}

// TestValidateChainSimpleModeFixture exercises harness-mode-parity's whole
// point: a simple-mode-produced chain (phase-less, plan-less RUN; an
// unregistered CHECK) must validate clean, matching the shape #38's pilot
// chain actually produced.
func TestValidateChainSimpleModeFixture(t *testing.T) {
	result, err := Validate(nil, filepath.Join("..", "..", "testdata", "chain-simple-mode"))
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !result.Valid {
		t.Fatalf("valid = false (findings=%v), want true", result.Findings)
	}
	if len(result.Findings) != 1 || result.Findings[0].Issue != "not_yet_implemented" {
		t.Fatalf("findings = %v, want exactly one not_yet_implemented finding", result.Findings)
	}
}

// TestValidateSimpleModeRunSkipsDBAndPhaseChecks: a RUN with mode: simple,
// phase: none, plan_id: none, and no matching runs-table row must not
// trigger broken_link/missing_key/stale_pointer — those are expected
// simple-mode shape, not defects (harness-mode-parity CONTEXT: Locked
// Decisions).
func TestValidateSimpleModeRunSkipsDBAndPhaseChecks(t *testing.T) {
	db, _ := freshDB(t)
	root := t.TempDir()
	writeMinimalSpec(t, root)

	runDir := filepath.Join(root, "runs", "work")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatalf("mkdir run dir: %v", err)
	}
	runID := ulid.Make().String()
	content := fmt.Sprintf("---\nid: %s\ntype: run\nphase: none\nlane: tiny\nmode: simple\nplan_id: none\ntrace_ids: []\ncreated: 2026-01-01\nupdated: 2026-01-01\n---\n\n# COOK RUN\n", runID)
	if err := os.WriteFile(filepath.Join(runDir, "20260101-1200-sample-task.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write run: %v", err)
	}

	result, err := Validate(db, root)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !result.Valid {
		t.Fatalf("valid = false (findings=%v), want true (simple mode has no story/plan/db row by design)", result.Findings)
	}
}

// TestValidateSimpleModeCheckSkipsDBStalePointer: a CHECK with mode:
// simple pointing at a real RUN file (doc-to-doc link, unaffected by mode)
// but with no matching checks-table row must not trigger stale_pointer.
func TestValidateSimpleModeCheckSkipsDBStalePointer(t *testing.T) {
	db, _ := freshDB(t)
	root := t.TempDir()
	writeMinimalSpec(t, root)

	runDir := filepath.Join(root, "runs", "work")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatalf("mkdir run dir: %v", err)
	}
	runID := ulid.Make().String()
	runContent := fmt.Sprintf("---\nid: %s\ntype: run\nphase: none\nlane: tiny\nmode: simple\nplan_id: none\ntrace_ids: []\ncreated: 2026-01-01\nupdated: 2026-01-01\n---\n\n# COOK RUN\n", runID)
	if err := os.WriteFile(filepath.Join(runDir, "20260101-1200-sample-task.md"), []byte(runContent), 0o644); err != nil {
		t.Fatalf("write run: %v", err)
	}

	checkDir := filepath.Join(root, "reports", "check")
	if err := os.MkdirAll(checkDir, 0o755); err != nil {
		t.Fatalf("mkdir check dir: %v", err)
	}
	checkContent := fmt.Sprintf("---\nid: %s\ntype: check\nphase: none\nlane: tiny\nmode: simple\nrun_id: %s\ncreated: 2026-01-01\nupdated: 2026-01-01\n---\n\n# CHECK REPORT\n", ulid.Make().String(), runID)
	if err := os.WriteFile(filepath.Join(checkDir, "20260101-1300-sample-task.md"), []byte(checkContent), 0o644); err != nil {
		t.Fatalf("write check: %v", err)
	}

	result, err := Validate(db, root)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !result.Valid {
		t.Fatalf("valid = false (findings=%v), want true (simple-mode check has no checks-table row by design)", result.Findings)
	}
}

// TestValidateSimpleModeStillRequiresValidID: mode: simple exempts
// phase/plan/DB checks, but a RUN's own id must still be a well-formed
// ULID — artifact hygiene stays universal (harness-mode-parity CONTEXT:
// Locked Decisions).
func TestValidateSimpleModeStillRequiresValidID(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "runs", "work")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatalf("mkdir run dir: %v", err)
	}
	content := "---\nid: not-a-real-ulid\ntype: run\nphase: none\nlane: tiny\nmode: simple\nplan_id: none\ntrace_ids: []\ncreated: 2026-01-01\nupdated: 2026-01-01\n---\n\n# COOK RUN\n"
	if err := os.WriteFile(filepath.Join(runDir, "20260101-1200-sample-task.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write run: %v", err)
	}

	result, err := Validate(nil, root)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.Valid {
		t.Fatalf("valid = true, want false (malformed RUN id, even in simple mode)")
	}
}
