package application

import (
	"strings"
	"testing"
)

const scaffoldedProgressSection = `# Plan: Demo

## Outcome
- result: demo

## Progress
<!-- Append-only durable entries record timestamp, phase, wave, task, task_status,
run_id, trace_id, exact verification/result, and changed surfaces or blocker. -->
- none

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- none

## Current State and Next Action
- active_phase: none
`

func TestAppendToPlanSectionReplacesLoneNonePlaceholder(t *testing.T) {
	got, err := AppendToPlanSection(scaffoldedProgressSection, "Progress", "- 2026-08-08, wave 1 done")
	if err != nil {
		t.Fatalf("AppendToPlanSection: %v", err)
	}
	want := `# Plan: Demo

## Outcome
- result: demo

## Progress
<!-- Append-only durable entries record timestamp, phase, wave, task, task_status,
run_id, trace_id, exact verification/result, and changed surfaces or blocker. -->
- 2026-08-08, wave 1 done

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- none

## Current State and Next Action
- active_phase: none
`
	if got != want {
		t.Fatalf("AppendToPlanSection =\n%s\nwant\n%s", got, want)
	}
}

func TestAppendToPlanSectionAppendsAfterExistingContent(t *testing.T) {
	content := `## Progress
- 2026-08-08T00:00Z, wave 1 done

## Decisions
- none
`
	got, err := AppendToPlanSection(content, "Progress", "- 2026-08-08T01:00Z, wave 2 done")
	if err != nil {
		t.Fatalf("AppendToPlanSection: %v", err)
	}
	want := `## Progress
- 2026-08-08T00:00Z, wave 1 done
- 2026-08-08T01:00Z, wave 2 done

## Decisions
- none
`
	if got != want {
		t.Fatalf("AppendToPlanSection =\n%s\nwant\n%s", got, want)
	}
}

func TestAppendToPlanSectionAppendsMultiLineEntry(t *testing.T) {
	content := `## Decisions
- none

## Validation
- none
`
	entry := "- **D1 — title.** Affected phase/task: p1.\n" +
		"  - Discovered: something.\n" +
		"  - Rationale: because.\n" +
		"  - Result: fixed."
	got, err := AppendToPlanSection(content, "Decisions", entry)
	if err != nil {
		t.Fatalf("AppendToPlanSection: %v", err)
	}
	want := `## Decisions
- **D1 — title.** Affected phase/task: p1.
  - Discovered: something.
  - Rationale: because.
  - Result: fixed.

## Validation
- none
`
	if got != want {
		t.Fatalf("AppendToPlanSection =\n%s\nwant\n%s", got, want)
	}
}

func TestAppendToPlanSectionMissingSectionErrorsWithoutMutation(t *testing.T) {
	content := "## Progress\n- none\n"
	got, err := AppendToPlanSection(content, "Decisions", "- new entry")
	if err == nil {
		t.Fatal("AppendToPlanSection: err = nil, want a not-found error")
	}
	if got != "" {
		t.Fatalf("AppendToPlanSection on error returned %q, want empty (caller must not write it)", got)
	}
	if err.Error() != "plan section not found: ## Decisions" {
		t.Fatalf("error = %q, want %q", err.Error(), "plan section not found: ## Decisions")
	}
}

func TestAppendToPlanSectionReorderedSections(t *testing.T) {
	// Decisions appears BEFORE Progress here — a hand-edited plan may not
	// follow the canonical section order (V3's own premise).
	content := `## Decisions
- none

## Progress
- 2026-08-08, first entry
`
	got, err := AppendToPlanSection(content, "Decisions", "- new decision")
	if err != nil {
		t.Fatalf("AppendToPlanSection: %v", err)
	}
	want := `## Decisions
- new decision

## Progress
- 2026-08-08, first entry
`
	if got != want {
		t.Fatalf("AppendToPlanSection =\n%s\nwant\n%s", got, want)
	}
}

func TestAppendToPlanSectionTrailingContentAfterLastHeading(t *testing.T) {
	content := `## Progress
- none

## Current State and Next Action
- active_phase: none
- exact_next_action: start wave 1
`
	got, err := AppendToPlanSection(content, "Current State and Next Action", "- exact_next_action: start wave 2")
	if err != nil {
		t.Fatalf("AppendToPlanSection: %v", err)
	}
	want := `## Progress
- none

## Current State and Next Action
- active_phase: none
- exact_next_action: start wave 1
- exact_next_action: start wave 2
`
	if got != want {
		t.Fatalf("AppendToPlanSection =\n%s\nwant\n%s", got, want)
	}
}

func TestAppendToPlanSectionLastHeadingNoTrailingNewline(t *testing.T) {
	content := "## Progress\n- none\n\n## Decisions\n- existing decision"
	got, err := AppendToPlanSection(content, "Decisions", "- new decision")
	if err != nil {
		t.Fatalf("AppendToPlanSection: %v", err)
	}
	want := "## Progress\n- none\n\n## Decisions\n- existing decision\n- new decision"
	if got != want {
		t.Fatalf("AppendToPlanSection =\n%q\nwant\n%q", got, want)
	}
}

