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

	v, err := QueryPlanSection("current-state", "")
	if err != nil {
		t.Fatalf("QueryPlanSection: %v", err)
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

	v, err := QueryPlanSection("phase", "p1-first")
	if err != nil {
		t.Fatalf("QueryPlanSection: %v", err)
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

	v, err := QueryPlanSection("current-state", "")
	if err != nil {
		t.Fatalf("QueryPlanSection: %v", err)
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

	v, err := QueryPlanSection("phase", "no-such-phase")
	if err != nil {
		t.Fatalf("QueryPlanSection: %v", err)
	}
	if !v.Degraded {
		t.Fatalf("QueryPlanSection: want degraded=true for an undefined phase")
	}
}

func TestQueryPlanSectionNoActivePlan(t *testing.T) {
	chdirFixture(t)

	_, err := QueryPlanSection("current-state", "")
	assertLifecycleValidationError(t, err, "no_active_plan",
		"query plan: no non-empty plan under docs/plans/active/*.md")
}

func TestQueryPlanSectionAmbiguousActivePlan(t *testing.T) {
	chdirFixture(t)
	writeFile(t, "docs/plans/active/a.md", planQueryFixture)
	writeFile(t, "docs/plans/active/b.md", planQueryFixture)

	_, err := QueryPlanSection("current-state", "")
	assertLifecycleValidationError(t, err, "ambiguous_active_plan",
		"query plan: 2 active plans found; this command requires exactly one")
}

func TestQueryPlanSectionUnknownSection(t *testing.T) {
	chdirFixture(t)
	writeFile(t, "docs/plans/active/demo.md", planQueryFixture)

	_, err := QueryPlanSection("bogus", "")
	assertLifecycleValidationError(t, err, "unknown_section",
		`query plan: unknown section "bogus" (want current-state|phase)`)
}

func TestQueryPlanSectionPhaseWithoutSlugIsMissingRequiredField(t *testing.T) {
	chdirFixture(t)
	writeFile(t, "docs/plans/active/demo.md", planQueryFixture)

	_, err := QueryPlanSection("phase", "")
	assertLifecycleValidationError(t, err, "missing_required_field",
		"query plan --section phase requires --phase {slug}")
}
