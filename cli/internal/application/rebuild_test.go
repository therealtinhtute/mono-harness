package application

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// TestRebuildFromMarkdownRoundTrip proves Wave 1's core goal (P3,
// docs/plans/active/harness-markdown-truth.md): wiping harness.db and
// rebuilding from a fixture repository containing only committed markdown
// reconstructs stories, the check-backed run, checks, handoffs, traces,
// decisions, and intakes to match the pre-wipe state.
//
// The plan's markdown is built from the exact same format*Entry functions
// the live write paths use (formatTraceProgressEntry,
// formatCheckValidationEntry, formatHandoffProgressEntry,
// formatDecisionEntry) so this test cannot silently drift from the real
// on-disk shape, and the pre-wipe DB rows are seeded with the id/at values
// baked into that markdown so the post-rebuild comparison is exact rather
// than approximate.
//
// Three fields are deliberately excluded from equality, matching
// RebuildFromMarkdown's documented, accepted tradeoffs: trace/decision ids
// (freshly minted on rebuild — markdown never carried the original),
// intakes.type/summary (no markdown source at all — synthesized), and
// runs.artifact_path/plan_id (not recoverable — seeded as "" / unset here
// so the comparison is exact rather than skipped).
func TestRebuildFromMarkdownRoundTrip(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	planID := ulid.Make().String()
	intakeID := ulid.Make().String()
	storyID := ulid.Make().String()
	runID := ulid.Make().String()
	checkID := ulid.Make().String()
	handoffID := ulid.Make().String()
	const slug = "p1-fixture"
	const lane = "high-risk"

	traceAt := "2026-07-01T10:00:00Z"
	checkAt := "2026-07-01T10:05:00Z"
	handoffAt := "2026-07-01T10:10:00Z"
	decisionAt := "2026-07-01T10:15:00Z"

	proofLinks := []domain.ProofLink{{Command: "go test ./...", OutputRef: "all green"}}
	openItems := []string{"ship wave 2"}
	nextAction := "start wave 2"

	planContent := "---\n" +
		"id: " + planID + "\n" +
		"type: plan\n" +
		"intake_id: " + intakeID + "\n" +
		"lane: " + lane + "\n" +
		"status: active\n" +
		"---\n\n" +
		"# Plan: Rebuild fixture\n\n" +
		"## Phases and Verification\n" +
		"- phases:\n" +
		"  - phase_slug: " + slug + "\n" +
		"    story_id: " + storyID + "\n" +
		"    status: checked\n" +
		"    goal: exercise the round trip\n" +
		"    waves:\n" +
		"      - wave: 1\n" +
		"        tasks:\n" +
		"          - task: seed the fixture\n\n" +
		"## Progress\n" +
		formatTraceProgressEntry(traceAt, 1, "seeded the fixture", runID, "seed the fixture", domain.TaskStatusDone) + "\n" +
		formatHandoffProgressEntry(handoffAt, handoffID, runID, checkID, nextAction, openItems, true) + "\n\n" +
		"## Decisions\n" +
		formatDecisionEntry(decisionAt, domain.Decision{Decision: "use a hand-seeded fixture", Rationale: "keeps the round trip exact", Phase: slug, Task: "seed the fixture"}) + "\n\n" +
		"## Validation\n" +
		formatCheckValidationEntry(checkAt, checkID, slug, runID, domain.VerdictApproved, domain.JudgeIndependent, "sonnet-5", proofLinks, domain.CheckModeFull) + "\n"

	planPath := filepath.Join("docs", "plans", "active", "rebuild-fixture.md")
	writeFile(t, planPath, planContent)

	seedRebuildFixtureRows(t, db, rebuildFixtureRows{
		planID: planID, intakeID: intakeID, lane: lane,
		storyID: storyID, slug: slug, status: domain.StoryChecked,
		runID:   runID,
		checkID: checkID, checkAt: checkAt, checkMode: domain.CheckModeFull, proofLinks: proofLinks,
		handoffID: handoffID, handoffAt: handoffAt, nextAction: nextAction, openItems: openItems,
		traceAt:    traceAt,
		decisionAt: decisionAt, decisionPhase: slug, decisionTask: "seed the fixture",
	})

	before := snapshotRebuildState(t, db, storyID, runID, checkID, handoffID, slug)

	dbPath := filepath.Join(t.TempDir(), "harness.db")
	rebuiltDB, err := infrastructure.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { rebuiltDB.Close() })
	if _, _, err := infrastructure.Migrate(rebuiltDB); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	result, err := RebuildFromMarkdown(rebuiltDB)
	if err != nil {
		t.Fatalf("RebuildFromMarkdown: %v", err)
	}
	if result.Stories != 1 || result.Intakes != 1 || result.Runs != 1 || result.Checks != 1 ||
		result.Handoffs != 1 || result.Traces != 1 || result.Decisions != 1 {
		t.Fatalf("RebuildResult = %+v, want exactly one of each", result)
	}

	after := snapshotRebuildState(t, rebuiltDB, storyID, runID, checkID, handoffID, slug)

	if before.story != after.story {
		t.Errorf("story mismatch:\nbefore: %+v\nafter:  %+v", before.story, after.story)
	}
	if before.run != after.run {
		t.Errorf("run mismatch:\nbefore: %+v\nafter:  %+v", before.run, after.run)
	}
	if before.check.id != after.check.id || before.check.runID != after.check.runID ||
		before.check.verdict != after.check.verdict || before.check.judge != after.check.judge ||
		before.check.judgeModel != after.check.judgeModel {
		t.Errorf("check scalar fields mismatch:\nbefore: %+v\nafter:  %+v", before.check, after.check)
	}
	if after.check.mode != domain.CheckModeFull {
		t.Errorf("rebuilt check mode = %q, want full", after.check.mode)
	}
	if before.check.mode != after.check.mode {
		t.Errorf("check mode mismatch:\nbefore: %+v\nafter:  %+v", before.check.mode, after.check.mode)
	}
	if !jsonEqual(t, before.check.proofLinksJSON, after.check.proofLinksJSON) {
		t.Errorf("check.proof_links mismatch:\nbefore: %s\nafter:  %s", before.check.proofLinksJSON, after.check.proofLinksJSON)
	}
	if before.handoff.id != after.handoff.id || before.handoff.runID != after.handoff.runID || before.handoff.checkID != after.handoff.checkID {
		t.Errorf("handoff scalar fields mismatch:\nbefore: %+v\nafter:  %+v", before.handoff, after.handoff)
	}
	if !jsonEqual(t, before.handoff.anchorsJSON, after.handoff.anchorsJSON) {
		t.Errorf("handoff.anchors mismatch:\nbefore: %s\nafter:  %s", before.handoff.anchorsJSON, after.handoff.anchorsJSON)
	}
	if before.trace != after.trace {
		t.Errorf("trace content mismatch:\nbefore: %+v\nafter:  %+v", before.trace, after.trace)
	}
	if before.decision != after.decision {
		t.Errorf("decision content mismatch:\nbefore: %+v\nafter:  %+v", before.decision, after.decision)
	}
	if before.intake.id != after.intake.id || before.intake.lane != after.intake.lane || before.intake.planID != after.intake.planID {
		t.Errorf("intake identity/lane/plan_id mismatch:\nbefore: %+v\nafter:  %+v", before.intake, after.intake)
	}
}

