package application

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

type ManagedDocsConflictError struct {
	Paths []string
}

func (e *ManagedDocsConflictError) Error() string {
	return "managed docs conflict: " + strings.Join(e.Paths, ", ")
}

type ManagedDocsResult struct {
	DocsWritten bool
}

type managedDocState struct {
	InstalledSHA256 string
	DocsVersion     string
}

type managedDocAction struct {
	RelativePath string
	TargetPath   string
	Content      []byte
	SHA256       string
	Op           string
	Write        bool
}

func SyncManagedDocs(db *sql.DB, changesetDir, docsRoot, conflictRoot string, docsFS fs.FS, docsVersion string, refresh, force bool) (ManagedDocsResult, error) {
	states, err := loadManagedDocStates(db)
	if err != nil {
		return ManagedDocsResult{}, err
	}
	if !refresh && !force && len(states) > 0 && infrastructure.Exists(docsRoot) {
		return ManagedDocsResult{}, nil
	}

	actions, conflicts, err := planManagedDocActions(states, docsRoot, docsFS, docsVersion, force)
	if err != nil {
		return ManagedDocsResult{}, err
	}
	if len(conflicts) > 0 {
		for _, action := range conflicts {
			path := filepath.Join(conflictRoot, filepath.FromSlash(action.RelativePath)+".upstream")
			if err := writeManagedFile(path, action.Content); err != nil {
				return ManagedDocsResult{}, fmt.Errorf("stage managed doc conflict %s: %w", action.RelativePath, err)
			}
		}
		paths := make([]string, len(conflicts))
		for i, action := range conflicts {
			paths[i] = action.RelativePath
		}
		sort.Strings(paths)
		return ManagedDocsResult{}, &ManagedDocsConflictError{Paths: paths}
	}

	result := ManagedDocsResult{}
	for _, action := range actions {
		if !action.Write {
			continue
		}
		if err := writeManagedFile(action.TargetPath, action.Content); err != nil {
			return result, fmt.Errorf("write managed doc %s: %w", action.RelativePath, err)
		}
		result.DocsWritten = true
	}

	lines, err := managedDocChangesetLines(db, actions, docsVersion)
	if err != nil {
		return result, err
	}
	if len(lines) > 0 {
		if _, _, err := AppendAndApply(db, changesetDir, lines); err != nil {
			return result, fmt.Errorf("record managed docs: %w", err)
		}
	}
	return result, nil
}

func loadManagedDocStates(db *sql.DB) (map[string]managedDocState, error) {
	rows, err := db.Query(`SELECT path, installed_sha256, docs_version FROM managed_docs`)
	if err != nil {
		return nil, fmt.Errorf("read managed docs: %w", err)
	}
	defer rows.Close()

	states := map[string]managedDocState{}
	for rows.Next() {
		var path string
		var state managedDocState
		if err := rows.Scan(&path, &state.InstalledSHA256, &state.DocsVersion); err != nil {
			return nil, fmt.Errorf("scan managed doc: %w", err)
		}
		states[path] = state
	}
	return states, rows.Err()
}

func planManagedDocActions(states map[string]managedDocState, docsRoot string, docsFS fs.FS, docsVersion string, force bool) (actions, conflicts []managedDocAction, err error) {
	err = fs.WalkDir(docsFS, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || path == "AGENTS.md" {
			return nil
		}
		content, err := fs.ReadFile(docsFS, path)
		if err != nil {
			return err
		}

		relative := filepath.ToSlash(filepath.Join("docs", filepath.FromSlash(path)))
		target := filepath.Join(docsRoot, filepath.FromSlash(path))
		upstreamHash := managedDocSHA256(content)
		local, readErr := os.ReadFile(target)
		localMissing := os.IsNotExist(readErr)
		if readErr != nil && !localMissing {
			return fmt.Errorf("read local managed doc %s: %w", relative, readErr)
		}
		localHash := ""
		if !localMissing {
			localHash = managedDocSHA256(local)
		}

		state, installed := states[relative]
		action := managedDocAction{RelativePath: relative, TargetPath: target, Content: content, SHA256: upstreamHash}
		switch {
		case force:
			action.Write = localMissing || localHash != upstreamHash
			if installed {
				action.Op = "update"
			} else {
				action.Op = "create"
			}
		case !installed && localMissing:
			action.Op = "create"
			action.Write = true
		case !installed && localHash == upstreamHash:
			action.Op = "create"
		case !installed:
			conflicts = append(conflicts, action)
			return nil
		case localMissing:
			action.Op = "update"
			action.Write = true
		case localHash == upstreamHash:
			if state.InstalledSHA256 != upstreamHash || state.DocsVersion != docsVersion {
				action.Op = "update"
			}
		case localHash == state.InstalledSHA256:
			action.Op = "update"
			action.Write = true
		case upstreamHash == state.InstalledSHA256:
			if state.DocsVersion != docsVersion {
				action.Op = "update"
				action.SHA256 = state.InstalledSHA256
			}
		default:
			conflicts = append(conflicts, action)
			return nil
		}
		if action.Op != "" {
			actions = append(actions, action)
		}
		return nil
	})
	return actions, conflicts, err
}

func managedDocChangesetLines(db *sql.DB, actions []managedDocAction, docsVersion string) ([]infrastructure.ChangesetLine, error) {
	at := time.Now().UTC().Format(time.RFC3339)
	lines := make([]infrastructure.ChangesetLine, 0, len(actions)+1)
	for _, action := range actions {
		fields := map[string]any{
			"installed_sha256": action.SHA256,
			"docs_version":     docsVersion,
			"updated_at":       at,
		}
		if action.Op == "create" {
			fields["path"] = action.RelativePath
		}
		lines = append(lines, infrastructure.ChangesetLine{
			Op: action.Op, Entity: "managed_doc", ID: action.RelativePath, Fields: fields, At: at,
		})
	}

	var current sql.NullString
	if err := db.QueryRow(`SELECT docs_version FROM meta LIMIT 1`).Scan(&current); err != nil {
		return nil, fmt.Errorf("read meta.docs_version: %w", err)
	}
	if current.String != docsVersion {
		lines = append(lines, infrastructure.ChangesetLine{
			Op: "update", Entity: "meta", ID: "meta", Fields: map[string]any{"docs_version": docsVersion}, At: at,
		})
	}
	return lines, nil
}

func writeManagedFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func managedDocSHA256(content []byte) string {
	sum := sha256.Sum256(content)
	return fmt.Sprintf("%x", sum[:])
}
