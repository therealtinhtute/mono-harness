package application

import (
	"os"
	"strings"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestCheckRecord(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := createLifecycleRun(t, db, changesetDir, "cli-domain")

	id, path, err := RecordCheck(db, changesetDir, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{
		{Command: "go test ./...", OutputRef: "ok", ArtifactPath: ".kit/runs/work/x.md"},
	})
	if err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}
	if path == "" {
		t.Fatal("changeset path is empty, want a written .jsonl file")
	}

	var verdict, proofLinks string
	if err := db.QueryRow(`SELECT verdict, proof_links FROM checks WHERE id = ?`, id).Scan(&verdict, &proofLinks); err != nil {
		t.Fatalf("query check: %v", err)
	}
	if verdict != domain.VerdictApproved {
		t.Fatalf("verdict = %q, want %q", verdict, domain.VerdictApproved)
	}
	if proofLinks == "[]" || proofLinks == "" {
		t.Fatalf("proof_links = %q, want a non-empty JSON array", proofLinks)
	}

	var latestCheckID string
	if err := db.QueryRow(`SELECT latest_check_id FROM meta LIMIT 1`).Scan(&latestCheckID); err != nil {
		t.Fatalf("query meta: %v", err)
	}
	if latestCheckID != id {
		t.Fatalf("meta.latest_check_id = %q, want %q", latestCheckID, id)
	}
	if got := queryStoryStatus(t, db, "cli-domain"); got != domain.StoryChecked {
		t.Fatalf("story status = %q, want checked", got)
	}
}

func TestCheckRecordRequestChangesAllowsEmptyProofLinks(t *testing.T) {
	db, changesetDir := freshDB(t)
	if _, _, err := CreateStory(db, changesetDir, "cli-domain", "ported domain commands work", ""); err != nil {
		t.Fatalf("CreateStory: %v", err)
	}
	runID, _, err := CreateRun(db, changesetDir, "cli-domain", "", "")
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	if got := queryStoryStatus(t, db, "cli-domain"); got != domain.StoryInProgress {
		t.Fatalf("pre-check story status = %q, want in-progress", got)
	}

	_, _, err = RecordCheck(db, changesetDir, runID, domain.VerdictRequestChanges, domain.JudgeIndependent, "test-model", nil)
	if err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}
	if got := queryStoryStatus(t, db, "cli-domain"); got != domain.StoryInProgress {
		t.Fatalf("story status = %q, want in-progress after REQUEST_CHANGES", got)
	}
}

func TestCheckRecordEmptyProofLinks(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	_, _, err := RecordCheck(db, changesetDir, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", nil)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "empty_proof_links" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: empty_proof_links}", err)
	}
}

func TestCheckRecordInvalidVerdict(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	_, _, err := RecordCheck(db, changesetDir, runID, "MAYBE", domain.JudgeIndependent, "test-model", []domain.ProofLink{{Command: "x"}})
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "invalid_verdict" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: invalid_verdict}", err)
	}
}

func TestCheckRecordUnknownRunID(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := RecordCheck(db, changesetDir, "01HZZZZZZZZZZZZZZZZZZZZZZZ", domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{{Command: "x"}})
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_run_id" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_run_id}", err)
	}
	if got := countRows(t, db, "checks"); got != 0 {
		t.Fatalf("checks rows = %d, want 0", got)
	}
}

// TestCheckRecordRequiresIndependentJudgeForHighRiskLane is the round trip
// for G2 (docs/audit/workflow-harness-ceremony-audit.md/V2): a check whose
// run resolves, via runs.plan_id -> intakes.plan_id, to a high-risk lane
// must be judged independent — same-session is rejected, but the identical
// run still checks out cleanly once the judge is independent.
func TestCheckRecordRequiresIndependentJudgeForHighRiskLane(t *testing.T) {
	db, changesetDir := freshDB(t)
	planID := ulid.Make().String()
	if _, _, err := CreateIntake(db, changesetDir, domain.IntakeHarnessImprovement, "high risk change", domain.LaneHighRisk, "", planID); err != nil {
		t.Fatalf("CreateIntake: %v", err)
	}
	if _, _, err := CreateStory(db, changesetDir, "hr-phase", "goal", ""); err != nil {
		t.Fatalf("CreateStory: %v", err)
	}
	runID, _, err := CreateRun(db, changesetDir, "hr-phase", "", planID)
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	_, _, err = RecordCheck(db, changesetDir, runID, domain.VerdictApproved, domain.JudgeSameSession, "test-model", []domain.ProofLink{
		{Command: "go test ./...", OutputRef: "ok"},
	})
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "independent_judge_required" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: independent_judge_required}", err)
	}
	if got := countRows(t, db, "checks"); got != 0 {
		t.Fatalf("checks rows = %d, want 0 (rejected check must not be written)", got)
	}

	id, _, err := RecordCheck(db, changesetDir, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{
		{Command: "go test ./...", OutputRef: "ok"},
	})
	if err != nil {
		t.Fatalf("RecordCheck with independent judge: %v", err)
	}
	if id == "" {
		t.Fatal("RecordCheck returned empty id")
	}
}

