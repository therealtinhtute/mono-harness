package application

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var gitignoreEntries = []string{"harness.db", "harness.db-wal", "harness.db-shm", ".kit/cache/", ".kit/conflicts/"}

const (
	agentsBlockStart = "<!-- ZHARNESS:BEGIN -->"
	agentsBlockEnd   = "<!-- ZHARNESS:END -->"
)

type ScaffoldResult struct {
	DocsWritten          bool
	AgentsShimWritten    bool
	AgentsShimNoticePath string
	GitignoreUpdated     bool
	DocsVersion          string
}

func ScaffoldDocs(db *sql.DB, changesetDir, root, kitDir string, docsFS fs.FS, docsVersion string, refresh, forceDocs bool) (ScaffoldResult, error) {
	result := ScaffoldResult{DocsVersion: docsVersion}

	managed, err := SyncManagedDocs(
		db,
		changesetDir,
		filepath.Join(root, "docs"),
		filepath.Join(root, kitDir, "conflicts"),
		docsFS,
		docsVersion,
		refresh,
		forceDocs,
	)
	if err != nil {
		return result, fmt.Errorf("scaffold docs: %w", err)
	}
	result.DocsWritten = managed.DocsWritten

	agentsWritten, err := writeAgentsManagedBlock(root, docsFS)
	if err != nil {
		return result, fmt.Errorf("agents block: %w", err)
	}
	result.AgentsShimWritten = agentsWritten

	updated, err := ensureGitignore(root)
	if err != nil {
		return result, fmt.Errorf("gitignore: %w", err)
	}
	result.GitignoreUpdated = updated
	return result, nil
}

func writeAgentsManagedBlock(root string, docsFS fs.FS) (bool, error) {
	content, err := fs.ReadFile(docsFS, "AGENTS.md")
	if err != nil {
		return false, err
	}
	block := agentsBlockStart + "\n" + strings.TrimSpace(string(content)) + "\n" + agentsBlockEnd
	path := filepath.Join(root, "AGENTS.md")
	existing, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return true, os.WriteFile(path, []byte(block+"\n"), 0o644)
	}
	if err != nil {
		return false, err
	}

	legacyMatches := false
	if legacy, legacyErr := os.ReadFile(filepath.Join(root, ".kit", "docs", "AGENTS.md")); legacyErr == nil {
		legacyMatches = bytes.Equal(existing, legacy)
	}

	text := string(existing)
	start := strings.Index(text, agentsBlockStart)
	end := strings.Index(text, agentsBlockEnd)
	if (start >= 0) != (end >= 0) || (start >= 0 && end < start) {
		return false, fmt.Errorf("AGENTS.md contains an incomplete zharness managed block")
	}

	var updated string
	switch {
	case start >= 0:
		end += len(agentsBlockEnd)
		updated = text[:start] + block + text[end:]
	case legacyMatches || strings.TrimSpace(text) == strings.TrimSpace(string(content)):
		updated = block + "\n"
	default:
		separator := "\n"
		if !strings.HasSuffix(text, "\n") {
			separator = "\n\n"
		}
		updated = text + separator + block + "\n"
	}
	if updated == text {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(updated), 0o644)
}

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

	var builder strings.Builder
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		builder.WriteString("\n")
	}
	for _, entry := range missing {
		builder.WriteString(entry)
		builder.WriteString("\n")
	}
	_, err = f.WriteString(builder.String())
	return true, err
}
