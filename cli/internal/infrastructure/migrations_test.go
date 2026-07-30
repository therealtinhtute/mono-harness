package infrastructure

import (
	"path/filepath"
	"sort"
	"testing"
)

// schemaMDTables is the current table list from cli/docs/SCHEMA.md.
var schemaMDTables = []string{
	"checks",
	"handoffs",
	"intakes",
	"interventions",
	"managed_docs",
	"meta",
	"runs",
	"stories",
	"traces",
}

func TestMigrate(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "harness.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	applied, schemaVersion, err := Migrate(db)
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	wantApplied := []string{"0001_init", "0002_meta_docs_version", "0003_drop_dead_surface", "0004_managed_docs", "0005_intake_plan_path", "0006_check_judge"}
	if len(applied) != len(wantApplied) {
		t.Fatalf("applied = %v, want %v", applied, wantApplied)
	}
	for i, name := range wantApplied {
		if applied[i] != name {
			t.Fatalf("applied = %v, want %v", applied, wantApplied)
		}
	}
	if schemaVersion != 6 {
		t.Fatalf("schemaVersion = %d, want 6", schemaVersion)
	}

	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan table name: %v", err)
		}
		got = append(got, name)
	}
	sort.Strings(got)

	if len(got) != len(schemaMDTables) {
		t.Fatalf("table count = %d, want %d (got %v)", len(got), len(schemaMDTables), got)
	}
	for i, name := range got {
		if name != schemaMDTables[i] {
			t.Fatalf("table[%d] = %q, want %q (full list: %v)", i, name, schemaMDTables[i], got)
		}
	}

	// Re-running Migrate on an already-current db must be a no-op.
	applied2, schemaVersion2, err := Migrate(db)
	if err != nil {
		t.Fatalf("Migrate (second run): %v", err)
	}
	if len(applied2) != 0 {
		t.Fatalf("second Migrate applied = %v, want none", applied2)
	}
	if schemaVersion2 != 6 {
		t.Fatalf("second Migrate schemaVersion = %d, want 6", schemaVersion2)
	}
}
