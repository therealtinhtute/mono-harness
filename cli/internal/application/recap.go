package application

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// RecapFacts is the git/WIP/judgment data watzup gathers itself (branch
// position, uncommitted counts, committed change themes, WIP items, risk
// rows, and the one recommended next action) and hands to RenderRecap.
// RenderRecap never runs git — it only formats what it's given.
type RecapFacts struct {
	Branch           string      `json:"branch"`
	Ahead            int         `json:"ahead"`
	Behind           int         `json:"behind"`
	UncommittedFiles int         `json:"uncommitted_files"`
	UncommittedAdds  int         `json:"uncommitted_adds"`
	UncommittedDels  int         `json:"uncommitted_dels"`
	HandoffSummary   string      `json:"handoff_summary"`
	Changes          []string    `json:"changes"`
	WIP              []string    `json:"wip"`
	Risks            []RecapRisk `json:"risks"`
	NextAction       string      `json:"next_action"`
}

// RecapRisk is one row of the Recap's Risk table.
type RecapRisk struct {
	Risk     string `json:"risk"`
	Severity string `json:"severity"`
	Action   string `json:"action"`
}

// recapForbiddenSubstrings ports watzup.md's former Output Contract
// Section 1 (Forbidden Phrases) verbatim — enforced here instead of an
// agent self-check pass.
var recapForbiddenSubstrings = []string{
	"git log", "git diff", "git status", "git branch", "git show",
	"--stat", "--shortstat", "--oneline", "HEAD~", "..HEAD", "--graph", "--decorate",
	"commit window", "diff stat", "last 10 commits", "last 50 commits", "analyzed", "scanned",
	"recap mode", "orient mode",
	"git ",
}

var recapScorePattern = regexp.MustCompile(`\d+\s*/\s*10\b`)

var recapValidSeverities = map[string]bool{"cao": true, "vừa": true, "thấp": true}

type recapTextField struct{ field, value string }

