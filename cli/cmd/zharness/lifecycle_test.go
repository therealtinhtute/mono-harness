package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// buildLifecycleTestBinary compiles the real zharness binary once per test
// run, so the lifecycle fixture drives public flags, JSON, and exit codes.
func buildLifecycleTestBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "zharness")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build zharness: %v\n%s", err, out)
	}
	return bin
}

func runZ(t *testing.T, bin, dir string, allowExitCode int, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("run %v: %v\nstderr: %s", args, err, stderr.String())
	}
	if exitCode != allowExitCode {
		t.Fatalf("run %v: exit %d, want %d\nstdout: %s\nstderr: %s", args, exitCode, allowExitCode, stdout.String(), stderr.String())
	}
	return stdout.Bytes()
}

type lifecyclePhase struct {
	Slug      string  `json:"slug"`
	Status    string  `json:"status"`
	DependsOn *string `json:"depends_on"`
}

func TestLifecycle_ScratchDirFullChain(t *testing.T) {
	bin := buildLifecycleTestBinary(t)
	root := t.TempDir()

	initOut := runZ(t, bin, root, 0, "init", "--json")
	var initResp struct {
		Status string `json:"status"`
	}
	mustUnmarshal(t, initOut, &initResp)
	if initResp.Status != "created" {
		t.Fatalf("init status = %q, want created", initResp.Status)
	}

	resumeOut := runZ(t, bin, root, 0, "resume", "--json")
	var resume0 struct {
		Readiness string `json:"readiness"`
		Drift     []any  `json:"drift"`
	}
	mustUnmarshal(t, resumeOut, &resume0)
	if resume0.Readiness != "clean" || len(resume0.Drift) != 0 {
		t.Fatalf("resume after init = %+v, want clean with no drift", resume0)
	}

	preflightOut := runZ(t, bin, root, 0, "preflight", "handoff", "--json")
	var handoffPreflight struct {
		Stage     string `json:"stage"`
		Mode      string `json:"mode"`
		Readiness string `json:"readiness"`
		Stop      any    `json:"stop"`
	}
	mustUnmarshal(t, preflightOut, &handoffPreflight)
	if handoffPreflight.Stage != "handoff" || handoffPreflight.Mode != "durable" || handoffPreflight.Readiness != "ready" || handoffPreflight.Stop != nil {
		t.Fatalf("handoff preflight = %+v, want accepted durable/ready invocation", handoffPreflight)
	}

	planPathRel := "docs/plans/active/lifecycle-smoke.md"
	planPath := filepath.Join(root, filepath.FromSlash(planPathRel))
	runZ(t, bin, root, 0, "scaffold", "plan", "--path", planPathRel, "--json")

	planIDOut := runZ(t, bin, root, 0, "id", "--json")
	var planIDResp struct {
		ID string `json:"id"`
	}
	mustUnmarshal(t, planIDOut, &planIDResp)
	if planIDResp.ID == "" {
		t.Fatal("id returned empty plan id")
	}

	intakeOut := runZ(t, bin, root, 0, "intake", "--type", "new-spec", "--summary", "two-phase lifecycle smoke test", "--lane", "tiny", "--plan-path", planPathRel, "--json")
	var intakeResp struct {
		ID string `json:"id"`
	}
	mustUnmarshal(t, intakeOut, &intakeResp)
	if intakeResp.ID == "" {
		t.Fatal("intake returned empty id")
	}

	for _, replacement := range []struct {
		old string
		new string
	}{
		{"{PLAN_ULID}", planIDResp.ID},
		{"{INTAKE_ULID}", intakeResp.ID},
		{"lane: {tiny|normal|high-risk}", "lane: tiny"},
		{"status: {active|completed}", "status: active"},
		{"{YYYY-MM-DD}", "2026-07-27"},
		{"{initiative title}", "Two-phase lifecycle smoke"},
		{"{observable initiative outcome}", "one plan stays synchronized through two dependent phases"},
		{"{checkable success signal}", "phase-a closes before phase-b starts and the final plan moves once"},
		{"{repository source, approved decision, or owner instruction}", "real-binary lifecycle fixture"},
		{"{falsifiable requirement}", "public commands advance both phases through planned, in-progress, checked, and done"},
		{"{authority}", "real-binary lifecycle fixture"},
		{"{explicitly excluded scope}", "legacy lifecycle markdown"},
	} {
		replacePlanAll(t, planPath, replacement.old, replacement.new)
	}
	assertPlanContains(t, planPath,
		"id: "+planIDResp.ID,
		"intake_id: "+intakeResp.ID,
		"approach: not-planned",
		"planning_status: not-planned",
		"phases: none",
		"lifecycle_status: not-planned",
		"exact_next_action: to-plan",
	)
	assertPlanNotContains(t, planPath, "{", "}")

	activePlans, err := filepath.Glob(filepath.Join(root, "docs", "plans", "active", "*.md"))
	if err != nil {
		t.Fatalf("glob active plans: %v", err)
	}
	if len(activePlans) != 1 || activePlans[0] != planPath {
		t.Fatalf("active plans = %v, want exactly %s", activePlans, planPath)
	}

	storyA := createLifecycleStory(t, bin, root, "phase-a", "prove non-final phase closure")
	storyB := createLifecycleStoryWithDependency(t, bin, root, "phase-b", "prove dependent final closure", "phase-a")

	replacePlan(t, planPath, `## Approach and Risks
- approach: not-planned
- constraints:
  - none
- risks:
  - none`, `## Approach and Risks
- approach: execute two dependent phases through public lifecycle commands
- constraints:
  - keep one plan active until the final phase closes
- risks:
  - phase status drift | mitigation: assert DB and plan after every transition`)
	phasePlan := fmt.Sprintf(`## Phases and Verification
<!-- Phase and task definitions are immutable. Only phase lifecycle status changes to mirror DB transitions. Append-only Progress is the sole task execution-status source; task definitions contain no status fields. -->
### Phase 1: phase-a
- phase_slug: phase-a
- story_id: %s
- status: planned
- goal: prove non-final phase closure
- depends_on: none
- waves:
  - wave: 1
    parallel: false
    tasks:
      - task: T1 execute phase-a
        touches: [fixture]
        avoids: [legacy lifecycle markdown]
    checks:
      - command: go test ./...
        expected: pass

### Phase 2: phase-b
- phase_slug: phase-b
- story_id: %s
- status: planned
- goal: prove dependent final closure
- depends_on: phase-a
- waves:
  - wave: 1
    parallel: false
    tasks:
      - task: T1 execute phase-b
        touches: [fixture]
        avoids: [legacy lifecycle markdown]
    checks:
      - command: go test ./...
        expected: pass`, storyA.ID, storyB.ID)
	replacePlan(t, planPath, `## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Do not add task status fields. Append-only Progress is the sole task execution-status source. Only each phase lifecycle status changes to mirror DB transitions: to-plan=planned; work after run create=in-progress; clean durable check=checked; closing handoff=done. Each planned phase records phase_slug, story_id, status, goal, depends_on, waves, tasks, and checks. -->
- planning_status: not-planned
- phases: none`, phasePlan)
	assertPlanPhaseStatus(t, planPath, 1, "phase-a", storyA.ID, "planned")
	assertPlanPhaseStatus(t, planPath, 2, "phase-b", storyB.ID, "planned")
	phases := queryLifecyclePhases(t, bin, root)
	assertLifecyclePhase(t, phases, "phase-a", "planned", "")
	assertLifecyclePhase(t, phases, "phase-b", "planned", "phase-a")

	beforeDB := mustReadFile(t, filepath.Join(root, "harness.db"))
	beforePlan := mustReadFile(t, planPath)
	for _, mode := range []string{"review", "bounded"} {
		out := runZ(t, bin, root, 0, "preflight", "check", "--mode", mode, "--json")
		var view struct {
			Stage     string `json:"stage"`
			Mode      string `json:"mode"`
			Readiness string `json:"readiness"`
			Stop      any    `json:"stop"`
		}
		mustUnmarshal(t, out, &view)
		if view.Stage != "check" || view.Mode != "reduced" || view.Readiness != "ready" || view.Stop != nil {
			t.Fatalf("check %s preflight = %+v, want reduced/ready response-only route", mode, view)
		}
	}
	if afterDB := mustReadFile(t, filepath.Join(root, "harness.db")); !bytes.Equal(afterDB, beforeDB) {
		t.Fatal("check review/bounded preflight mutated harness.db")
	}
	if afterPlan := mustReadFile(t, planPath); !bytes.Equal(afterPlan, beforePlan) {
		t.Fatal("check review/bounded preflight mutated the active plan")
	}

	runA := createLifecycleRun(t, bin, root, "phase-a", planIDResp.ID)
	replacePlan(t, planPath, "- story_id: "+storyA.ID+"\n- status: planned", "- story_id: "+storyA.ID+"\n- status: in-progress")
	replacePlan(t, planPath, `## Current State and Next Action
- active_phase: none
- lifecycle_status: not-planned
- latest_run_id: none
- latest_trace_ids: []
- latest_check_id: none
- latest_handoff_id: none
- blockers: none
- open_items: [to-plan must define stable phases, stories, waves, tasks, and checks]
- exact_next_action: to-plan`, "## Current State and Next Action\n- active_phase: phase-a\n- lifecycle_status: in-progress\n- latest_run_id: "+runA.ID+"\n- latest_trace_ids: []\n- latest_check_id: none\n- latest_handoff_id: none\n- blockers: none\n- open_items: [check phase-a]\n- exact_next_action: check full phase phase-a")
	phases = queryLifecyclePhases(t, bin, root)
	assertLifecyclePhase(t, phases, "phase-a", "in-progress", "")
	assertPlanPhaseStatus(t, planPath, 1, "phase-a", storyA.ID, "in-progress")

	// trace add and check record now write their own `## Progress`/
	// `## Validation` entries (P3, "CLI owns the pen") — the test asserts
	// against that real CLI-authored content instead of splicing its own
	// hand-authored text in ahead of it.
	traceA := createLifecycleTrace(t, bin, root, runA.ID, "phase-a wave complete", "T1", "DONE")
	replacePlan(t, planPath, "- latest_trace_ids: []", "- latest_trace_ids: ["+traceA.ID+"]")
	assertPlanContains(t, planPath, "run: `"+runA.ID+"`", "task_status: `DONE`", "summary: phase-a wave complete")

	requestA := createLifecycleCheck(t, bin, root, runA.ID, "REQUEST_CHANGES", "phase-a finding")
	replacePlan(t, planPath, "- latest_check_id: none", "- latest_check_id: "+requestA.ID)
	replacePlan(t, planPath, "- blockers: none", "- blockers: [phase-a changes requested]")
	replacePlan(t, planPath, "- open_items: [check phase-a]", "- open_items: [resolve phase-a findings]")
	replacePlan(t, planPath, "- exact_next_action: check full phase phase-a", "- exact_next_action: work full phase phase-a")
	phases = queryLifecyclePhases(t, bin, root)
	assertLifecyclePhase(t, phases, "phase-a", "in-progress", "")
	assertPlanPhaseStatus(t, planPath, 1, "phase-a", storyA.ID, "in-progress")
	assertPlanContains(t, planPath, "check: `"+requestA.ID+"`", "verdict: `REQUEST_CHANGES`", "lifecycle_status: in-progress")

	checkA := createLifecycleCheck(t, bin, root, runA.ID, "APPROVED", "phase-a pass")
	replacePlan(t, planPath, "- story_id: "+storyA.ID+"\n- status: in-progress", "- story_id: "+storyA.ID+"\n- status: checked")
	replacePlan(t, planPath, "- lifecycle_status: in-progress", "- lifecycle_status: checked")
	replacePlan(t, planPath, "- latest_check_id: "+requestA.ID, "- latest_check_id: "+checkA.ID)
	replacePlan(t, planPath, "- blockers: [phase-a changes requested]", "- blockers: none")
	replacePlan(t, planPath, "- open_items: [resolve phase-a findings]", "- open_items: [close phase-a and start phase-b]")
	replacePlan(t, planPath, "- exact_next_action: work full phase phase-a", "- exact_next_action: handoff")
	phases = queryLifecyclePhases(t, bin, root)
	assertLifecyclePhase(t, phases, "phase-a", "checked", "")
	assertPlanPhaseStatus(t, planPath, 1, "phase-a", storyA.ID, "checked")
	assertPlanContains(t, planPath, "check: `"+checkA.ID+"`", "verdict: `APPROVED`", "latest_check_id: "+checkA.ID)

	handoffA := createLifecycleHandoff(t, bin, root, runA.ID, checkA.ID)
	replacePlan(t, planPath, "- story_id: "+storyA.ID+"\n- status: checked", "- story_id: "+storyA.ID+"\n- status: done")
	replacePlan(t, planPath, "- lifecycle_status: checked", "- lifecycle_status: done")
	replacePlan(t, planPath, "- latest_handoff_id: none", "- latest_handoff_id: "+handoffA.ID)
	replacePlan(t, planPath, "- open_items: [close phase-a and start phase-b]", "- open_items: [start phase-b]")
	replacePlan(t, planPath, "- exact_next_action: handoff", "- exact_next_action: work full phase phase-b")
	phases = queryLifecyclePhases(t, bin, root)
	assertLifecyclePhase(t, phases, "phase-a", "done", "")
	assertLifecyclePhase(t, phases, "phase-b", "planned", "phase-a")
	assertPlanPhaseStatus(t, planPath, 1, "phase-a", storyA.ID, "done")
	assertPlanPhaseStatus(t, planPath, 2, "phase-b", storyB.ID, "planned")
	assertPlanContains(t, planPath, "status: active", "latest_handoff_id: "+handoffA.ID, "exact_next_action: work full phase phase-b")
	completedPlan := filepath.Join(root, "docs", "plans", "completed", "lifecycle-smoke.md")
	if _, err := os.Stat(completedPlan); !os.IsNotExist(err) {
		t.Fatalf("plan completed before final phase: %v", err)
	}

	runB := createLifecycleRun(t, bin, root, "phase-b", planIDResp.ID)
	replacePlan(t, planPath, "- story_id: "+storyB.ID+"\n- status: planned", "- story_id: "+storyB.ID+"\n- status: in-progress")
	replacePlan(t, planPath, "- active_phase: phase-a", "- active_phase: phase-b")
	replacePlan(t, planPath, "- lifecycle_status: done", "- lifecycle_status: in-progress")
	replacePlan(t, planPath, "- latest_run_id: "+runA.ID, "- latest_run_id: "+runB.ID)
	replacePlan(t, planPath, "- latest_trace_ids: ["+traceA.ID+"]", "- latest_trace_ids: []")
	replacePlan(t, planPath, "- latest_check_id: "+checkA.ID, "- latest_check_id: none")
	replacePlan(t, planPath, "- open_items: [start phase-b]", "- open_items: [check phase-b]")
	replacePlan(t, planPath, "- exact_next_action: work full phase phase-b", "- exact_next_action: check full phase phase-b")
	phases = queryLifecyclePhases(t, bin, root)
	assertLifecyclePhase(t, phases, "phase-a", "done", "")
	assertLifecyclePhase(t, phases, "phase-b", "in-progress", "phase-a")
	assertPlanPhaseStatus(t, planPath, 2, "phase-b", storyB.ID, "in-progress")

	traceB := createLifecycleTrace(t, bin, root, runB.ID, "phase-b wave complete", "T1", "DONE")
	replacePlan(t, planPath, "- latest_trace_ids: []", "- latest_trace_ids: ["+traceB.ID+"]")
	assertPlanContains(t, planPath, "run: `"+runB.ID+"`", "summary: phase-b wave complete")

	checkB := createLifecycleCheck(t, bin, root, runB.ID, "APPROVED", "phase-b pass")
	replacePlan(t, planPath, "- story_id: "+storyB.ID+"\n- status: in-progress", "- story_id: "+storyB.ID+"\n- status: checked")
	replacePlan(t, planPath, "- lifecycle_status: in-progress", "- lifecycle_status: checked")
	replacePlan(t, planPath, "- latest_check_id: none", "- latest_check_id: "+checkB.ID)
	replacePlan(t, planPath, "- open_items: [check phase-b]", "- open_items: [close phase-b]")
	replacePlan(t, planPath, "- exact_next_action: check full phase phase-b", "- exact_next_action: handoff")
	phases = queryLifecyclePhases(t, bin, root)
	assertLifecyclePhase(t, phases, "phase-a", "done", "")
	assertLifecyclePhase(t, phases, "phase-b", "checked", "phase-a")
	assertPlanPhaseStatus(t, planPath, 2, "phase-b", storyB.ID, "checked")
	assertPlanContains(t, planPath, "check: `"+checkB.ID+"`", "verdict: `APPROVED`", "latest_check_id: "+checkB.ID)

	handoffB := createLifecycleHandoff(t, bin, root, runB.ID, checkB.ID)
	replacePlan(t, planPath, "- story_id: "+storyB.ID+"\n- status: checked", "- story_id: "+storyB.ID+"\n- status: done")
	replacePlan(t, planPath, "- active_phase: phase-b", "- active_phase: none")
	replacePlan(t, planPath, "- lifecycle_status: checked", "- lifecycle_status: done")
	replacePlan(t, planPath, "- latest_handoff_id: "+handoffA.ID, "- latest_handoff_id: "+handoffB.ID)
	replacePlan(t, planPath, "- open_items: [close phase-b]", "- open_items: none")
	replacePlan(t, planPath, "- exact_next_action: handoff", "- exact_next_action: initiative complete")
	phases = queryLifecyclePhases(t, bin, root)
	assertLifecyclePhase(t, phases, "phase-a", "done", "")
	assertLifecyclePhase(t, phases, "phase-b", "done", "phase-a")
	assertPlanPhaseStatus(t, planPath, 2, "phase-b", storyB.ID, "done")
	assertPlanContains(t, planPath, "latest_handoff_id: "+handoffB.ID, "lifecycle_status: done")

	replacePlan(t, planPath, "status: active", "status: completed")
	closedPlan := mustReadFile(t, planPath)
	if err := os.MkdirAll(filepath.Dir(completedPlan), 0o755); err != nil {
		t.Fatalf("create completed plans dir: %v", err)
	}
	if err := os.Rename(planPath, completedPlan); err != nil {
		t.Fatalf("move active plan to completed: %v", err)
	}
	movedPlan := mustReadFile(t, completedPlan)
	if !bytes.Equal(movedPlan, closedPlan) {
		t.Fatal("completed plan differs from the active plan that was moved")
	}
	assertPlanContains(t, completedPlan,
		"id: "+planIDResp.ID,
		"status: completed",
		"latest_handoff_id: "+handoffB.ID,
		"Append-only Progress is the sole task execution-status source",
		"task_status: `DONE`",
	)
	assertPlanNotContains(t, completedPlan,
		"\n        status:",
		"\n        status: pending",
		"\n        status: in-progress",
	)

	activePlans, err = filepath.Glob(filepath.Join(root, "docs", "plans", "active", "*.md"))
	if err != nil {
		t.Fatalf("glob active plans after closure: %v", err)
	}
	completedPlans, err := filepath.Glob(filepath.Join(root, "docs", "plans", "completed", "*.md"))
	if err != nil {
		t.Fatalf("glob completed plans: %v", err)
	}
	if len(activePlans) != 0 || len(completedPlans) != 1 || completedPlans[0] != completedPlan {
		t.Fatalf("plans after closure: active=%v completed=%v", activePlans, completedPlans)
	}

	assertNoLegacyLifecycleMarkdown(t, root)

	validateOut := runZ(t, bin, root, 0, "validate", "--json")
	var validateResp struct {
		Valid    bool `json:"valid"`
		Findings []struct {
			Link  string `json:"link"`
			Issue string `json:"issue"`
		} `json:"findings"`
	}
	mustUnmarshal(t, validateOut, &validateResp)
	if !validateResp.Valid || len(validateResp.Findings) != 0 {
		t.Fatalf("validate = %+v, want valid with no findings", validateResp)
	}

	auditOut := runZ(t, bin, root, 0, "audit", "--json")
	var auditResp struct {
		PointerDrift       []any `json:"pointer_drift"`
		ContractViolations []struct {
			Issue string `json:"issue"`
		} `json:"contract_violations"`
		UnlinkedProofs []any `json:"unlinked_proofs"`
	}
	mustUnmarshal(t, auditOut, &auditResp)
	// R15: the scaffolded docs/ARCHITECTURE.md question form is intentionally
	// reported as unanswered until a human answers it — it is the one expected
	// finding in a fresh consumer repository, not drift or proof loss.
	if len(auditResp.PointerDrift) != 0 || len(auditResp.UnlinkedProofs) != 0 || len(auditResp.ContractViolations) != 1 {
		t.Fatalf("audit = %+v, want no drift, no unlinked proofs, and exactly the architecture_elicitation_unanswered violation", auditResp)
	}
	if len(auditResp.ContractViolations) == 1 && auditResp.ContractViolations[0].Issue != "architecture_elicitation_unanswered" {
		t.Fatalf("audit = %+v, want only architecture_elicitation_unanswered", auditResp)
	}
}

