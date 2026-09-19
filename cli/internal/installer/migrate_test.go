package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const legacyValidation = `- 2026-09-19T08:29Z — phase ` + "`p1`" + ` — verdict: APPROVED — mode: gate
  - ` + "`true`" + ` — ok
  - judge: independent
`

func legacyPlan(order []string) string {
	body := map[string]string{
		"Outcome":                       "- result: x\n\n",
		"Authority and Requirements":    "- R1: y\n\n",
		"Non-goals":                     "- NG1: z\n\n",
		"Approach and Risks":            "- approach: a\n\n",
		"Phases and Verification":       "- phases:\n  - phase_slug: `p1`\n    - story_id: `p1-1`\n    - status: checked\n    - goal: R1\n\n",
		"Progress":                      "- 2026-09-19T08:00Z — p1 — task_status=DONE — ok\n\n",
		"Decisions":                     "- 2026-09-19 — p1 — chose a\n\n",
		"Validation":                    legacyValidation + "\n",
		"Current State and Next Action": "- active_phase: p1\n- lifecycle_status: checked\n",
	}
	var b strings.Builder
	b.WriteString("---\nid: p-1\nintake_id: i-1\nlane: normal\nstatus: active\n---\n\n# Plan: p\n\n")
	for _, h := range order {
		b.WriteString("## " + h + "\n" + body[h])
	}
	return b.String()
}

func headingsOf(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if strings.HasPrefix(l, "## ") {
			out = append(out, l)
		}
	}
	return out
}

func TestMigratePlan_LegacyToFiveSections(t *testing.T) {
	for name, order := range map[string][]string{
		"canonical order": legacyPlanSections,
		"current state before validation": {"Outcome", "Authority and Requirements", "Non-goals", "Approach and Risks",
			"Phases and Verification", "Current State and Next Action", "Validation", "Decisions", "Progress"},
	} {
		t.Run(name, func(t *testing.T) {
			in := legacyPlan(order)
			out, changed, err := MigratePlan([]byte(in))
			if err != nil || !changed {
				t.Fatalf("MigratePlan = changed %v, err %v", changed, err)
			}
			got := string(out)
			want := []string{"## Goal", "## Phases and Verification", "## Log", "## Validation", "## Current State and Next Action"}
			if h := headingsOf(got); strings.Join(h, "|") != strings.Join(want, "|") {
				t.Fatalf("headings = %v", h)
			}
			for _, s := range []string{"### Authority and Requirements\n- R1: y", "### Non-goals\n", "### Decisions\n- 2026-09-19 — p1 — chose a",
				"## Phases and Verification\n- approach: a\n\n- phases:", "  - phase_slug: `p1`\n    status: checked\n", "## Validation\n" + legacyValidation} {
				if !strings.Contains(got, s) {
					t.Errorf("missing %q in:\n%s", s, got)
				}
			}
			for _, s := range []string{"intake_id", "story_id", "- status:"} {
				if strings.Contains(got, s) {
					t.Errorf("still contains %q", s)
				}
			}
			if !strings.HasPrefix(got, "---\nid: p-1\nlane: normal\nstatus: active\n---\n\n# Plan: p\n\n## Goal\n") {
				t.Errorf("frontmatter or preamble changed:\n%s", got)
			}
			_, _, before, _ := splitPlan(in)
			_, _, after, _ := splitPlan(got)
			if sha([]byte(before["Validation"])) != sha([]byte(after["Validation"])) {
				t.Fatal("Validation bytes changed")
			}

			again, changed, err := MigratePlan(out)
			if err != nil || changed || string(again) != got {
				t.Fatalf("second run: changed %v, err %v", changed, err)
			}
		})
	}
}

func TestMigratePlan_UnknownSetsUntouched(t *testing.T) {
	short := legacyPlanSections[:8]
	dup := append(append([]string{}, legacyPlanSections...), "Progress")
	for name, in := range map[string]string{
		"missing section":   legacyPlan(short),
		"extra heading":     legacyPlan(legacyPlanSections) + "## Notes\nx\n",
		"duplicate heading": legacyPlan(dup),
		"renamed heading":   strings.Replace(legacyPlan(legacyPlanSections), "## Current State and Next Action", "## Current State", 1),
		"no headings":       "# plan a",
	} {
		t.Run(name, func(t *testing.T) {
			out, changed, err := MigratePlan([]byte(in))
			if err != nil || changed || string(out) != in {
				t.Fatalf("MigratePlan changed an unknown plan: changed %v, err %v", changed, err)
			}
		})
	}
}

func TestUpdate_MigratesActivePlan(t *testing.T) {
	root := tempRepo(t, true)
	mustInstall(t, root)
	plan := "docs/plans/active/p.md"
	pf(t, root, plan, legacyPlan(legacyPlanSections))

	var sb strings.Builder
	if err := RunUpdate(updateOptions{Root: root, Version: "test"}, &sb); err != nil {
		t.Fatalf("update: %v\n%s", err, sb.String())
	}
	if !strings.Contains(sb.String(), "migrated       "+plan) {
		t.Fatalf("no migrated line:\n%s", sb.String())
	}
	b, err := os.ReadFile(filepath.Join(root, plan))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "## Log\n") {
		t.Fatalf("plan not migrated:\n%s", b)
	}

	sb.Reset()
	if err := RunUpdate(updateOptions{Root: root, Version: "test"}, &sb); err != nil {
		t.Fatalf("second update: %v\n%s", err, sb.String())
	}
	if strings.Contains(sb.String(), "migrated") || strings.Contains(sb.String(), "notice") {
		t.Fatalf("second update touched the plan:\n%s", sb.String())
	}
}

func TestUpdate_UnknownPlanLeftWithNotice(t *testing.T) {
	root := tempRepo(t, true)
	mustInstall(t, root)
	plan := "docs/plans/active/p.md"
	in := legacyPlan(legacyPlanSections) + "## Notes\nx\n"
	pf(t, root, plan, in)

	var sb strings.Builder
	if err := RunUpdate(updateOptions{Root: root, Version: "test"}, &sb); err != nil {
		t.Fatalf("update: %v\n%s", err, sb.String())
	}
	if !strings.Contains(sb.String(), "notice     "+plan+" has an unrecognized section set") {
		t.Fatalf("no notice:\n%s", sb.String())
	}
	if b, _ := os.ReadFile(filepath.Join(root, plan)); string(b) != in {
		t.Fatal("unknown plan was modified")
	}
}
