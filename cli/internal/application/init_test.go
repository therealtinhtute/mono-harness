package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// fixtureDocsFS is a minimal stand-in for the real embedded doc set, small
// enough to assert file-by-file without depending on cli/docs/embedded's
// actual content.
var fixtureDocsFS = fstest.MapFS{
	"AGENTS.md":         {Data: []byte("shim content")},
	"AUTHORITY.md":      {Data: []byte("authority content")},
	"playbooks/work.md": {Data: []byte("work playbook")},
}

func countChangesets(t *testing.T, changesetDir string) int {
	t.Helper()
	entries, err := os.ReadDir(changesetDir)
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", changesetDir, err)
	}
	return len(entries)
}

// TestScaffoldDocs_FreshScaffold covers idempotency matrix cell A (no
// .kit/docs at all): every embedded file is copied, the AGENTS.md shim is
// written to repo root (absent there), .gitignore is created with both
// required entries, and docs_version is stamped.
func TestScaffoldDocs_FreshScaffold(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	kitDir := filepath.Join(root, ".kit")

	result, err := ScaffoldDocs(db, changesetDir, root, kitDir, fixtureDocsFS, "0.2.0", false)
	if err != nil {
		t.Fatalf("ScaffoldDocs: %v", err)
	}
	if !result.DocsWritten {
		t.Error("DocsWritten = false, want true on a fresh scaffold")
	}
	if !result.AgentsShimWritten {
		t.Error("AgentsShimWritten = false, want true when root AGENTS.md is absent")
	}
	if !result.GitignoreUpdated {
		t.Error("GitignoreUpdated = false, want true on a fresh .gitignore")
	}

	for path, file := range fixtureDocsFS {
		got, err := os.ReadFile(filepath.Join(kitDir, "docs", path))
		if err != nil {
			t.Fatalf("read scaffolded %s: %v", path, err)
		}
		if string(got) != string(file.Data) {
			t.Errorf("%s content = %q, want %q", path, got, file.Data)
		}
	}

	shim, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read root AGENTS.md: %v", err)
	}
	if string(shim) != "shim content" {
		t.Errorf("root AGENTS.md = %q, want %q", shim, "shim content")
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
	if docsVersion != "0.2.0" {
		t.Errorf("meta.docs_version = %q, want %q", docsVersion, "0.2.0")
	}
}

// TestScaffoldDocs_SecondRunIsNoop covers the "no file churn" requirement:
// calling ScaffoldDocs again with the same docsVersion and refresh=false
// must not rewrite docs, must not touch an already-written AGENTS.md, must
// not touch an already-complete .gitignore, and — critically — must not
// write a new changeset (AppendAndApply mints a fresh file every call, so
// an unconditional docs_version stamp would dirty git status on every init).
func TestScaffoldDocs_SecondRunIsNoop(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	kitDir := filepath.Join(root, ".kit")

	if _, err := ScaffoldDocs(db, changesetDir, root, kitDir, fixtureDocsFS, "0.2.0", false); err != nil {
		t.Fatalf("first ScaffoldDocs: %v", err)
	}
	changesetsAfterFirst := countChangesets(t, changesetDir)

	result, err := ScaffoldDocs(db, changesetDir, root, kitDir, fixtureDocsFS, "0.2.0", false)
	if err != nil {
		t.Fatalf("second ScaffoldDocs: %v", err)
	}
	if result.DocsWritten {
		t.Error("DocsWritten = true on second run, want false (docs already present, refresh=false)")
	}
	if result.AgentsShimWritten {
		t.Error("AgentsShimWritten = true on second run, want false (shim already present)")
	}
	if result.AgentsShimNoticePath == "" {
		t.Error("AgentsShimNoticePath = empty on second run, want the canonical path notice")
	}
	if result.GitignoreUpdated {
		t.Error("GitignoreUpdated = true on second run, want false (.gitignore already complete)")
	}

	if got := countChangesets(t, changesetDir); got != changesetsAfterFirst {
		t.Errorf("changeset count after second run = %d, want %d (no new changeset for an unchanged docs_version)", got, changesetsAfterFirst)
	}
}

// TestScaffoldDocs_AddDocsOnly covers idempotency matrix cell B: an
// existing project dir with no .kit/docs subdir yet gets docs added,
// independent of anything else already in kitDir.
func TestScaffoldDocs_AddDocsOnly(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	kitDir := filepath.Join(root, ".kit")
	if err := os.MkdirAll(kitDir, 0o755); err != nil {
		t.Fatalf("pre-create kitDir: %v", err)
	}

	result, err := ScaffoldDocs(db, changesetDir, root, kitDir, fixtureDocsFS, "0.2.0", false)
	if err != nil {
		t.Fatalf("ScaffoldDocs: %v", err)
	}
	if !result.DocsWritten {
		t.Error("DocsWritten = false, want true when .kit exists but .kit/docs does not")
	}
}

