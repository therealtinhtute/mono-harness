package embedded

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// TestBuildManifest_MatchesDiskTree guards against a doc being added under
// cli/docs/embedded on disk but not to the //go:embed directive in
// cli/docs/embedded/embed.go (or vice versa): the manifest and the actual
// tree init/resume rely on for staleness detection must agree exactly.
func TestBuildManifest_MatchesDiskTree(t *testing.T) {
	m, err := BuildManifest("test")
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	manifestPaths := append([]string(nil), m.Paths...)
	sort.Strings(manifestPaths)

	root := "../../docs/embedded"
	var diskPaths []string
	err = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(p) == ".go" {
			return nil // embed.go carries the directive, it is not itself embedded
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		diskPaths = append(diskPaths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	sort.Strings(diskPaths)

	if len(manifestPaths) != len(diskPaths) {
		t.Fatalf("manifest has %d paths, disk has %d\nmanifest: %v\ndisk: %v", len(manifestPaths), len(diskPaths), manifestPaths, diskPaths)
	}
	for i := range manifestPaths {
		if manifestPaths[i] != diskPaths[i] {
			t.Fatalf("manifest[%d] = %q, disk[%d] = %q (mismatch)\nmanifest: %v\ndisk: %v", i, manifestPaths[i], i, diskPaths[i], manifestPaths, diskPaths)
		}
	}
}
