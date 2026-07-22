package infrastructure

import (
	"database/sql"
	"fmt"
)

type migration struct {
	Version int
	Name    string
	SQL     string
}

// migrations holds the versioned schema history. v1 is the frozen
// SCHEMA.md table set (11 tables); table order within it satisfies FK
// forward-references (SQLite itself doesn't require this, but it matches
// how the tables read in SCHEMA.md). Later versions are additive.
var migrations = []migration{
	{
		Version: 1,
		Name:    "0001_init",
		SQL: `
CREATE TABLE stories (
	id TEXT PRIMARY KEY,
	slug TEXT UNIQUE NOT NULL,
	goal TEXT NOT NULL,
	status TEXT NOT NULL,
	depends_on TEXT REFERENCES stories(slug),
	created_at TEXT NOT NULL
);

CREATE TABLE runs (
	id TEXT PRIMARY KEY,
	story_slug TEXT NOT NULL REFERENCES stories(slug),
	plan_id TEXT,
	trace_ids TEXT NOT NULL DEFAULT '[]',
	artifact_path TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE TABLE checks (
	id TEXT PRIMARY KEY,
	run_id TEXT NOT NULL REFERENCES runs(id),
	verdict TEXT NOT NULL,
	proof_links TEXT NOT NULL DEFAULT '[]',
	artifact_path TEXT,
	created_at TEXT NOT NULL
);

CREATE TABLE meta (
	schema_version INTEGER NOT NULL,
	current_phase TEXT REFERENCES stories(slug),
	entry_phase TEXT REFERENCES stories(slug),
	latest_run_id TEXT REFERENCES runs(id),
	latest_check_id TEXT REFERENCES checks(id),
	last_applied_changeset TEXT
);

CREATE TABLE handoffs (
	id TEXT PRIMARY KEY,
	run_id TEXT REFERENCES runs(id),
	check_id TEXT REFERENCES checks(id),
	anchors TEXT NOT NULL DEFAULT '{}',
	created_at TEXT NOT NULL
);

CREATE TABLE intakes (
	id TEXT PRIMARY KEY,
	type TEXT NOT NULL,
	summary TEXT NOT NULL,
	lane TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE TABLE decisions (
	id TEXT PRIMARY KEY,
	summary TEXT NOT NULL,
	rationale TEXT NOT NULL,
	rejected TEXT,
	created_at TEXT NOT NULL
);

CREATE TABLE backlog (
	id TEXT PRIMARY KEY,
	summary TEXT NOT NULL,
	priority TEXT,
	created_at TEXT NOT NULL
);

CREATE TABLE tools (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	purpose TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE TABLE interventions (
	id TEXT PRIMARY KEY,
	verdict_id TEXT NOT NULL REFERENCES checks(id),
	reason TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE TABLE traces (
	id TEXT PRIMARY KEY,
	run_id TEXT REFERENCES runs(id),
	wave INTEGER NOT NULL,
	summary TEXT NOT NULL,
	created_at TEXT NOT NULL
);
`,
	},
	{
		Version: 2,
		Name:    "0002_meta_docs_version",
		SQL:     `ALTER TABLE meta ADD COLUMN docs_version TEXT;`,
	},
	{
		Version: 3,
		Name:    "0003_drop_dead_surface",
		SQL: `
DROP TABLE decisions;
DROP TABLE backlog;
DROP TABLE tools;
`,
	},
}

// CurrentSchemaVersion returns the highest version among known migrations.
func CurrentSchemaVersion() int {
	v := 0
	for _, m := range migrations {
		if m.Version > v {
			v = m.Version
		}
	}
	return v
}

// Migrate applies every pending migration (version > the db's current
// meta.schema_version) inside one transaction and returns the migration
// names applied plus the resulting schema version.
func Migrate(db *sql.DB) (applied []string, schemaVersion int, err error) {
	current, err := readSchemaVersion(db)
	if err != nil {
		return nil, 0, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, 0, fmt.Errorf("begin migrate tx: %w", err)
	}
	defer tx.Rollback()

	for _, m := range migrations {
		if m.Version <= current {
			continue
		}
		if _, err := tx.Exec(m.SQL); err != nil {
			return nil, 0, fmt.Errorf("migration %s: %w", m.Name, err)
		}
		applied = append(applied, m.Name)
		current = m.Version
	}

	if len(applied) > 0 {
		if err := upsertSchemaVersion(tx, current); err != nil {
			return nil, 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("commit migrate tx: %w", err)
	}

	return applied, current, nil
}

func readSchemaVersion(db *sql.DB) (int, error) {
	var tableCount int
	if err := db.QueryRow(
		`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='meta'`,
	).Scan(&tableCount); err != nil {
		return 0, fmt.Errorf("check meta table: %w", err)
	}
	if tableCount == 0 {
		return 0, nil
	}

	var version sql.NullInt64
	err := db.QueryRow(`SELECT schema_version FROM meta LIMIT 1`).Scan(&version)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read schema_version: %w", err)
	}
	return int(version.Int64), nil
}

func upsertSchemaVersion(tx *sql.Tx, version int) error {
	res, err := tx.Exec(`UPDATE meta SET schema_version = ?`, version)
	if err != nil {
		return fmt.Errorf("update schema_version: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		if _, err := tx.Exec(`INSERT INTO meta (schema_version) VALUES (?)`, version); err != nil {
			return fmt.Errorf("insert meta row: %w", err)
		}
	}
	return nil
}