func TestAppendToPlanSectionPreservesCRLF(t *testing.T) {
	content := "## Progress\r\n- none\r\n\r\n## Decisions\r\n- none\r\n"
	got, err := AppendToPlanSection(content, "Progress", "- 2026-08-08, wave 1 done")
	if err != nil {
		t.Fatalf("AppendToPlanSection: %v", err)
	}
	want := "## Progress\r\n- 2026-08-08, wave 1 done\r\n\r\n## Decisions\r\n- none\r\n"
	if got != want {
		t.Fatalf("AppendToPlanSection =\n%q\nwant\n%q", got, want)
	}
}

func TestAppendToPlanSectionCRLFEntryNormalizedToMatchFile(t *testing.T) {
	content := "## Progress\r\n- none\r\n"
	// Caller supplies an entry with CRLF line endings (e.g. built from a
	// Windows-authored string); the file's own CRLF style still governs
	// the single, consistent output.
	got, err := AppendToPlanSection(content, "Progress", "- line one\r\n- line two")
	if err != nil {
		t.Fatalf("AppendToPlanSection: %v", err)
	}
	want := "## Progress\r\n- line one\r\n- line two\r\n"
	if got != want {
		t.Fatalf("AppendToPlanSection =\n%q\nwant\n%q", got, want)
	}
}

func TestAppendToPlanSectionSkipsHTMLCommentWhenDetectingPlaceholder(t *testing.T) {
	content := `## Progress
<!-- multi-line
comment block -->
- none

## Decisions
- none
`
	got, err := AppendToPlanSection(content, "Progress", "- real entry")
	if err != nil {
		t.Fatalf("AppendToPlanSection: %v", err)
	}
	want := `## Progress
<!-- multi-line
comment block -->
- real entry

## Decisions
- none
`
	if got != want {
		t.Fatalf("AppendToPlanSection =\n%s\nwant\n%s", got, want)
	}
}

func TestAppendToPlanSectionDoesNotTreatCommentAsSubstantiveContent(t *testing.T) {
	// A section whose ONLY content is a comment (no "- none" placeholder
	// at all) must still accept an append, inserted after the comment.
	content := `## Progress
<!-- append-only notes -->

## Decisions
- none
`
	got, err := AppendToPlanSection(content, "Progress", "- first real entry")
	if err != nil {
		t.Fatalf("AppendToPlanSection: %v", err)
	}
	want := `## Progress
<!-- append-only notes -->
- first real entry

## Decisions
- none
`
	if got != want {
		t.Fatalf("AppendToPlanSection =\n%s\nwant\n%s", got, want)
	}
}

func TestAppendToPlanSectionDoesNotDropSubHeadings(t *testing.T) {
	// docs/plans/completed/*.md's actual phase blocks nest ### headings
	// under ## Phases and Verification — the appender must treat ### as
	// section content, not a section boundary.
	content := `## Phases and Verification
### Phase 1: Alpha
- phase_slug: alpha

## Progress
- none
`
	got, err := AppendToPlanSection(content, "Phases and Verification", "### Phase 2: Beta\n- phase_slug: beta")
	if err != nil {
		t.Fatalf("AppendToPlanSection: %v", err)
	}
	want := `## Phases and Verification
### Phase 1: Alpha
- phase_slug: alpha
### Phase 2: Beta
- phase_slug: beta

## Progress
- none
`
	if got != want {
		t.Fatalf("AppendToPlanSection =\n%s\nwant\n%s", got, want)
	}
}

// TestAppendToPlanSectionRealHandWrittenPlanIsNotCorrupted feeds a real
// excerpt of docs/plans/completed/eval-layer.md's ## Progress section
// (hand-written, not scaffolded) through the appender and proves every
// original byte survives intact, with the new entry appended after it —
// the plan's own "hand-written plans are the primary case" requirement.
func TestAppendToPlanSectionRealHandWrittenPlanIsNotCorrupted(t *testing.T) {
	content := "## Progress\n\n" +
		"- `2026-07-30T09:12Z` — phase `link-integrity` started. wave: —. task: phase-start. task_status: `in-progress`. run: `01KYS4NST8ACHAJAC9S5V12PBF`. changed surfaces: none yet. verification: `zharness query phases --json` → `link-integrity: in-progress`.\n" +
		"- `2026-07-30T09:16Z` — wave 1, task T1.1. task_status: `DONE`. run: `01KYS4NST8ACHAJAC9S5V12PBF`. changed surfaces: `scripts/verify-doc-links.sh` (new). verification: `bash scripts/verify-doc-links.sh; echo \"exit=$?\"` → 11 findings, `exit=1`.\n\n" +
		"## Decisions\n\n" +
		"- **D1 — title.** Affected phase/task: link-integrity / T1.1.\n" +
		"  - Discovered during T1.3: the first real run returned 25 findings.\n"

	newEntry := "- `2026-08-08T00:00Z` — new wave complete."
	got, err := AppendToPlanSection(content, "Progress", newEntry)
	if err != nil {
		t.Fatalf("AppendToPlanSection: %v", err)
	}
	if !containsOrderedSubstrings(got,
		"phase `link-integrity` started",
		"wave 1, task T1.1",
		newEntry,
		"## Decisions",
		"D1 — title",
		"Discovered during T1.3",
	) {
		t.Fatalf("AppendToPlanSection lost or reordered original content:\n%s", got)
	}
}

func containsOrderedSubstrings(s string, parts ...string) bool {
	pos := 0
	for _, p := range parts {
		idx := strings.Index(s[pos:], p)
		if idx == -1 {
			return false
		}
		pos += idx + len(p)
	}
	return true
}
