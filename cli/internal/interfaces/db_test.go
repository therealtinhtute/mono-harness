package interfaces

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// runDBCommand executes a zharness command through the real CLI surface,
// returning stdout, the command error (nil on success), and stderr for
// diagnostics — unlike runIDCommand/executeReadOnlyJSONCommand, it does not
// fail the test on error, since several tests here exercise error paths.
func runDBCommand(t *testing.T, args ...string) (stdout string, err error) {
	t.Helper()
	jsonOutput = false
	cmd := NewRootCmd("dev")
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), err
}

func TestDBRebuildRequiresConfirmation(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}
	before, statErr := os.Stat(dbPath)
	if statErr != nil {
		t.Fatalf("Stat before rebuild attempt: %v", statErr)
	}

	_, err := runDBCommand(t, "db", "rebuild", "--json")
	ce, ok := err.(*cliError)
	if !ok || ce.Code != "confirmation_required" {
		t.Fatalf("db rebuild without --yes = %v, want confirmation_required cliError", err)
	}

	after, statErr := os.Stat(dbPath)
	if statErr != nil {
		t.Fatalf("Stat after rebuild attempt: %v", statErr)
	}
	if before.ModTime() != after.ModTime() || before.Size() != after.Size() {
		t.Fatalf("db rebuild without --yes mutated the database: before=%+v after=%+v", before, after)
	}
}

func TestDBRebuildDoesNotTouchDocs(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}
	agentsBefore, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("ReadFile AGENTS.md after init: %v", err)
	}

	if _, err := runDBCommand(t, "db", "rebuild", "--yes", "--json"); err != nil {
		t.Fatalf("db rebuild --yes: %v", err)
	}

	agentsAfter, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("ReadFile AGENTS.md after rebuild: %v", err)
	}
	if string(agentsBefore) != string(agentsAfter) {
		t.Fatal("db rebuild modified AGENTS.md; it must be database-only, unlike init")
	}
}

