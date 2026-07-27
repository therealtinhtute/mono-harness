package infrastructure

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type databaseFileState struct {
	data    []byte
	size    int64
	modTime time.Time
}

func TestOpenReadOnlyDoesNotCreateMissingDatabase(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing.db")
	if _, err := OpenReadOnly(path); err == nil {
		t.Fatal("OpenReadOnly() missing database error = nil")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("OpenReadOnly() created %s", path)
	}
}

func TestOpenReadOnlyRejectsWrites(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "harness.db")
	writable, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, _, err := Migrate(writable); err != nil {
		writable.Close()
		t.Fatalf("Migrate() error = %v", err)
	}
	if err := writable.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	readOnly, err := OpenReadOnly(path)
	if err != nil {
		t.Fatalf("OpenReadOnly() error = %v", err)
	}
	defer readOnly.Close()
	if _, err := readOnly.Exec(`UPDATE meta SET docs_version = 'changed'`); err == nil {
		t.Fatal("read-only database accepted an update")
	}
}

func TestOpenReadOnlyDoesNotCreateMissingWALSidecars(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "harness.db")
	writable, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, _, err := Migrate(writable); err != nil {
		writable.Close()
		t.Fatalf("Migrate() error = %v", err)
	}
	if _, err := writable.Exec(`UPDATE meta SET docs_version = 'persisted'`); err != nil {
		writable.Close()
		t.Fatalf("seed meta: %v", err)
	}
	if err := writable.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	before := captureDatabaseFileState(t, path)
	if len(before.data) < 20 {
		t.Fatalf("database header is %d bytes, want at least 20", len(before.data))
	}
	if before.data[18] != 2 || before.data[19] != 2 {
		t.Fatalf("database header journal mode = %v, want WAL", before.data[18:20])
	}
	assertPathMissing(t, path+"-wal")
	assertPathMissing(t, path+"-shm")

	readOnly, err := OpenReadOnly(path)
	if err != nil {
		t.Fatalf("OpenReadOnly() error = %v", err)
	}
	var docsVersion string
	if err := readOnly.QueryRow(`SELECT docs_version FROM meta`).Scan(&docsVersion); err != nil {
		readOnly.Close()
		t.Fatalf("read persisted state: %v", err)
	}
	if docsVersion != "persisted" {
		readOnly.Close()
		t.Fatalf("docs_version = %q, want persisted", docsVersion)
	}
	if err := readOnly.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	assertDatabaseFileState(t, path, before)
	assertPathMissing(t, path+"-wal")
	assertPathMissing(t, path+"-shm")
}

func TestOpenReadOnlyEscapesCanonicalFileURI(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("Windows filenames cannot contain '?' characters")
	}

	base := t.TempDir()
	sourcePath := filepath.Join(base, "source.db")
	writable, err := Open(sourcePath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, _, err := Migrate(writable); err != nil {
		writable.Close()
		t.Fatalf("Migrate() error = %v", err)
	}
	if _, err := writable.Exec(`UPDATE meta SET docs_version = 'escaped-uri'`); err != nil {
		writable.Close()
		t.Fatalf("seed meta: %v", err)
	}
	if err := writable.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	root := filepath.Join(base, "path with spaces ? hash # percent %")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(root, "harness ?#%.db")
	if err := os.Rename(sourcePath, path); err != nil {
		t.Fatalf("Rename fixture: %v", err)
	}
	before := captureDatabaseFileState(t, path)
	assertPathMissing(t, path+"-wal")
	assertPathMissing(t, path+"-shm")

	readOnly, err := OpenReadOnly(path)
	if err != nil {
		t.Fatalf("OpenReadOnly() error = %v", err)
	}
	var docsVersion string
	if err := readOnly.QueryRow(`SELECT docs_version FROM meta`).Scan(&docsVersion); err != nil {
		readOnly.Close()
		t.Fatalf("read escaped URI state: %v", err)
	}
	if err := readOnly.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if docsVersion != "escaped-uri" {
		t.Fatalf("docs_version = %q, want escaped-uri", docsVersion)
	}
	assertDatabaseFileState(t, path, before)
	assertPathMissing(t, path+"-wal")
	assertPathMissing(t, path+"-shm")
}

