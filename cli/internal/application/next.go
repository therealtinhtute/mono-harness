package application

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// NextView mirrors `zharness next`'s shape: the resolved execution mode,
// the selected phase (full mode only, once resolved), and an optional
// stop naming a blocker + its recovery.
type NextView struct {
	Mode        string    `json:"mode"`
	ActivePhase *string   `json:"active_phase"`
	Stop        *StopInfo `json:"stop,omitempty"`
}

// StopInfo is one blocking state from work.md's former Full-Mode-Detection
// table: a taxonomy code, a human message naming the specific blocker, and
// the exact recovery command/action.
type StopInfo struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Recovery string `json:"recovery"`
}

const (
	nextSpecPath    = ".kit/planning/SPEC.md"
	nextRoadmapPath = ".kit/planning/ROADMAP.md"
)

func phaseContextPath(slug string) string {
	return filepath.Join(".kit", "planning", "phases", slug, slug+"-CONTEXT.md")
}

func phasePlanPath(slug string) string {
	return filepath.Join(".kit", "planning", "phases", slug, slug+"-PLAN.md")
}

var placeholderMarkers = []string{"TBD", "TODO", "similar to", "implement later"}

var roadmapPhaseHeader = regexp.MustCompile(`(?m)^## Phase \d+:\s*(\S+)`)

