package application

import (
	"regexp"
	"strings"
)

// planPhaseStatusLine matches a phase block's own `status:` field line
// (`    status: planned`), scoped to search within an already-located
// phase block range so it can never match a different phase's status line.
var planPhaseStatusLine = regexp.MustCompile(`(?m)^([ \t]*)status: [^\r\n]*\r?$`)

// SetPlanPhaseStatus replaces the `status:` line inside the phase block
// identified by phase_slug == slug with newStatus, trying the `###
// phase_slug:` heading form first and the `- phase_slug:` list-item form
// second — same precedence as extractPlanPhaseBlock (plan_query.go), so
// the read and write paths cannot independently drift on which block a
// slug resolves to. found is false if no phase block with that slug
// exists, or the block has no status line to replace — content is
// returned unchanged in that case, mirroring QueryPlanSection's Degraded
// behavior: a malformed or absent block is not repaired by invention.
func SetPlanPhaseStatus(content, slug, newStatus string) (out string, found bool) {
	crlf := strings.Contains(content, "\r\n")
	normalized := content
	if crlf {
		normalized = strings.ReplaceAll(content, "\r\n", "\n")
	}

	start, end, ok := planPhaseBlockRange(normalized, slug)
	if !ok {
		return content, false
	}

	block := normalized[start:end]
	loc := planPhaseStatusLine.FindStringSubmatchIndex(block)
	if loc == nil {
		return content, false
	}
	indent := block[loc[2]:loc[3]]
	newBlock := block[:loc[0]] + indent + "status: " + newStatus + block[loc[1]:]

	out = normalized[:start] + newBlock + normalized[end:]
	if crlf {
		out = strings.ReplaceAll(out, "\n", "\r\n")
	}
	return out, true
}

// planPhaseBlockRange returns the raw (untrimmed) [start,end) byte range of
// one phase's block within normalized content — the same boundary logic as
// extractPlanPhaseHeadingBlock/extractPlanPhaseListBlock (plan_query.go),
// but returning offsets instead of a trimmed copy, since SetPlanPhaseStatus
// splices a replacement back into the original content rather than reading
// a standalone slice.
func planPhaseBlockRange(normalized, slug string) (start, end int, found bool) {
	if start, end, ok := planPhaseHeadingBlockRange(normalized, slug); ok {
		return start, end, true
	}
	return planPhaseListBlockRange(normalized, slug)
}

func planPhaseHeadingBlockRange(normalized, slug string) (start, end int, found bool) {
	matches := planPhaseHeading.FindAllStringSubmatchIndex(normalized, -1)
	end = len(normalized)
	for i, m := range matches {
		if normalized[m[2]:m[3]] != slug {
			continue
		}
		start = m[1]
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		return start, end, true
	}
	return 0, 0, false
}

func planPhaseListBlockRange(normalized, slug string) (start, end int, found bool) {
	matches := planPhaseListItem.FindAllStringSubmatchIndex(normalized, -1)
	end = len(normalized)
	for i, m := range matches {
		if normalized[m[4]:m[5]] != slug {
			continue
		}
		indent := len(normalized[m[2]:m[3]])
		start = m[1]
		for j := i + 1; j < len(matches); j++ {
			if len(normalized[matches[j][2]:matches[j][3]]) <= indent {
				end = matches[j][0]
				break
			}
		}
		return start, end, true
	}
	return 0, 0, false
}