// RenderRecap builds watzup's Vietnamese Recap text deterministically from
// the harness-derived ResumeView plus agent-gathered RecapFacts.
// Forbidden-phrase safety, the risk-table shape, the severity ladder, and
// the drifted-state recovery override are enforced here (return a
// *domain.ValidationError on violation) instead of relying on an agent
// self-check paragraph.
func RenderRecap(view ResumeView, facts RecapFacts) (string, error) {
	if err := validateRecapFacts(facts); err != nil {
		return "", err
	}

	if isEmptyState(view, facts) {
		return "Nhánh sạch — không có thay đổi nào so với main.\n" +
			"Next: Bắt đầu task mới hoặc kéo thay đổi mới nhất.\n", nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Recap — %s (%s)\n\n", facts.Branch, time.Now().Format("2006-01-02"))

	b.WriteString("Trạng thái\n")
	fmt.Fprintf(&b, "- Nhánh: %s, %s\n", facts.Branch, recapPositionSummary(facts.Ahead, facts.Behind))
	if facts.UncommittedFiles > 0 {
		fmt.Fprintf(&b, "- Uncommitted: %d files, +%d/-%d lines\n", facts.UncommittedFiles, facts.UncommittedAdds, facts.UncommittedDels)
	}
	fmt.Fprintf(&b, "- Readiness: %s\n", view.Readiness)
	b.WriteString("\n")

	if lines := recapContextLines(view, facts); len(lines) > 0 {
		b.WriteString("Context\n")
		for _, l := range lines {
			fmt.Fprintf(&b, "- %s\n", l)
		}
		b.WriteString("\n")
	}

	if lines := recapChangeLines(facts); len(lines) > 0 {
		b.WriteString("Thay đổi\n")
		for _, l := range lines {
			fmt.Fprintf(&b, "- %s\n", l)
		}
		b.WriteString("\n")
	}

	if len(facts.Risks) > 0 {
		b.WriteString("Risks\n")
		b.WriteString("| Risk | Mức độ | Action |\n")
		b.WriteString("|------|--------|--------|\n")
		for _, r := range facts.Risks {
			fmt.Fprintf(&b, "| %s | %s | %s |\n", r.Risk, r.Severity, r.Action)
		}
		b.WriteString("\n")
	}

	next := facts.NextAction
	if view.Readiness == "drifted" && len(view.Drift) > 0 {
		next = view.Drift[0].Recovery
	}
	fmt.Fprintf(&b, "Next: %s\n", next)

	return b.String(), nil
}

func recapPositionSummary(ahead, behind int) string {
	switch {
	case ahead > 0 && behind > 0:
		return fmt.Sprintf("%d commits ahead, %d behind main", ahead, behind)
	case ahead > 0:
		return fmt.Sprintf("%d commits ahead of main", ahead)
	case behind > 0:
		return fmt.Sprintf("%d commits behind main", behind)
	default:
		return "up to date with main"
	}
}

func recapContextLines(view ResumeView, facts RecapFacts) []string {
	lines := []string{}
	if facts.HandoffSummary != "" {
		lines = append(lines, "Handoff: "+facts.HandoffSummary)
	} else {
		lines = append(lines, "Không có handoff")
	}
	if view.Position.CurrentPhase != nil {
		runStatus, checkStatus := "chưa có", "chưa có"
		if view.LatestRunID != nil {
			runStatus = "recorded"
		}
		if view.LatestCheckID != nil {
			checkStatus = "recorded"
		}
		extra := ""
		if view.Readiness == "drifted" && len(view.Drift) > 0 {
			extra = fmt.Sprintf(" | drift: %s", view.Drift[0].Type)
		}
		lines = append(lines, fmt.Sprintf("Phase: %s | latest run: %s | latest check: %s%s", *view.Position.CurrentPhase, runStatus, checkStatus, extra))
	}
	return lines
}

// recapChangeLines caps the combined committed+WIP list at 5 items — a
// defensive backstop for the 25-line output target; the agent is expected
// to already cap themes/WIP before calling.
func recapChangeLines(facts RecapFacts) []string {
	var lines []string
	lines = append(lines, facts.Changes...)
	for _, w := range facts.WIP {
		lines = append(lines, "[WIP] "+w)
	}
	if len(lines) > 5 {
		lines = lines[:5]
	}
	return lines
}

func isEmptyState(view ResumeView, facts RecapFacts) bool {
	return facts.Ahead == 0 && facts.Behind == 0 && facts.UncommittedFiles == 0 &&
		facts.HandoffSummary == "" && len(facts.Changes) == 0 && len(facts.WIP) == 0 &&
		view.Readiness == "clean"
}

func validateRecapFacts(facts RecapFacts) error {
	for i, r := range facts.Risks {
		if !recapValidSeverities[r.Severity] {
			return &domain.ValidationError{
				Code:    "invalid_severity",
				Message: fmt.Sprintf("resume: risks[%d].severity %q must be one of cao|vừa|thấp", i, r.Severity),
			}
		}
	}

	texts := []recapTextField{
		{"branch", facts.Branch},
		{"handoff_summary", facts.HandoffSummary},
		{"next_action", facts.NextAction},
	}
	for i, c := range facts.Changes {
		texts = append(texts, recapTextField{fmt.Sprintf("changes[%d]", i), c})
	}
	for i, w := range facts.WIP {
		texts = append(texts, recapTextField{fmt.Sprintf("wip[%d]", i), w})
	}
	for i, r := range facts.Risks {
		texts = append(texts, recapTextField{fmt.Sprintf("risks[%d].risk", i), r.Risk})
		texts = append(texts, recapTextField{fmt.Sprintf("risks[%d].action", i), r.Action})
	}

	for _, t := range texts {
		if phrase, found := findForbiddenPhrase(t.value); found {
			return &domain.ValidationError{
				Code:    "forbidden_phrase",
				Message: fmt.Sprintf("resume: facts.%s contains forbidden phrase %q", t.field, phrase),
			}
		}
	}
	return nil
}

func findForbiddenPhrase(s string) (string, bool) {
	for _, f := range recapForbiddenSubstrings {
		if strings.Contains(s, f) {
			return f, true
		}
	}
	if m := recapScorePattern.FindString(s); m != "" {
		return m, true
	}
	return "", false
}
