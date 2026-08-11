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

// TestDBRebuildRecoversInterleavedMachineChangeset is the CLI-level
// counterpart of TestChangesetStatusFlagsInterleavedMachineChangesetNeverApplied:
// it reproduces the same two-machine data-loss scenario and proves `db
// rebuild --yes` — not just detection — actually recovers the row that
// direct incremental apply could not (docs/audit/workflow-harness-ceremony-audit.md, F5).
func TestDBRebuildRecoversInterleavedMachineChangeset(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}

	// init already advanced the fence to a real "now" ULID (managed_docs
	// scaffold), so ids must be minted above that floor rather than
	// hardcoded — NextChangesetID reads both the DB fence and any existing
	// file in dir, guaranteeing id1 < id2 < id3 regardless of wall-clock time.
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	var initialFence string
	if err := db.QueryRow(`SELECT last_applied_changeset FROM meta`).Scan(&initialFence); err != nil {
		t.Fatalf("read initial fence: %v", err)
	}
	id1, err := infrastructure.NextChangesetID(changesetDir, initialFence) // machine A, wave 1
	if err != nil {
		t.Fatalf("NextChangesetID id1: %v", err)
	}
	a1ID, a2ID, b1ID := ulid.Make().String(), ulid.Make().String(), ulid.Make().String()

	path1, err := infrastructure.WriteChangesetWithID(changesetDir, id1, []infrastructure.ChangesetLine{{
		Op: "create", Entity: "story", ID: a1ID,
		Fields: map[string]any{"slug": "a1", "goal": "machine A wave 1", "status": "planned", "created_at": "2026-08-07T08:00:00Z"},
		At:     "2026-08-07T08:00:00Z",
	}})
	if err != nil {
		t.Fatalf("WriteChangesetWithID id1: %v", err)
	}
	if _, _, err := infrastructure.ApplyChangeset(db, path1); err != nil {
		t.Fatalf("ApplyChangeset id1: %v", err)
	}

	id2, err := infrastructure.NextChangesetID(changesetDir, "") // machine B — merged in, never applied
	if err != nil {
		t.Fatalf("NextChangesetID id2: %v", err)
	}
	if _, err := infrastructure.WriteChangesetWithID(changesetDir, id2, []infrastructure.ChangesetLine{{
		Op: "create", Entity: "story", ID: b1ID,
		Fields: map[string]any{"slug": "b1", "goal": "machine B wave 1", "status": "planned", "created_at": "2026-08-07T08:00:30Z"},
		At:     "2026-08-07T08:00:30Z",
	}}); err != nil {
		t.Fatalf("WriteChangesetWithID id2: %v", err)
	}

	id3, err := infrastructure.NextChangesetID(changesetDir, "") // machine A, wave 2
	if err != nil {
		t.Fatalf("NextChangesetID id3: %v", err)
	}
	path3, err := infrastructure.WriteChangesetWithID(changesetDir, id3, []infrastructure.ChangesetLine{{
		Op: "create", Entity: "story", ID: a2ID,
		Fields: map[string]any{"slug": "a2", "goal": "machine A wave 2", "status": "planned", "created_at": "2026-08-07T08:01:00Z"},
		At:     "2026-08-07T08:01:00Z",
	}})
	if err != nil {
		t.Fatalf("WriteChangesetWithID id3: %v", err)
	}
	if _, _, err := infrastructure.ApplyChangeset(db, path3); err != nil {
		t.Fatalf("ApplyChangeset id3: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Before rebuild: b1 is missing, exactly the loss the audit measured.
	slugs := queryStorySlugs(t)
	if contains(slugs, "b1") {
		t.Fatalf("story b1 present before rebuild = %v, want it missing (setup did not reproduce the loss)", slugs)
	}

	out, err := runDBCommand(t, "db", "rebuild", "--yes", "--json")
	if err != nil {
		t.Fatalf("db rebuild --yes: %v (output=%s)", err, out)
	}
	var result struct {
		Status        string `json:"status"`
		SchemaVersion int    `json:"schema_version"`
		Replayed      int    `json:"replayed"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("decode db rebuild output %q: %v", out, err)
	}
	// Replayed also counts init's own managed-docs changesets, not just the
	// three story creates this test wrote — the story presence check below
	// is the real assertion.
	if result.Status != "rebuilt" || result.Replayed < 3 {
		t.Fatalf("db rebuild result = %+v, want status=rebuilt replayed>=3", result)
	}

	// After rebuild: all three stories are present, replayed in ULID order.
	slugs = queryStorySlugs(t)
	for _, want := range []string{"a1", "b1", "a2"} {
		if !contains(slugs, want) {
			t.Fatalf("story %s missing after rebuild; slugs = %v", want, slugs)
		}
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
