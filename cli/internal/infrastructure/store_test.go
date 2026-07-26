package infrastructure

import (
	"os"
	"path/filepath"
	"testing"
)

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
