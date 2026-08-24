package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

const planLifecycleFixture = `---
id: 01TESTPLANLIFECYCLEXULIDXX
type: plan
status: active
created: 2026-01-01
updated: 2026-01-01
---

# Plan: Lifecycle fixture

## Phases and Verification
- phases:
  - phase_slug: p0-only
    status: planned

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- none

## Current State and Next Action
- active_phase: p0-only
`

const planLifecycleHeadingFixture = "---\nid: 01TESTPLANLIFECYCLEXULIDXX\ntype: plan\nstatus: active\ncreated: 2026-01-01\nupdated: 2026-01-01\n---\n\n# Plan: Lifecycle fixture\n\n## Phases and Verification\n### phase_slug: `p0-only`\n- status: planned\n\n## Decisions\n<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->\n- none\n\n## Current State and Next Action\n- active_phase: p0-only\n"

const planLifecycleNoPhaseFixture = `---
id: 01TESTPLANLIFECYCLEXULIDXX
type: plan
status: active
created: 2026-01-01
updated: 2026-01-01
---

# Plan: Lifecycle fixture

## Phases and Verification
- phases:
  - status: planned

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- none

## Current State and Next Action
- active_phase: none
`

func writePlanLifecycleFixture(t *testing.T) string {
	return writePlanLifecycleFixtureContent(t, planLifecycleFixture)
}

func writePlanLifecycleFixtureContent(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join("docs", "plans", "active", "lifecycle.md")
	writeFile(t, path, content)
	return path
}

func assertActiveDirEmpty(t *testing.T) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join("docs", "plans", "active"))
	if err != nil {
		t.Fatalf("ReadDir docs/plans/active: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("docs/plans/active/ has %d entries after transition, want 0: %v", len(entries), entries)
	}
}

func TestParseActivePlanPhaseOrderPreservesCanonicalForms(t *testing.T) {
	content := strings.Join([]string{
		"- phase_slug: list-first",
		"### phase_slug: `heading-middle`",
		"- phase_slug: list-last",
	}, "\n")

	got := strings.Join(parseActivePlanPhaseOrder(content), ",")
	if got != "list-first,heading-middle,list-last" {
		t.Fatalf("phase order = %q, want list-first,heading-middle,list-last", got)
	}
}

func TestParseActivePlanPhaseOrderIgnoresHistoricalHeadingForm(t *testing.T) {
	content := strings.Join([]string{
		"### Phase 1 — `legacy`",
		"- phase_slug: canonical",
	}, "\n")

	got := strings.Join(parseActivePlanPhaseOrder(content), ",")
	if got != "canonical" {
		t.Fatalf("phase order = %q, want canonical", got)
	}
}

func TestPlanCompleteMovesCleanHeadingPlan(t *testing.T) {
	chdirFixture(t)
	path := writePlanLifecycleFixtureContent(t, planLifecycleHeadingFixture)
	if got := strings.Join(parseActivePlanPhaseOrder(planLifecycleHeadingFixture), ","); got != "p0-only" {
		t.Fatalf("heading phase order = %q, want p0-only", got)
	}
	db := freshDB(t)
	seedStory(t, db, "p0-only", domain.StoryDone)

	dest, stop, err := PlanComplete(db)
	if err != nil {
		t.Fatalf("PlanComplete: %v", err)
	}
	if stop != nil {
		t.Fatalf("PlanComplete: unexpected stop %+v", stop)
	}
	if dest != filepath.Join("docs", "plans", "completed", "lifecycle.md") {
		t.Fatalf("dest = %q, want docs/plans/completed/lifecycle.md", dest)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("active plan %s still exists after completion: %v", path, err)
	}
}

func TestPlanCompleteRefusedWithHeadingOpenPhase(t *testing.T) {
	chdirFixture(t)
	path := writePlanLifecycleFixtureContent(t, planLifecycleHeadingFixture)
	db := freshDB(t)
	seedStory(t, db, "p0-only", domain.StoryPlanned)

	_, _, err := PlanComplete(db)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "open_phase" {
		t.Fatalf("PlanComplete: err = %v, want *domain.ValidationError{Code: open_phase}", err)
	}
	if !strings.Contains(ve.Message, "p0-only") {
		t.Fatalf("message = %q, want it to name the open phase", ve.Message)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("active plan %s was moved despite refusal: %v", path, statErr)
	}
}

