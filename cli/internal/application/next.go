package application

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// NextView mirrors `zharness next`'s shape: the resolved execution mode,
// the selected phase (full mode only, once resolved), and an optional
// stop naming a blocker + its recovery.
type NextView struct {
	Mode        string    `json:"mode"`
	ActivePhase *string   `json:"active_phase"`
	Stop        *StopInfo `json:"stop,omitempty"`
}

var placeholderMarkers = []string{"TBD", "TODO", "similar to", "implement later"}

var activePlanPhaseSlug = regexp.MustCompile(`(?m)^[ \t]*-[ \t]+phase_slug:[ \t]*([^ \t\r\n]+)[ \t]*$`)

// Next resolves work.md's former Mode-Resolution + Full-Mode-Detection
// tables in Go. Two rows are deliberately NOT encoded here and stay
// agent-side (named in the stop taxonomy's absence, not silently
// dropped):
//   - contract-drift: needs a working-tree diff against the phase's
//     Allowed/Forbidden Surfaces — this CLI has no git access, the same
//     constraint `resume`'s git/WIP block already carries.
//   - stale-plan (file/symbol no longer exists): an active plan legitimately
//     references files it will create by being executed, which don't exist
//     yet — a naive existence check false-positives on almost every real plan.
//     Left to the manual-check review pass instead.
func Next(db *sql.DB, argument string) (NextView, error) {
	mode, explicitPhase, ok := parseNextArgument(argument)
	if !ok {
		return NextView{}, &domain.ValidationError{
			Code:    "unknown_argument",
			Message: fmt.Sprintf("next: unrecognized argument %q (want: (none) | auto | simple [@file] | full | full phase <slug> | phase <slug>)", argument),
		}
	}

	if mode == "simple" {
		return NextView{Mode: "simple"}, nil
	}

	plan, stop, err := ResolveActivePlan()
	if err != nil {
		return NextView{}, err
	}
	if stop != nil {
		if stop.Code == "ambiguous" {
			return NextView{Mode: mode, Stop: stop}, nil
		}
		// stop.Code == "none": auto mode legitimately falls back to
		// simple with no active plan; full mode keeps next's own
		// historical no-plan code and message.
		if mode == "auto" {
			return NextView{Mode: "simple"}, nil
		}
		return NextView{Mode: "full", Stop: &StopInfo{
			Code:     "no-plan",
			Message:  "No non-empty active initiative plan exists under docs/plans/active/.",
			Recovery: "run `brainstorm lock` to create docs/plans/active/{slug}.md",
		}}, nil
	}

	return resolveFullMode(db, plan, explicitPhase)
}

func parseNextArgument(argument string) (mode, explicitPhase string, ok bool) {
	fields := strings.Fields(strings.TrimSpace(argument))
	if len(fields) == 0 {
		return "auto", "", true
	}
	switch fields[0] {
	case "auto":
		if len(fields) == 1 {
			return "auto", "", true
		}
		return "", "", false
	case "simple":
		return "simple", "", true
	case "full":
		if len(fields) >= 3 && fields[1] == "phase" {
			return "full", fields[2], true
		}
		if len(fields) == 1 {
			return "full", "", true
		}
		return "", "", false
	case "phase":
		if len(fields) == 2 {
			return "full", fields[1], true
		}
		return "", "", false
	default:
		return "", "", false
	}
}

func resolveFullMode(db *sql.DB, plan activePlan, explicitPhase string) (NextView, error) {
	slugs := parseActivePlanPhaseOrder(plan.content)
	if len(slugs) == 0 {
		return NextView{Mode: "full", Stop: &StopInfo{
			Code:     "no-phase",
			Message:  fmt.Sprintf("Active plan %s defines no phase_slug entries.", plan.path),
			Recovery: "to-plan",
		}}, nil
	}

	slug := explicitPhase
	if slug != "" && !containsSlug(slugs, slug) {
		return NextView{Mode: "full", ActivePhase: &slug, Stop: &StopInfo{
			Code:     "no-phase",
			Message:  fmt.Sprintf("Phase %q is not defined in active plan %s.", slug, plan.path),
			Recovery: "to-plan",
		}}, nil
	}
	if slug == "" {
		candidate, err := selectActivePhase(db, slugs)
		if err != nil {
			return NextView{}, err
		}
		if candidate == "" {
			return NextView{Mode: "full", Stop: &StopInfo{
				Code:     "all-phases-done",
				Message:  fmt.Sprintf("Every phase in active plan %s is done.", plan.path),
				Recovery: "run `handoff`, or `brainstorm lock` a new initiative",
			}}, nil
		}
		slug = candidate
	}

	if marker, found := findPlaceholder(plan.content); found {
		return NextView{Mode: "full", ActivePhase: &slug, Stop: &StopInfo{
			Code:     "placeholder-plan",
			Message:  fmt.Sprintf("Active plan %s still contains a placeholder marker: %q", plan.path, marker),
			Recovery: "to-plan",
		}}, nil
	}

	return NextView{Mode: "full", ActivePhase: &slug}, nil
}

// selectActivePhase returns the first plan-ordered phase whose story is not
// done. A phase with no story row yet counts as incomplete. Without a DB, the
// first plan phase is the only deterministic candidate.
func selectActivePhase(db *sql.DB, slugs []string) (string, error) {
	if db == nil {
		return slugs[0], nil
	}
	for _, slug := range slugs {
		_, status, exists, err := storyByExactSlug(db, slug)
		if err != nil {
			return "", err
		}
		if !exists || status != domain.StoryDone {
			return slug, nil
		}
	}
	return "", nil
}

func parseActivePlanPhaseOrder(content string) []string {
	matches := activePlanPhaseSlug.FindAllStringSubmatch(content, -1)
	slugs := make([]string, 0, len(matches))
	for _, match := range matches {
		slugs = append(slugs, match[1])
	}
	return slugs
}

func containsSlug(slugs []string, target string) bool {
	for _, slug := range slugs {
		if slug == target {
			return true
		}
	}
	return false
}

func findPlaceholder(plan string) (marker string, found bool) {
	for _, marker := range placeholderMarkers {
		if strings.Contains(plan, marker) {
			return marker, true
		}
	}
	return "", false
}
