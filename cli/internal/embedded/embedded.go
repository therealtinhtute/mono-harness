// Package embedded exposes the canonical doc set (playbooks + shim +
// authority docs) bundled into the zharness binary, plus a manifest used
// by init to scaffold a project and by resume to detect drift.
package embedded

import (
	"io/fs"

	docsembedded "github.com/therealtinhtute/skills/cli/docs/embedded"
)

// FS is the embedded doc set: AGENTS.md, AUTHORITY.md, CONTEXT_RULES.md,
// and playbooks/*.md, rooted at cli/docs/embedded.
var FS = docsembedded.FS

// Manifest lists every embedded doc path alongside the docs_version stamp
// identifying which CLI build produced them.
type Manifest struct {
	Paths       []string
	DocsVersion string
}

// BuildManifest walks FS and returns every embedded doc path plus the
// supplied docs version (the CLI's own version string; "dev" for
// unreleased builds).
func BuildManifest(docsVersion string) (Manifest, error) {
	var paths []string
	err := fs.WalkDir(FS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			paths = append(paths, p)
		}
		return nil
	})
	if err != nil {
		return Manifest{}, err
	}
	return Manifest{Paths: paths, DocsVersion: docsVersion}, nil
}

// PlaybookCount returns how many playbook files are embedded.
func PlaybookCount() (int, error) {
	entries, err := fs.ReadDir(FS, "playbooks")
	if err != nil {
		return 0, err
	}
	return len(entries), nil
}
