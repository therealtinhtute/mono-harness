package embedded

import (
	"os"
	"path/filepath"
	"testing"
)

// TestProjectionDrift_RootDocsMatchEmbed guards the managed root projection.
// AGENTS.md is excluded because init merges only its marked block; WORKFLOW
// and playbooks must remain byte-identical to their embedded sources.
func TestProjectionDrift_RootDocsMatchEmbed(t *testing.T) {
	m, err := BuildManifest("test")
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}

	root := "../../../docs"
	for _, p := range m.Paths {
		if p == "AGENTS.md" {
			continue
		}
		want, err := FS.ReadFile(p)
		if err != nil {
			t.Fatalf("read embedded %s: %v", p, err)
		}
		got, err := os.ReadFile(filepath.Join(root, p))
		if err != nil {
			t.Fatalf("docs/%s missing or unreadable: %v (run `zharness init --refresh-docs`)", p, err)
		}
		if string(got) != string(want) {
			t.Errorf("docs/%s has drifted from the embed — run `zharness init --refresh-docs`", p)
		}
	}
}
