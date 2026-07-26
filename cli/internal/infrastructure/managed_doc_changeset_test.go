package infrastructure

import (
	"path/filepath"
	"testing"
)

func TestManagedDocChangesetReplay(t *testing.T) {
	t.Parallel()

	db := freshDB(t)
	path, err := WriteChangeset(filepath.Join(t.TempDir(), "changesets"), []ChangesetLine{{
		Op:     "create",
		Entity: "managed_doc",
		ID:     "docs/WORKFLOW.md",
		Fields: map[string]any{
			"path":             "docs/WORKFLOW.md",
			"installed_sha256": "abc123",
			"docs_version":     "0.6.0",
			"updated_at":       "2026-07-26T00:00:00Z",
		},
		At: "2026-07-26T00:00:00Z",
	}})
	if err != nil {
		t.Fatalf("WriteChangeset() error = %v", err)
	}

	if _, _, err := ApplyChangeset(db, path); err != nil {
		t.Fatalf("ApplyChangeset() error = %v", err)
	}
	var gotPath, gotHash, gotVersion string
	if err := db.QueryRow(`SELECT path, installed_sha256, docs_version FROM managed_docs WHERE id = ?`, "docs/WORKFLOW.md").Scan(&gotPath, &gotHash, &gotVersion); err != nil {
		t.Fatalf("query managed_docs: %v", err)
	}
	if gotPath != "docs/WORKFLOW.md" || gotHash != "abc123" || gotVersion != "0.6.0" {
		t.Fatalf("managed doc = (%q, %q, %q)", gotPath, gotHash, gotVersion)
	}
}
