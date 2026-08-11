package interfaces

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const queryPlanFixture = "# Plan: Demo\n" +
	"\n" +
	"## Phases and Verification\n" +
	"### phase_slug: `p1-first`\n" +
	"- status: planned\n" +
	"- goal: first phase\n" +
	"\n" +
	"## Current State and Next Action\n" +
	"- active_phase: none\n"

func writeQueryPlanFixture(t *testing.T) {
	t.Helper()
	path := filepath.Join("docs", "plans", "active", "demo.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(queryPlanFixture), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// TestQueryPlanCurrentStateAndPhaseRoundTrip proves the CLI wiring for the
// ceremony audit's P3 proposal (docs/audit/workflow-harness-ceremony-audit.md):
// a plan-slice read that needs no harness.db at all.
func TestQueryPlanCurrentStateAndPhaseRoundTrip(t *testing.T) {
	t.Chdir(t.TempDir())
	writeQueryPlanFixture(t)

	out, err := runDBCommand(t, "query", "plan", "--section", "current-state", "--json")
	if err != nil {
		t.Fatalf("query plan --section current-state: %v (output=%s)", err, out)
	}
	var v struct {
		Content  string `json:"content"`
		Degraded bool   `json:"degraded"`
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("decode output %q: %v", out, err)
	}
	if v.Degraded || v.Content != "- active_phase: none" {
		t.Fatalf("current-state = %+v, want the Current State body undegraded", v)
	}

	out, err = runDBCommand(t, "query", "plan", "--section", "phase", "--phase", "p1-first", "--json")
	if err != nil {
		t.Fatalf("query plan --section phase: %v (output=%s)", err, out)
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("decode output %q: %v", out, err)
	}
	if v.Degraded {
		t.Fatalf("phase query unexpectedly degraded: %+v", v)
	}
}

// queryPlanListFormFixture mirrors the list form to-plan actually produces
// (no `###` headings) — see plan_query_test.go's planQueryListFixture in
// the application layer for the full rationale.
const queryPlanListFormFixture = "# Plan: Demo\n" +
	"\n" +
	"## Phases and Verification\n" +
	"- phases:\n" +
	"  - phase_slug: p1-first\n" +
	"    status: planned\n" +
	"    goal: first phase\n" +
	"\n" +
	"## Current State and Next Action\n" +
	"- active_phase: none\n"

func TestQueryPlanListFormPhaseRoundTrip(t *testing.T) {
	t.Chdir(t.TempDir())
	path := filepath.Join("docs", "plans", "active", "demo.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(queryPlanListFormFixture), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	out, err := runDBCommand(t, "query", "plan", "--section", "phase", "--phase", "p1-first", "--json")
	if err != nil {
		t.Fatalf("query plan --section phase: %v (output=%s)", err, out)
	}
	var v struct {
		Content  string `json:"content"`
		Degraded bool   `json:"degraded"`
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("decode output %q: %v", out, err)
	}
	if v.Degraded {
		t.Fatalf("list-form phase query unexpectedly degraded: %+v", v)
	}
}

func TestQueryPlanUnknownSectionIsUserError(t *testing.T) {
	t.Chdir(t.TempDir())
	writeQueryPlanFixture(t)

	_, err := runDBCommand(t, "query", "plan", "--section", "bogus", "--json")
	if err == nil {
		t.Fatalf("query plan --section bogus: want an error")
	}
}

func TestQueryPlanNoActivePlanIsUserError(t *testing.T) {
	t.Chdir(t.TempDir())

	_, err := runDBCommand(t, "query", "plan", "--section", "current-state", "--json")
	if err == nil {
		t.Fatalf("query plan with no active plan: want an error")
	}
}
