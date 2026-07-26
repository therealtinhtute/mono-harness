package application

import (
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

func TestScaffoldDocsFreshRootProjection(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()

	result, err := ScaffoldDocs(db, changesetDir, root, ".kit", fixtureDocsFS, "0.6.0", false, false)
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

func TestScaffoldDocsSecondRunIsNoop(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	if _, err := ScaffoldDocs(db, changesetDir, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("first ScaffoldDocs() error = %v", err)
	}
	before := countChangesets(t, changesetDir)

	result, err := ScaffoldDocs(db, changesetDir, root, ".kit", fixtureDocsFS, "0.6.0", false, false)
	if err != nil {
		t.Fatalf("second ScaffoldDocs() error = %v", err)
	}
	if result.DocsWritten || result.AgentsShimWritten || result.GitignoreUpdated {
		t.Fatalf("second ScaffoldDocs() result = %+v, want no writes", result)
	}
	if after := countChangesets(t, changesetDir); after != before {
		t.Fatalf("changeset count = %d, want %d", after, before)
	}
}

func TestScaffoldDocsPreservesAgentsContentOutsideManagedBlock(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	local := "# Local rules\n\nkeep this byte-for-byte\n"
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(local), 0o644); err != nil {
		t.Fatalf("seed AGENTS.md: %v", err)
	}

	if _, err := ScaffoldDocs(db, changesetDir, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
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
	if _, err := ScaffoldDocs(db, changesetDir, root, ".kit", fixtureDocsFSV2, "0.7.0", true, false); err != nil {
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
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	if _, err := ScaffoldDocs(db, changesetDir, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("initial ScaffoldDocs() error = %v", err)
	}

	result, err := ScaffoldDocs(db, changesetDir, root, ".kit", fixtureDocsFSV2, "0.7.0", true, false)
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
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	if _, err := ScaffoldDocs(db, changesetDir, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("initial ScaffoldDocs() error = %v", err)
	}
	path := filepath.Join(root, "docs", "WORKFLOW.md")
	local := []byte("# Local workflow\n")
	if err := os.WriteFile(path, local, 0o644); err != nil {
		t.Fatalf("write local edit: %v", err)
	}

	if _, err := ScaffoldDocs(db, changesetDir, root, ".kit", fixtureDocsFS, "0.7.0", true, false); err != nil {
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
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	if _, err := ScaffoldDocs(db, changesetDir, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("initial ScaffoldDocs() error = %v", err)
	}
	path := filepath.Join(root, "docs", "WORKFLOW.md")
	local := []byte("# Local workflow\n")
	if err := os.WriteFile(path, local, 0o644); err != nil {
		t.Fatalf("write local edit: %v", err)
	}
	beforeChangesets := countChangesets(t, changesetDir)

	_, err := ScaffoldDocs(db, changesetDir, root, ".kit", fixtureDocsFSV2, "0.7.0", true, false)
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
	if after := countChangesets(t, changesetDir); after != beforeChangesets {
		t.Fatalf("conflict wrote changeset: before=%d after=%d", beforeChangesets, after)
	}
}

func TestScaffoldDocsForceOverwritesConflict(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	if _, err := ScaffoldDocs(db, changesetDir, root, ".kit", fixtureDocsFS, "0.6.0", false, false); err != nil {
		t.Fatalf("initial ScaffoldDocs() error = %v", err)
	}
	path := filepath.Join(root, "docs", "WORKFLOW.md")
	if err := os.WriteFile(path, []byte("local edit"), 0o644); err != nil {
		t.Fatalf("write local edit: %v", err)
	}

	if _, err := ScaffoldDocs(db, changesetDir, root, ".kit", fixtureDocsFSV2, "0.7.0", true, true); err != nil {
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
