package application

import "strings"

// AppendToPlanSection inserts entry into the named `## {section}` heading's
// body, positioned immediately before the next level-2 (`## `) heading — or
// at end of file, if section is the last heading present. Follows the
// targeted-line-scan precedent at next.go's parseActivePlanPhaseOrder (V3,
// docs/plans/active/harness-memory-ceremony-convergence.md) rather than a
// full markdown parser: this repository's plans don't nest `## ` sequences
// inside code fences, so a line-level scan is sufficient and matches the
// codebase's existing minimal-dependency style.
//
// entry is supplied with LF line endings; content's own line-ending style
// (LF or CRLF) is detected and preserved in the returned content, so a
// hand-edited CRLF plan is not silently converted to LF.
//
// If the section's body consists solely of the literal placeholder line
// "- none" (the scaffold template's bootstrap value, see scaffold.go), that
// line is replaced by entry. Otherwise entry is appended after the
// section's existing content, before any next heading. HTML comment blocks
// (`<!-- ... -->`, single- or multi-line — every scaffolded section carries
// one) are recognized so they are never mistaken for "- none" or split by
// an insertion.
//
// Returns an error, with content unchanged, if section's `## {section}`
// heading is not present anywhere in content — a plan missing an expected
// section is not repaired by invention; the caller decides how to recover.
// Section order in the file does not matter: the next-heading search looks
// for the next `## ` line after the target heading, whatever section that
// happens to be.
func AppendToPlanSection(content, section, entry string) (string, error) {
	crlf := strings.Contains(content, "\r\n")
	normalized := content
	if crlf {
		normalized = strings.ReplaceAll(content, "\r\n", "\n")
	}
	entry = strings.ReplaceAll(entry, "\r\n", "\n")

	lines := strings.Split(normalized, "\n")
	headingAt := -1
	for i, line := range lines {
		if name, ok := headingName(line); ok && name == section {
			headingAt = i
			break
		}
	}
	if headingAt == -1 {
		return "", &sectionNotFoundError{section: section}
	}

	bodyStart := headingAt + 1
	bodyEnd := len(lines)
	for i := bodyStart; i < len(lines); i++ {
		if _, ok := headingName(lines[i]); ok {
			bodyEnd = i
			break
		}
	}

	body := lines[bodyStart:bodyEnd]
	insertAt, replaceCount := planSectionInsertionPoint(body)
	entryLines := strings.Split(entry, "\n")

	newBody := make([]string, 0, len(body)+len(entryLines))
	newBody = append(newBody, body[:insertAt]...)
	newBody = append(newBody, entryLines...)
	newBody = append(newBody, body[insertAt+replaceCount:]...)

	result := make([]string, 0, len(lines)+len(entryLines))
	result = append(result, lines[:bodyStart]...)
	result = append(result, newBody...)
	result = append(result, lines[bodyEnd:]...)

	out := strings.Join(result, "\n")
	if crlf {
		out = strings.ReplaceAll(out, "\n", "\r\n")
	}
	return out, nil
}

// planSectionInsertionPoint decides where within body (the lines strictly
// between a section heading and the next one) a new entry belongs.
// Returns the index to insert before, and how many existing lines to
// replace at that index (0 for a pure insertion, 1 to replace a lone
// "- none" placeholder).
func planSectionInsertionPoint(body []string) (insertAt, replaceCount int) {
	substantive := substantiveLineIndices(body)
	if len(substantive) == 1 && strings.TrimSpace(body[substantive[0]]) == "- none" {
		return substantive[0], 1
	}
	for i := len(body) - 1; i >= 0; i-- {
		if strings.TrimSpace(body[i]) != "" {
			return i + 1, 0
		}
	}
	return 0, 0
}

// substantiveLineIndices returns indices of body lines that are neither
// blank nor part of an HTML comment block — the lines that matter for
// deciding whether a section's only content is the bootstrap placeholder.
func substantiveLineIndices(body []string) []int {
	var indices []int
	inComment := false
	for i, line := range body {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if inComment {
			if strings.Contains(line, "-->") {
				inComment = false
			}
			continue
		}
		hasOpen := strings.Contains(trimmed, "<!--")
		hasClose := strings.Contains(trimmed, "-->")
		switch {
		case hasOpen && hasClose:
			continue // self-contained one-line comment
		case hasOpen:
			inComment = true
			continue
		default:
			indices = append(indices, i)
		}
	}
	return indices
}

// headingName reports the level-2 (`## `) heading name of line, if line is
// one. A level-3+ heading (`### `) is not a match: the third character
// differs from the required space, so HasPrefix on "## " already excludes
// it — no separate check needed.
func headingName(line string) (name string, ok bool) {
	trimmed := strings.TrimRight(line, " \t")
	if !strings.HasPrefix(trimmed, "## ") {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")), true
}

type sectionNotFoundError struct{ section string }

func (e *sectionNotFoundError) Error() string {
	return "plan section not found: ## " + e.section
}
