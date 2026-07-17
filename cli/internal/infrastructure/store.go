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

// Exists reports whether a db file is already present at path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
