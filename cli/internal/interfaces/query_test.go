package interfaces

import (
	"bytes"
	"os"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func TestQueryValidateAuditResumeOpenDatabaseReadOnly(t *testing.T) {
	commands := []struct {
		name string
		args []string
	}{
		{name: "query", args: []string{"query", "state", "--json"}},
		{name: "validate", args: []string{"validate", "--json"}},
		{name: "audit", args: []string{"audit", "--json"}},
		{name: "resume", args: []string{"resume", "--json"}},
	}

	for _, command := range commands {
		t.Run(command.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			db, err := infrastructure.Open(dbPath)
			if err != nil {
				t.Fatalf("Open: %v", err)
			}
			if _, _, err := infrastructure.Migrate(db); err != nil {
				db.Close()
				t.Fatalf("Migrate: %v", err)
			}
			var mode string
			if err := db.QueryRow(`PRAGMA journal_mode=DELETE`).Scan(&mode); err != nil {
				db.Close()
				t.Fatalf("set journal mode: %v", err)
			}
			if err := db.Close(); err != nil {
				t.Fatalf("Close: %v", err)
			}
			before, err := os.Stat(dbPath)
			if err != nil {
				t.Fatalf("Stat before command: %v", err)
			}

			jsonOutput = false
			resumeFacts = ""
			cmd := NewRootCmd("dev")
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(command.args)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("%s command: %v (output=%s)", command.name, err, output.String())
			}

			readOnly, err := infrastructure.OpenReadOnly(dbPath)
			if err != nil {
				t.Fatalf("OpenReadOnly: %v", err)
			}
			if err := readOnly.QueryRow(`PRAGMA journal_mode`).Scan(&mode); err != nil {
				readOnly.Close()
				t.Fatalf("read journal mode: %v", err)
			}
			if err := readOnly.Close(); err != nil {
				t.Fatalf("close read-only db: %v", err)
			}
			if mode != "delete" {
				t.Fatalf("journal_mode = %q, want delete; command reopened the DB writable", mode)
			}
			after, err := os.Stat(dbPath)
			if err != nil {
				t.Fatalf("Stat after command: %v", err)
			}
			if before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
				t.Fatalf("command mutated DB metadata: before=%+v after=%+v", before, after)
			}
		})
	}
}
