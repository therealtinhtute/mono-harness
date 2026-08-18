package application

import (
	"testing"

	"github.com/oklog/ulid/v2"
)

func TestPlanIndexStalenessNeverIndexed(t *testing.T) {
	chdirFixture(t)
	db, _ := freshDB(t)
	path := writeActivePlanFixture(t, "fixture-plan")

	onDisk, indexed, stale, row, err := planIndexStaleness(db, path)
	if err != nil {
		t.Fatalf("planIndexStaleness: %v", err)
	}
	if onDisk == "" {
		t.Fatal("onDisk hash is empty, want a computed sha256")
	}
	if indexed {
		t.Fatal("indexed = true, want false for a path with no plan_index row")
	}
	if !stale {
		t.Fatal("stale = false, want true for a never-indexed path")
	}
	if row != (planIndexRow{}) {
		t.Fatalf("row = %+v, want zero value when never indexed", row)
	}
}

func TestPlanIndexStalenessMatchesIndexedHash(t *testing.T) {
	chdirFixture(t)
	db, _ := freshDB(t)
	path := writeActivePlanFixture(t, "fixture-plan")

	onDisk, _, _, _, err := planIndexStaleness(db, path)
	if err != nil {
		t.Fatalf("planIndexStaleness (compute hash): %v", err)
	}
	id := ulid.Make().String()
	if _, err := db.Exec(`INSERT INTO plan_index (id, path, sha256, status, updated_at) VALUES (?, ?, ?, ?, ?)`,
		id, path, onDisk, "active", "2026-08-18T00:00:00Z"); err != nil {
		t.Fatalf("insert plan_index: %v", err)
	}

	_, indexed, stale, row, err := planIndexStaleness(db, path)
	if err != nil {
		t.Fatalf("planIndexStaleness: %v", err)
	}
	if !indexed {
		t.Fatal("indexed = false, want true after inserting a row")
	}
	if stale {
		t.Fatal("stale = true, want false when on-disk hash matches the indexed hash")
	}
	if row.ID != id {
		t.Fatalf("row.ID = %q, want %q", row.ID, id)
	}
}

func TestPlanIndexStalenessDiffersFromIndexedHash(t *testing.T) {
	chdirFixture(t)
	db, _ := freshDB(t)
	path := writeActivePlanFixture(t, "fixture-plan")

	id := ulid.Make().String()
	if _, err := db.Exec(`INSERT INTO plan_index (id, path, sha256, status, updated_at) VALUES (?, ?, ?, ?, ?)`,
		id, path, "stale-hash-does-not-match", "active", "2026-08-18T00:00:00Z"); err != nil {
		t.Fatalf("insert plan_index: %v", err)
	}

	_, indexed, stale, _, err := planIndexStaleness(db, path)
	if err != nil {
		t.Fatalf("planIndexStaleness: %v", err)
	}
	if !indexed {
		t.Fatal("indexed = false, want true after inserting a row")
	}
	if !stale {
		t.Fatal("stale = false, want true when on-disk hash differs from the indexed hash")
	}
}
