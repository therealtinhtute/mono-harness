package interfaces

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/installer"
)

func run(t *testing.T, args ...string) error {
	t.Helper()
	root := NewRootCmd("test")
	root.SetArgs(args)
	return root.Execute()
}

func TestUpdateCheck_Flags(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	repo := t.TempDir()
	var sb strings.Builder
	if err := installer.Install(repo, "test", &sb); err != nil {
		t.Fatalf("install: %v\n%s", err, sb.String())
	}

	if err := run(t, "update", "--all"); err == nil || !strings.Contains(err.Error(), "--all requires --check") {
		t.Fatalf("--all without --check = %v", err)
	}
	if err := run(t, "update", "--check", "--root", repo); err != nil {
		t.Fatalf("clean --check = %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "docs/WORKFLOW.md"), []byte("edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(t, "update", "--check", "--root", repo); err != installer.ErrDrift {
		t.Fatalf("drifted --check = %v, want ErrDrift", err)
	}
	if err := run(t, "update", "--check", "--all"); err != installer.ErrDrift {
		t.Fatalf("drifted --check --all = %v, want ErrDrift", err)
	}
	if b, _ := os.ReadFile(filepath.Join(repo, "docs/WORKFLOW.md")); string(b) != "edited\n" {
		t.Fatal("--check wrote to the repository")
	}
}