func createLifecycleStory(t *testing.T, bin, root, slug, goal string) struct {
	ID     string `json:"id"`
	Status string `json:"status"`
} {
	t.Helper()
	out := runZ(t, bin, root, 0, "story", "--slug", slug, "--goal", goal, "--json")
	var response struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	mustUnmarshal(t, out, &response)
	if response.ID == "" || response.Status != "planned" {
		t.Fatalf("story %s = %+v, want non-empty planned story", slug, response)
	}
	return response
}

func createLifecycleStoryWithDependency(t *testing.T, bin, root, slug, goal, dependsOn string) struct {
	ID     string `json:"id"`
	Status string `json:"status"`
} {
	t.Helper()
	out := runZ(t, bin, root, 0, "story", "--slug", slug, "--goal", goal, "--depends-on", dependsOn, "--json")
	var response struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	mustUnmarshal(t, out, &response)
	if response.ID == "" || response.Status != "planned" {
		t.Fatalf("story %s = %+v, want non-empty planned story", slug, response)
	}
	return response
}

func createLifecycleRun(t *testing.T, bin, root, slug, planID string) struct {
	ID string `json:"id"`
} {
	t.Helper()
	out := runZ(t, bin, root, 0, "run", "create", "--slug", slug, "--plan-id", planID, "--json")
	var response struct {
		ID string `json:"id"`
	}
	mustUnmarshal(t, out, &response)
	if response.ID == "" {
		t.Fatalf("run create for %s returned empty id", slug)
	}
	return response
}

