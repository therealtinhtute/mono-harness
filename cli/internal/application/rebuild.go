package application

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// RebuildResult reports how many rows RebuildFromMarkdown reconstructed per
// table, mirroring `db rebuild`'s existing "replayed" count shape.
type RebuildResult struct {
	Stories   int `json:"stories"`
	Intakes   int `json:"intakes"`
	Runs      int `json:"runs"`
	Checks    int `json:"checks"`
	Handoffs  int `json:"handoffs"`
	Traces    int `json:"traces"`
	Decisions int `json:"decisions"`
	Memories  int `json:"memories"`
}

// rebuildPlanGlobs covers every plan a rebuild must scan — active and
// completed — reusing the same path constants ResolveActivePlan and
// completePlan already anchor on (plan_resolve.go, plan_lifecycle.go), so
// this can't independently drift from where a plan actually lives.
var rebuildPlanGlobs = []string{nextActivePlansGlob, filepath.Join(activePlanCompletedDir, "*.md")}

// RebuildFromMarkdown reconstructs stories, intakes, runs, checks,
// handoffs, traces, and decisions from every committed plan under
// docs/plans/{active,completed}/*.md alone — no read of .kit/changesets/
// (P3 wave 1, docs/plans/active/harness-markdown-truth.md). db must
// already be freshly migrated and empty.
//
// Three fields genuinely cannot be recovered from markdown, each handled
// by degrading gracefully rather than inventing durable-looking data:
//
//   - meta pointers (current_phase, latest_run_id, latest_check_id,
//     docs_version) are left unset. Nothing in committed markdown proves
//     which run/check is "latest" without re-deriving ordering the DB used
//     to own, and an unset meta is already a supported, tested drift state
//     (resume's unknown_phase/out_of_order/stale_docs tolerance).
//   - traces and decisions never carried their own row ID in markdown
//     (formatTraceProgressEntry/formatDecisionEntry emit no id — only
//     checks and handoffs embed one). Rebuilt trace/decision rows get
//     freshly minted IDs; their content is preserved, their identity is
//     not — nothing else in the schema references trace_id or decision_id.
//   - a run is only reconstructed when a Validation (check) entry
//     backreferences it, because that is the one entry shape that also
//     carries the run's story slug (` + "`phase:`" + ` in
//     formatCheckValidationEntry) — the only source that can satisfy
//     runs.story_slug NOT NULL. A run mentioned only by a trace or handoff
//     entry, never checked, has no recoverable story_slug and is not
//     reconstructed; any trace/decision/handoff that backreferences such a
//     run gets a NULL run_id instead of a dangling one — the same
//     "no durable footprint, correctly dropped" tradeoff already accepted
//     for runs generally.
//
// Similarly, intakes.type and intakes.summary have no home in a plan's own
// frontmatter (only intake_id and lane do) — reconstructed intake rows get
// a synthesized type/summary and a freshly minted id; only plan_id and
// lane are functionally load-bearing (resolveLaneForRun's join keys on
// intakes.plan_id, not intakes.id), and both are recovered exactly.
func RebuildFromMarkdown(db *sql.DB) (RebuildResult, error) {
	var result RebuildResult

	paths, err := rebuildPlanPaths()
	if err != nil {
		return result, err
	}

	contents := make(map[string]string, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return result, fmt.Errorf("db rebuild: read %s: %w", path, err)
		}
		contents[path] = normalizeLineEndings(string(data))
	}

	// Pass 1: stories, so every plan's phase-block stories exist before any
	// run (FK'd to stories.slug) or decision (FK'd to stories.slug via
	// phase) is inserted, regardless of which plan file references which
	// story. depends_on is deliberately left unset here and backfilled in
	// its own pass immediately after: stories.depends_on REFERENCES
	// stories(slug) under foreign_keys=ON, and a phase's dependency can
	// name a slug from another plan file, or a later phase in the same
	// plan, that has not been inserted yet — inserting every row first and
	// only then wiring depends_on removes any ordering requirement between
	// plans or between phases within one plan.
	for _, path := range paths {
		if err := rebuildStoriesFromPlan(db, contents[path], &result); err != nil {
			return result, fmt.Errorf("db rebuild: stories from %s: %w", path, err)
		}
	}
	for _, path := range paths {
		if err := rebuildStoryDependsOn(db, contents[path]); err != nil {
			return result, fmt.Errorf("db rebuild: story depends_on from %s: %w", path, err)
		}
	}

	// Pass 2: intakes, checks (+ their backing runs), traces, handoffs,
	// decisions. Checks run first among these because they are the only
	// entry shape that can resolve a run's story_slug.
	runStorySlug := map[string]string{}
	for _, path := range paths {
		if err := rebuildIntakeFromFrontmatter(db, path, contents[path], &result); err != nil {
			return result, fmt.Errorf("db rebuild: intake from %s: %w", path, err)
		}
		if err := rebuildChecksFromValidation(db, contents[path], runStorySlug, &result); err != nil {
			return result, fmt.Errorf("db rebuild: checks from %s: %w", path, err)
		}
		if err := rebuildProgressEntries(db, contents[path], runStorySlug, &result); err != nil {
			return result, fmt.Errorf("db rebuild: progress from %s: %w", path, err)
		}
		if err := rebuildDecisionsFromSection(db, contents[path], &result); err != nil {
			return result, fmt.Errorf("db rebuild: decisions from %s: %w", path, err)
		}
	}

	// Memories live outside docs/plans/{active,completed}/*.md entirely
	// (docs/memory/*.md, one file per entry), so they're reconstructed in
	// their own pass rather than folded into the plan-scanning loop above
	// (R1/R4, docs/plans/active/durable-memory.md).
	if err := rebuildMemoriesFromMarkdown(db, &result); err != nil {
		return result, fmt.Errorf("db rebuild: memories: %w", err)
	}

	return result, nil
}

