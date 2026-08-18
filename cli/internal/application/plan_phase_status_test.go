package application

import (
	"os"
	"strings"
	"testing"
)

const phasesListFormFixture = `---
id: 01TESTPLANFIXTUREXULIDXXX
type: plan
status: active
---

# Plan: Fixture

## Phases and Verification
- phases:
  - phase_slug: p1-first
    story_id: 01STORYONEXXXXXXXXXXXXXXX
    status: planned
    goal: first phase
    waves:
      - wave: 1
        tasks:
          - task: do the thing
  - phase_slug: p2-second
    story_id: 01STORYTWOXXXXXXXXXXXXXXX
    status: planned
    goal: second phase

## Progress
- none
`

const phasesHeadingFormFixture = `---
id: 01TESTPLANFIXTUREXULIDXXX
type: plan
status: active
---

# Plan: Fixture

## Phases and Verification

### phase_slug: ` + "`p1-first`" + `
story_id: 01STORYONEXXXXXXXXXXXXXXX
status: planned
goal: first phase

### phase_slug: ` + "`p2-second`" + `
story_id: 01STORYTWOXXXXXXXXXXXXXXX
status: planned
goal: second phase

## Progress
- none
`

func TestSetPlanPhaseStatusListForm(t *testing.T) {
	out, found := SetPlanPhaseStatus(phasesListFormFixture, "p1-first", "in-progress")
	if !found {
		t.Fatal("found = false, want true")
	}
	if !strings.Contains(out, "    status: in-progress\n") {
		t.Fatalf("output missing updated status line:\n%s", out)
	}
	if !strings.Contains(out, "p2-second") || !strings.Contains(out, "status: planned") {
		t.Fatal("sibling phase p2-second's status was disturbed")
	}
	if strings.Count(out, "status: planned") != 1 {
		t.Fatalf("want exactly one remaining 'status: planned' (p2-second), got:\n%s", out)
	}
}

func TestSetPlanPhaseStatusHeadingForm(t *testing.T) {
	out, found := SetPlanPhaseStatus(phasesHeadingFormFixture, "p2-second", "done")
	if !found {
		t.Fatal("found = false, want true")
	}
	if !strings.Contains(out, "status: done") {
		t.Fatalf("output missing updated status line:\n%s", out)
	}
	if !strings.Contains(out, "p1-first") {
		t.Fatal("p1-first block was lost")
	}
	// p1-first's own status must be untouched.
	idx := strings.Index(out, "p1-first")
	block := out[idx:strings.Index(out, "p2-second")]
	if !strings.Contains(block, "status: planned") {
		t.Fatalf("p1-first status was disturbed:\n%s", block)
	}
}

func TestSetPlanPhaseStatusUnknownSlug(t *testing.T) {
	out, found := SetPlanPhaseStatus(phasesListFormFixture, "no-such-phase", "done")
	if found {
		t.Fatal("found = true, want false for unknown slug")
	}
	if out != phasesListFormFixture {
		t.Fatal("content must be returned unchanged when slug is not found")
	}
}

func TestSetPlanPhaseStatusPreservesCRLF(t *testing.T) {
	crlf := strings.ReplaceAll(phasesListFormFixture, "\n", "\r\n")
	out, found := SetPlanPhaseStatus(crlf, "p1-first", "checked")
	if !found {
		t.Fatal("found = false, want true")
	}
	if !strings.Contains(out, "\r\n") {
		t.Fatal("CRLF line endings were not preserved")
	}
	if strings.Contains(out, "status: checked\n") && !strings.Contains(out, "status: checked\r\n") {
		t.Fatalf("updated line lost CRLF:\n%q", out)
	}
}

// TestPreparePlanPhaseStatusWritesThroughToDisk proves preparePlanPhaseStatus
// (plan_write.go) — the write closure story create/run create/check
// record/handoff actually call — locates the active plan, rewrites its
// matching phase block's status line on disk, and refreshes plan_index,
// end to end (not just the pure SetPlanPhaseStatus string transform above).
func TestPreparePlanPhaseStatusWritesThroughToDisk(t *testing.T) {
	chdirFixture(t)
	path := "docs/plans/active/fixture-plan.md"
	writeFile(t, path, phasesListFormFixture)
	db := freshDB(t)

	write, err := preparePlanPhaseStatus(db, "p1-first", "in-progress")
	if err != nil {
		t.Fatalf("preparePlanPhaseStatus: %v", err)
	}
	if err := write(); err != nil {
		t.Fatalf("write(): %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "    status: in-progress\n") {
		t.Fatalf("on-disk plan missing updated status:\n%s", data)
	}

	if n := countRows(t, db, "plan_index"); n != 1 {
		t.Fatalf("plan_index rows = %d, want 1 (refreshed by the write closure)", n)
	}
}

// TestPreparePlanPhaseStatusNoMatchingPhaseIsNoop covers a story with no
// matching phase block (e.g. an ad hoc `story create` outside `## Phases
// and Verification`) — the write closure must be a harmless no-op rather
// than an error, and must not touch the file at all.
func TestPreparePlanPhaseStatusNoMatchingPhaseIsNoop(t *testing.T) {
	chdirFixture(t)
	path := "docs/plans/active/fixture-plan.md"
	writeFile(t, path, phasesListFormFixture)
	db := freshDB(t)

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	write, err := preparePlanPhaseStatus(db, "no-such-phase", "in-progress")
	if err != nil {
		t.Fatalf("preparePlanPhaseStatus: %v", err)
	}
	if err := write(); err != nil {
		t.Fatalf("write(): %v", err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(before) != string(after) {
		t.Fatal("no-op write must not modify the plan file")
	}
	if n := countRows(t, db, "plan_index"); n != 0 {
		t.Fatalf("plan_index rows = %d, want 0 (no write happened)", n)
	}
}
