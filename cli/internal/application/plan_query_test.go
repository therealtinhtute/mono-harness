package application

import (
	"strings"
	"testing"
)

const planQueryFixture = "# Plan: Demo\n" +
	"\n" +
	"## Outcome\n" +
	"- result: demo\n" +
	"\n" +
	"## Phases and Verification\n" +
	"### phase_slug: `p1-first`\n" +
	"- status: planned\n" +
	"- goal: first phase\n" +
	"- waves:\n" +
	"  - wave 1:\n" +
	"    - task: do the thing | check: it works\n" +
	"\n" +
	"### phase_slug: `p2-last`\n" +
	"- status: planned\n" +
	"- goal: last phase\n" +
	"- waves:\n" +
	"  - wave 1:\n" +
	"    - task: do the other thing | check: it also works\n" +
	"\n" +
	"## Progress\n" +
	"- none\n" +
	"\n" +
	"## Current State and Next Action\n" +
	"- active_phase: none\n" +
	"- exact_next_action: to-plan\n"

func TestExtractPlanSectionFindsCurrentState(t *testing.T) {
	body, ok := extractPlanSection(planQueryFixture, "Current State and Next Action")
	if !ok {
		t.Fatalf("extractPlanSection: not found")
	}
	want := "- active_phase: none\n- exact_next_action: to-plan"
	if body != want {
		t.Fatalf("body = %q, want %q", body, want)
	}
}

func TestExtractPlanSectionMissingHeadingReturnsNotFound(t *testing.T) {
	if _, ok := extractPlanSection(planQueryFixture, "Decisions"); ok {
		t.Fatalf("extractPlanSection: want not found for a heading absent from the fixture")
	}
}

func TestExtractPlanSectionHandlesCRLF(t *testing.T) {
	crlf := "# Plan\r\n\r\n## Outcome\r\n- result: demo\r\n\r\n## Current State and Next Action\r\n- active_phase: none\r\n"
	body, ok := extractPlanSection(crlf, "Current State and Next Action")
	if !ok {
		t.Fatalf("extractPlanSection: not found")
	}
	if body != "- active_phase: none" {
		t.Fatalf("body = %q, want %q", body, "- active_phase: none")
	}
}

func TestExtractPlanSectionCapturesTrailingContentAfterLastHeading(t *testing.T) {
	content := "## Current State and Next Action\n- active_phase: none\n- blockers: none\n"
	body, ok := extractPlanSection(content, "Current State and Next Action")
	if !ok {
		t.Fatalf("extractPlanSection: not found")
	}
	want := "- active_phase: none\n- blockers: none"
	if body != want {
		t.Fatalf("body = %q, want %q", body, want)
	}
}

func TestExtractPlanPhaseBlockFindsNonLastPhase(t *testing.T) {
	body, ok := extractPlanPhaseBlock(planQueryFixture, "p1-first")
	if !ok {
		t.Fatalf("extractPlanPhaseBlock: not found")
	}
	if got := body; !containsAll(got, "goal: first phase", "do the thing") {
		t.Fatalf("body = %q, want the p1-first block only", got)
	}
	if containsAll(body, "last phase") {
		t.Fatalf("body = %q, leaked p2-last's content", body)
	}
}

func TestExtractPlanPhaseBlockLastPhaseStopsAtNextTopLevelHeading(t *testing.T) {
	body, ok := extractPlanPhaseBlock(planQueryFixture, "p2-last")
	if !ok {
		t.Fatalf("extractPlanPhaseBlock: not found")
	}
	if !containsAll(body, "goal: last phase", "do the other thing") {
		t.Fatalf("body = %q, want the p2-last block", body)
	}
	if containsAll(body, "## Progress", "- none") {
		t.Fatalf("body = %q, leaked past the phase block into ## Progress", body)
	}
}

func TestExtractPlanPhaseBlockUnknownSlugReturnsNotFound(t *testing.T) {
	if _, ok := extractPlanPhaseBlock(planQueryFixture, "no-such-phase"); ok {
		t.Fatalf("extractPlanPhaseBlock: want not found for an undefined slug")
	}
}

func TestExtractPlanPhaseBlockHandlesCRLF(t *testing.T) {
	crlf := "## Phases and Verification\r\n### phase_slug: `only`\r\n- status: planned\r\n\r\n## Progress\r\n- none\r\n"
	body, ok := extractPlanPhaseBlock(crlf, "only")
	if !ok {
		t.Fatalf("extractPlanPhaseBlock: not found")
	}
	if body != "- status: planned" {
		t.Fatalf("body = %q, want %q", body, "- status: planned")
	}
}