func rebuildPlanPaths() ([]string, error) {
	var paths []string
	for _, glob := range rebuildPlanGlobs {
		matches, err := filepath.Glob(glob)
		if err != nil {
			return nil, fmt.Errorf("db rebuild: glob %s: %w", glob, err)
		}
		paths = append(paths, matches...)
	}
	sort.Strings(paths)
	return paths, nil
}

// planFieldLegacyAlias names the field key docs/plans/completed/pr60-review-fixes.md
// (R18) uses in place of the canonical key to-plan writes — that plan
// predates to-plan's field-naming convention and is never rewritten.
var planFieldLegacyAlias = map[string]string{
	"story_id":   "story",
	"depends_on": "depends on",
}

// planFieldValue finds the first "key: value" line anywhere in block — the
// same single-match assumption SetPlanPhaseStatus (plan_phase_status.go)
// already relies on for a phase block's own status line, extended to the
// block's other scalar fields (story_id, goal, depends_on), which to-plan
// places immediately after phase_slug, before any nested waves/tasks. If key
// itself isn't found, planFieldLegacyAlias's historical name is tried next.
// A value starting with a backtick is trimmed to the text between its first
// backtick pair, since the legacy form quotes the value and may append
// trailing prose after it (e.g. "`slug` — explanation").
func planFieldValue(block, key string) (string, bool) {
	value, ok := planFieldValueExact(block, key)
	if !ok {
		if alias, hasAlias := planFieldLegacyAlias[key]; hasAlias {
			value, ok = planFieldValueExact(block, alias)
		}
	}
	if !ok {
		return "", false
	}
	if strings.HasPrefix(value, "`") {
		if end := strings.Index(value[1:], "`"); end != -1 {
			value = value[1 : end+1]
		}
	}
	return value, true
}

func planFieldValueExact(block, key string) (string, bool) {
	re := regexp.MustCompile(`(?m)^[ \t]*(?:-[ \t]*)?` + regexp.QuoteMeta(key) + `: ?(.*?)[ \t]*\r?$`)
	m := re.FindStringSubmatch(block)
	if m == nil {
		return "", false
	}
	return m[1], true
}

// planPhaseSlugs enumerates every phase_slug declared in a plan's `##
// Phases and Verification` section, in file order, trying both the
// heading and list-item forms (plan_query.go's two supported shapes).
func planPhaseSlugs(sectionBody string) []string {
	var slugs []string
	for _, m := range planPhaseHeading.FindAllStringSubmatch(sectionBody, -1) {
		slugs = append(slugs, m[1])
	}
	for _, m := range planPhaseListItem.FindAllStringSubmatch(sectionBody, -1) {
		slugs = append(slugs, m[2])
	}
	return slugs
}