// Next resolves work.md's former Mode-Resolution + Full-Mode-Detection
// tables in Go. Two rows are deliberately NOT encoded here and stay
// agent-side (named in the stop taxonomy's absence, not silently
// dropped):
//   - contract-drift: needs a working-tree diff against the phase's
//     Allowed/Forbidden Surfaces — this CLI has no git access, the same
//     constraint `resume`'s git/WIP block already carries.
//   - stale-plan (file/symbol no longer exists): a PLAN.md legitimately
//     references files it will create by being executed, which don't
//     exist yet — a naive existence check false-positives on almost
//     every real plan. Left to the manual-check review pass instead.
func Next(db *sql.DB, argument string) (NextView, error) {
	mode, explicitPhase, ok := parseNextArgument(argument)
	if !ok {
		return NextView{}, &domain.ValidationError{
			Code:    "unknown_argument",
			Message: fmt.Sprintf("next: unrecognized argument %q (want: (none) | auto | simple [@file] | full | full phase <slug> | phase <slug>)", argument),
		}
	}

	if mode == "auto" {
		hasSpec := !fileMissingOrEmpty(nextSpecPath)
		hasBrainstorm, err := anyBrainstormReports()
		if err != nil {
			return NextView{}, err
		}
		switch {
		case hasSpec && hasBrainstorm:
			return NextView{Mode: "auto", Stop: &StopInfo{
				Code:     "ambiguous",
				Message:  "Both a SPEC.md and a brainstorm report exist — unclear which should drive this invocation.",
				Recovery: "ask the user: continue the locked SPEC via `work full`, or `brainstorm refine` if the brainstorm file supersedes it",
			}}, nil
		case hasSpec:
			mode = "full"
		default:
			mode = "simple"
		}
	}

	if mode == "simple" {
		return NextView{Mode: "simple"}, nil
	}

	return resolveFullMode(db, explicitPhase)
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

func resolveFullMode(db *sql.DB, explicitPhase string) (NextView, error) {
	if fileMissingOrEmpty(nextSpecPath) {
		return NextView{Mode: "full", Stop: &StopInfo{
			Code:     "no-spec",
			Message:  "No .kit/planning/SPEC.md found (or it is empty). Lock the problem first.",
			Recovery: "brainstorm",
		}}, nil
	}
	if !infrastructure.Exists(nextRoadmapPath) {
		return NextView{Mode: "full", Stop: &StopInfo{
			Code:     "no-plan",
			Message:  "SPEC exists, no plan. Generate the roadmap first.",
			Recovery: "to-plan full",
		}}, nil
	}

	slug := explicitPhase
	if slug == "" {
		candidate, multiple, err := selectActivePhase(db)
		if err != nil {
			return NextView{}, err
		}
		if multiple {
			return NextView{Mode: "full", Stop: &StopInfo{
				Code:     "multiple-incomplete",
				Message:  "More than one roadmap phase has incomplete work.",
				Recovery: "ask the user which phase to run (recommend the first incomplete phase by roadmap order)",
			}}, nil
		}
		if candidate == "" {
			return NextView{Mode: "full", Stop: &StopInfo{
				Code:     "all-phases-done",
				Message:  "Every roadmap phase is done.",
				Recovery: "run `handoff`, or `brainstorm` a new initiative",
			}}, nil
		}
		slug = candidate
	}

	var missing []string
	if !infrastructure.Exists(phasePlanPath(slug)) {
		missing = append(missing, phasePlanPath(slug))
	}
	if !infrastructure.Exists(phaseContextPath(slug)) {
		missing = append(missing, phaseContextPath(slug))
	}
	if len(missing) > 0 {
		return NextView{Mode: "full", ActivePhase: &slug, Stop: &StopInfo{
			Code:     "no-phase",
			Message:  fmt.Sprintf("Phase artifacts missing for %q: %s", slug, strings.Join(missing, ", ")),
			Recovery: fmt.Sprintf("to-plan phase %s", slug),
		}}, nil
	}

	planBytes, err := os.ReadFile(phasePlanPath(slug))
	if err != nil {
		return NextView{}, fmt.Errorf("next: read %s: %w", phasePlanPath(slug), err)
	}
	if marker, found := findPlaceholder(string(planBytes)); found {
		return NextView{Mode: "full", ActivePhase: &slug, Stop: &StopInfo{
			Code:     "placeholder-plan",
			Message:  fmt.Sprintf("Phase plan %s still contains a placeholder marker: %q", phasePlanPath(slug), marker),
			Recovery: fmt.Sprintf("to-plan phase %s", slug),
		}}, nil
	}

	return NextView{Mode: "full", ActivePhase: &slug}, nil
}

// selectActivePhase parses ROADMAP.md for its ordered phase list and
// cross-references each slug's story status in the DB. A phase with no
// story row yet, or a story not yet `done`, counts as incomplete — this
// reads the harness's own authoritative completion field rather than
// scanning PLAN.md prose for ad-hoc "all waves complete" markers.
func selectActivePhase(db *sql.DB) (slug string, multipleIncomplete bool, err error) {
	slugs, err := parseRoadmapPhaseOrder(nextRoadmapPath)
	if err != nil {
		return "", false, err
	}

	var incomplete []string
	for _, s := range slugs {
		if db == nil {
			// No harness.db yet (matches resume's "no-harness" state) — every
			// phase counts as not-yet-done since nothing has been recorded.
			incomplete = append(incomplete, s)
			continue
		}
		_, status, exists, err := storyByExactSlug(db, s)
		if err != nil {
			return "", false, err
		}
		if !exists || status != domain.StoryDone {
			incomplete = append(incomplete, s)
		}
	}

	switch len(incomplete) {
	case 0:
		return "", false, nil
	case 1:
		return incomplete[0], false, nil
	default:
		return "", true, nil
	}
}

func parseRoadmapPhaseOrder(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("next: read %s: %w", path, err)
	}
	matches := roadmapPhaseHeader.FindAllStringSubmatch(string(data), -1)
	slugs := make([]string, 0, len(matches))
	for _, m := range matches {
		slugs = append(slugs, m[1])
	}
	return slugs, nil
}

func findPlaceholder(plan string) (marker string, found bool) {
	for _, m := range placeholderMarkers {
		if strings.Contains(plan, m) {
			return m, true
		}
	}
	return "", false
}

func anyBrainstormReports() (bool, error) {
	matches, err := filepath.Glob(".kit/reports/brainstorm/*.md")
	if err != nil {
		return false, fmt.Errorf("next: glob brainstorm reports: %w", err)
	}
	return len(matches) > 0, nil
}

func fileMissingOrEmpty(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return true
	}
	return info.Size() == 0
}