// TestCheckRecordAllowsSameSessionJudgeWhenLaneUnresolvable proves the gate
// is additive: a run with no plan_id trail (the ordinary case until
// playbooks start passing --plan-id) behaves exactly as before this
// feature existed — same-session is accepted.
func TestCheckRecordAllowsSameSessionJudgeWhenLaneUnresolvable(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := createLifecycleRun(t, db, changesetDir, "no-plan-id-trail")

	id, _, err := RecordCheck(db, changesetDir, runID, domain.VerdictApproved, domain.JudgeSameSession, "test-model", []domain.ProofLink{
		{Command: "go test ./...", OutputRef: "ok"},
	})
	if err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}
	if id == "" {
		t.Fatal("RecordCheck returned empty id")
	}
}

// TestCheckRecordAllowsSameSessionJudgeForNonHighRiskLane proves the gate
// is lane-specific: a resolvable but non-high-risk lane does not restrict
// the judge.
func TestCheckRecordAllowsSameSessionJudgeForNonHighRiskLane(t *testing.T) {
	db, changesetDir := freshDB(t)
	planID := ulid.Make().String()
	if _, _, err := CreateIntake(db, changesetDir, domain.IntakeMaintenance, "tiny change", domain.LaneTiny, "", planID); err != nil {
		t.Fatalf("CreateIntake: %v", err)
	}
	if _, _, err := CreateStory(db, changesetDir, "tiny-phase", "goal", ""); err != nil {
		t.Fatalf("CreateStory: %v", err)
	}
	runID, _, err := CreateRun(db, changesetDir, "tiny-phase", "", planID)
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	id, _, err := RecordCheck(db, changesetDir, runID, domain.VerdictApproved, domain.JudgeSameSession, "test-model", []domain.ProofLink{
		{Command: "go test ./...", OutputRef: "ok"},
	})
	if err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}
	if id == "" {
		t.Fatal("RecordCheck returned empty id")
	}
}

// TestCheckRecordWritesPlanValidationEntry proves P3 wave 2 for check
// record: verdict, judge, and each proof link land as a `## Validation`
// entry with nested sub-bullets, in the same operation as the DB write.
func TestCheckRecordWritesPlanValidationEntry(t *testing.T) {
	chdirFixture(t)
	planPath := writeActivePlanFixture(t, "demo")
	db, changesetDir := freshDB(t)
	runID := createLifecycleRun(t, db, changesetDir, "cli-domain")

	id, _, err := RecordCheck(db, changesetDir, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{
		{Command: "go test ./...", OutputRef: "ok"},
		{Command: "bash scripts/verify-doc-links.sh", OutputRef: "0 findings"},
	})
	if err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}

	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "verdict: `APPROVED`") || !strings.Contains(content, "check: `"+id+"`") {
		t.Fatalf("plan Validation missing verdict/check id:\n%s", content)
	}
	if !strings.Contains(content, "cli-domain") {
		t.Fatalf("plan Validation missing phase:\n%s", content)
	}
	if !strings.Contains(content, "  - `go test ./...` → ok") {
		t.Fatalf("plan Validation missing first proof-link sub-bullet:\n%s", content)
	}
	if !strings.Contains(content, "  - `bash scripts/verify-doc-links.sh` → 0 findings") {
		t.Fatalf("plan Validation missing second proof-link sub-bullet:\n%s", content)
	}
	if !strings.Contains(content, "## Decisions\n<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->\n- none") {
		t.Fatalf("plan Decisions section corrupted:\n%s", content)
	}
}

// TestCheckRecordMalformedPlanBlocksDBWrite is check record's version of
// the atomicity proof: since its entry text isn't computable until after
// the id/at mint (AppendNewEntityAndApply's clock-precision design), the
// pre-check (planSectionWritable) must still catch a missing section
// before the DB write, not just after.
func TestCheckRecordMalformedPlanBlocksDBWrite(t *testing.T) {
	chdirFixture(t)
	planPath := writeActivePlanFixture(t, "demo")
	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	corrupted := strings.Replace(string(data), "## Validation", "## Renamed", 1)
	if err := os.WriteFile(planPath, []byte(corrupted), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	db, changesetDir := freshDB(t)
	runID := createLifecycleRun(t, db, changesetDir, "cli-domain")

	_, _, err = RecordCheck(db, changesetDir, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{
		{Command: "go test ./...", OutputRef: "ok"},
	})
	if err == nil {
		t.Fatal("RecordCheck = nil error, want a plan-section-not-found failure")
	}
	if got := countRows(t, db, "checks"); got != 0 {
		t.Fatalf("checks rows = %d, want 0 — DB write must not proceed when the plan can't be written to", got)
	}
	after, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("ReadFile after failed RecordCheck: %v", err)
	}
	if string(after) != corrupted {
		t.Fatal("plan file changed despite the failed write")
	}
}
