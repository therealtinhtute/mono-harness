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
