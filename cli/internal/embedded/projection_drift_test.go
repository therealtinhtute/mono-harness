package embedded

import (
	"os"
	"path/filepath"
	"testing"
)

// TestProjectionDrift_KitDocsMatchesEmbed guards the single-source-of-truth
// contract: this repo's own tracked `.kit/docs/` projection must stay
// byte-identical to the embed that produced it (`zharness init
// --refresh-docs`). It exists because `.kit/docs/*` is committed to git and
// therefore hand-editable — the exact drift precedent from issue #24. The
// embed is canonical; `.kit/docs/` is generated output, never edited
// directly.
func TestProjectionDrift_KitDocsMatchesEmbed(t *testing.T) {
	m, err := BuildManifest("test")
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}

	root := "../../../.kit/docs"
	for _, p := range m.Paths {
		want, err := FS.ReadFile(p)
		if err != nil {
			t.Fatalf("read embedded %s: %v", p, err)
		}
		got, err := os.ReadFile(filepath.Join(root, p))
		if err != nil {
			t.Fatalf(".kit/docs/%s missing or unreadable: %v (run `zharness init --refresh-docs`)", p, err)
		}
		if string(got) != string(want) {
			t.Errorf(".kit/docs/%s has drifted from the embed — run `zharness init --refresh-docs` (never hand-edit .kit/docs/*)", p)
		}
	}
}
