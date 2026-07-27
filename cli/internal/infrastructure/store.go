package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"

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

type DatabaseNotFoundError struct {
	Path string
}

func (e *DatabaseNotFoundError) Error() string {
	return fmt.Sprintf("database not found at %s", e.Path)
}

func IsDatabaseNotFound(err error) bool {
	var notFound *DatabaseNotFoundError
	return errors.As(err, &notFound)
}

// ReadOnlyDB retains the repository's shared lock for the complete SQL handle
// lifetime. Close always closes SQLite before releasing the directory lock.
type ReadOnlyDB struct {
	*sql.DB
	lock *RepositoryLock

	closeOnce sync.Once
	closeErr  error
}

func (db *ReadOnlyDB) Raw() *sql.DB {
	if db == nil {
		return nil
	}
	return db.DB
}

func (db *ReadOnlyDB) Close() error {
	if db == nil {
		return nil
	}
	db.closeOnce.Do(func() {
		var dbErr error
		if db.DB != nil {
			dbErr = db.DB.Close()
		}
		db.closeErr = errors.Join(dbErr, db.lock.Close())
	})
	return db.closeErr
}

// OpenReadOnly acquires the shared repository-root lock before inspecting the
// database or its WAL sidecars and retains it until ReadOnlyDB.Close.
func OpenReadOnly(path string) (*ReadOnlyDB, error) {
	lock, err := AcquireRepositoryLock(context.Background(), filepath.Dir(path), RepositoryLockShared)
	if err != nil {
		return nil, fmt.Errorf("open db read-only: %w", err)
	}
	db, err := openReadOnlyUnderExistingLock(path)
	if err != nil {
		_ = lock.Close()
		return nil, err
	}
	db.lock = lock
	return db, nil
}

// OpenReadOnlyUnderExistingLock opens a read-only database without acquiring a
// nested repository lock. Callers must already hold the repository-root lock.
func OpenReadOnlyUnderExistingLock(path string) (*ReadOnlyDB, error) {
	return openReadOnlyUnderExistingLock(path)
}

func openReadOnlyUnderExistingLock(path string) (*ReadOnlyDB, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("open db read-only: resolve path: %w", err)
	}
	canonicalPath, err := filepath.EvalSymlinks(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &DatabaseNotFoundError{Path: absolutePath}
		}
		return nil, fmt.Errorf("open db read-only: resolve path: %w", err)
	}

	// SQLite creates missing WAL sidecars even for mode=ro. immutable=1 avoids
	// that for a clean database, but modernc/SQLite ignores committed frames in
	// an existing WAL under immutable mode. Probe and open the canonical target
	// so symlinks cannot hide its sidecars or select stale immutable reads.
	immutable := false
	walInfo, err := os.Stat(canonicalPath + "-wal")
	switch {
	case os.IsNotExist(err), err == nil && walInfo.Size() == 0:
		immutable = true
	case err != nil:
		return nil, fmt.Errorf("open db read-only: inspect wal: %w", err)
	default:
		if _, err := os.Stat(canonicalPath + "-shm"); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("open db read-only: non-empty WAL exists without SHM; refusing to ignore WAL content or create a sidecar")
			}
			return nil, fmt.Errorf("open db read-only: inspect shm: %w", err)
		}
	}

	query := url.Values{"mode": {"ro"}}
	if immutable {
		query.Set("immutable", "1")
	}
	dsn := (&url.URL{Scheme: "file", Path: canonicalPath, RawQuery: query.Encode()}).String()
	raw, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("open db read-only: %w", err)
	}
	if err := raw.Ping(); err != nil {
		raw.Close()
		return nil, fmt.Errorf("open db read-only: %w", err)
	}
	return &ReadOnlyDB{DB: raw}, nil
}

// Exists reports whether a db file is already present at path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
