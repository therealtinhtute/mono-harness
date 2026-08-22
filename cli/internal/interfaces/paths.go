package interfaces

import (
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const (
	dbPath       = "harness.db"
	legacyDBPath = ".kit/harness.db"
	kitDir       = ".kit"
	kitRoot      = ".kit"
	docsDir      = "docs"
	conflictDir  = ".kit/conflicts"
)

var (
	gitRootOnce sync.Once
	gitRootDir  string
)

// gitRepositoryRoot shells out to git rev-parse --show-toplevel once per
// process, following the exec.Command("git", ...) precedent in
// plan_resolve.go. It returns "" outside a repository or when the git
// binary is unavailable.
func gitRepositoryRoot() string {
	gitRootOnce.Do(func() {
		out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			return
		}
		if root := strings.TrimSpace(string(out)); root != "" {
			gitRootDir = root
		}
	})
	return gitRootDir
}

// resolveDBPath anchors the database against the git repository root so a
// subcommand behaves identically regardless of which subdirectory of the
// repository it is invoked from (R22). Outside a git repository it falls
// back to the bare filename resolved against the process cwd, preserving
// pre-R22 behavior for non-git checkouts.
func resolveDBPath() string {
	if root := gitRepositoryRoot(); root != "" {
		return filepath.Join(root, dbPath)
	}
	return dbPath
}

// resolveLegacyDBPath gives legacyDBPath the same git-root treatment, since
// it names the same class of path.
func resolveLegacyDBPath() string {
	if root := gitRepositoryRoot(); root != "" {
		return filepath.Join(root, legacyDBPath)
	}
	return legacyDBPath
}
