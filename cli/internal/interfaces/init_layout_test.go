package interfaces

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func TestInitReplaysExistingChangesetsBeforeDocsStamp(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	path, err := infrastructure.WriteChangeset(changesetDir, []infrastructure.ChangesetLine{{
		Op: "create", Entity: "story", ID: "01JZZZZZZZZZZZZZZZZZZZZZZZ",
		Fields: map[string]any{"slug": "replayed", "goal": "prove replay", "status": "planned", "created_at": "2026-07-26T00:00:00Z"},
		At:     "2026-07-26T00:00:00Z",
	}})
	if err != nil {
		t.Fatalf("WriteChangeset() error = %v", err)
	}
	if filepath.Dir(path) != changesetDir {
		t.Fatalf("changeset path = %s", path)
	}

	jsonOutput = false
	cmd := NewRootCmd("dev")
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"init", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command error = %v", err)
	}

	db, err := infrastructure.OpenReadOnly(dbPath)
	if err != nil {
		t.Fatalf("OpenReadOnly() error = %v", err)
	}
	defer db.Close()
	var slug string
	if err := db.QueryRow(`SELECT slug FROM stories WHERE id = ?`, "01JZZZZZZZZZZZZZZZZZZZZZZZ").Scan(&slug); err != nil {
		t.Fatalf("replayed story missing: %v", err)
	}
	if slug != "replayed" {
		t.Fatalf("story slug = %q", slug)
	}
	pending, _, _, err := infrastructure.ChangesetStatus(db.Raw(), changesetDir)
	if err != nil {
		t.Fatalf("ChangesetStatus() error = %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("pending changesets = %v", pending)
	}
}

func TestInitForceRemovesSidecarsAndReplays(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := infrastructure.WriteChangeset(changesetDir, []infrastructure.ChangesetLine{{
		Op: "create", Entity: "story", ID: "01JXXXXXXXXXXXXXXXXXXXXXXX",
		Fields: map[string]any{"slug": "force-replay", "goal": "prove force replay", "status": "planned", "created_at": "2026-07-26T00:00:00Z"},
		At:     "2026-07-26T00:00:00Z",
	}}); err != nil {
		t.Fatalf("WriteChangeset() error = %v", err)
	}

	jsonOutput = false
	cmd := NewRootCmd("dev")
	cmd.SetArgs([]string{"init"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("initial init error = %v", err)
	}
	for _, path := range []string{dbPath + "-wal", dbPath + "-shm"} {
		if err := os.WriteFile(path, []byte("stale sidecar"), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	cmd = NewRootCmd("dev")
	cmd.SetArgs([]string{"init", "--force"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("forced init error = %v", err)
	}
	db, err := infrastructure.OpenReadOnly(dbPath)
	if err != nil {
		t.Fatalf("OpenReadOnly() error = %v", err)
	}
	defer db.Close()
	var slug string
	if err := db.QueryRow(`SELECT slug FROM stories WHERE id = ?`, "01JXXXXXXXXXXXXXXXXXXXXXXX").Scan(&slug); err != nil {
		t.Fatalf("force rebuild did not replay story: %v", err)
	}
	if slug != "force-replay" {
		t.Fatalf("story slug = %q", slug)
	}
}

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
