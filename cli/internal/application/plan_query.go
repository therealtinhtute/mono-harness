package application

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// PlanSectionView is the `query plan --section ...` view: a slice of the
// active plan's markdown, not the whole file. It complements the
// compressed-index queries (traces/decisions/checks) for the two things
// about a plan that stay markdown-only — never table-backed, because
// they're either a free-text snapshot or a task's own definition rather
// than append-only history: the `## Current State and Next Action`
// snapshot, and one phase's `### phase_slug: ...` block (its waves, tasks,
// and checks) inside `## Phases and Verification` (P3,
// docs/audit/workflow-harness-ceremony-audit.md).
//
// Degraded is true when the requested section or phase block could not be
// found — Content then carries the full plan file instead of failing the
// call, so a malformed hand-edited plan degrades an agent's read to what
// it already had to do before this command existed, rather than blocking
// it outright.
type PlanSectionView struct {
	Path     string `json:"path"`
	Section  string `json:"section"`
	Phase    string `json:"phase,omitempty"`
	Content  string `json:"content"`
	Degraded bool   `json:"degraded"`
}

var planPhaseHeading = regexp.MustCompile("(?m)^### phase_slug: `([^`\r\n]+)`[ \t]*\r?$")

// planPhaseListItem matches the scaffold template's list form
// (`  - phase_slug: {slug}`, under `## Phases and Verification` →
// `- phases:`) — the shape `zharness scaffold plan` and `to-plan` actually
// produce when a plan is hand- or agent-filled without `###` headings.
// Group 1 is the item's leading whitespace, used to find its sibling
// boundary; group 2 is the slug.
var planPhaseListItem = regexp.MustCompile(`(?m)^([ \t]*)- phase_slug: ([a-zA-Z0-9][a-zA-Z0-9_-]*)[ \t]*\r?$`)

// QueryPlanSection resolves the single active plan under
// docs/plans/active/*.md and returns the requested slice. section is
// "current-state" or "phase" ("phase" requires phase, a phase_slug).
func QueryPlanSection(section, phase string) (PlanSectionView, error) {
	if section != "current-state" && section != "phase" {
		return PlanSectionView{}, &domain.ValidationError{
			Code:    "unknown_section",
			Message: fmt.Sprintf("query plan: unknown section %q (want current-state|phase)", section),
		}
	}
	if section == "phase" && phase == "" {
		return PlanSectionView{}, &domain.ValidationError{
			Code:    "missing_required_field",
			Message: "query plan --section phase requires --phase {slug}",
		}
	}

	plans, err := findActivePlans()
	if err != nil {
		return PlanSectionView{}, err
	}
	switch len(plans) {
	case 0:
		return PlanSectionView{}, &domain.ValidationError{
			Code:    "no_active_plan",
			Message: "query plan: no non-empty plan under docs/plans/active/*.md",
		}
	case 1:
		// fall through
	default:
		return PlanSectionView{}, &domain.ValidationError{
			Code:    "ambiguous_active_plan",
			Message: fmt.Sprintf("query plan: %d active plans found; this command requires exactly one", len(plans)),
		}
	}
	plan := plans[0]

	var body string
	var found bool
	if section == "current-state" {
		body, found = extractPlanSection(plan.content, "Current State and Next Action")
	} else {
		body, found = extractPlanPhaseBlock(plan.content, phase)
	}

	view := PlanSectionView{Path: plan.path, Section: section, Phase: phase}
	if !found {
		view.Content = plan.content
		view.Degraded = true
		return view, nil
	}
	view.Content = body
	return view, nil
}

// extractPlanSection returns the body of a `## {name}` heading — every line
// after it up to (not including) the next `## ` heading, or end of file.
// Shares findPlanSectionBody (plan_section.go) with AppendToPlanSection so
// the read and write paths cannot independently drift on where a section
// starts and ends (V3).
func extractPlanSection(content, name string) (string, bool) {
	lines := strings.Split(normalizeLineEndings(content), "\n")
	start, end, ok := findPlanSectionBody(lines, name)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(strings.Join(lines[start:end], "\n")), true
}

// extractPlanPhaseBlock returns one phase's block, trying the `###
// phase_slug: \`{slug}\`` heading form first and the scaffold template's
// `- phase_slug: {slug}` list form second — a plan may use either
// depending on how to-plan filled it in. The heading form keeps precedence
// when a plan happens to contain both (an unlikely hand-edit, but the
// heading form is the more explicit signal).
func extractPlanPhaseBlock(content, slug string) (string, bool) {
	normalized := normalizeLineEndings(content)
	if body, ok := extractPlanPhaseHeadingBlock(normalized, slug); ok {
		return body, true
	}
	return extractPlanPhaseListBlock(normalized, slug)
}

// extractPlanPhaseHeadingBlock returns one phase's `### phase_slug:
// \`{slug}\`` block — every line after that heading up to (not including)
// the next `### phase_slug:` heading or the next `## ` heading, whichever
// comes first. Not scoped to inside `## Phases and Verification`
// specifically: phase_slug headings are unambiguous on their own, and
// scoping would only add a failure mode for a plan whose sections got
// reordered by hand. content must already be normalized to LF.
func extractPlanPhaseHeadingBlock(normalized, slug string) (string, bool) {
	matches := planPhaseHeading.FindAllStringSubmatchIndex(normalized, -1)

	start := -1
	end := len(normalized)
	for i, m := range matches {
		if normalized[m[2]:m[3]] != slug {
			continue
		}
		start = m[1] // end of the matched heading line (before its newline)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		break
	}
	if start == -1 {
		return "", false
	}
	return trimPlanPhaseBlockBody(normalized[start:end]), true
}

// extractPlanPhaseListBlock returns one phase's `  - phase_slug: {slug}`
// list-item block — every line after that item up to (not including) the
// next sibling `- phase_slug:` item at the same or lesser indentation, or
// the next `## ` heading. Sibling boundary is indentation-based rather
// than "next match" because a phase's own body (waves, tasks) may itself
// contain deeper-indented lines that are not siblings. content must
// already be normalized to LF.
func extractPlanPhaseListBlock(normalized, slug string) (string, bool) {
	matches := planPhaseListItem.FindAllStringSubmatchIndex(normalized, -1)

	start := -1
	end := len(normalized)
	for i, m := range matches {
		if normalized[m[4]:m[5]] != slug {
			continue
		}
		indent := len(normalized[m[2]:m[3]])
		start = m[1] // end of the matched list-item line (before its newline)
		for j := i + 1; j < len(matches); j++ {
			if len(normalized[matches[j][2]:matches[j][3]]) <= indent {
				end = matches[j][0]
				break
			}
		}
		break
	}
	if start == -1 {
		return "", false
	}
	return trimPlanPhaseBlockBody(normalized[start:end]), true
}

// trimPlanPhaseBlockBody clips a candidate phase body at the next top-level
// "## " heading — which can still fall inside a naively-sliced range when
// the requested phase is the last one defined — then trims surrounding
// whitespace.
func trimPlanPhaseBlockBody(body string) string {
	if idx := strings.Index(body, "\n## "); idx != -1 {
		body = body[:idx]
	} else if strings.HasPrefix(body, "## ") {
		body = ""
	}
	return strings.TrimSpace(body)
}

func normalizeLineEndings(content string) string {
	return strings.ReplaceAll(content, "\r\n", "\n")
}
