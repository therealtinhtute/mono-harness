// Package embedded holds the canonical doc set shipped inside the zharness
// binary via go:embed. The directive must live in this directory because
// go:embed patterns cannot traverse ".." out of the package directory.
package embedded

import "embed"

//go:embed AGENTS.md WORKFLOW.md playbooks
var FS embed.FS

// Templates holds project.identity.md, which install and update copy to
// docs/PROJECT.md only when that file is absent. It is a separate embed.FS
// from FS on purpose: FS is walked by BuildManifest and fresh-overwritten
// on update, whereas the identity template is written once.
//
//go:embed templates
var Templates embed.FS
