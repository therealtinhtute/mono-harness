package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/embedded"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// TestInit_FreshScratchDir_FullIntegration ties T1-T4 together against the
// real embedded doc set (not the fixtureDocsFS fake init_test.go's unit
// tests use): a fresh scratch dir, migrated db, and ScaffoldDocs run once —
// then asserts the full .kit/docs tree, AGENTS.md shim, .gitignore entries,
// docs_version stamp, and a clean Resume readiness, exactly as `zharness
// init` would leave a brand-new project.
func TestInit_FreshScratchDir_FullIntegration(t *testing.T) {
	root := t.TempDir()
	kitDir := filepath.Join(root, ".kit")
	changesetDir := filepath.Join(kitDir, "changesets")

	if err := os.MkdirAll(kitDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", kitDir, err)
	}

	db, err := infrastructure.Open(filepath.Join(kitDir, "harness.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	if _, _, err := infrastructure.Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	scaffold, err := ScaffoldDocs(db, changesetDir, root, kitDir, embedded.FS, "test-integration", false)
	if err != nil {
		t.Fatalf("ScaffoldDocs: %v", err)
	}
	if !scaffold.DocsWritten || !scaffold.AgentsShimWritten || !scaffold.GitignoreUpdated {
		t.Fatalf("scaffold result = %+v, want all three side effects on a fresh scratch dir", scaffold)
	}

	manifest, err := embedded.BuildManifest("test-integration")
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	for _, p := range manifest.Paths {
		want, err := embedded.FS.ReadFile(p)
		if err != nil {
			t.Fatalf("read embedded %s: %v", p, err)
		}
		got, err := os.ReadFile(filepath.Join(kitDir, "docs", p))
		if err != nil {
			t.Fatalf("scaffolded doc %s missing on disk: %v", p, err)
		}
		if string(got) != string(want) {
			t.Errorf("scaffolded %s content does not match embedded source", p)
		}
	}

	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Errorf("root AGENTS.md missing: %v", err)
	}

	gitignore, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	for _, entry := range gitignoreEntries {
		if !strings.Contains(string(gitignore), entry) {
			t.Errorf(".gitignore missing entry %q, got:\n%s", entry, gitignore)
		}
	}

	var docsVersion string
	if err := db.QueryRow("SELECT docs_version FROM meta").Scan(&docsVersion); err != nil {
		t.Fatalf("query meta.docs_version: %v", err)
	}
	if docsVersion != "test-integration" {
		t.Errorf("meta.docs_version = %q, want %q", docsVersion, "test-integration")
	}

	view, err := Resume(db, "test-integration")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if view.Readiness != "clean" {
		t.Errorf("Resume readiness = %q, want %q (fresh scaffold, no runs/checks recorded yet)", view.Readiness, "clean")
	}
	if len(view.Drift) != 0 {
		t.Errorf("Resume drift = %v, want none on a fresh scaffold", view.Drift)
	}
}
