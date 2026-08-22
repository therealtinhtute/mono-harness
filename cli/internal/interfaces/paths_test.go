package interfaces

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

// resetGitRootCache re-arms the per-process git-root discovery so each test
// resolves against its own temporary repository rather than one cached by a
// previous test.
func resetGitRootCache() {
	gitRootOnce = sync.Once{}
	gitRootDir = ""
}

// newGitRepo creates a temporary git repository with one commit so
// git rev-parse --show-toplevel succeeds inside it, and returns its root.
func newGitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = append(cmd.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	run("init", "-q", root)
	run("config", "user.name", "t")
	run("config", "user.email", "t@example.com")
	if err := os.WriteFile(filepath.Join(root, "probe.txt"), []byte("probe\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "probe.txt")
	run("commit", "-q", "-m", "probe")
	return root
}

func TestDBPathResolvesAgainstGitRootFromSubdirectory(t *testing.T) {
	resetGitRootCache()
	root := newGitRepo(t)
	subdir := filepath.Join(root, "cli", "internal", "application")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(subdir)

	got := resolveDBPath()
	want := filepath.Join(root, "harness.db")
	if got != want {
		t.Fatalf("resolved db path from subdirectory: got %q, want %q", got, want)
	}
}

func TestLegacyDBPathResolvesAgainstGitRootFromSubdirectory(t *testing.T) {
	resetGitRootCache()
	root := newGitRepo(t)
	subdir := filepath.Join(root, "some", "nested", "dir")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(subdir)

	got := resolveLegacyDBPath()
	want := filepath.Join(root, ".kit", "harness.db")
	if got != want {
		t.Fatalf("resolved legacy db path from subdirectory: got %q, want %q", got, want)
	}
}

func TestDBPathFallsBackToBareFilenameOutsideGitRepository(t *testing.T) {
	resetGitRootCache()
	dir := t.TempDir() // no .git ancestor anywhere above it in a fresh temp dir
	t.Chdir(dir)

	if got, want := resolveDBPath(), "harness.db"; got != want {
		t.Fatalf("fallback outside a git repository: got %q, want %q", got, want)
	}
	if got, want := resolveLegacyDBPath(), filepath.Join(".kit", "harness.db"); got != want {
		t.Fatalf("legacy fallback outside a git repository: got %q, want %q", got, want)
	}
}
