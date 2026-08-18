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

func writePlanLifecycleFixture(t *testing.T) string {
	t.Helper()
	path := filepath.Join("docs", "plans", "active", "lifecycle.md")
	writeFile(t, path, planLifecycleFixture)
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
