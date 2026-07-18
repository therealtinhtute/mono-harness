package infrastructure

import "testing"

// TestMetaDocsVersionRoundTrip verifies the 0002_meta_docs_version migration
// and the metaColumns allowlist entry together: a changeset setting
// docs_version applies cleanly and the value reads back via direct SQL.
// This also establishes the read pattern T5's integration test reuses.
func TestMetaDocsVersionRoundTrip(t *testing.T) {
	db := freshDB(t)
	dir := t.TempDir()

	path, err := WriteChangeset(dir, []ChangesetLine{
		{
			Op:     "update",
			Entity: "meta",
			ID:     "meta",
			Fields: map[string]any{"docs_version": "0.2.0"},
			At:     "2026-07-18T09:30:00Z",
		},
	})
	if err != nil {
		t.Fatalf("WriteChangeset: %v", err)
	}

	if _, _, err := ApplyChangeset(db, path); err != nil {
		t.Fatalf("ApplyChangeset: %v", err)
	}

	var got string
	if err := db.QueryRow("SELECT docs_version FROM meta").Scan(&got); err != nil {
		t.Fatalf("query meta.docs_version: %v", err)
	}
	if got != "0.2.0" {
		t.Fatalf("meta.docs_version = %q, want %q", got, "0.2.0")
	}
}