func TestOpenReadOnlyPreservesZeroByteWALWithoutCreatingSHM(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "harness.db")
	writable, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, _, err := Migrate(writable); err != nil {
		writable.Close()
		t.Fatalf("Migrate() error = %v", err)
	}
	if _, err := writable.Exec(`UPDATE meta SET docs_version = 'zero-wal'`); err != nil {
		writable.Close()
		t.Fatalf("seed meta: %v", err)
	}
	if err := writable.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := os.WriteFile(path+"-wal", nil, 0o644); err != nil {
		t.Fatalf("write zero-byte WAL: %v", err)
	}
	beforeDB := captureDatabaseFileState(t, path)
	beforeWAL := captureDatabaseFileState(t, path+"-wal")
	assertPathMissing(t, path+"-shm")

	readOnly, err := OpenReadOnly(path)
	if err != nil {
		t.Fatalf("OpenReadOnly() error = %v", err)
	}
	var docsVersion string
	if err := readOnly.QueryRow(`SELECT docs_version FROM meta`).Scan(&docsVersion); err != nil {
		readOnly.Close()
		t.Fatalf("read persisted state: %v", err)
	}
	if err := readOnly.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if docsVersion != "zero-wal" {
		t.Fatalf("docs_version = %q, want zero-wal", docsVersion)
	}
	assertDatabaseFileState(t, path, beforeDB)
	assertDatabaseFileState(t, path+"-wal", beforeWAL)
	assertPathMissing(t, path+"-shm")
}

func TestOpenReadOnlyReadsLatestCommittedWALState(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "harness.db")
	writer, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	writer.SetMaxOpenConns(1)
	defer writer.Close()
	if _, _, err := Migrate(writer); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	if _, err := writer.Exec(`PRAGMA wal_autocheckpoint=0`); err != nil {
		t.Fatalf("disable WAL autocheckpoint: %v", err)
	}
	if _, err := writer.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		t.Fatalf("checkpoint schema: %v", err)
	}
	before := captureDatabaseFileState(t, path)
	if _, err := writer.Exec(`INSERT INTO stories (id, slug, goal, status, created_at) VALUES
		('01ARZ3NDEKTSV4RRFFQ69G5FAV', 'wal-only', 'latest WAL state', 'planned', '2026-07-27T00:00:00Z')`); err != nil {
		t.Fatalf("seed WAL-only story: %v", err)
	}
	assertDatabaseFileState(t, path, before)
	assertPathExists(t, path+"-wal")
	assertPathExists(t, path+"-shm")

	immutable, err := sql.Open(driverName, "file:"+path+"?mode=ro&immutable=1")
	if err != nil {
		t.Fatalf("open immutable control: %v", err)
	}
	var immutableCount int
	if err := immutable.QueryRow(`SELECT COUNT(*) FROM stories WHERE slug = 'wal-only'`).Scan(&immutableCount); err != nil {
		immutable.Close()
		t.Fatalf("query immutable control: %v", err)
	}
	if err := immutable.Close(); err != nil {
		t.Fatalf("close immutable control: %v", err)
	}
	if immutableCount != 0 {
		t.Fatalf("immutable control count = %d, want 0 to prove the row exists only in WAL", immutableCount)
	}

	readOnly, err := OpenReadOnly(path)
	if err != nil {
		t.Fatalf("OpenReadOnly() error = %v", err)
	}
	var goal string
	if err := readOnly.QueryRow(`SELECT goal FROM stories WHERE slug = 'wal-only'`).Scan(&goal); err != nil {
		readOnly.Close()
		t.Fatalf("read latest WAL state: %v", err)
	}
	if err := readOnly.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if goal != "latest WAL state" {
		t.Fatalf("goal = %q, want latest WAL state", goal)
	}
	assertDatabaseFileState(t, path, before)
}

