// Package embedded holds the canonical doc set shipped inside the zharness
// binary via go:embed. The directive must live in this directory because
// go:embed patterns cannot traverse ".." out of the package directory.
package embedded

import "embed"

//go:embed AGENTS.md AUTHORITY.md CONTEXT_RULES.md playbooks
var FS embed.FS
