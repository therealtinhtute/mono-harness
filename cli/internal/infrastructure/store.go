package infrastructure

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

const driverName = "sqlite"

// Open opens (creating if absent) the SQLite database at path.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open(driverName, path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable wal: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign_keys: %w", err)
	}
	return db, nil
}

// OpenReadOnly opens an existing SQLite database without creating files,
// enabling WAL, or permitting writes. Read-only routing commands use it so
// inspection cannot mutate harness state as a side effect.
func OpenReadOnly(path string) (*sql.DB, error) {
	db, err := sql.Open(driverName, "file:"+path+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("open db read-only: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("open db read-only: %w", err)
	}
	return db, nil
}

// Exists reports whether a db file is already present at path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