// TestRebuildFromMarkdownDropsUncheckedRun proves the documented tradeoff:
// a run backreferenced only by a trace/handoff entry, never by a
// Validation (check) entry, cannot resolve runs.story_slug and is
// correctly dropped rather than reconstructed with an invented slug.
func TestRebuildFromMarkdownDropsUncheckedRun(t *testing.T) {
	chdirFixture(t)
	runID := ulid.Make().String()
	planContent := "---\n" +
		"id: " + ulid.Make().String() + "\n" +
		"type: plan\n" +
		"status: active\n" +
		"---\n\n" +
		"# Plan: Unchecked run fixture\n\n" +
		"## Phases and Verification\n" +
		"- phases: none\n\n" +
		"## Progress\n" +
		formatTraceProgressEntry("2026-07-01T10:00:00Z", 1, "orphan trace", runID, "", "") + "\n"
	writeFile(t, filepath.Join("docs", "plans", "active", "orphan.md"), planContent)

	db := freshDB(t)
	result, err := RebuildFromMarkdown(db)
	if err != nil {
		t.Fatalf("RebuildFromMarkdown: %v", err)
	}
	if result.Runs != 0 {
		t.Fatalf("Runs = %d, want 0 (never check-backed)", result.Runs)
	}
	if result.Traces != 1 {
		t.Fatalf("Traces = %d, want 1 (row still reconstructs, with a NULL run_id)", result.Traces)
	}
	var runIDCol sql.NullString
	if err := db.QueryRow(`SELECT run_id FROM traces LIMIT 1`).Scan(&runIDCol); err != nil {
		t.Fatalf("query trace: %v", err)
	}
	if runIDCol.Valid {
		t.Fatalf("trace.run_id = %q, want NULL (backreferenced run was never reconstructed)", runIDCol.String)
	}
}