func createLifecycleTrace(t *testing.T, bin, root, runID, summary, task, taskStatus string) struct {
	ID string `json:"id"`
} {
	t.Helper()
	args := []string{"trace", "add", "--wave", "1", "--summary", summary, "--run-id", runID}
	if task != "" {
		args = append(args, "--task", task)
	}
	if taskStatus != "" {
		args = append(args, "--task-status", taskStatus)
	}
	out := runZ(t, bin, root, 0, append(args, "--json")...)
	var response struct {
		ID string `json:"id"`
	}
	mustUnmarshal(t, out, &response)
	if response.ID == "" {
		t.Fatal("trace add returned empty id")
	}
	return response
}

func createLifecycleCheck(t *testing.T, bin, root, runID, verdict, outputRef string) struct {
	ID      string `json:"id"`
	Verdict string `json:"verdict"`
} {
	t.Helper()
	proofLinks := fmt.Sprintf(`[{"command":"true","output_ref":%q}]`, outputRef)
	out := runZ(t, bin, root, 0, "check", "record", "--verdict", verdict, "--run-id", runID, "--judge", "independent", "--judge-model", "test-model", "--proof-links", proofLinks, "--json")
	var response struct {
		ID      string `json:"id"`
		Verdict string `json:"verdict"`
	}
	mustUnmarshal(t, out, &response)
	if response.ID == "" || response.Verdict != verdict {
		t.Fatalf("check record = %+v, want non-empty %s check", response, verdict)
	}
	return response
}