// TestDBRebuildRecoversFromMarkdownNotChangesets is the CLI-level successor
// to the old changeset-replay recovery test (P3,
// docs/plans/active/harness-markdown-truth.md, Wave 1 Task 2): `db rebuild
// --yes` now reconstructs from committed plan markdown alone, so a story
// whose only committed record is a phase block in docs/plans/active/*.md —
// even if the DB row backing it was lost (e.g. an interleaved-machine
// changeset that was written but never applied, docs/audit/workflow-harness-ceremony-audit.md
// F5) — is recovered; a story that exists only as an applied DB row, with
// no phase block anywhere in committed markdown, is correctly NOT recovered
// (NG4: rebuild is destructive to DB-only state that never made it into
// markdown, by design).
func TestDBRebuildRecoversFromMarkdownNotChangesets(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}

	if err := os.MkdirAll(filepath.Join("docs", "plans", "active"), 0o755); err != nil {
		t.Fatalf("MkdirAll active plans: %v", err)
	}
	a1ID, b1ID, dbOnlyID := ulid.Make().String(), ulid.Make().String(), ulid.Make().String()
	plan := "---\nid: " + ulid.Make().String() + "\ntype: plan\nstatus: active\n---\n\n" +
		"# Plan: Rebuild recovery fixture\n\n" +
		"## Phases and Verification\n" +
		"- phases:\n" +
		"  - phase_slug: a1\n" +
		"    story_id: " + a1ID + "\n" +
		"    status: planned\n" +
		"    goal: machine A wave 1\n" +
		"  - phase_slug: b1\n" +
		"    story_id: " + b1ID + "\n" +
		"    status: planned\n" +
		"    goal: machine B wave 1\n\n" +
		"## Progress\n- none\n"
	if err := os.WriteFile(filepath.Join("docs", "plans", "active", "demo.md"), []byte(plan), 0o644); err != nil {
		t.Fatalf("WriteFile active plan: %v", err)
	}

	db, err := infrastructure.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	// a1 was applied normally; b1's changeset never made it into the db
	// (the interleaved-machine loss scenario) even though its phase block
	// is committed; dbOnly has a DB row with no markdown counterpart at all.
	if _, err := db.Exec(`INSERT INTO stories (id, slug, goal, status, created_at) VALUES (?, 'a1', 'machine A wave 1', 'planned', '2026-08-07T08:00:00Z')`, a1ID); err != nil {
		t.Fatalf("seed a1: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO stories (id, slug, goal, status, created_at) VALUES (?, 'db-only', 'never committed to markdown', 'planned', '2026-08-07T08:00:30Z')`, dbOnlyID); err != nil {
		t.Fatalf("seed db-only: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Before rebuild: b1 is missing (the loss), db-only is present.
	slugs := queryStorySlugs(t)
	if contains(slugs, "b1") {
		t.Fatalf("story b1 present before rebuild = %v, want it missing (setup did not reproduce the loss)", slugs)
	}
	if !contains(slugs, "db-only") {
		t.Fatalf("story db-only present before rebuild = %v, want it present (setup did not seed it)", slugs)
	}

	out, err := runDBCommand(t, "db", "rebuild", "--yes", "--json")
	if err != nil {
		t.Fatalf("db rebuild --yes: %v (output=%s)", err, out)
	}
	var result struct {
		Status        string `json:"status"`
		SchemaVersion int    `json:"schema_version"`
		Stories       int    `json:"stories"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("decode db rebuild output %q: %v", out, err)
	}
	if result.Status != "rebuilt" || result.Stories != 2 {
		t.Fatalf("db rebuild result = %+v, want status=rebuilt stories=2 (a1, b1)", result)
	}

	// After rebuild: a1 and b1 (both markdown-backed) are present; db-only
	// (no markdown counterpart) is correctly gone.
	slugs = queryStorySlugs(t)
	for _, want := range []string{"a1", "b1"} {
		if !contains(slugs, want) {
			t.Fatalf("story %s missing after rebuild; slugs = %v", want, slugs)
		}
	}
	if contains(slugs, "db-only") {
		t.Fatalf("story db-only present after rebuild = %v, want it gone (no markdown backing it, per NG4)", slugs)
	}
}

func TestDBStatusReportsSchemaFenceRowsAndContextCost(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := os.MkdirAll(filepath.Join("docs", "plans", "active"), 0o755); err != nil {
		t.Fatalf("MkdirAll active plans: %v", err)
	}
	// A real phase block and a real Current State section, both smaller
	// than the whole plan, so a bounded per-stage read is provably cheaper
	// than a full-plan read (R7, docs/audit/sdlc-token-cache-audit.md).
	plan := []byte("---\nid: x\n---\n# Plan\n\n## Outcome\nresult\n\n" +
		"## Phases and Verification\n### phase_slug: `alpha`\n- status: in-progress\n- goal: phase alpha goal\n\n" +
		"## Current State and Next Action\n- active_phase: alpha\n- lifecycle_status: in-progress\n")
	if err := os.WriteFile(filepath.Join("docs", "plans", "active", "demo.md"), plan, 0o644); err != nil {
		t.Fatalf("WriteFile active plan: %v", err)
	}

	db, err := infrastructure.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	storyID := ulid.Make().String()
	if _, err := db.Exec(`INSERT INTO stories (id, slug, goal, status, created_at)
		VALUES (?, 'alpha', 'goal', 'planned', '2026-08-07T00:00:00Z')`, storyID); err != nil {
		t.Fatalf("seed story: %v", err)
	}
	if _, err := db.Exec(`UPDATE meta SET current_phase = 'alpha'`); err != nil {
		t.Fatalf("seed current_phase: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	out, err := runDBCommand(t, "db", "status", "--json")
	if err != nil {
		t.Fatalf("db status: %v", err)
	}
	var view struct {
		SchemaVersion int            `json:"schema_version"`
		Rows          map[string]int `json:"rows"`
		ContextCost   struct {
			ActivePlanPath  string `json:"active_plan_path"`
			ActivePlanBytes int    `json:"active_plan_bytes"`
			Stages          map[string]struct {
				PlaybookBytes        int `json:"playbook_bytes"`
				EstimatedTokensToday int `json:"estimated_tokens_today"`
			} `json:"stages"`
			Note string `json:"note"`
		} `json:"context_cost_estimate"`
	}
	if err := json.Unmarshal([]byte(out), &view); err != nil {
		t.Fatalf("decode db status output %q: %v", out, err)
	}

	if view.SchemaVersion < 1 {
		t.Fatalf("schema_version = %d, want >= 1", view.SchemaVersion)
	}
	if view.Rows["stories"] != 1 {
		t.Fatalf("rows[stories] = %d, want 1", view.Rows["stories"])
	}
	if view.ContextCost.ActivePlanPath == "" || view.ContextCost.ActivePlanBytes != len(plan) {
		t.Fatalf("context cost plan = (%q, %d), want (non-empty, %d)", view.ContextCost.ActivePlanPath, view.ContextCost.ActivePlanBytes, len(plan))
	}

	fullPlanEstimate := func(stage string) int {
		s, ok := view.ContextCost.Stages[stage]
		if !ok {
			t.Fatalf("context cost stages missing %q", stage)
		}
		return (s.PlaybookBytes + len(plan)) / 4
	}

	// watzup/work/handoff read a bounded section (R7), so their estimate
	// must be strictly cheaper than a full-plan read of the same plan —
	// this plan's Outcome/phase-block/Current State sections are each
	// smaller than the whole file by construction above.
	for _, stage := range []string{"watzup", "work", "handoff"} {
		s, ok := view.ContextCost.Stages[stage]
		if !ok {
			t.Fatalf("context cost stages missing %q", stage)
		}
		if s.EstimatedTokensToday >= fullPlanEstimate(stage) {
			t.Fatalf("%s estimated_tokens_today = %d, want less than the full-plan estimate %d (section-read path, R7)", stage, s.EstimatedTokensToday, fullPlanEstimate(stage))
		}
		if s.EstimatedTokensToday <= s.PlaybookBytes/4 {
			t.Fatalf("%s estimated_tokens_today = %d, want more than playbook-only %d (its section must contribute real bytes)", stage, s.EstimatedTokensToday, s.PlaybookBytes/4)
		}
	}

	// brainstorm/to-plan/check still read the plan in full — their
	// estimate must be unchanged, exactly (playbook_bytes+plan_bytes)/4.
	for _, stage := range []string{"brainstorm", "to-plan", "check"} {
		s, ok := view.ContextCost.Stages[stage]
		if !ok {
			t.Fatalf("context cost stages missing %q", stage)
		}
		if want := fullPlanEstimate(stage); s.EstimatedTokensToday != want {
			t.Fatalf("%s estimated_tokens_today = %d, want %d (unchanged full-plan read)", stage, s.EstimatedTokensToday, want)
		}
	}

	if view.ContextCost.Note == "" {
		t.Fatal("context cost note is empty, want the heuristic caveat")
	}
}

func TestDBStatusIsReadOnly(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}
	before, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("Stat before db status: %v", err)
	}
	if _, err := runDBCommand(t, "db", "status", "--json"); err != nil {
		t.Fatalf("db status: %v", err)
	}
	after, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("Stat after db status: %v", err)
	}
	if before.ModTime() != after.ModTime() || before.Size() != after.Size() {
		t.Fatalf("db status mutated the database: before=%+v after=%+v", before, after)
	}
}

func queryStorySlugs(t *testing.T) []string {
	t.Helper()
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		t.Fatalf("Open for slug query: %v", err)
	}
	defer db.Close()
	rows, err := db.Query(`SELECT slug FROM stories ORDER BY created_at`)
	if err != nil {
		t.Fatalf("query slugs: %v", err)
	}
	defer rows.Close()
	var slugs []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatalf("scan slug: %v", err)
		}
		slugs = append(slugs, s)
	}
	return slugs
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