// TestRebuildFromMarkdownLegacyCheckEntryWithoutMode proves the mode
// column round-trip degrades honestly for pre-mode Validation entries:
// an entry carrying no `mode:` segment still reconstructs the check, with
// an empty mode that is never treated as full.
func TestRebuildFromMarkdownLegacyCheckEntryWithoutMode(t *testing.T) {
	chdirFixture(t)
	storyID := ulid.Make().String()
	runID := ulid.Make().String()
	checkID := ulid.Make().String()
	const slug = "legacy-check-fixture"
	planContent := "---\n" +
		"id: " + ulid.Make().String() + "\n" +
		"type: plan\n" +
		"status: active\n" +
		"---\n\n" +
		"# Plan: Legacy check entry fixture\n\n" +
		"## Phases and Verification\n" +
		"### phase_slug: `" + slug + "`\n" +
		"- story_id: " + storyID + "\n" +
		"- status: checked\n" +
		"- goal: prove legacy check entries rebuild with empty mode\n" +
		"- depends_on: none\n\n" +
		"## Progress\n\n## Decisions\n\n" +
		"## Validation\n" +
		"- `2026-07-02T10:05:00Z` — check. verdict: `APPROVED`. check: `" + checkID + "`. run: `" + runID + "`. phase: `" + slug + "`. judge: `independent` (sonnet-5).\n"
	writeFile(t, filepath.Join("docs", "plans", "active", "legacy-check-fixture.md"), planContent)

	db := freshDB(t)
	result, err := RebuildFromMarkdown(db)
	if err != nil {
		t.Fatalf("RebuildFromMarkdown: %v", err)
	}
	if result.Checks != 1 {
		t.Fatalf("RebuildResult.Checks = %d, want 1", result.Checks)
	}
	var mode string
	if err := db.QueryRow(`SELECT mode FROM checks WHERE id = ?`, checkID).Scan(&mode); err != nil {
		t.Fatalf("query rebuilt check: %v", err)
	}
	if mode != "" {
		t.Fatalf("rebuilt legacy check mode = %q, want empty string", mode)
	}
}

// TestRebuildFromMarkdownListItemStoryFields proves R17: a phase block
// under the `### phase_slug:` heading form (plan_query.go's other
// supported shape) writes its scalar fields as markdown list items —
// `- story_id: ...` — the form to-plan actually emits for this shape.
// planFieldValue's regex anchors on `^[ \t]*story_id:` and so misses the
// leading `- `, and rebuildStoriesFromPlan silently drops the story
// through its malformed-block continue branch.
func TestRebuildFromMarkdownListItemStoryFields(t *testing.T) {
	chdirFixture(t)
	storyID := ulid.Make().String()
	const slug = "list-item-fixture"
	planContent := "---\n" +
		"id: " + ulid.Make().String() + "\n" +
		"type: plan\n" +
		"status: active\n" +
		"---\n\n" +
		"# Plan: List-item fixture\n\n" +
		"## Phases and Verification\n" +
		"### phase_slug: `" + slug + "`\n" +
		"- story_id: " + storyID + "\n" +
		"- status: planned\n" +
		"- goal: prove list-item fields rebuild\n" +
		"- depends_on: none\n\n" +
		"## Progress\n\n## Decisions\n\n## Validation\n"
	writeFile(t, filepath.Join("docs", "plans", "active", "list-item-fixture.md"), planContent)

	db := freshDB(t)
	result, err := RebuildFromMarkdown(db)
	if err != nil {
		t.Fatalf("RebuildFromMarkdown: %v", err)
	}
	if result.Stories != 1 {
		t.Fatalf("RebuildResult.Stories = %d, want 1", result.Stories)
	}
	exists, err := storyRowExistsByID(db, storyID)
	if err != nil {
		t.Fatalf("storyRowExistsByID: %v", err)
	}
	if !exists {
		t.Fatalf("story %s not found after rebuild: list-item field form was dropped", storyID)
	}
}

