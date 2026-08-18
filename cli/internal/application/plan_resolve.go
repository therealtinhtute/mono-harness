package application

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// StopInfo is one blocking state from work.md's former Full-Mode-Detection
// table: a taxonomy code, a human message naming the specific blocker, and
// the exact recovery command/action. Candidates and OmittedField are set
// only for Code:"ambiguous" — the R3 3-tier disambiguation packet (Tier 1:
// per-candidate frontmatter preview; Tier 2: the bounded packet itself,
// declaring what it refused to read instead of silently truncating).
type StopInfo struct {
	Code         string          `json:"code"`
	Message      string          `json:"message"`
	Recovery     string          `json:"recovery"`
	Candidates   []PlanCandidate `json:"candidates,omitempty"`
	OmittedField string          `json:"omitted,omitempty"`
}

// PlanCandidate is one active-plan file surfaced by an ambiguous Stop.
// Updated/OrderedBy come from R3's Tier 1 read (the frontmatter `updated:`
// line, one of the plan's first 10 frontmatter lines) or, per R4, from the
// candidate's last commit date when frontmatter is missing or unparseable
// — never from reading the plan body.
type PlanCandidate struct {
	Path          string `json:"path"`
	Updated       string `json:"updated,omitempty"`
	OrderedBy     string `json:"ordered_by"`
	FrontmatterOK bool   `json:"frontmatter_ok"`
}

const nextActivePlansGlob = "docs/plans/active/*.md"

type activePlan struct {
	path    string
	content string
}

func findActivePlans() ([]activePlan, error) {
	matches, err := filepath.Glob(nextActivePlansGlob)
	if err != nil {
		return nil, fmt.Errorf("resolve plan: glob active plans: %w", err)
	}
	sort.Strings(matches)

	plans := make([]activePlan, 0, len(matches))
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("resolve plan: read %s: %w", path, err)
		}
		content := string(data)
		if strings.TrimSpace(content) == "" {
			continue
		}
		plans = append(plans, activePlan{path: path, content: content})
	}
	return plans, nil
}

// ResolveActivePlan is the single entry point every caller that needs
// exactly-one-active-plan must use (R2, docs/audit/consumer-adoption-audit.md
// D1). It returns the resolved plan on success, or a Stop describing why
// none could be resolved — never a bare error for the ambiguous/absent
// cases, so no caller can degrade silently or invent its own message.
func ResolveActivePlan() (activePlan, *StopInfo, error) {
	plans, err := findActivePlans()
	if err != nil {
		return activePlan{}, nil, err
	}
	switch len(plans) {
	case 0:
		return activePlan{}, &StopInfo{
			Code:     "none",
			Message:  fmt.Sprintf("docs/plans/active/ contains no non-empty plan: this operation requires exactly one active plan (docs/audit/consumer-adoption-audit.md, D1). Run `brainstorm lock` to create docs/plans/active/{slug}.md."),
			Recovery: "run `brainstorm lock` to create docs/plans/active/{slug}.md",
		}, nil
	case 1:
		return plans[0], nil, nil
	default:
		return activePlan{}, buildAmbiguousStop(plans), nil
	}
}

const frontmatterPreviewLines = 10

// buildAmbiguousStop is R3's 3-tier ladder, Tiers 1 and 2. Tier 0 (index/
// traces already in the agent's own context, ~0 tokens) is not this
// function's job — it costs nothing precisely because it requires no new
// read here. This function reads only each candidate's first 10
// frontmatter lines (Tier 1), never the plan body, and packs the result
// into a bounded packet (Tier 2) that names what it omitted rather than
// silently truncating.
func buildAmbiguousStop(plans []activePlan) *StopInfo {
	candidates := make([]PlanCandidate, len(plans))
	for i, plan := range plans {
		candidates[i] = resolveCandidate(plan)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidateSortKey(candidates[i]).After(candidateSortKey(candidates[j]))
	})

	paths := make([]string, len(candidates))
	summaries := make([]string, len(candidates))
	for i, c := range candidates {
		paths[i] = c.Path
		signal := c.Updated
		if signal == "" {
			signal = "no ordering signal"
		}
		summaries[i] = fmt.Sprintf("%s (%s via %s)", c.Path, signal, c.OrderedBy)
	}

	return &StopInfo{
		Code: "ambiguous",
		Message: fmt.Sprintf(
			"docs/plans/active/ contains %d plans (%s): exactly one active plan may exist (docs/audit/consumer-adoption-audit.md, D1). candidates, newest first: %s. omitted: plan bodies were not read to disambiguate (R3, docs/plans/active/harness-markdown-truth.md). Run `zharness plan complete` or `zharness plan abandon` on all but one.",
			len(plans), strings.Join(paths, ", "), strings.Join(summaries, "; "),
		),
		Recovery:     "run `zharness plan complete` or `zharness plan abandon` on all but one",
		Candidates:   candidates,
		OmittedField: "plan bodies",
	}
}

// resolveCandidate reads plan's first 10 frontmatter lines and its
// `updated:` value (Tier 1). When frontmatter is absent or has no
// `updated:` line, it falls back to the candidate's last commit date (R4)
// so ordering never depends on reading the file body — and marks
// FrontmatterOK so a future `validate` finding can report the gap.
func resolveCandidate(plan activePlan) PlanCandidate {
	lines, frontmatterOK := frontmatterPreview(plan.content, frontmatterPreviewLines)
	candidate := PlanCandidate{Path: plan.path, FrontmatterOK: frontmatterOK}

	if frontmatterOK {
		if updated, ok := frontmatterPreviewField(lines, "updated"); ok {
			candidate.Updated = updated
			candidate.OrderedBy = "frontmatter_updated"
			return candidate
		}
	}
	if commitTime, err := gitLogCommitTime(plan.path); err == nil && commitTime != "" {
		candidate.Updated = commitTime
		candidate.OrderedBy = "git_log_fallback"
		return candidate
	}
	candidate.OrderedBy = "unordered"
	return candidate
}

// frontmatterPreview returns up to limit lines strictly between the first
// two `---` delimiters at the top of content. ok is false when content
// does not open with a `---` frontmatter block or the block never closes
// — an unparseable-frontmatter candidate, per R4.
func frontmatterPreview(content string, limit int) (lines []string, ok bool) {
	all := strings.Split(content, "\n")
	if len(all) == 0 || strings.TrimSpace(all[0]) != "---" {
		return nil, false
	}
	end := -1
	for i := 1; i < len(all); i++ {
		if strings.TrimSpace(all[i]) == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return nil, false
	}
	body := all[1:end]
	if len(body) > limit {
		body = body[:limit]
	}
	return body, true
}

// frontmatterPreviewField looks up a `key: value` line within an
// already-bounded frontmatter preview (frontmatterPreview's output).
func frontmatterPreviewField(lines []string, key string) (value string, ok bool) {
	prefix := key + ":"
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix)), true
		}
	}
	return "", false
}

// gitLogCommitTime returns path's last commit time (R4's fallback
// ordering signal) via `git log -1 --format=%cI -- <path>` — commit
// metadata only, never the file body.
func gitLogCommitTime(path string) (string, error) {
	out, err := exec.Command("git", "log", "-1", "--format=%cI", "--", path).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// candidateSortKey parses a candidate's Updated string (RFC3339 from git
// log, or YYYY-MM-DD from frontmatter) for the newest-first sort. An
// unparseable or empty value sorts last.
func candidateSortKey(c PlanCandidate) time.Time {
	if c.Updated == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, c.Updated); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02", c.Updated); err == nil {
		return t
	}
	return time.Time{}
}