func createLifecycleHandoff(t *testing.T, bin, root, runID, checkID string) struct {
	ID string `json:"id"`
} {
	t.Helper()
	out := runZ(t, bin, root, 0, "handoff", "record", "--run-id", runID, "--check-id", checkID, "--open-items", "[]", "--close-phase", "--json")
	var response struct {
		ID string `json:"id"`
	}
	mustUnmarshal(t, out, &response)
	if response.ID == "" {
		t.Fatal("handoff record returned empty id")
	}
	return response
}

func queryLifecyclePhases(t *testing.T, bin, root string) []lifecyclePhase {
	t.Helper()
	out := runZ(t, bin, root, 0, "query", "phases", "--json")
	var phases []lifecyclePhase
	mustUnmarshal(t, out, &phases)
	return phases
}

func assertLifecyclePhase(t *testing.T, phases []lifecyclePhase, slug, status, dependsOn string) {
	t.Helper()
	for _, phase := range phases {
		if phase.Slug != slug {
			continue
		}
		if phase.Status != status {
			t.Fatalf("phase %s status = %q, want %q", slug, phase.Status, status)
		}
		if dependsOn == "" {
			if phase.DependsOn != nil {
				t.Fatalf("phase %s depends_on = %v, want nil", slug, phase.DependsOn)
			}
		} else if phase.DependsOn == nil || *phase.DependsOn != dependsOn {
			t.Fatalf("phase %s depends_on = %v, want %q", slug, phase.DependsOn, dependsOn)
		}
		return
	}
	t.Fatalf("phase %s not found in %+v", slug, phases)
}