func rebuildStoriesFromPlan(db *sql.DB, content string, result *RebuildResult) error {
	body, ok := extractPlanSection(content, "Phases and Verification")
	if !ok {
		return nil
	}
	for _, slug := range planPhaseSlugs(body) {
		block, ok := extractPlanPhaseBlock(content, slug)
		if !ok {
			continue
		}
		storyID, ok := planFieldValue(block, "story_id")
		if !ok || storyID == "" {
			continue // malformed/hand-edited phase block: degrade, don't invent an id
		}
		exists, err := storyRowExistsByID(db, storyID)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		status, _ := planFieldValue(block, "status")
		if status == "" {
			status = domain.StoryPlanned
		}
		goal, _ := planFieldValue(block, "goal")

		if _, err := db.Exec(
			`INSERT INTO stories (id, slug, goal, status, created_at) VALUES (?, ?, ?, ?, ?)`,
			storyID, slug, goal, status, rebuildStamp(),
		); err != nil {
			return fmt.Errorf("insert story %s: %w", slug, err)
		}
		result.Stories++
	}
	return nil
}

func storyRowExistsByID(db *sql.DB, id string) (bool, error) {
	var found string
	err := db.QueryRow(`SELECT id FROM stories WHERE id = ?`, id).Scan(&found)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// rebuildStoryDependsOn backfills stories.depends_on once every story row
// from every plan file already exists (see the ordering note at its call
// site in RebuildFromMarkdown) — a slug whose depends_on names a story this
// rebuild never found degrades to NULL rather than violating the FK.
func rebuildStoryDependsOn(db *sql.DB, content string) error {
	body, ok := extractPlanSection(content, "Phases and Verification")
	if !ok {
		return nil
	}
	for _, slug := range planPhaseSlugs(body) {
		block, ok := extractPlanPhaseBlock(content, slug)
		if !ok {
			continue
		}
		dependsOn, ok := planFieldValue(block, "depends_on")
		if !ok || dependsOn == "" {
			continue
		}
		exists, err := storyRowExistsBySlug(db, dependsOn)
		if err != nil {
			return err
		}
		if !exists {
			continue // dependency this rebuild never found: degrade, don't invent
		}
		if _, err := db.Exec(`UPDATE stories SET depends_on = ? WHERE slug = ?`, dependsOn, slug); err != nil {
			return fmt.Errorf("set depends_on for %s: %w", slug, err)
		}
	}
	return nil
}

// rebuildIntakeFromFrontmatter recovers what a plan's own frontmatter
// actually carries — intake_id (the intake row's own id — CreateIntake
// mints it independently of the plan's own id, and a plan only ever
// records it, never the reverse, so intake_id is a real, preserved
// identity, unlike trace/decision ids), the plan's own frontmatter id
// (which CreateIntake's caller passes as intakes.plan_id — the field
// resolveLaneForRun actually joins on, per intake.go's doc comment), and
// lane — and synthesizes only what has no markdown home at all (see
// RebuildFromMarkdown's doc comment).
func rebuildIntakeFromFrontmatter(db *sql.DB, path, content string, result *RebuildResult) error {
	lines, ok := frontmatterPreview(content, frontmatterPreviewLines)
	if !ok {
		return nil
	}
	intakeID, ok := frontmatterPreviewField(lines, "intake_id")
	if !ok || intakeID == "" {
		return nil
	}
	planID, ok := frontmatterPreviewField(lines, "id")
	if !ok || planID == "" {
		return nil
	}
	lane, _ := frontmatterPreviewField(lines, "lane")

	exists, err := rowExistsByID(db, "intakes", intakeID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	if _, err := db.Exec(
		`INSERT INTO intakes (id, type, summary, lane, plan_id, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		intakeID, "reconstructed",
		fmt.Sprintf("reconstructed by db rebuild from %s frontmatter (intake_id, lane); original type/summary are not recorded in markdown", path),
		lane, planID, rebuildStamp(),
	); err != nil {
		return fmt.Errorf("insert intake for %s: %w", path, err)
	}
	result.Intakes++
	return nil
}

// checkValidationHeader matches formatCheckValidationEntry's (check_record.go)
// generated header line. Proof-link sub-bullets are parsed separately by
// checkProofLinkLine, since they're 0..N indented continuation lines, not
// part of the header.
var checkValidationHeader = regexp.MustCompile(
	"^- `([^`]+)` — check\\. verdict: `([^`]+)`\\. check: `([^`]+)`\\. run: `([^`]+)`\\.(?: phase: `([^`]+)`\\.)? judge: `([^`]+)` \\(([^)]*)\\)\\.$",
)

var checkProofLinkLine = regexp.MustCompile("^  - `([^`]+)`(?: → (.*))?$")

func rebuildChecksFromValidation(db *sql.DB, content string, runStorySlug map[string]string, result *RebuildResult) error {
	body, ok := extractPlanSection(content, "Validation")
	if !ok {
		return nil
	}
	for _, block := range splitTopLevelEntries(body) {
		lines := strings.Split(block, "\n")
		m := checkValidationHeader.FindStringSubmatch(lines[0])
		if m == nil {
			continue // hand-authored or otherwise non-CLI-generated entry: degrade, skip
		}
		at, verdict, checkID, runID, phase, judge, judgeModel := m[1], m[2], m[3], m[4], m[5], m[6], m[7]

		if phase != "" {
			if _, err := ensureRunRow(db, runID, phase, runStorySlug, result); err != nil {
				return err
			}
		}
		if _, resolvable := runStorySlug[runID]; !resolvable {
			continue // no recoverable story_slug for this run: the check itself can't satisfy runs FK either
		}

		exists, err := rowExistsByID(db, "checks", checkID)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		var proofLinks []map[string]string
		for _, sub := range lines[1:] {
			pm := checkProofLinkLine.FindStringSubmatch(sub)
			if pm == nil {
				continue
			}
			proofLinks = append(proofLinks, map[string]string{"command": pm[1], "output_ref": pm[2]})
		}
		proofLinksJSON, err := json.Marshal(proofLinks)
		if err != nil {
			return err
		}

		if _, err := db.Exec(
			`INSERT INTO checks (id, run_id, verdict, proof_links, judge, judge_model, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			checkID, runID, verdict, string(proofLinksJSON), judge, judgeModel, at,
		); err != nil {
			return fmt.Errorf("insert check %s: %w", checkID, err)
		}
		result.Checks++
	}
	return nil
}

// ensureRunRow inserts a run row for runID the first time it's seen backed
// by a resolvable story slug, recording the resolution in runStorySlug so
// later trace/decision/handoff entries referencing the same runID can
// decide whether to keep or drop that backreference (see
// RebuildFromMarkdown's doc comment). artifact_path and plan_id are not
// recoverable from markdown; artifact_path defaults to "" (satisfies its
// NOT NULL column, matching domain.Run.Validate()'s own accepted empty
// case), plan_id stays NULL.
func ensureRunRow(db *sql.DB, runID, storySlug string, runStorySlug map[string]string, result *RebuildResult) (bool, error) {
	if existing, ok := runStorySlug[runID]; ok {
		return existing == storySlug, nil
	}
	storyExists, err := storyRowExistsBySlug(db, storySlug)
	if err != nil {
		return false, err
	}
	if !storyExists {
		return false, nil // phase referenced a story this rebuild never found: degrade, don't invent
	}
	exists, err := rowExistsByID(db, "runs", runID)
	if err != nil {
		return false, err
	}
	if !exists {
		if _, err := db.Exec(
			`INSERT INTO runs (id, story_slug, artifact_path, created_at) VALUES (?, ?, '', ?)`,
			runID, storySlug, rebuildStamp(),
		); err != nil {
			return false, fmt.Errorf("insert run %s: %w", runID, err)
		}
		result.Runs++
	}
	runStorySlug[runID] = storySlug
	return true, nil
}

func storyRowExistsBySlug(db *sql.DB, slug string) (bool, error) {
	var found string
	err := db.QueryRow(`SELECT slug FROM stories WHERE slug = ?`, slug).Scan(&found)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func rowExistsByID(db *sql.DB, table, id string) (bool, error) {
	var found string
	err := db.QueryRow(`SELECT id FROM `+table+` WHERE id = ?`, id).Scan(&found) //nolint:gosec // table is always a compile-time literal from this file's own callers, never user input
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// traceProgressHeader and handoffProgressHeader match
// formatTraceProgressEntry (trace.go) and formatHandoffProgressEntry
// (handoff.go). task's value is not backtick-wrapped in the source format
// (unlike every other field), so a task string containing a literal "."
// truncates at the first one — a known, documented limit of this
// non-greedy match, not a rebuild-time error: that one field degrades,
// the row still reconstructs.
var traceProgressHeader = regexp.MustCompile(
	"^- `([^`]+)` — wave (\\d+)(?:, task ([^`]+?))?\\.(?: task_status: `([^`]+)`\\.)?(?: run: `([^`]+)`\\.)? summary: (.*)\\.$",
)

var handoffProgressHeader = regexp.MustCompile(
	"^- `([^`]+)` — handoff recorded\\. handoff: `([^`]+)`\\.(?: run: `([^`]+)`\\.)?(?: check: `([^`]+)`\\.)?( phase closed\\.)?(?: next action: (.*?)\\.)?(?: open items: (.*)\\.)?$",
)

func rebuildProgressEntries(db *sql.DB, content string, runStorySlug map[string]string, result *RebuildResult) error {
	body, ok := extractPlanSection(content, "Progress")
	if !ok {
		return nil
	}
	for _, block := range splitTopLevelEntries(body) {
		line := strings.Split(block, "\n")[0]

		if m := handoffProgressHeader.FindStringSubmatch(line); m != nil {
			if err := rebuildHandoffRow(db, m, runStorySlug, result); err != nil {
				return err
			}
			continue
		}
		if m := traceProgressHeader.FindStringSubmatch(line); m != nil {
			if err := rebuildTraceRow(db, m, runStorySlug, result); err != nil {
				return err
			}
			continue
		}
		// hand-authored or otherwise non-CLI-generated entry: degrade, skip
	}
	return nil
}

func rebuildHandoffRow(db *sql.DB, m []string, runStorySlug map[string]string, result *RebuildResult) error {
	at, handoffID, runID, checkID := m[1], m[2], m[3], m[4]

	exists, err := rowExistsByID(db, "handoffs", handoffID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	runIDArg := resolvedRunIDArg(runID, runStorySlug)
	checkIDArg := any(nil)
	if checkID != "" {
		checkExists, err := rowExistsByID(db, "checks", checkID)
		if err != nil {
			return err
		}
		if checkExists {
			checkIDArg = checkID
		}
	}

	// Mirrors formatHandoffProgressEntry's caller (RecordHandoff,
	// handoff.go) key-for-key — latest_run_id/latest_check_id duplicate the
	// run_id/check_id columns inside anchors too, matching the live write
	// exactly, so a rebuilt row's anchors JSON round-trips byte-for-byte
	// whenever every field was actually present in the markdown line. The
	// one unrecoverable case is open_items: the live write always sets the
	// key (to an empty slice when there were none), but the markdown line
	// omits "open items: " entirely in that case, so a handoff originally
	// recorded with zero open items reconstructs without that key.
	anchors := map[string]any{}
	if runIDArg != nil {
		anchors["latest_run_id"] = runIDArg
	}
	if checkIDArg != nil {
		anchors["latest_check_id"] = checkIDArg
	}
	if len(m) > 6 && m[6] != "" {
		anchors["exact_next_action"] = m[6]
	}
	if len(m) > 7 && m[7] != "" {
		anchors["open_items"] = strings.Split(m[7], "; ")
	}
	anchorsJSON, err := json.Marshal(anchors)
	if err != nil {
		return err
	}

	if _, err := db.Exec(
		`INSERT INTO handoffs (id, run_id, check_id, anchors, created_at) VALUES (?, ?, ?, ?, ?)`,
		handoffID, runIDArg, checkIDArg, string(anchorsJSON), at,
	); err != nil {
		return fmt.Errorf("insert handoff %s: %w", handoffID, err)
	}
	result.Handoffs++
	return nil
}

func rebuildTraceRow(db *sql.DB, m []string, runStorySlug map[string]string, result *RebuildResult) error {
	at, waveStr, task, taskStatus, runID, summary := m[1], m[2], m[3], m[4], m[5], m[6]
	var wave int
	if _, err := fmt.Sscanf(waveStr, "%d", &wave); err != nil {
		return nil // malformed wave number: degrade, skip rather than fail the whole rebuild
	}

	runIDArg := resolvedRunIDArg(runID, runStorySlug)
	var taskArg, taskStatusArg any
	if task != "" {
		taskArg = task
	}
	if taskStatus != "" {
		taskStatusArg = taskStatus
	}

	if _, err := db.Exec(
		`INSERT INTO traces (id, run_id, wave, summary, created_at, task, task_status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		ulid.Make().String(), runIDArg, wave, summary, at, taskArg, taskStatusArg,
	); err != nil {
		return fmt.Errorf("insert trace: %w", err)
	}
	result.Traces++
	return nil
}

// decisionEntry matches formatDecisionEntry (decision.go). decision and
// rationale are free text with no backtick delimiter, so — like task above
// — a value containing the exact literal marker text this regex anchors on
// (" (phase: `", ". rationale: ") truncates early rather than failing the
// whole rebuild; that row's free text degrades, the row itself still
// reconstructs with a valid decision/rationale prefix.
var decisionEntry = regexp.MustCompile(
	"^- `([^`]+)` — (.+?)(?: \\(phase: `([^`]+)`\\))?(?:, task: ([^.]+))?\\. rationale: (.*)\\.$",
)

func rebuildDecisionsFromSection(db *sql.DB, content string, result *RebuildResult) error {
	body, ok := extractPlanSection(content, "Decisions")
	if !ok {
		return nil
	}
	for _, block := range splitTopLevelEntries(body) {
		line := strings.Split(block, "\n")[0]
		m := decisionEntry.FindStringSubmatch(line)
		if m == nil {
			continue // hand-authored or otherwise non-CLI-generated entry: degrade, skip
		}
		at, decisionText, phase, task, rationale := m[1], m[2], m[3], m[4], m[5]

		var phaseArg any
		if phase != "" {
			exists, err := storyRowExistsBySlug(db, phase)
			if err != nil {
				return err
			}
			if exists {
				phaseArg = phase
			}
		}
		var taskArg any
		if task != "" {
			taskArg = task
		}

		if _, err := db.Exec(
			`INSERT INTO decisions (id, phase, task, decision, rationale, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
			ulid.Make().String(), phaseArg, taskArg, decisionText, rationale, at,
		); err != nil {
			return fmt.Errorf("insert decision: %w", err)
		}
		result.Decisions++
	}
	return nil
}

func resolvedRunIDArg(runID string, runStorySlug map[string]string) any {
	if runID == "" {
		return nil
	}
	if _, ok := runStorySlug[runID]; !ok {
		return nil // no recoverable story_slug for this run: drop the backreference, not the row
	}
	return runID
}

// splitTopLevelEntries splits a section body into blocks, each starting at
// a line beginning with "- " at column 0 (a top-level bullet) and
// including any indented continuation lines that follow it (e.g. a check
// entry's proof-link sub-bullets) up to the next top-level bullet.
func splitTopLevelEntries(body string) []string {
	lines := strings.Split(body, "\n")
	var blocks []string
	var current []string
	flush := func() {
		if len(current) > 0 {
			blocks = append(blocks, strings.Join(current, "\n"))
			current = nil
		}
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "- ") {
			flush()
		}
		if len(current) == 0 && strings.TrimSpace(line) == "" {
			continue
		}
		current = append(current, line)
	}
	flush()
	var entries []string
	for _, b := range blocks {
		if strings.TrimSpace(b) == "- none" {
			continue
		}
		entries = append(entries, b)
	}
	return entries
}

// rebuildStamp is used for rows whose created_at has no recoverable source
// in markdown (stories, intakes, runs) — every other reconstructed row
// (checks, handoffs, traces, decisions) reuses its own entry's `` `AT` ``
// timestamp instead.
func rebuildStamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
