package application

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// gitignoreEntries are the paths every project's .gitignore must cover so a
// local db/cache never gets committed. Append-only: existing file content is
// never rewritten or reordered.
var gitignoreEntries = []string{".kit/harness.db", ".kit/cache/"}

// ScaffoldResult reports what ScaffoldDocs actually did, for init's
// text/--json output.
type ScaffoldResult struct {
	DocsWritten          bool
	AgentsShimWritten    bool
	AgentsShimNoticePath string // set when a root AGENTS.md already existed
	GitignoreUpdated     bool
	DocsVersion          string
}

// ScaffoldDocs implements the doc-scaffolding half of `init`/`init
// --refresh-docs`: copies the embedded doc set into {kitDir}/docs, writes a
// root AGENTS.md shim only if one is not already present, ensures
// .gitignore covers the db + cache paths, and stamps meta.docs_version.
//
// Per the locked idempotency matrix, the doc copy + version stamp only run
// when {kitDir}/docs is absent or refresh is true; the AGENTS.md shim and
// .gitignore checks are cheap, idempotent, and always run so a second plain
// `init` never leaves either half-applied.
func ScaffoldDocs(db *sql.DB, changesetDir, root, kitDir string, docsFS fs.FS, docsVersion string, refresh bool) (ScaffoldResult, error) {
	result := ScaffoldResult{DocsVersion: docsVersion}

	docsDir := filepath.Join(kitDir, "docs")
	if !infrastructure.Exists(docsDir) || refresh {
		if err := copyFS(docsFS, docsDir); err != nil {
			return result, fmt.Errorf("scaffold docs: %w", err)
		}
		result.DocsWritten = true

		if err := stampDocsVersion(db, changesetDir, docsVersion); err != nil {
			return result, fmt.Errorf("stamp docs_version: %w", err)
		}
	}

	shimWritten, noticePath, err := writeAgentsShim(root, docsDir, docsFS)
	if err != nil {
		return result, fmt.Errorf("agents shim: %w", err)
	}
	result.AgentsShimWritten = shimWritten
	result.AgentsShimNoticePath = noticePath

	updated, err := ensureGitignore(root)
	if err != nil {
		return result, fmt.Errorf("gitignore: %w", err)
	}
	result.GitignoreUpdated = updated

	return result, nil
}

// copyFS writes every file in src to destDir, creating parent directories
// as needed. Existing files are overwritten — the caller decides when this
// runs (fresh scaffold or an explicit --refresh-docs), never blindly.
func copyFS(src fs.FS, destDir string) error {
	return fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(destDir, p)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(src, p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// writeAgentsShim writes the embedded AGENTS.md to the repo root only if no
// root AGENTS.md already exists. Never overwrites one that's already there
// — instead it returns the path to the canonical copy under docsDir so the
// caller can print a notice.
func writeAgentsShim(root, docsDir string, docsFS fs.FS) (written bool, noticePath string, err error) {
	rootAgents := filepath.Join(root, "AGENTS.md")
	if infrastructure.Exists(rootAgents) {
		return false, filepath.Join(docsDir, "AGENTS.md"), nil
	}

	data, err := fs.ReadFile(docsFS, "AGENTS.md")
	if err != nil {
		return false, "", err
	}
	if err := os.WriteFile(rootAgents, data, 0o644); err != nil {
		return false, "", err
	}
	return true, "", nil
}

// ensureGitignore appends any of gitignoreEntries missing from the root
// .gitignore. Existing content is never rewritten or reordered; a missing
// file is created containing only the required entries.
func ensureGitignore(root string) (bool, error) {
	path := filepath.Join(root, ".gitignore")
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	present := map[string]bool{}
	for _, line := range strings.Split(string(existing), "\n") {
		present[strings.TrimSpace(line)] = true
	}

	var missing []string
	for _, entry := range gitignoreEntries {
		if !present[entry] {
			missing = append(missing, entry)
		}
	}
	if len(missing) == 0 {
		return false, nil
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return false, err
	}
	defer f.Close()

	var b strings.Builder
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		b.WriteString("\n")
	}
	for _, entry := range missing {
		b.WriteString(entry)
		b.WriteString("\n")
	}
	_, err = f.WriteString(b.String())
	return true, err
}

// stampDocsVersion updates meta.docs_version via a changeset only when it
// actually differs from the current value, so a repeat init/refresh with
// the same CLI version writes zero changesets (STATE.md idempotency).
func stampDocsVersion(db *sql.DB, changesetDir, docsVersion string) error {
	var existing sql.NullString
	if err := db.QueryRow(`SELECT docs_version FROM meta LIMIT 1`).Scan(&existing); err != nil {
		return fmt.Errorf("read meta.docs_version: %w", err)
	}
	if existing.String == docsVersion {
		return nil
	}

	at := time.Now().UTC().Format(time.RFC3339)
	_, _, err := AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "update", Entity: "meta", ID: "meta", Fields: map[string]any{"docs_version": docsVersion}, At: at},
	})
	return err
}