func assertPlanPhaseStatus(t *testing.T, path string, number int, slug, storyID, status string) {
	t.Helper()
	needle := fmt.Sprintf("### Phase %d: %s\n- phase_slug: %s\n- story_id: %s\n- status: %s", number, slug, slug, storyID, status)
	assertPlanContains(t, path, needle)
}

func replacePlan(t *testing.T, path, old, new string) {
	t.Helper()
	data := mustReadFile(t, path)
	if count := bytes.Count(data, []byte(old)); count != 1 {
		t.Fatalf("replace %q in %s: found %d occurrences, want 1", old, path, count)
	}
	data = bytes.Replace(data, []byte(old), []byte(new), 1)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write plan %s: %v", path, err)
	}
}

func replacePlanAll(t *testing.T, path, old, new string) {
	t.Helper()
	data := mustReadFile(t, path)
	if !bytes.Contains(data, []byte(old)) {
		t.Fatalf("replace %q in %s: no occurrences", old, path)
	}
	data = bytes.ReplaceAll(data, []byte(old), []byte(new))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write plan %s: %v", path, err)
	}
}

func assertPlanContains(t *testing.T, path string, phrases ...string) {
	t.Helper()
	data := mustReadFile(t, path)
	for _, phrase := range phrases {
		if !bytes.Contains(data, []byte(phrase)) {
			t.Fatalf("plan %s missing %q\n%s", path, phrase, data)
		}
	}
}

func assertPlanNotContains(t *testing.T, path string, phrases ...string) {
	t.Helper()
	data := mustReadFile(t, path)
	for _, phrase := range phrases {
		if bytes.Contains(data, []byte(phrase)) {
			t.Fatalf("plan %s unexpectedly contains %q\n%s", path, phrase, data)
		}
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func assertNoLegacyLifecycleMarkdown(t *testing.T, root string) {
	t.Helper()
	for _, legacyPath := range []string{
		".kit/planning",
		".kit/plans",
		".kit/runs",
		".kit/reports",
		".kit/HANDOFF.md",
		".kit/docs",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(legacyPath))); err == nil {
			t.Errorf("legacy lifecycle path exists: %s", legacyPath)
		} else if !os.IsNotExist(err) {
			t.Errorf("stat legacy lifecycle path %s: %v", legacyPath, err)
		}
	}
	if err := filepath.Walk(filepath.Join(root, ".kit"), func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			t.Errorf("unexpected lifecycle markdown under .kit: %s", path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk .kit: %v", err)
	}
}

func mustUnmarshal(t *testing.T, data []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("unmarshal %s: %v", data, err)
	}
}