// TestRebuildFromMarkdownEmDashHeadingLegacyFields proves R18: even once the
// em-dash heading form ("### Phase N — `{slug}`", plan_query.go) is
// discovered, docs/plans/completed/pr60-review-fixes.md's own field names --
// `story:` with a backtick-wrapped value, `depends on:` with trailing prose
// after the value -- are not the `story_id:`/`depends_on:` keys
// planFieldValue matches, so rebuildStoriesFromPlan still takes its
// malformed-block continue branch and drops the story.
func TestRebuildFromMarkdownEmDashHeadingLegacyFields(t *testing.T) {
	chdirFixture(t)
	storyID := ulid.Make().String()
	const slug = "em-dash-fixture"
	planContent := "---\n" +
		"id: " + ulid.Make().String() + "\n" +
		"type: plan\n" +
		"status: active\n" +
		"---\n\n" +
		"# Plan: Em-dash fixture\n\n" +
		"## Phases and Verification\n" +
		"### Phase 1 — `" + slug + "`\n" +
		"- story: `" + storyID + "`\n" +
		"- status: planned\n" +
		"- depends on: `none` — no upstream phase\n" +
		"- goal: prove em-dash heading legacy fields rebuild\n\n" +
		"## Progress\n\n## Decisions\n\n## Validation\n"
	writeFile(t, filepath.Join("docs", "plans", "active", "em-dash-fixture.md"), planContent)

	db := freshDB(t)
	result, err := RebuildFromMarkdown(db)
	if err != nil {
		t.Fatalf("RebuildFromMarkdown: %v", err)
	}
	if result.Stories != 1 {
		t.Fatalf("RebuildResult.Stories = %d, want 1", result.Stories)
	}
	exists, err := storyRowExistsByID(db, storyID)
	if err != nil {
		t.Fatalf("storyRowExistsByID: %v", err)
	}
	if !exists {
		t.Fatalf("story %s not found after rebuild: em-dash heading form / legacy field names were dropped", storyID)
	}
}

type rebuildFixtureRows struct {
	planID, intakeID, lane      string
	storyID, slug, status       string
	runID                       string
	checkID, checkAt            string
	checkMode                   string
	proofLinks                  []domain.ProofLink
	handoffID, handoffAt        string
	nextAction                  string
	openItems                   []string
	traceAt                     string
	decisionAt                  string
	decisionPhase, decisionTask string
}

