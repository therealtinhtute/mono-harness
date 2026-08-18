package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestRecordDecisionsBatchAndQueryRoundTrip(t *testing.T) {
	db := freshDB(t)
	runID := seedRun(t, db)
	seedStory(t, db, "p1", "planned")

	ids, err := RecordDecisions(db, runID, []domain.Decision{
		{Decision: "batch one decision", Rationale: "reason one", Phase: "p1", Task: "wave 1 task"},
		{Decision: "batch two decision", Rationale: "reason two"},
	})
	if err != nil {
		t.Fatalf("RecordDecisions: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("RecordDecisions ids = %v, want 2", ids)
	}
	if ids[0] == ids[1] {
		t.Fatalf("RecordDecisions minted duplicate ids: %v", ids)
	}
	if got := countRows(t, db, "decisions"); got != 2 {
		t.Fatalf("decisions rows = %d, want 2", got)
	}

	all, err := QueryDecisions(db, "", 0)
	if err != nil {
		t.Fatalf("QueryDecisions: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("QueryDecisions = %+v, want 2 rows", all)
	}
	if all[0].Decision != "batch one decision" || all[0].Rationale != "reason one" {
		t.Fatalf("QueryDecisions[0] = %+v, want the first batch decision", all[0])
	}
	if all[0].Phase == nil || *all[0].Phase != "p1" {
		t.Fatalf("QueryDecisions[0].Phase = %v, want p1", all[0].Phase)
	}
	if all[0].Task == nil || *all[0].Task != "wave 1 task" {
		t.Fatalf("QueryDecisions[0].Task = %v, want %q", all[0].Task, "wave 1 task")
	}
	if all[0].RunID == nil || *all[0].RunID != runID {
		t.Fatalf("QueryDecisions[0].RunID = %v, want %s", all[0].RunID, runID)
	}
	if all[1].Phase != nil || all[1].Task != nil {
		t.Fatalf("QueryDecisions[1] = %+v, want nil phase/task (none supplied)", all[1])
	}

	filtered, err := QueryDecisions(db, "p1", 0)
	if err != nil {
		t.Fatalf("QueryDecisions phase filter: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Decision != "batch one decision" {
		t.Fatalf("QueryDecisions phase filter = %+v, want only the p1 decision", filtered)
	}

	tailed, err := QueryDecisions(db, "", 1)
	if err != nil {
		t.Fatalf("QueryDecisions tail=1: %v", err)
	}
	if len(tailed) != 1 || tailed[0].Decision != "batch two decision" {
		t.Fatalf("QueryDecisions tail=1 = %+v, want only the most recent decision", tailed)
	}
}

func TestRecordDecisionsEmptyBatch(t *testing.T) {
	db := freshDB(t)

	_, err := RecordDecisions(db, "", nil)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "empty_decisions" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: empty_decisions}", err)
	}
	if got := countRows(t, db, "decisions"); got != 0 {
		t.Fatalf("decisions rows = %d, want 0", got)
	}
}

func TestRecordDecisionsUnknownRunID(t *testing.T) {
	db := freshDB(t)

	_, err := RecordDecisions(db, "01HZZZZZZZZZZZZZZZZZZZZZZZ", []domain.Decision{
		{Decision: "d", Rationale: "r"},
	})
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_run_id" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_run_id}", err)
	}
	if got := countRows(t, db, "decisions"); got != 0 {
		t.Fatalf("decisions rows = %d, want 0 (nothing written on validation failure)", got)
	}
}

func TestRecordDecisionsUnknownPhase(t *testing.T) {
	db := freshDB(t)

	_, err := RecordDecisions(db, "", []domain.Decision{
		{Decision: "d", Rationale: "r", Phase: "no-such-phase"},
	})
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_phase" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_phase}", err)
	}
	if got := countRows(t, db, "decisions"); got != 0 {
		t.Fatalf("decisions rows = %d, want 0", got)
	}
}

