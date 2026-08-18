package application

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	db, changesetDir := freshDB(t)

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
		formatCheckValidationEntry(checkAt, checkID, slug, runID, domain.VerdictApproved, domain.JudgeIndependent, "sonnet-5", proofLinks) + "\n"

	planPath := filepath.Join("docs", "plans", "active", "rebuild-fixture.md")
	writeFile(t, planPath, planContent)

	seedRebuildFixtureRows(t, db, changesetDir, rebuildFixtureRows{
		planID: planID, intakeID: intakeID, lane: lane,
		storyID: storyID, slug: slug, status: domain.StoryChecked,
		runID: runID,
		checkID: checkID, checkAt: checkAt, proofLinks: proofLinks,
		handoffID: handoffID, handoffAt: handoffAt, nextAction: nextAction, openItems: openItems,
		traceAt: traceAt,
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

	db, _ := freshDB(t)
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

type rebuildFixtureRows struct {
	planID, intakeID, lane      string
	storyID, slug, status       string
	runID                       string
	checkID, checkAt            string
	proofLinks                  []domain.ProofLink
	handoffID, handoffAt        string
	nextAction                  string
	openItems                   []string
	traceAt                     string
	decisionAt                  string
	decisionPhase, decisionTask string
}

func seedRebuildFixtureRows(t *testing.T, db *sql.DB, changesetDir string, f rebuildFixtureRows) {
	t.Helper()
	const at = "2026-07-01T09:00:00Z"

	mustApply(t, db, changesetDir, "story", f.storyID, map[string]any{
		"slug": f.slug, "goal": "exercise the round trip", "status": f.status, "created_at": at,
	}, at)
	mustApply(t, db, changesetDir, "run", f.runID, map[string]any{
		"story_slug": f.slug, "artifact_path": "", "created_at": at,
	}, at)
	mustApply(t, db, changesetDir, "intake", f.intakeID, map[string]any{
		"type": "reconstructed", "summary": "seeded for round-trip fixture", "lane": f.lane, "plan_id": f.planID, "created_at": at,
	}, at)

	proofLinksAny := make([]any, len(f.proofLinks))
	for i, pl := range f.proofLinks {
		proofLinksAny[i] = map[string]any{"command": pl.Command, "output_ref": pl.OutputRef}
	}
	mustApply(t, db, changesetDir, "check", f.checkID, map[string]any{
		"run_id": f.runID, "verdict": domain.VerdictApproved, "judge": domain.JudgeIndependent, "judge_model": "sonnet-5",
		"proof_links": proofLinksAny, "created_at": f.checkAt,
	}, f.checkAt)

	anchors := map[string]any{
		"latest_run_id": f.runID, "latest_check_id": f.checkID,
		"exact_next_action": f.nextAction, "open_items": f.openItems,
	}
	mustApply(t, db, changesetDir, "handoff", f.handoffID, map[string]any{
		"run_id": f.runID, "check_id": f.checkID, "anchors": anchors, "created_at": f.handoffAt,
	}, f.handoffAt)

	mustApply(t, db, changesetDir, "trace", ulid.Make().String(), map[string]any{
		"run_id": f.runID, "wave": 1, "summary": "seeded the fixture", "task": "seed the fixture", "task_status": domain.TaskStatusDone, "created_at": f.traceAt,
	}, f.traceAt)

	mustApply(t, db, changesetDir, "decision", ulid.Make().String(), map[string]any{
		"phase": f.decisionPhase, "task": f.decisionTask, "decision": "use a hand-seeded fixture", "rationale": "keeps the round trip exact", "created_at": f.decisionAt,
	}, f.decisionAt)
}

func mustApply(t *testing.T, db *sql.DB, changesetDir, entity, id string, fields map[string]any, at string) {
	t.Helper()
	if _, _, err := AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "create", Entity: entity, ID: id, Fields: fields, At: at},
	}); err != nil {
		t.Fatalf("seed %s %s: %v", entity, id, err)
	}
}

type rebuildStorySnapshot struct {
	slug, goal, status, dependsOn string
}

type rebuildRunSnapshot struct {
	storySlug, artifactPath string
}

type rebuildCheckSnapshot struct {
	id, runID, verdict, judge, judgeModel, proofLinksJSON string
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
	if err := db.QueryRow(`SELECT run_id, verdict, judge, judge_model, proof_links FROM checks WHERE id = ?`, checkID).
		Scan(&snap.check.runID, &snap.check.verdict, &snap.check.judge, &snap.check.judgeModel, &snap.check.proofLinksJSON); err != nil {
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

var _ = fmt.Sprintf // keep fmt import if unused paths change