func TestPlanCompleteRefusedWithNoRecognizedPhases(t *testing.T) {
	chdirFixture(t)
	path := writePlanLifecycleFixtureContent(t, planLifecycleNoPhaseFixture)
	db := freshDB(t)

	_, _, err := PlanComplete(db)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "no_phases" {
		t.Fatalf("PlanComplete: err = %v, want *domain.ValidationError{Code: no_phases}", err)
	}
	if !strings.Contains(ve.Message, "no recognized phase declarations") {
		t.Fatalf("message = %q, want it to name the missing phase declarations", ve.Message)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("active plan %s was moved despite refusal: %v", path, statErr)
	}
}

func TestPlanCompleteMovesCleanPlan(t *testing.T) {
	chdirFixture(t)
	writePlanLifecycleFixture(t)
	db := freshDB(t)
	seedStory(t, db, "p0-only", domain.StoryDone)

	dest, stop, err := PlanComplete(db)
	if err != nil {
		t.Fatalf("PlanComplete: %v", err)
	}
	if stop != nil {
		t.Fatalf("PlanComplete: unexpected stop %+v", stop)
	}
	if dest != filepath.Join("docs", "plans", "completed", "lifecycle.md") {
		t.Fatalf("dest = %q, want docs/plans/completed/lifecycle.md", dest)
	}

	assertActiveDirEmpty(t)
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", dest, err)
	}
	content := string(data)
	if !strings.Contains(content, "status: completed") {
		t.Fatalf("moved plan missing status: completed:\n%s", content)
	}
	if strings.Contains(content, "updated: 2026-01-01") {
		t.Fatalf("moved plan did not refresh updated:\n%s", content)
	}
	if !strings.Contains(content, "plan completed.") {
		t.Fatalf("moved plan missing recorded transition in ## Decisions:\n%s", content)
	}
}

func TestPlanCompleteRefusedWithOpenPhase(t *testing.T) {
	chdirFixture(t)
	path := writePlanLifecycleFixture(t)
	db := freshDB(t)
	seedStory(t, db, "p0-only", domain.StoryPlanned)

	_, _, err := PlanComplete(db)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "open_phase" {
		t.Fatalf("PlanComplete: err = %v, want *domain.ValidationError{Code: open_phase}", err)
	}
	if !strings.Contains(ve.Message, "p0-only") {
		t.Fatalf("message = %q, want it to name the open phase", ve.Message)
	}

	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("active plan %s was moved despite refusal: %v", path, statErr)
	}
}

func TestPlanAbandonRequiresReason(t *testing.T) {
	chdirFixture(t)
	path := writePlanLifecycleFixture(t)

	_, _, err := PlanAbandon("")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "missing_required_field" {
		t.Fatalf("PlanAbandon: err = %v, want *domain.ValidationError{Code: missing_required_field}", err)
	}

	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("active plan %s was moved despite missing --reason: %v", path, statErr)
	}
}

func TestPlanAbandonMovesPlanRegardlessOfPhaseStatus(t *testing.T) {
	chdirFixture(t)
	writePlanLifecycleFixture(t)

	dest, stop, err := PlanAbandon("superseded by a different initiative")
	if err != nil {
		t.Fatalf("PlanAbandon: %v", err)
	}
	if stop != nil {
		t.Fatalf("PlanAbandon: unexpected stop %+v", stop)
	}

	assertActiveDirEmpty(t)
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", dest, err)
	}
	content := string(data)
	if !strings.Contains(content, "status: abandoned") {
		t.Fatalf("moved plan missing status: abandoned:\n%s", content)
	}
	if !strings.Contains(content, "superseded by a different initiative") {
		t.Fatalf("moved plan missing recorded reason:\n%s", content)
	}
}
