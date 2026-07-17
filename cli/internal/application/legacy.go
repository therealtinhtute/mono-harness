package application

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// ErrLegacyFieldUnmapped marks a legacy input with no destination under
// STATE.md's Legacy Field Mapping (CONTRACT.md `legacy_field_unmapped`).
type ErrLegacyFieldUnmapped struct {
	Field  string
	Reason string
}

func (e *ErrLegacyFieldUnmapped) Error() string {
	return fmt.Sprintf("legacy field %q: %s", e.Field, e.Reason)
}

// knownLegacyFields is STATE.md's full Legacy Field Mapping table (10
// rows): 4 map 1:1 to a meta column, 4 are dropped/derived convention
// paths, 1 (handoff) is superseded by a queryable list, 1 (last_updated)
// is dropped. All 10 are "mapped" (even the dropped ones have a defined
// disposition) — only a name absent from this set is unmapped.
var knownLegacyFields = map[string]bool{
	"current_phase":       true,
	"entry_phase":         true,
	"spec":                true,
	"roadmap":             true,
	"active_context":      true,
	"active_plan":         true,
	"latest_cook_run":     true,
	"latest_check_report": true,
	"handoff":             true,
	"last_updated":        true,
}

// parseFlatYAML reads workflow-state.yml's flat `key: value` lines. The
// file has no nesting/lists, so a full YAML parser isn't warranted.
func parseFlatYAML(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	fields := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		idx := strings.Index(line, ": ")
		if idx < 0 {
			if strings.HasSuffix(strings.TrimSpace(line), ":") {
				key := strings.TrimSuffix(strings.TrimSpace(line), ":")
				fields[key] = ""
			}
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+2:])
		fields[key] = val
	}
	return fields, nil
}

// checkKnownLegacyFields rejects any key not named in STATE.md's mapping.
func checkKnownLegacyFields(fields map[string]string) error {
	for k := range fields {
		if !knownLegacyFields[k] {
			return &ErrLegacyFieldUnmapped{Field: k, Reason: "no destination in STATE.md"}
		}
	}
	return nil
}

// normalizePhase treats the documented "none" sentinel (STATE.md: "none
// if no phase started") the same as an empty/absent value.
func normalizePhase(v string) string {
	if v == "none" {
		return ""
	}
	return v
}

var phaseHeadingRe = regexp.MustCompile(`(?m)^##\s+Phase\s+\d+:\s*(\S+)\s*$`)
var goalLineRe = regexp.MustCompile(`(?m)^\*\*Goal:\*\*\s*(.+)$`)

// parseRoadmapGoal extracts the `**Goal:**` line under `## Phase N: {slug}`
// in ROADMAP.md. Returns ok=false if the roadmap or the phase heading
// can't be found — a soft miss, not an error (a phase absent from the
// roadmap still gets a story row, with a placeholder goal).
func parseRoadmapGoal(roadmapPath, slug string) (string, bool) {
	data, err := os.ReadFile(roadmapPath)
	if err != nil {
		return "", false
	}
	text := string(data)

	headings := phaseHeadingRe.FindAllStringSubmatchIndex(text, -1)
	for i, m := range headings {
		if text[m[2]:m[3]] != slug {
			continue
		}
		start := m[1]
		end := len(text)
		if i+1 < len(headings) {
			end = headings[i+1][0]
		}
		if gm := goalLineRe.FindStringSubmatch(text[start:end]); gm != nil {
			return strings.TrimSpace(gm[1]), true
		}
		return "", false
	}
	return "", false
}

var artifactFilenameRe = regexp.MustCompile(`^(\d{8})-(\d{4})-(.+)\.md$`)

// parseArtifactFilename extracts the phase slug and timestamp embedded in
// a `{YYYYMMDD-HHmm}-{slug}.md` artifact path (the locked naming
// convention in the work/check reference templates).
func parseArtifactFilename(path string) (slug, createdAtRFC3339 string, ok bool) {
	base := path
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		base = path[i+1:]
	}
	m := artifactFilenameRe.FindStringSubmatch(base)
	if m == nil {
		return "", "", false
	}
	t, err := time.Parse("20060102 1504", m[1]+" "+m[2])
	if err != nil {
		return "", "", false
	}
	return m[3], t.UTC().Format(time.RFC3339), true
}