func seedRebuildFixtureRows(t *testing.T, db *sql.DB, f rebuildFixtureRows) {
	t.Helper()
	const at = "2026-07-01T09:00:00Z"

	if _, err := db.Exec(
		`INSERT INTO stories (id, slug, goal, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		f.storyID, f.slug, "exercise the round trip", f.status, at,
	); err != nil {
		t.Fatalf("seed story: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO runs (id, story_slug, artifact_path, created_at) VALUES (?, ?, ?, ?)`,
		f.runID, f.slug, "", at,
	); err != nil {
		t.Fatalf("seed run: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO intakes (id, type, summary, lane, plan_path, plan_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		f.intakeID, "reconstructed", "seeded for round-trip fixture", f.lane, "", f.planID, at,
	); err != nil {
		t.Fatalf("seed intake: %v", err)
	}

	proofLinksAny := make([]map[string]string, len(f.proofLinks))
	for i, pl := range f.proofLinks {
		proofLinksAny[i] = map[string]string{"command": pl.Command, "output_ref": pl.OutputRef}
	}
	proofLinksJSON, err := json.Marshal(proofLinksAny)
	if err != nil {
		t.Fatalf("marshal proof_links: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO checks (id, run_id, verdict, judge, judge_model, proof_links, mode, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		f.checkID, f.runID, domain.VerdictApproved, domain.JudgeIndependent, "sonnet-5", string(proofLinksJSON), f.checkMode, f.checkAt,
	); err != nil {
		t.Fatalf("seed check: %v", err)
	}

	anchors := map[string]any{
		"latest_run_id": f.runID, "latest_check_id": f.checkID,
		"exact_next_action": f.nextAction, "open_items": f.openItems,
	}
	anchorsJSON, err := json.Marshal(anchors)
	if err != nil {
		t.Fatalf("marshal anchors: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO handoffs (id, run_id, check_id, anchors, created_at) VALUES (?, ?, ?, ?, ?)`,
		f.handoffID, f.runID, f.checkID, string(anchorsJSON), f.handoffAt,
	); err != nil {
		t.Fatalf("seed handoff: %v", err)
	}

	if _, err := db.Exec(
		`INSERT INTO traces (id, run_id, wave, summary, task, task_status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		ulid.Make().String(), f.runID, 1, "seeded the fixture", "seed the fixture", domain.TaskStatusDone, f.traceAt,
	); err != nil {
		t.Fatalf("seed trace: %v", err)
	}

	if _, err := db.Exec(
		`INSERT INTO decisions (id, phase, task, decision, rationale, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		ulid.Make().String(), f.decisionPhase, f.decisionTask, "use a hand-seeded fixture", "keeps the round trip exact", f.decisionAt,
	); err != nil {
		t.Fatalf("seed decision: %v", err)
	}
}

type rebuildStorySnapshot struct {
	slug, goal, status, dependsOn string
}

type rebuildRunSnapshot struct {
	storySlug, artifactPath string
}

type rebuildCheckSnapshot struct {
	id, runID, verdict, judge, judgeModel, proofLinksJSON, mode string
}

type rebuildHandoffSnapshot struct {
	id, runID, checkID, anchorsJSON string
}

type rebuildTraceSnapshot struct {
	runID, summary, task, taskStatus, createdAt string
	wave                                        int
}

type rebuildDecisionSnapshot struct {
	phase, task, decision, rationale, createdAt string
}

type rebuildIntakeSnapshot struct {
	id, lane, planID string
}

type rebuildStateSnapshot struct {
	story    rebuildStorySnapshot
	run      rebuildRunSnapshot
	check    rebuildCheckSnapshot
	handoff  rebuildHandoffSnapshot
	trace    rebuildTraceSnapshot
	decision rebuildDecisionSnapshot
	intake   rebuildIntakeSnapshot
}

func snapshotRebuildState(t *testing.T, db *sql.DB, storyID, runID, checkID, handoffID, slug string) rebuildStateSnapshot {
	t.Helper()
	var snap rebuildStateSnapshot

	var dependsOn sql.NullString
	if err := db.QueryRow(`SELECT slug, goal, status, depends_on FROM stories WHERE id = ?`, storyID).
		Scan(&snap.story.slug, &snap.story.goal, &snap.story.status, &dependsOn); err != nil {
		t.Fatalf("snapshot story: %v", err)
	}
	snap.story.dependsOn = dependsOn.String

	if err := db.QueryRow(`SELECT story_slug, artifact_path FROM runs WHERE id = ?`, runID).
		Scan(&snap.run.storySlug, &snap.run.artifactPath); err != nil {
		t.Fatalf("snapshot run: %v", err)
	}

	snap.check.id = checkID
	if err := db.QueryRow(`SELECT run_id, verdict, judge, judge_model, proof_links, mode FROM checks WHERE id = ?`, checkID).
		Scan(&snap.check.runID, &snap.check.verdict, &snap.check.judge, &snap.check.judgeModel, &snap.check.proofLinksJSON, &snap.check.mode); err != nil {
		t.Fatalf("snapshot check: %v", err)
	}

	snap.handoff.id = handoffID
	var handoffCheckID sql.NullString
	if err := db.QueryRow(`SELECT run_id, check_id, anchors FROM handoffs WHERE id = ?`, handoffID).
		Scan(&snap.handoff.runID, &handoffCheckID, &snap.handoff.anchorsJSON); err != nil {
		t.Fatalf("snapshot handoff: %v", err)
	}
	snap.handoff.checkID = handoffCheckID.String

	if err := db.QueryRow(`SELECT run_id, wave, summary, task, task_status, created_at FROM traces WHERE run_id = ?`, runID).
		Scan(&snap.trace.runID, &snap.trace.wave, &snap.trace.summary, &snap.trace.task, &snap.trace.taskStatus, &snap.trace.createdAt); err != nil {
		t.Fatalf("snapshot trace: %v", err)
	}

	if err := db.QueryRow(`SELECT phase, task, decision, rationale, created_at FROM decisions WHERE phase = ?`, slug).
		Scan(&snap.decision.phase, &snap.decision.task, &snap.decision.decision, &snap.decision.rationale, &snap.decision.createdAt); err != nil {
		t.Fatalf("snapshot decision: %v", err)
	}

	var planID sql.NullString
	if err := db.QueryRow(`SELECT plan_id FROM runs WHERE id = ?`, runID).Scan(&planID); err != nil {
		t.Fatalf("snapshot run.plan_id: %v", err)
	}
	if err := db.QueryRow(`SELECT id, lane, plan_id FROM intakes WHERE lane = ?`, "high-risk").
		Scan(&snap.intake.id, &snap.intake.lane, &snap.intake.planID); err != nil {
		t.Fatalf("snapshot intake: %v", err)
	}

	return snap
}

func jsonEqual(t *testing.T, a, b string) bool {
	t.Helper()
	var av, bv any
	if err := json.Unmarshal([]byte(a), &av); err != nil {
		t.Fatalf("unmarshal %q: %v", a, err)
	}
	if err := json.Unmarshal([]byte(b), &bv); err != nil {
		t.Fatalf("unmarshal %q: %v", b, err)
	}
	return reflect.DeepEqual(av, bv)
}
