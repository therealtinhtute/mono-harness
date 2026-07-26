// Package embedded holds the canonical doc set shipped inside the zharness
// binary via go:embed. The directive must live in this directory because
// go:embed patterns cannot traverse ".." out of the package directory.
package embedded

import "embed"

//go:embed AGENTS.md WORKFLOW.md playbooks
var FS embed.FS

// Templates holds the artifact-skeleton set (run/check/handoff/spec/plan)
// emitted on demand by `zharness scaffold`. It is a separate embed.FS from
// FS on purpose: FS is walked by BuildManifest and scaffolded into
// .kit/docs at init, whereas templates are never projected — they are
// written only when the scaffold command asks for one.
//
//go:embed templates
var Templates embed.FS
