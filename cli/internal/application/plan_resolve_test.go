package application

import (
	"strconv"
	"strings"
	"testing"
)

// planFrontmatter builds a plan whose frontmatter carries `updated:`, so
// the ambiguous ladder's Tier 1 read resolves without falling back to git
// log — bodyLines pads the plan body far past the frontmatter block to
// prove Tier 1/Tier 2 never scale with file size.
func planFrontmatter(updated string, bodyLines int) string {
	var b strings.Builder
	b.WriteString("---\nid: 01TESTPLANRESOLVEXULIDXXX\ntype: plan\nstatus: active\nupdated: " + updated + "\n---\n\n# Plan\n\n## Decisions\n- none\n")
	for i := 0; i < bodyLines; i++ {
		b.WriteString("- filler line " + strconv.Itoa(i) + " padding this plan body well past its frontmatter\n")
	}
	return b.String()
}

// TestResolveActivePlanAmbiguousPacketStaysBounded proves R3's Tier 2
// guarantee: an ambiguous Stop between a 1,621-line and a 410-line plan
// still fits comfortably under 500 tokens (~4 chars/token, so ~2000
// chars), because only the first 10 frontmatter lines per candidate are
// ever read — never the body.
func TestResolveActivePlanAmbiguousPacketStaysBounded(t *testing.T) {
	chdirFixture(t)
	writeFile(t, "docs/plans/active/big.md", planFrontmatter("2026-08-10", 1621))
	writeFile(t, "docs/plans/active/small.md", planFrontmatter("2026-08-15", 410))

	_, stop, err := ResolveActivePlan()
	if err != nil {
		t.Fatalf("ResolveActivePlan: %v", err)
	}
	if stop == nil || stop.Code != "ambiguous" {
		t.Fatalf("stop = %+v, want code=ambiguous", stop)
	}

	const tokenBudget = 500
	const approxCharsPerToken = 4
	if len(stop.Message) > tokenBudget*approxCharsPerToken {
		t.Fatalf("ambiguous message is %d chars (~%d tokens), want under %d tokens:\n%s",
			len(stop.Message), len(stop.Message)/approxCharsPerToken, tokenBudget, stop.Message)
	}
	if !strings.Contains(stop.Message, "big.md") || !strings.Contains(stop.Message, "small.md") {
		t.Fatalf("message = %q, want it to name both candidates", stop.Message)
	}
	if stop.OmittedField == "" {
		t.Fatalf("stop.OmittedField is empty, want it to declare plan bodies were never read")
	}
	if len(stop.Candidates) != 2 {
		t.Fatalf("stop.Candidates has %d entries, want 2", len(stop.Candidates))
	}
	for _, c := range stop.Candidates {
		if !c.FrontmatterOK || c.OrderedBy != "frontmatter_updated" {
			t.Fatalf("candidate %+v, want frontmatter_ok=true ordered_by=frontmatter_updated", c)
		}
	}
	if stop.Candidates[0].Path != "docs/plans/active/small.md" {
		t.Fatalf("candidates = %+v, want small.md (updated 2026-08-15) ordered first", stop.Candidates)
	}
}

// TestResolveActivePlanAmbiguousFallsBackToGitLogOrdering proves R4: a
// candidate with no frontmatter block at all still resolves — ordering
// falls back to git log instead of erroring or reading the body — and is
// marked so a future `validate` finding can report the missing
// frontmatter.
func TestResolveActivePlanAmbiguousFallsBackToGitLogOrdering(t *testing.T) {
	chdirFixture(t)
	writeFile(t, "docs/plans/active/no-frontmatter.md", "# Plan with no frontmatter block\n\nbody text\n")
	writeFile(t, "docs/plans/active/with-frontmatter.md", planFrontmatter("2026-08-15", 5))

	_, stop, err := ResolveActivePlan()
	if err != nil {
		t.Fatalf("ResolveActivePlan: %v", err)
	}
	if stop == nil || stop.Code != "ambiguous" {
		t.Fatalf("stop = %+v, want code=ambiguous", stop)
	}

	var found bool
	for _, c := range stop.Candidates {
		if c.Path != "docs/plans/active/no-frontmatter.md" {
			continue
		}
		found = true
		if c.FrontmatterOK {
			t.Fatalf("candidate %+v, want frontmatter_ok=false", c)
		}
		if c.OrderedBy == "frontmatter_updated" {
			t.Fatalf("candidate %+v, want ordered_by != frontmatter_updated", c)
		}
	}
	if !found {
		t.Fatalf("stop.Candidates = %+v, missing docs/plans/active/no-frontmatter.md", stop.Candidates)
	}
}