// TestRecordDecisionsBatchIsAtomicOnValidationFailure proves a batch with
// one invalid element writes nothing — not a partial batch — matching the
// all-or-nothing shape of every other multi-row writer in this package
// (e.g. check record's status+meta update).
func TestRecordDecisionsBatchIsAtomicOnValidationFailure(t *testing.T) {
	db := freshDB(t)

	_, err := RecordDecisions(db, "", []domain.Decision{
		{Decision: "valid first element", Rationale: "reason"},
		{Decision: "second element", Rationale: ""},
	})
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "missing_required_field" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: missing_required_field}", err)
	}
	if got := countRows(t, db, "decisions"); got != 0 {
		t.Fatalf("decisions rows = %d, want 0 — the valid first element must not be written alone", got)
	}
}

// TestRecordDecisionsWritesPlanDecisionsEntryPerElement proves the whole
// batch lands as separate, individually-formatted lines in the plan's
// `## Decisions` section, in one operation (P3 wave 2).
func TestRecordDecisionsWritesPlanDecisionsEntryPerElement(t *testing.T) {
	chdirFixture(t)
	planPath := writeActivePlanFixture(t, "demo")
	db := freshDB(t)
	seedStory(t, db, "p1", "planned")

	ids, err := RecordDecisions(db, "", []domain.Decision{
		{Decision: "picked A over B", Rationale: "matched constraint", Phase: "p1", Task: "wave 1 task"},
		{Decision: "deferred C", Rationale: "out of scope"},
	})
	if err != nil {
		t.Fatalf("RecordDecisions: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("RecordDecisions ids = %v, want 2", ids)
	}

	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "picked A over B") || !strings.Contains(content, "matched constraint") {
		t.Fatalf("plan Decisions missing first entry:\n%s", content)
	}
	if !strings.Contains(content, "deferred C") || !strings.Contains(content, "out of scope") {
		t.Fatalf("plan Decisions missing second entry:\n%s", content)
	}
	if !strings.Contains(content, "## Progress\n<!-- Append-only durable entries record timestamp, phase, wave, task, task_status,\nrun_id, trace_id, exact verification/result, and changed surfaces or blocker. -->\n- none") {
		t.Fatalf("plan Progress section corrupted:\n%s", content)
	}
}

// TestRecordDecisionsAtomicityIncludesPlanWrite extends the atomicity
// proof to the plan-write path: a malformed plan (Decisions section
// missing) must block the DB write too, not just leave the plan corrupted.
func TestRecordDecisionsAtomicityIncludesPlanWrite(t *testing.T) {
	chdirFixture(t)
	planPath := writeActivePlanFixture(t, "demo")
	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	corrupted := strings.Replace(string(data), "## Decisions", "## Renamed", 1)
	if err := os.WriteFile(planPath, []byte(corrupted), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	db := freshDB(t)
	_, err = RecordDecisions(db, "", []domain.Decision{{Decision: "d", Rationale: "r"}})
	if err == nil {
		t.Fatal("RecordDecisions = nil error, want a plan-section-not-found failure")
	}
	if got := countRows(t, db, "decisions"); got != 0 {
		t.Fatalf("decisions rows = %d, want 0", got)
	}
	after, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("ReadFile after failed RecordDecisions: %v", err)
	}
	if string(after) != corrupted {
		t.Fatal("plan file changed despite the failed write")
	}
}

// TestRecordDecisionsReadOnlyPlanBlocksDBWrite is R8's forced-failure proof:
// a markdown write that fails at the filesystem level (not just a malformed
// section) must still leave zero DB rows behind it
// (docs/plans/active/harness-markdown-truth.md).
func TestRecordDecisionsReadOnlyPlanBlocksDBWrite(t *testing.T) {
	chdirFixture(t)
	planPath := writeActivePlanFixture(t, "demo")
	makeDirReadOnly(t, filepath.Dir(planPath))

	db := freshDB(t)

	_, err := RecordDecisions(db, "", []domain.Decision{{Decision: "d", Rationale: "r"}})
	if err == nil {
		t.Fatal("RecordDecisions = nil error, want a read-only plan write failure")
	}
	if got := countRows(t, db, "decisions"); got != 0 {
		t.Fatalf("decisions rows = %d, want 0 — DB write must not proceed when the plan write fails", got)
	}
}
