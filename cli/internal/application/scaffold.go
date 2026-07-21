package application

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// scaffoldKinds maps a scaffold kind to its embedded template path. The
// four kinds mirror the artifact skeletons the work/check/handoff/brainstorm
// playbooks previously carried inline as prose.
var scaffoldKinds = map[string]string{
	"run":     "templates/run.md",
	"check":   "templates/check.md",
	"handoff": "templates/handoff.md",
	"spec":    "templates/spec.md",
}

// ScaffoldArtifact writes the embedded skeleton for kind to path and
// returns the bytes written. It refuses to overwrite an existing non-empty
// file (file_exists) so a half-filled artifact is never clobbered; an
// empty placeholder file is allowed to be filled in. Pure file I/O — no DB
// row, no changeset. The artifact's harness row is still written by the
// kind's own record command (run create / check record / handoff record /
// intake), keeping scaffold a template emitter, not a state mutation.
func ScaffoldArtifact(templates fs.FS, kind, path string) ([]byte, error) {
	rel, ok := scaffoldKinds[kind]
	if !ok {
		return nil, &domain.ValidationError{Code: "unknown_kind", Message: "scaffold: unknown kind " + kind + " (want run|check|handoff|spec)"}
	}
	if path == "" {
		return nil, &domain.ValidationError{Code: "missing_required_field", Message: "scaffold: --path is required"}
	}
	if info, err := os.Stat(path); err == nil && info.Size() > 0 {
		return nil, &domain.ValidationError{Code: "file_exists", Message: "scaffold: " + path + " already exists and is non-empty; refusing to overwrite"}
	}
	data, err := fs.ReadFile(templates, rel)
	if err != nil {
		return nil, fmt.Errorf("scaffold: read template %s: %w", rel, err)
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("scaffold: mkdir %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return nil, fmt.Errorf("scaffold: write %s: %w", path, err)
	}
	return data, nil
}
