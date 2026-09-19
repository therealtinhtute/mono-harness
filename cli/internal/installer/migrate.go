package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const activePlansDir = "docs/plans/active"

// legacyPlanSections is the 9-section plan format that predates the 5-section
// one. A plan is migrated only when its `## ` headings are exactly this set.
var legacyPlanSections = []string{
	"Outcome", "Authority and Requirements", "Non-goals", "Approach and Risks",
	"Phases and Verification", "Progress", "Decisions", "Validation",
	"Current State and Next Action",
}

var (
	legacyIDLine     = regexp.MustCompile(`^(intake_id|story_id):`)
	legacyStoryLine  = regexp.MustCompile(`^\s*- story_id:`)
	legacyStatusLine = regexp.MustCompile(`(?m)^([ \t]*)- (status:.*)$`)
)

// MigratePlan rewrites a 9-section plan into the 5-section format by merging
// sections: Outcome, Authority and Requirements and Non-goals become Goal;
// Approach and Risks opens Phases and Verification; Progress and Decisions
// become Log. It drops intake_id and story_id lines and turns each phase's
// `- status:` bullet into the bare `status:` line the completed-plan guard
// reads. Any other heading set is returned unchanged. The Validation body is
// copied verbatim and checked, so a migration can never alter proof evidence.
func MigratePlan(in []byte) ([]byte, bool, error) {
	front, preamble, sections, ok := splitPlan(string(in))
	if !ok || len(sections) != len(legacyPlanSections) {
		return in, false, nil
	}
	for _, name := range legacyPlanSections {
		if _, found := sections[name]; !found {
			return in, false, nil
		}
	}

	var fm []string
	for _, l := range strings.SplitAfter(front, "\n") {
		if !legacyIDLine.MatchString(l) {
			fm = append(fm, l)
		}
	}
	var phases []string
	for _, l := range strings.SplitAfter(sections["Phases and Verification"], "\n") {
		if legacyStoryLine.MatchString(l) {
			continue
		}
		phases = append(phases, legacyStatusLine.ReplaceAllString(l, "$1$2"))
	}

	var b strings.Builder
	b.WriteString(strings.Join(fm, ""))
	b.WriteString(preamble)
	b.WriteString("## Goal\n" + sections["Outcome"])
	b.WriteString("### Authority and Requirements\n" + sections["Authority and Requirements"])
	b.WriteString("### Non-goals\n" + sections["Non-goals"])
	b.WriteString("## Phases and Verification\n" + sections["Approach and Risks"] + strings.Join(phases, ""))
	b.WriteString("## Log\n" + sections["Progress"])
	b.WriteString("### Decisions\n" + sections["Decisions"])
	b.WriteString("## Validation\n" + sections["Validation"])
	b.WriteString("## Current State and Next Action\n" + sections["Current State and Next Action"])
	out := []byte(b.String())

	_, _, after, _ := splitPlan(string(out))
	if after["Validation"] != sections["Validation"] {
		return in, false, fmt.Errorf("plan migration would change ## Validation")
	}
	return out, true, nil
}

// splitPlan separates optional frontmatter, the text before the first `## `
// heading, and each `## ` section body keyed by heading text. Every body ends
// with a newline so sections can be concatenated in any order. A repeated
// heading reports !ok.
func splitPlan(s string) (front, preamble string, sections map[string]string, ok bool) {
	if strings.HasPrefix(s, "---\n") {
		if end := strings.Index(s[4:], "\n---\n"); end >= 0 {
			front, s = s[:4+end+5], s[4+end+5:]
		}
	}
	sections = map[string]string{}
	var cur string
	var body strings.Builder
	inSection := false
	flush := func() bool {
		if !inSection {
			preamble = body.String()
		} else {
			if _, dup := sections[cur]; dup {
				return false
			}
			t := body.String()
			if !strings.HasSuffix(t, "\n") {
				t += "\n"
			}
			sections[cur] = t
		}
		body.Reset()
		return true
	}
	for _, l := range strings.SplitAfter(s, "\n") {
		if h, isHeading := strings.CutPrefix(strings.TrimSuffix(l, "\n"), "## "); isHeading {
			if !flush() {
				return "", "", nil, false
			}
			cur, inSection = h, true
			continue
		}
		body.WriteString(l)
	}
	if !flush() {
		return "", "", nil, false
	}
	return front, preamble, sections, true
}

// planMigration is the pending write, if any, for the one active plan.
type planMigration struct {
	path   string
	data   []byte
	notice string
}

// preparePlanMigration decides what update does to the active plan before
// anything is written: migrate it, leave it with a notice, or refuse.
func preparePlanMigration(root string) (planMigration, error) {
	entries, err := os.ReadDir(filepath.Join(root, activePlansDir))
	if err != nil {
		return planMigration{}, nil
	}
	var plans []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		plans = append(plans, filepath.ToSlash(filepath.Join(activePlansDir, e.Name())))
	}
	sort.Strings(plans)
	switch len(plans) {
	case 0:
		return planMigration{}, nil
	case 1:
	default:
		return planMigration{notice: fmt.Sprintf("%d active plans; none migrated", len(plans))}, nil
	}
	in, err := os.ReadFile(filepath.Join(root, plans[0]))
	if err != nil {
		return planMigration{}, err
	}
	out, changed, err := MigratePlan(in)
	if err != nil {
		return planMigration{}, fmt.Errorf("%s: %w", plans[0], err)
	}
	if changed {
		return planMigration{path: plans[0], data: out}, nil
	}
	if _, _, secs, ok := splitPlan(string(in)); !ok || !isSlimPlan(secs) {
		return planMigration{notice: plans[0] + " has an unrecognized section set; not migrated"}, nil
	}
	return planMigration{}, nil
}

func isSlimPlan(secs map[string]string) bool {
	for _, h := range []string{"Goal", "Phases and Verification", "Log", "Validation", "Current State and Next Action"} {
		if _, ok := secs[h]; !ok {
			return false
		}
	}
	return len(secs) == 5
}
