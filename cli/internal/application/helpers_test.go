package application

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// chdirFixture creates a fresh temp dir and chdirs the test into it (t.Chdir
// restores the original cwd on cleanup) — every plan/phase path under test
// is cwd-relative.
func chdirFixture(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
}

// makeDirReadOnly chmods dir to read-only (no write/create permission),
// restoring it before the test's TempDir cleanup runs so that cleanup can
// still remove files inside it. Used to force a plan markdown write to
// fail without touching the file being written (writeFileAtomically writes
// a sibling temp file into dir, so removing dir's write permission is what
// actually blocks the write).
func makeDirReadOnly(t *testing.T, dir string) {
	t.Helper()
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatalf("Chmod(%s, 0o555): %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(dir, 0o755); err != nil {
			t.Fatalf("Chmod(%s, 0o755) restore: %v", dir, err)
		}
	})
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func writeActivePlan(t *testing.T, name string, slugs ...string) string {
	t.Helper()
	var content strings.Builder
	content.WriteString("# Active plan\n\n## Phases and Verification\n")
	for i, slug := range slugs {
		fmt.Fprintf(&content, "### Phase %d: %s\n- phase_slug: %s\n- goal: goal\n\n", i+1, slug, slug)
	}
	path := filepath.Join("docs", "plans", "active", name+".md")
	writeFile(t, path, content.String())
	return path
}

// seedStory writes a story row with an explicit slug + status, since seedRun
// always hardcodes the "cli-domain" slug.
func seedStory(t *testing.T, db *sql.DB, slug, status string) {
	t.Helper()
	if _, err := db.Exec(
		`INSERT OR IGNORE INTO stories (id, slug, goal, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		ulid.Make().String(), slug, "goal", status, "2026-07-22T00:00:00Z",
	); err != nil {
		t.Fatalf("seed story %s: %v", slug, err)
	}
}

// scaffoldedPlanFixture is a minimal plan with every append-only section
// in its bootstrap ("- none") state, matching scaffold.go's real template
// closely enough to exercise the plan-write path (P3 wave 2).
const scaffoldedPlanFixture = `---
id: 01TESTPLANFIXTUREXULIDXXX
type: plan
intake_id: 01TESTINTAKEFIXTUREXULIDX
lane: normal
status: active
created: 2026-07-22
updated: 2026-07-22
---

# Plan: Fixture

## Progress
<!-- Append-only durable entries record timestamp, phase, wave, task, task_status,
run_id, trace_id, exact verification/result, and changed surfaces or blocker. -->
- none

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- none

## Validation
<!-- Append-only durable entries record timestamp, phase, exact command/result/output,
run_id, check_id, verdict, and proof_gaps. -->
- none

## Current State and Next Action
- active_phase: none
`

// writeActivePlanFixture writes scaffoldedPlanFixture at
// docs/plans/active/{slug}.md, relative to the test's current directory —
// callers must have already chdir'd into a scratch root (chdirFixture).
// Returns the plan's path for direct reads/assertions.
func writeActivePlanFixture(t *testing.T, slug string) string {
	t.Helper()
	path := filepath.Join("docs", "plans", "active", slug+".md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(scaffoldedPlanFixture), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
	return path
}

// freshDB opens a fresh, migrated db scoped to a per-test temp dir.
func freshDB(t *testing.T) (db *sql.DB) {
	t.Helper()
	root := t.TempDir()
	db, err := infrastructure.Open(filepath.Join(root, "harness.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, _, err := infrastructure.Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func countRows(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

// assertRowExists proves a row for id exists in table.
func assertRowExists(t *testing.T, db *sql.DB, table, id string) {
	t.Helper()
	var found string
	if err := db.QueryRow("SELECT id FROM "+table+" WHERE id = ?", id).Scan(&found); err != nil {
		t.Fatalf("row for %s not found in %s: %v", id, table, err)
	}
}

// seedRun writes a story + run row directly (bypassing domain validation
// and the Create* application functions under test) so FK-dependent
// fixtures (checks.run_id, traces.run_id) have something real to point
// at. FKs are enforced (PRAGMA foreign_keys=ON, see store.go), so a bare
// literal ID would fail at insert time.
func seedRun(t *testing.T, db *sql.DB) (runID string) {
	t.Helper()
	at := "2026-07-17T12:00:00Z"
	storyID := ulid.Make().String()
	if _, err := db.Exec(
		`INSERT OR IGNORE INTO stories (id, slug, goal, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		storyID, "cli-domain", "ported domain commands work", "planned", at,
	); err != nil {
		t.Fatalf("seed story: %v", err)
	}

	runID = ulid.Make().String()
	if _, err := db.Exec(
		`INSERT INTO runs (id, story_slug, artifact_path, created_at) VALUES (?, ?, ?, ?)`,
		runID, "cli-domain", ".kit/runs/work/x.md", at,
	); err != nil {
		t.Fatalf("seed run: %v", err)
	}
	return runID
}

// seedCheck extends seedRun with a checks row.
func seedCheck(t *testing.T, db *sql.DB) (checkID string) {
	t.Helper()
	runID := seedRun(t, db)
	at := "2026-07-17T12:05:00Z"
	checkID = ulid.Make().String()
	if _, err := db.Exec(
		`INSERT INTO checks (id, run_id, verdict, created_at) VALUES (?, ?, ?, ?)`,
		checkID, runID, "APPROVED", at,
	); err != nil {
		t.Fatalf("seed check: %v", err)
	}
	return checkID
}