// TestScaffoldDocs_ExistingRootAgentsNeverOverwritten proves a pre-existing
// root AGENTS.md (not one we wrote) is left untouched, per the locked
// "never overwrite" decision.
func TestScaffoldDocs_ExistingRootAgentsNeverOverwritten(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	kitDir := filepath.Join(root, ".kit")
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("user's own content"), 0o644); err != nil {
		t.Fatalf("seed root AGENTS.md: %v", err)
	}

	result, err := ScaffoldDocs(db, changesetDir, root, kitDir, fixtureDocsFS, "0.2.0", false)
	if err != nil {
		t.Fatalf("ScaffoldDocs: %v", err)
	}
	if result.AgentsShimWritten {
		t.Error("AgentsShimWritten = true, want false (root AGENTS.md already existed)")
	}
	if result.AgentsShimNoticePath == "" {
		t.Error("AgentsShimNoticePath = empty, want the canonical shim location")
	}

	got, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read root AGENTS.md: %v", err)
	}
	if string(got) != "user's own content" {
		t.Errorf("root AGENTS.md was overwritten: got %q", got)
	}
}

// TestScaffoldDocs_RefreshRestoresCanonicalContent covers T4: `refresh=true`
// overwrites locally modified docs with the canonical embedded content and
// bumps the docs_version stamp, even though {kitDir}/docs already exists
// (the one case plain init, refresh=false, leaves untouched). It must not
// merge user edits — canonical overwrite is the contract.
func TestScaffoldDocs_RefreshRestoresCanonicalContent(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	kitDir := filepath.Join(root, ".kit")

	if _, err := ScaffoldDocs(db, changesetDir, root, kitDir, fixtureDocsFS, "0.2.0", false); err != nil {
		t.Fatalf("initial ScaffoldDocs: %v", err)
	}

	workPath := filepath.Join(kitDir, "docs", "playbooks", "work.md")
	if err := os.WriteFile(workPath, []byte("user-modified content"), 0o644); err != nil {
		t.Fatalf("simulate local edit: %v", err)
	}

	result, err := ScaffoldDocs(db, changesetDir, root, kitDir, fixtureDocsFS, "0.3.0", true)
	if err != nil {
		t.Fatalf("refresh ScaffoldDocs: %v", err)
	}
	if !result.DocsWritten {
		t.Error("DocsWritten = false on refresh, want true even though docs already existed")
	}

	got, err := os.ReadFile(workPath)
	if err != nil {
		t.Fatalf("read refreshed work.md: %v", err)
	}
	want := string(fixtureDocsFS["playbooks/work.md"].Data)
	if string(got) != want {
		t.Errorf("work.md after refresh = %q, want canonical content %q (user edit should be overwritten, not merged)", got, want)
	}

	var docsVersion string
	if err := db.QueryRow("SELECT docs_version FROM meta").Scan(&docsVersion); err != nil {
		t.Fatalf("query meta.docs_version: %v", err)
	}
	if docsVersion != "0.3.0" {
		t.Errorf("meta.docs_version after refresh = %q, want %q", docsVersion, "0.3.0")
	}
}

// TestScaffoldDocs_RefreshLeavesOtherMetaPointersUntouched proves refresh
// only ever writes docs_version — current_phase/entry_phase/latest_run_id
// (the other meta columns `resume` reads) must be unaffected, so `resume`
// output is otherwise unchanged across a refresh.
func TestScaffoldDocs_RefreshLeavesOtherMetaPointersUntouched(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	kitDir := filepath.Join(root, ".kit")

	storyID := "01JZZZZZZZZZZZZZZZZZZZZZZZ"
	seedAt := "2026-07-18T10:00:00Z"
	if _, _, err := AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "create", Entity: "story", ID: storyID, Fields: map[string]any{
			"slug": "cli-embed-scaffold", "goal": "test fixture", "status": "in-progress", "created_at": seedAt,
		}, At: seedAt},
		{Op: "update", Entity: "meta", ID: "meta", Fields: map[string]any{
			"current_phase": "cli-embed-scaffold", "entry_phase": "cli-embed-scaffold",
		}, At: seedAt},
	}); err != nil {
		t.Fatalf("seed meta pointers: %v", err)
	}

	if _, err := ScaffoldDocs(db, changesetDir, root, kitDir, fixtureDocsFS, "0.2.0", false); err != nil {
		t.Fatalf("initial ScaffoldDocs: %v", err)
	}
	if _, err := ScaffoldDocs(db, changesetDir, root, kitDir, fixtureDocsFS, "0.3.0", true); err != nil {
		t.Fatalf("refresh ScaffoldDocs: %v", err)
	}

	var currentPhase, entryPhase string
	if err := db.QueryRow(`SELECT current_phase, entry_phase FROM meta`).Scan(&currentPhase, &entryPhase); err != nil {
		t.Fatalf("query meta pointers: %v", err)
	}
	if currentPhase != "cli-embed-scaffold" || entryPhase != "cli-embed-scaffold" {
		t.Errorf("meta pointers after refresh = (%q, %q), want unchanged (%q, %q)", currentPhase, entryPhase, "cli-embed-scaffold", "cli-embed-scaffold")
	}
}