func TestOpenReadOnlyResolvesSymlinkBeforeInspectingWALSidecars(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	targetDir := filepath.Join(root, "target")
	linkDir := filepath.Join(root, "link")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("MkdirAll target: %v", err)
	}
	if err := os.MkdirAll(linkDir, 0o755); err != nil {
		t.Fatalf("MkdirAll link: %v", err)
	}
	targetPath := filepath.Join(targetDir, "harness.db")
	linkPath := filepath.Join(linkDir, "harness.db")
	writer, err := Open(targetPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	writer.SetMaxOpenConns(1)
	defer writer.Close()
	if _, _, err := Migrate(writer); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	if _, err := writer.Exec(`PRAGMA wal_autocheckpoint=0`); err != nil {
		t.Fatalf("disable WAL autocheckpoint: %v", err)
	}
	if _, err := writer.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		t.Fatalf("checkpoint schema: %v", err)
	}
	before := captureDatabaseFileState(t, targetPath)
	if _, err := writer.Exec(`INSERT INTO stories (id, slug, goal, status, created_at) VALUES
		('01ARZ3NDEKTSV4RRFFQ69G5FAW', 'symlink-wal-only', 'resolved target WAL state', 'planned', '2026-07-27T00:00:00Z')`); err != nil {
		t.Fatalf("seed WAL-only story: %v", err)
	}
	if err := os.Symlink(targetPath, linkPath); err != nil {
		t.Fatalf("Symlink: %v", err)
	}
	assertPathExists(t, targetPath+"-wal")
	assertPathExists(t, targetPath+"-shm")
	assertPathMissing(t, linkPath+"-wal")
	assertPathMissing(t, linkPath+"-shm")

	readOnly, err := OpenReadOnly(linkPath)
	if err != nil {
		t.Fatalf("OpenReadOnly() error = %v", err)
	}
	var goal string
	if err := readOnly.QueryRow(`SELECT goal FROM stories WHERE slug = 'symlink-wal-only'`).Scan(&goal); err != nil {
		readOnly.Close()
		t.Fatalf("read WAL state through symlink: %v", err)
	}
	if err := readOnly.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if goal != "resolved target WAL state" {
		t.Fatalf("goal = %q, want resolved target WAL state", goal)
	}
	assertDatabaseFileState(t, targetPath, before)
	assertPathMissing(t, linkPath+"-wal")
	assertPathMissing(t, linkPath+"-shm")
}

func TestOpenReadOnlyRefusesNonEmptyWALWithoutSHM(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "harness.db")
	writable, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, _, err := Migrate(writable); err != nil {
		writable.Close()
		t.Fatalf("Migrate() error = %v", err)
	}
	if err := writable.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := os.WriteFile(path+"-wal", []byte("non-empty WAL"), 0o644); err != nil {
		t.Fatalf("write WAL fixture: %v", err)
	}

	_, err = OpenReadOnly(path)
	if err == nil || !strings.Contains(err.Error(), "non-empty WAL exists without SHM") {
		t.Fatalf("OpenReadOnly() error = %v, want missing SHM refusal", err)
	}
	assertPathMissing(t, path+"-shm")
}

func captureDatabaseFileState(t *testing.T, path string) databaseFileState {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%s): %v", path, err)
	}
	return databaseFileState{data: data, size: info.Size(), modTime: info.ModTime()}
}

func assertDatabaseFileState(t *testing.T, path string, want databaseFileState) {
	t.Helper()
	got := captureDatabaseFileState(t, path)
	if !bytes.Equal(got.data, want.data) || got.size != want.size || !got.modTime.Equal(want.modTime) {
		t.Fatalf("database changed: before(size=%d mod=%s) after(size=%d mod=%s)", want.size, want.modTime, got.size, got.modTime)
	}
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("Stat(%s): %v", path, err)
	}
}

func assertPathMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("Stat(%s) error = %v, want not exist", path, err)
	}
}
