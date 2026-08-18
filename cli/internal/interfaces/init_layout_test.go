package interfaces

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func TestInitRefusesLegacyDatabaseFork(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.MkdirAll(filepath.Dir(legacyDBPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	db, err := infrastructure.Open(legacyDBPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, _, err := infrastructure.Migrate(db); err != nil {
		db.Close()
		t.Fatalf("Migrate() error = %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	jsonOutput = false
	cmd := NewRootCmd("dev")
	cmd.SetArgs([]string{"init"})
	err = cmd.Execute()
	cliErr, ok := err.(*cliError)
	if !ok || cliErr.Code != "layout_migration_required" {
		t.Fatalf("init error = %T %v, want layout_migration_required", err, err)
	}
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Fatal("init created root database beside legacy database")
	}
}
