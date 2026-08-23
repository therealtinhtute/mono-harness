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
	claudeMdImport   = "@AGENTS.md"
)

type ScaffoldResult struct {
	DocsWritten          bool
	AgentsShimWritten    bool
	ClaudeShimWritten    bool
	AgentsShimNoticePath string
	GitignoreUpdated     bool
	DocsVersion          string
}

func ScaffoldDocs(db *sql.DB, root, kitDir string, docsFS fs.FS, docsVersion string, refresh, forceDocs bool) (ScaffoldResult, error) {
	result := ScaffoldResult{DocsVersion: docsVersion}

	managed, err := SyncManagedDocs(
		db,
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

	agentsBody, err := fs.ReadFile(docsFS, "AGENTS.md")
	if err != nil {
		return result, fmt.Errorf("agents block: %w", err)
	}
	agentsWritten, err := writeManagedBlock(root, "AGENTS.md", string(agentsBody))
	if err != nil {
		return result, fmt.Errorf("agents block: %w", err)
	}
	// Claude Code reads CLAUDE.md and never AGENTS.md (anthropics/claude-code#34235),
	// so the managed contract reaches it as an import rather than a second copy.
	claudeWritten, err := writeManagedBlock(root, "CLAUDE.md", claudeMdImport)
	if err != nil {
		return result, fmt.Errorf("claude import block: %w", err)
	}
	result.AgentsShimWritten = agentsWritten
	result.ClaudeShimWritten = claudeWritten

	if err := writeScaffoldOnceDocs(root); err != nil {
		return result, fmt.Errorf("scaffold-once docs: %w", err)
	}

	updated, err := ensureGitignore(root)
	if err != nil {
		return result, fmt.Errorf("gitignore: %w", err)
	}
	result.GitignoreUpdated = updated
	return result, nil
}

func writeManagedBlock(root, relPath, content string) (bool, error) {
	block := agentsBlockStart + "\n" + strings.TrimSpace(content) + "\n" + agentsBlockEnd
	path := filepath.Join(root, relPath)
	existing, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return true, os.WriteFile(path, []byte(block+"\n"), 0o644)
	}
	if err != nil {
		return false, err
	}

	// Only AGENTS.md was ever projected to .kit/docs/, so only it can be a
	// harness-authored leftover worth replacing wholesale. Probing that path for
	// any other file would hand the replace branch a consumer-owned copy.
	legacyMatches := false
	if relPath == "AGENTS.md" {
		if legacy, legacyErr := os.ReadFile(filepath.Join(root, ".kit", "docs", relPath)); legacyErr == nil {
			legacyMatches = bytes.Equal(existing, legacy)
		}
	}

	text := string(existing)
	start := strings.Index(text, agentsBlockStart)
	end := strings.Index(text, agentsBlockEnd)
	if (start >= 0) != (end >= 0) || (start >= 0 && end < start) {
		return false, fmt.Errorf("%s contains an incomplete zharness managed block", relPath)
	}

	var updated string
	switch {
	case start >= 0:
		end += len(agentsBlockEnd)
		updated = text[:start] + block + text[end:]
	case legacyMatches || strings.TrimSpace(text) == strings.TrimSpace(content):
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

// scaffoldOnceDocs bodies write "~" where the markdown needs a backtick, since
// a Go raw string literal cannot contain one.
var scaffoldOnceDocs = []struct{ path, body string }{
	{"docs/README.md", strings.ReplaceAll(docsReadmeBody, "~", "`")},
	{"docs/decisions/README.md", strings.ReplaceAll(decisionsReadmeBody, "~", "`")},
	{"docs/decisions/templates/decision.md", strings.ReplaceAll(decisionTemplateBody, "~", "`")},
	{"docs/ARCHITECTURE.md", architectureDocBody},
}

// writeScaffoldOnceDocs seeds the authored-docs entrypoint. Unlike the managed
// set it records no managed_docs row, so a consumer edit is never compared,
// never staged as a conflict, and never refreshed.
func writeScaffoldOnceDocs(root string) error {
	for _, doc := range scaffoldOnceDocs {
		path := filepath.Join(root, filepath.FromSlash(doc.path))
		if _, err := os.Stat(path); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(doc.body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

const architectureDocBody = `# Architecture

## Problem and audience

What problem does this project solve, and who is its audience?

## Main use cases

What are the main use cases, described in domain language?

## Non-standard domain nouns

Which domain nouns have meanings here that are non-standard?

## Silent invariants

Which invariants can fail silently?

## Boundaries

Which boundaries must not be crossed?
`

const docsReadmeBody = `# Documentation

The authored documentation map for this repository. Every hand-written document
should be reachable from this page.

~zharness init~ wrote this file once, because it was absent. It is yours now —
the harness never refreshes, overwrites, or deletes it.

## Where to go

| You want to | Read |
|---|---|
| Run a workflow stage | docs/WORKFLOW.md, then the one playbook it names |
| Know why something is built the way it is | docs/decisions/README.md |
| See what is being built right now | the active plan under docs/plans/active/ |

Add a row for each authored document as you write it.

## Ownership

Three classes. The class decides who is allowed to edit the file.

- **managed** — projected from the binary's embedded doc set and hash-tracked.
  Edit the embedded source and cut a release; a local edit is staged under
  .kit/conflicts/ rather than silently overwritten. Covers docs/WORKFLOW.md and
  docs/playbooks/.
- **scaffold-once** — written by ~zharness init~ only when absent, then owned by
  you. Covers this file, docs/decisions/README.md, and
  docs/decisions/templates/decision.md.
- **authored** — written by hand, never embedded, never regenerated. Everything
  else under docs/.

An existing path under docs/ that is missing from this page is a defect in this
page.
`

const decisionsReadmeBody = `# Decision Records

One file per decision that is expensive to reverse. Cheap, reversible choices do
not belong here — the code is their record.

Copy docs/decisions/templates/decision.md to start a new one, numbered in
sequence: 0001-short-slug.md, 0002-….

| # | Decision | Status |
|---|---|---|
| — | none yet | — |

Every record carries the same five headings: Status, Context, Decision,
Consequences, Authority. A record is never rewritten once accepted — supersede
it with a new one and mark the old one Superseded.
`

const decisionTemplateBody = `# NNNN — Short imperative title

## Status

Proposed | Accepted | Superseded by NNNN. Include the date it was accepted.

## Context

The forces in play: the constraint, the incident, the measurement. Enough that a
reader who was not present can tell why the question was open at all. State what
was actually observed, not what was assumed.

## Decision

What was decided, in the active voice. Name the alternative that was rejected
and the reason — a record without a rejected alternative is a description, not a
decision.

## Consequences

What this makes easy, and what it makes hard. Include the costs; a record that
lists only benefits was written to justify rather than to inform.

## Authority

Where the claims come from: commits, files with path:line citations, measured
output, external documentation, or the owner's call and its date.
`

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
