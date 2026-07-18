package application

import (
	"testing"
)

// TestResumeStaleDocsMatch covers the "match" case from cli-stale-drift-PLAN.md
// T1: written docs_version equals the running CLI's version → no drift.
func TestResumeStaleDocsMatch(t *testing.T) {
	db, changesetDir := freshDB(t)
	setMeta(t, db, changesetDir, map[string]any{"docs_version": "0.2.0"})

	view, err := Resume(db, "0.2.0")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if len(view.Drift) != 0 {
		t.Fatalf("drift = %v, want none (versions match)", view.Drift)
	}
	if view.Readiness != "clean" {
		t.Fatalf("readiness = %q, want clean", view.Readiness)
	}
}

// TestResumeStaleDocsDiffers covers the "differ" case: written docs_version
// diverges from the running CLI's version, neither side is "dev" → fires
// stale_docs with the exact named recovery.
func TestResumeStaleDocsDiffers(t *testing.T) {
	db, changesetDir := freshDB(t)
	setMeta(t, db, changesetDir, map[string]any{"docs_version": "0.2.0"})

	view, err := Resume(db, "0.3.0")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if view.Readiness != "drifted" {
		t.Fatalf("readiness = %q, want drifted", view.Readiness)
	}
	if len(view.Drift) != 1 || view.Drift[0].Type != "stale_docs" {
		t.Fatalf("drift = %v, want exactly one stale_docs finding", view.Drift)
	}
	if view.Drift[0].Recovery != StaleDocsRecovery {
		t.Fatalf("recovery = %q, want %q", view.Drift[0].Recovery, StaleDocsRecovery)
	}
}

// TestResumeStaleDocsDevWritten covers the "dev" case (written side): a
// project scaffolded by a dev build must never fire staleness, even against
// a differing released CLI version — dogfooding must not drown in drift.
func TestResumeStaleDocsDevWritten(t *testing.T) {
	db, changesetDir := freshDB(t)
	setMeta(t, db, changesetDir, map[string]any{"docs_version": "dev"})

	view, err := Resume(db, "0.3.0")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if len(view.Drift) != 0 {
		t.Fatalf("drift = %v, want none (written side is dev)", view.Drift)
	}
}

// TestResumeStaleDocsDevCLI covers the "dev" case (CLI side): running a dev
// build of the CLI against docs stamped by a released version must not fire
// staleness either — the exemption is symmetric.
func TestResumeStaleDocsDevCLI(t *testing.T) {
	db, changesetDir := freshDB(t)
	setMeta(t, db, changesetDir, map[string]any{"docs_version": "0.2.0"})

	view, err := Resume(db, "dev")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if len(view.Drift) != 0 {
		t.Fatalf("drift = %v, want none (CLI side is dev)", view.Drift)
	}
}

// TestResumeStaleDocsMissing covers the "missing" case: a project that has
// never had docs_version stamped (imported legacy project, or scaffolded
// pre-embed) must not fire drift — blocking old projects is not acceptable
// per cli-stale-drift-CONTEXT.md's Locked Decisions.
func TestResumeStaleDocsMissing(t *testing.T) {
	db, _ := freshDB(t)

	view, err := Resume(db, "0.3.0")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if len(view.Drift) != 0 {
		t.Fatalf("drift = %v, want none (docs_version never stamped)", view.Drift)
	}
}

// TestResumeStaleDocsPreMigrationSchema proves Resume never queries the
// docs_version column on a db whose schema_version predates migration
// 0002_meta_docs_version — reading it there would be a raw SQL error ("no
// such column"), not a graceful unversioned state. Simulated by forcing
// schema_version back to 1 on an otherwise fully-migrated db (the column
// still exists, but the guard must trigger on the version number alone,
// proving the short-circuit fires before any docs_version read is attempted
// — the exact condition a real un-migrated project after a CLI upgrade
// would hit).
func TestResumeStaleDocsPreMigrationSchema(t *testing.T) {
	db, changesetDir := freshDB(t)
	setMeta(t, db, changesetDir, map[string]any{"docs_version": "0.2.0"})
	if _, err := db.Exec(`UPDATE meta SET schema_version = 1`); err != nil {
		t.Fatalf("force schema_version=1: %v", err)
	}

	view, err := Resume(db, "0.3.0")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if len(view.Drift) != 0 {
		t.Fatalf("drift = %v, want none (schema_version below docsVersionMinSchema)", view.Drift)
	}
}
