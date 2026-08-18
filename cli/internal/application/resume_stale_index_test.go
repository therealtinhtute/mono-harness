package application

import (
	"testing"

	"github.com/oklog/ulid/v2"
)

// TestResumeStaleIndexNeverIndexed covers the "never indexed" case: an
// active plan exists on disk but plan_index has no row for it yet (a freshly
// scaffolded plan, before any trace/decision/check/handoff has written to
// it) — Resume must say so via stale_index rather than silently reporting
// clean (R9, docs/plans/active/harness-markdown-truth.md).
func TestResumeStaleIndexNeverIndexed(t *testing.T) {
	chdirFixture(t)
	writeActivePlanFixture(t, "fixture-plan")
	db := freshDB(t)

	view, err := Resume(db, "dev")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if view.Readiness != "drifted" {
		t.Fatalf("readiness = %q, want drifted", view.Readiness)
	}
	if len(view.Drift) != 1 || view.Drift[0].Type != "stale_index" {
		t.Fatalf("drift = %v, want exactly one stale_index finding", view.Drift)
	}
}

// TestResumeStaleIndexOutOfDate covers the "indexed but stale" case: a
// plan_index row exists for the active plan's path, but its sha256 no
// longer matches the file's current content — Resume must still flag it,
// proving the comparison is a real hash check, not a presence check.
func TestResumeStaleIndexOutOfDate(t *testing.T) {
	chdirFixture(t)
	path := writeActivePlanFixture(t, "fixture-plan")
	db := freshDB(t)

	id := ulid.Make().String()
	if _, err := db.Exec(`INSERT INTO plan_index (id, path, sha256, status, updated_at) VALUES (?, ?, ?, ?, ?)`,
		id, path, "stale-hash-does-not-match", "active", "2026-08-18T00:00:00Z"); err != nil {
		t.Fatalf("insert plan_index: %v", err)
	}

	view, err := Resume(db, "dev")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if len(view.Drift) != 1 || view.Drift[0].Type != "stale_index" {
		t.Fatalf("drift = %v, want exactly one stale_index finding", view.Drift)
	}
}

// TestResumeStaleIndexMatchesIndexedHash covers the "in sync" case: the
// indexed sha256 matches the file's current content → no stale_index drift,
// and Resume must not have written anything (it opens the db read-only in
// production via interfaces/resume.go's OpenReadOnly).
func TestResumeStaleIndexMatchesIndexedHash(t *testing.T) {
	chdirFixture(t)
	path := writeActivePlanFixture(t, "fixture-plan")
	db := freshDB(t)

	onDisk, _, _, _, err := planIndexStaleness(db, path)
	if err != nil {
		t.Fatalf("planIndexStaleness (compute hash): %v", err)
	}
	id := ulid.Make().String()
	if _, err := db.Exec(`INSERT INTO plan_index (id, path, sha256, status, updated_at) VALUES (?, ?, ?, ?, ?)`,
		id, path, onDisk, "active", "2026-08-18T00:00:00Z"); err != nil {
		t.Fatalf("insert plan_index: %v", err)
	}

	view, err := Resume(db, "dev")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if len(view.Drift) != 0 {
		t.Fatalf("drift = %v, want none (indexed hash matches on-disk content)", view.Drift)
	}
	if view.Readiness != "clean" {
		t.Fatalf("readiness = %q, want clean", view.Readiness)
	}
}

// TestResumeStaleIndexNoActivePlan covers the "no active plan" case: an
// unstarted or ambiguous project has nothing to check plan_index against —
// Resume must not error or fire drift just because ResolveActivePlan
// couldn't resolve one.
func TestResumeStaleIndexNoActivePlan(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	view, err := Resume(db, "dev")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if len(view.Drift) != 0 {
		t.Fatalf("drift = %v, want none (no active plan to check)", view.Drift)
	}
}
