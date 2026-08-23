package application

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

var fixtureDocsFS = fstest.MapFS{
	"AGENTS.md":         {Data: []byte("## Harness\n\nmanaged instructions\n")},
	"WORKFLOW.md":       {Data: []byte("# Workflow\n\nversion one\n")},
	"playbooks/work.md": {Data: []byte("# Work\n\nrun safely\n")},
}

var fixtureDocsFSV2 = fstest.MapFS{
	"AGENTS.md":         {Data: []byte("## Harness\n\nmanaged instructions\n")},
	"WORKFLOW.md":       {Data: []byte("# Workflow\n\nversion two\n")},
	"playbooks/work.md": {Data: []byte("# Work\n\nrun safely\n")},
}

func TestScaffoldDocsFreshRootProjection(t *testing.T) {
	db := freshDB(t)
	root := t.TempDir()

	result, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFS, "0.6.0", false, false)
	if err != nil {
		t.Fatalf("ScaffoldDocs() error = %v", err)
	}
	if !result.DocsWritten || !result.AgentsShimWritten || !result.GitignoreUpdated {
		t.Fatalf("ScaffoldDocs() result = %+v", result)
	}
	for path, want := range map[string]string{
		"docs/WORKFLOW.md":       "# Workflow\n\nversion one\n",
		"docs/playbooks/work.md": "# Work\n\nrun safely\n",
	} {
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if string(got) != want {
			t.Fatalf("%s = %q, want %q", path, got, want)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatal("embedded AGENTS.md was projected into docs")
	}

	agents, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(agents), agentsBlockStart) || !strings.Contains(string(agents), "managed instructions") || !strings.Contains(string(agents), agentsBlockEnd) {
		t.Fatalf("AGENTS.md managed block missing: %s", agents)
	}

	var managedCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM managed_docs`).Scan(&managedCount); err != nil {
		t.Fatalf("count managed_docs: %v", err)
	}
	if managedCount != 2 {
		t.Fatalf("managed_docs count = %d, want 2", managedCount)
	}
	var docsVersion string
	if err := db.QueryRow(`SELECT docs_version FROM meta`).Scan(&docsVersion); err != nil {
		t.Fatalf("read meta.docs_version: %v", err)
	}
	if docsVersion != "0.6.0" {
		t.Fatalf("docs_version = %q, want 0.6.0", docsVersion)
	}

	gitignore, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	for _, entry := range gitignoreEntries {
		if !strings.Contains(string(gitignore), entry) {
			t.Fatalf(".gitignore missing %q", entry)
		}
	}
}

func TestArchitectureScaffoldUsesUnansweredQuestions(t *testing.T) {
	if len(scaffoldOnceDocs) != 4 {
		t.Fatalf("scaffold-once entries = %d, want exactly four", len(scaffoldOnceDocs))
	}

	var architecture struct{ path, body string }
	for _, doc := range scaffoldOnceDocs {
		if doc.path == "docs/ARCHITECTURE.md" {
			architecture = doc
			break
		}
	}
	if architecture.path == "" {
		t.Fatal("architecture scaffold-once entry is missing")
	}
	if got := strings.Count(architecture.body, "?"); got != 5 {
		t.Fatalf("architecture questions = %d, want five", got)
	}
	if strings.Contains(architecture.body, "repository") || strings.Contains(architecture.body, "zharness") {
		t.Fatalf("architecture scaffold contains repository-specific text: %q", architecture.body)
	}

	db := freshDB(t)
	root := t.TempDir()
	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("initial ScaffoldDocs() error = %v", err)
	}
	path := filepath.Join(root, filepath.FromSlash(architecture.path))
	initial, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read architecture scaffold: %v", err)
	}
	if !bytes.Equal(initial, []byte(architecture.body)) {
		t.Fatalf("architecture scaffold content differs from declared body")
	}

	authored := []byte("# Answered\n")
	if err := os.WriteFile(path, authored, 0o644); err != nil {
		t.Fatalf("write answered architecture: %v", err)
	}
	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFSV2, "0.7.0", true, true); err != nil {
		t.Fatalf("forced refresh ScaffoldDocs() error = %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read architecture after refresh: %v", err)
	}
	if !bytes.Equal(got, authored) {
		t.Fatalf("architecture scaffold was rewritten: got %q, want %q", got, authored)
	}
}

func TestScaffoldOnceDocsSurviveForcedRefresh(t *testing.T) {
	db := freshDB(t)
	root := t.TempDir()
	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("initial ScaffoldDocs() error = %v", err)
	}

	for _, doc := range scaffoldOnceDocs {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(doc.path))); err != nil {
			t.Fatalf("scaffold-once %s not created: %v", doc.path, err)
		}
	}

	readme := filepath.Join(root, "docs", "README.md")
	authored := []byte("# Our docs\n\nconsumer-authored, must never be touched\n")
	if err := os.WriteFile(readme, authored, 0o644); err != nil {
		t.Fatalf("write consumer README: %v", err)
	}

	// force=true is what overwrites a conflicting *managed* doc; a scaffold-once
	// doc has no managed_docs row, so it is never compared and never rewritten.
	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFSV2, "0.7.0", true, true); err != nil {
		t.Fatalf("forced refresh ScaffoldDocs() error = %v", err)
	}

	got, err := os.ReadFile(readme)
	if err != nil {
		t.Fatalf("read README after forced refresh: %v", err)
	}
	if !bytes.Equal(got, authored) {
		t.Fatalf("consumer README rewritten: got %q, want %q", got, authored)
	}

	for _, doc := range scaffoldOnceDocs {
		var rows int
		if err := db.QueryRow(`SELECT COUNT(*) FROM managed_docs WHERE path = ?`, doc.path).Scan(&rows); err != nil {
			t.Fatalf("count managed_docs for %s: %v", doc.path, err)
		}
		if rows != 0 {
			t.Fatalf("managed_docs rows for scaffold-once %s = %d, want 0", doc.path, rows)
		}
	}

	if _, err := os.Stat(filepath.Join(root, ".kit", "conflicts", "docs", "README.md.upstream")); !os.IsNotExist(err) {
		t.Fatal("scaffold-once README staged as a conflict")
	}
}

func TestScaffoldDocsWritesClaudeMdImport(t *testing.T) {
	db := freshDB(t)
	root := t.TempDir()
	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("ScaffoldDocs() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	want := agentsBlockStart + "\n" + claudeMdImport + "\n" + agentsBlockEnd + "\n"
	if string(got) != want {
		t.Fatalf("CLAUDE.md = %q, want %q", got, want)
	}
	if strings.Contains(string(got), "managed instructions") {
		t.Fatal("CLAUDE.md holds a copy of the managed body instead of an import")
	}
}

func TestScaffoldDocsClaudeMdImportPreservesHumanText(t *testing.T) {
	db := freshDB(t)
	root := t.TempDir()
	local := "# Project rules\n\nkeep this byte-for-byte\n"
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte(local), 0o644); err != nil {
		t.Fatalf("seed CLAUDE.md: %v", err)
	}

	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("ScaffoldDocs() error = %v", err)
	}
	first, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	if !strings.HasPrefix(string(first), local) {
		t.Fatalf("human text changed: %q", first)
	}
	if !strings.Contains(string(first), claudeMdImport) {
		t.Fatalf("import missing: %q", first)
	}

	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFSV2, "0.7.0", true, false); err != nil {
		t.Fatalf("second ScaffoldDocs() error = %v", err)
	}
	second, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("re-read CLAUDE.md: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("second init changed CLAUDE.md: first=%q second=%q", first, second)
	}
	if strings.Count(string(second), agentsBlockStart) != 1 {
		t.Fatalf("marker duplicated: %q", second)
	}
}

func TestScaffoldDocsKeepsClaudeMdWithLegacyKitCopy(t *testing.T) {
	db := freshDB(t)
	root := t.TempDir()
	local := "# Project rules\n\nthis repo keeps a backup under .kit/docs\n"
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte(local), 0o644); err != nil {
		t.Fatalf("seed CLAUDE.md: %v", err)
	}
	legacyDir := filepath.Join(root, ".kit", "docs")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(.kit/docs) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, "CLAUDE.md"), []byte(local), 0o644); err != nil {
		t.Fatalf("seed .kit/docs/CLAUDE.md: %v", err)
	}

	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("ScaffoldDocs() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	if !strings.HasPrefix(string(got), local) {
		t.Fatalf("consumer CLAUDE.md replaced by the managed block: %q", got)
	}
	if !strings.Contains(string(got), claudeMdImport) {
		t.Fatalf("import missing: %q", got)
	}
}

func TestScaffoldDocsSecondRunIsNoop(t *testing.T) {
	db := freshDB(t)
	root := t.TempDir()
	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("first ScaffoldDocs() error = %v", err)
	}
	result, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFS, "0.6.0", false, false)
	if err != nil {
		t.Fatalf("second ScaffoldDocs() error = %v", err)
	}
	if result.DocsWritten || result.AgentsShimWritten || result.GitignoreUpdated {
		t.Fatalf("second ScaffoldDocs() result = %+v, want no writes", result)
	}
}

func TestScaffoldDocsPreservesAgentsContentOutsideManagedBlock(t *testing.T) {
	db := freshDB(t)
	root := t.TempDir()
	local := "# Local rules\n\nkeep this byte-for-byte\n"
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(local), 0o644); err != nil {
		t.Fatalf("seed AGENTS.md: %v", err)
	}

	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("ScaffoldDocs() error = %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.HasPrefix(string(got), local) || !strings.Contains(string(got), agentsBlockStart) {
		t.Fatalf("AGENTS.md local prefix changed: %q", got)
	}

	before := string(got)
	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFSV2, "0.7.0", true, false); err != nil {
		t.Fatalf("refresh ScaffoldDocs() error = %v", err)
	}
	after, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read refreshed AGENTS.md: %v", err)
	}
	if !strings.HasPrefix(string(after), local) || strings.Count(string(after), agentsBlockStart) != 1 {
		t.Fatalf("AGENTS.md outside block or marker count changed: before=%q after=%q", before, after)
	}
}

func TestScaffoldDocsRefreshUpdatesUntouchedFile(t *testing.T) {
	db := freshDB(t)
	root := t.TempDir()
	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("initial ScaffoldDocs() error = %v", err)
	}

	result, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFSV2, "0.7.0", true, false)
	if err != nil {
		t.Fatalf("refresh ScaffoldDocs() error = %v", err)
	}
	if !result.DocsWritten {
		t.Fatal("DocsWritten = false, want true")
	}
	got, err := os.ReadFile(filepath.Join(root, "docs", "WORKFLOW.md"))
	if err != nil {
		t.Fatalf("read WORKFLOW.md: %v", err)
	}
	if string(got) != string(fixtureDocsFSV2["WORKFLOW.md"].Data) {
		t.Fatalf("WORKFLOW.md = %q", got)
	}
}

func TestScaffoldDocsRefreshPreservesLocalOnlyEdit(t *testing.T) {
	db := freshDB(t)
	root := t.TempDir()
	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("initial ScaffoldDocs() error = %v", err)
	}
	path := filepath.Join(root, "docs", "WORKFLOW.md")
	local := []byte("# Local workflow\n")
	if err := os.WriteFile(path, local, 0o644); err != nil {
		t.Fatalf("write local edit: %v", err)
	}

	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFS, "0.7.0", true, false); err != nil {
		t.Fatalf("refresh ScaffoldDocs() error = %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read local edit: %v", err)
	}
	if string(got) != string(local) {
		t.Fatalf("local-only edit overwritten: %q", got)
	}
}

func TestScaffoldDocsRefreshStagesConflictWithoutOverwrite(t *testing.T) {
	db := freshDB(t)
	root := t.TempDir()
	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("initial ScaffoldDocs() error = %v", err)
	}
	path := filepath.Join(root, "docs", "WORKFLOW.md")
	local := []byte("# Local workflow\n")
	if err := os.WriteFile(path, local, 0o644); err != nil {
		t.Fatalf("write local edit: %v", err)
	}
	_, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFSV2, "0.7.0", true, false)
	var conflict *ManagedDocsConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("ScaffoldDocs() error = %v, want ManagedDocsConflictError", err)
	}
	got, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("read local file: %v", readErr)
	}
	if string(got) != string(local) {
		t.Fatalf("conflicting local file overwritten: %q", got)
	}
	staged, readErr := os.ReadFile(filepath.Join(root, ".kit", "conflicts", "docs", "WORKFLOW.md.upstream"))
	if readErr != nil {
		t.Fatalf("read staged upstream file: %v", readErr)
	}
	if string(staged) != string(fixtureDocsFSV2["WORKFLOW.md"].Data) {
		t.Fatalf("staged upstream = %q", staged)
	}
	var docsVersion string
	if err := db.QueryRow(`SELECT docs_version FROM meta`).Scan(&docsVersion); err != nil {
		t.Fatalf("read meta.docs_version: %v", err)
	}
	if docsVersion != "0.6.0" {
		t.Fatalf("conflict updated docs_version = %q, want unchanged 0.6.0", docsVersion)
	}
}

func TestScaffoldDocsForceOverwritesConflict(t *testing.T) {
	db := freshDB(t)
	root := t.TempDir()
	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("initial ScaffoldDocs() error = %v", err)
	}
	path := filepath.Join(root, "docs", "WORKFLOW.md")
	if err := os.WriteFile(path, []byte("local edit"), 0o644); err != nil {
		t.Fatalf("write local edit: %v", err)
	}

	if _, err := ScaffoldDocs(db, root, ".kit", fixtureDocsFSV2, "0.7.0", true, true); err != nil {
		t.Fatalf("forced ScaffoldDocs() error = %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read WORKFLOW.md: %v", err)
	}
	if string(got) != string(fixtureDocsFSV2["WORKFLOW.md"].Data) {
		t.Fatalf("forced WORKFLOW.md = %q", got)
	}
}