// planQueryListFixture mirrors the literal shape `zharness scaffold plan`
// produces once `to-plan` fills it in without `###` headings — the shape
// audited in docs/audit/sdlc-token-cache-audit.md §4, which the pre-fix
// regex silently degraded on every plan built the documented way.
const planQueryListFixture = "# Plan: Demo\n" +
	"\n" +
	"## Outcome\n" +
	"- result: demo\n" +
	"\n" +
	"## Phases and Verification\n" +
	"- planning_status: planned\n" +
	"- phases:\n" +
	"  - phase_slug: p1-first\n" +
	"    story_id: 01STORYULID0000000000001\n" +
	"    status: planned\n" +
	"    goal: first phase\n" +
	"    depends_on: none\n" +
	"    waves:\n" +
	"      - wave: 1\n" +
	"        tasks:\n" +
	"          - task: do the thing\n" +
	"            check: it works\n" +
	"  - phase_slug: p2-last\n" +
	"    story_id: 01STORYULID0000000000002\n" +
	"    status: planned\n" +
	"    goal: last phase\n" +
	"    depends_on: ['p1-first']\n" +
	"    waves:\n" +
	"      - wave: 1\n" +
	"        tasks:\n" +
	"          - task: do the other thing\n" +
	"            check: it also works\n" +
	"\n" +
	"## Progress\n" +
	"- none\n" +
	"\n" +
	"## Current State and Next Action\n" +
	"- active_phase: none\n" +
	"- exact_next_action: to-plan\n"

func TestExtractPlanPhaseBlockListFormFindsNonLastPhase(t *testing.T) {
	body, ok := extractPlanPhaseBlock(planQueryListFixture, "p1-first")
	if !ok {
		t.Fatalf("extractPlanPhaseBlock: not found in list-form fixture")
	}
	if !containsAll(body, "goal: first phase", "do the thing") {
		t.Fatalf("body = %q, want the p1-first block only", body)
	}
	if containsAll(body, "last phase") {
		t.Fatalf("body = %q, leaked p2-last's content", body)
	}
}

func TestExtractPlanPhaseBlockListFormLastPhaseStopsAtNextTopLevelHeading(t *testing.T) {
	body, ok := extractPlanPhaseBlock(planQueryListFixture, "p2-last")
	if !ok {
		t.Fatalf("extractPlanPhaseBlock: not found in list-form fixture")
	}
	if !containsAll(body, "goal: last phase", "do the other thing") {
		t.Fatalf("body = %q, want the p2-last block", body)
	}
	if containsAll(body, "## Progress", "- none") {
		t.Fatalf("body = %q, leaked past the phase block into ## Progress", body)
	}
}

func TestExtractPlanPhaseBlockListFormUnknownSlugReturnsNotFound(t *testing.T) {
	if _, ok := extractPlanPhaseBlock(planQueryListFixture, "no-such-phase"); ok {
		t.Fatalf("extractPlanPhaseBlock: want not found for an undefined slug in list-form fixture")
	}
}

func TestExtractPlanPhaseBlockListFormHandlesCRLF(t *testing.T) {
	crlf := "## Phases and Verification\r\n" +
		"- phases:\r\n" +
		"  - phase_slug: only\r\n" +
		"    status: planned\r\n" +
		"\r\n" +
		"## Progress\r\n" +
		"- none\r\n"
	body, ok := extractPlanPhaseBlock(crlf, "only")
	if !ok {
		t.Fatalf("extractPlanPhaseBlock: not found")
	}
	if body != "status: planned" {
		t.Fatalf("body = %q, want %q", body, "status: planned")
	}
}

func TestExtractPlanPhaseBlockHeadingFormTakesPrecedenceOverListForm(t *testing.T) {
	// The list-form "dup" sits under a real "## " heading so the
	// heading-form block's own boundary (next "### phase_slug:" or next
	// "## ") correctly excludes it — proving heading-form content, not
	// list-form content, is what came back.
	mixed := "## Phases and Verification\n" +
		"### phase_slug: `dup`\n" +
		"- source: heading\n" +
		"\n" +
		"## Progress\n" +
		"- phases:\n" +
		"  - phase_slug: dup\n" +
		"    source: list\n"
	body, ok := extractPlanPhaseBlock(mixed, "dup")
	if !ok {
		t.Fatalf("extractPlanPhaseBlock: not found")
	}
	if !containsAll(body, "source: heading") || containsAll(body, "source: list") {
		t.Fatalf("body = %q, want the heading-form block to take precedence", body)
	}
}

func TestQueryPlanSectionListFormPhaseRoundTrip(t *testing.T) {
	chdirFixture(t)
	writeFile(t, "docs/plans/active/demo.md", planQueryListFixture)

	v, stop, err := QueryPlanSection("phase", "p1-first")
	if err != nil {
		t.Fatalf("QueryPlanSection: %v", err)
	}
	if stop != nil {
		t.Fatalf("QueryPlanSection: unexpected stop %+v", stop)
	}
	if v.Degraded {
		t.Fatalf("QueryPlanSection: unexpectedly degraded on list-form plan")
	}
	if !containsAll(v.Content, "goal: first phase") {
		t.Fatalf("Content = %q", v.Content)
	}
	if containsAll(v.Content, "last phase") {
		t.Fatalf("Content = %q, leaked p2-last's content", v.Content)
	}
}

func containsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}

func TestQueryPlanSectionCurrentStateRoundTrip(t *testing.T) {
	chdirFixture(t)
	writeFile(t, "docs/plans/active/demo.md", planQueryFixture)

	v, stop, err := QueryPlanSection("current-state", "")
	if err != nil {
		t.Fatalf("QueryPlanSection: %v", err)
	}
	if stop != nil {
		t.Fatalf("QueryPlanSection: unexpected stop %+v", stop)
	}
	if v.Degraded {
		t.Fatalf("QueryPlanSection: unexpectedly degraded")
	}
	if v.Content != "- active_phase: none\n- exact_next_action: to-plan" {
		t.Fatalf("Content = %q", v.Content)
	}
	if v.Path != "docs/plans/active/demo.md" {
		t.Fatalf("Path = %q", v.Path)
	}
}

func TestQueryPlanSectionPhaseRoundTrip(t *testing.T) {
	chdirFixture(t)
	writeFile(t, "docs/plans/active/demo.md", planQueryFixture)

	v, stop, err := QueryPlanSection("phase", "p1-first")
	if err != nil {
		t.Fatalf("QueryPlanSection: %v", err)
	}
	if stop != nil {
		t.Fatalf("QueryPlanSection: unexpected stop %+v", stop)
	}
	if v.Degraded {
		t.Fatalf("QueryPlanSection: unexpectedly degraded")
	}
	if !containsAll(v.Content, "goal: first phase") {
		t.Fatalf("Content = %q", v.Content)
	}
}

func TestQueryPlanSectionDegradesOnMissingSection(t *testing.T) {
	chdirFixture(t)
	writeFile(t, "docs/plans/active/demo.md", "# Plan\n\n## Outcome\n- result: demo\n")

	v, stop, err := QueryPlanSection("current-state", "")
	if err != nil {
		t.Fatalf("QueryPlanSection: %v", err)
	}
	if stop != nil {
		t.Fatalf("QueryPlanSection: unexpected stop %+v", stop)
	}
	if !v.Degraded {
		t.Fatalf("QueryPlanSection: want degraded=true for a plan missing the section")
	}
	if v.Content == "" {
		t.Fatalf("QueryPlanSection: degraded response carried no fallback content")
	}
}

func TestQueryPlanSectionDegradesOnMissingPhase(t *testing.T) {
	chdirFixture(t)
	writeFile(t, "docs/plans/active/demo.md", planQueryFixture)

	v, stop, err := QueryPlanSection("phase", "no-such-phase")
	if err != nil {
		t.Fatalf("QueryPlanSection: %v", err)
	}
	if stop != nil {
		t.Fatalf("QueryPlanSection: unexpected stop %+v", stop)
	}
	if !v.Degraded {
		t.Fatalf("QueryPlanSection: want degraded=true for an undefined phase")
	}
}

func TestQueryPlanSectionNoActivePlan(t *testing.T) {
	chdirFixture(t)

	_, stop, err := QueryPlanSection("current-state", "")
	if err != nil {
		t.Fatalf("QueryPlanSection: %v", err)
	}
	if stop == nil || stop.Code != "none" {
		t.Fatalf("stop = %+v, want code=none", stop)
	}
}

func TestQueryPlanSectionAmbiguousActivePlan(t *testing.T) {
	chdirFixture(t)
	writeFile(t, "docs/plans/active/a.md", planQueryFixture)
	writeFile(t, "docs/plans/active/b.md", planQueryFixture)

	_, stop, err := QueryPlanSection("current-state", "")
	if err != nil {
		t.Fatalf("QueryPlanSection: %v", err)
	}
	if stop == nil || stop.Code != "ambiguous" {
		t.Fatalf("stop = %+v, want code=ambiguous", stop)
	}
}

func TestQueryPlanSectionUnknownSection(t *testing.T) {
	chdirFixture(t)
	writeFile(t, "docs/plans/active/demo.md", planQueryFixture)

	_, _, err := QueryPlanSection("bogus", "")
	assertLifecycleValidationError(t, err, "unknown_section",
		`query plan: unknown section "bogus" (want current-state|phase)`)
}

func TestQueryPlanSectionPhaseWithoutSlugIsMissingRequiredField(t *testing.T) {
	chdirFixture(t)
	writeFile(t, "docs/plans/active/demo.md", planQueryFixture)

	_, _, err := QueryPlanSection("phase", "")
	assertLifecycleValidationError(t, err, "missing_required_field",
		"query plan --section phase requires --phase {slug}")
}
