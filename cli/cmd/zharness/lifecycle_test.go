package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
)

// buildLifecycleTestBinary compiles the real zharness binary once per test
// run into a temp file, so the lifecycle test below drives the actual CLI
// surface (flag parsing, JSON encoding, exit codes) rather than internal
// Go calls — the shape cli-stale-drift-PLAN.md's T4 risk note asks for
// ("keep it runnable against an installed binary, not only go test"), and
// what cli-release can reuse as its own smoke test.
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

// runZ runs bin with args in dir, failing the test on a non-zero exit
// unless allowExitCode says otherwise (validate legitimately exits 1 on
// any real finding, which this scratch fixture is built to avoid — but
// callers that expect a specific finding pass it explicitly).
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

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
}

// TestLifecycle_ScratchDirFullChain drives init->intake->story->run
// registration->trace add->check record->handoff record->resume/validate/
// audit through the real built binary on a scratch dir, per
// cli-stale-drift-PLAN.md's T4.
//
// Known, deliberate scope limit (user-confirmed at gate time, see
// .kit/runs/work/20260718-1718-cli-stale-drift.md): no command in the
// current CLI surface ever writes meta.current_phase or transitions
// story.status past "planned" — STATE.md's Writer/Reader Ownership table
// documents `to-plan`/`work`/`check record` moving these pointers, but no
// production code does it (only story creation and legacy import ever
// write status; run registration, even work's own documented two-line
// changeset, never touches current_phase). Since resume's readiness only
// reads Position.Status, and Position.Status only populates when
// current_phase is set, readiness stays "clean" through this entire
// lifecycle — it never reaches PLAN.md's literal "in-progress" or
// "checked" through the CLI as it stands today. Tracked as backlog item
// 01KXTBG4JZYTW528Y5XZQK8FEH, not this phase's fix — this test asserts
// the real, current behavior rather than the aspirational one.
func TestLifecycle_ScratchDirFullChain(t *testing.T) {
	bin := buildLifecycleTestBinary(t)
	root := t.TempDir()

	// --- init ---
	initOut := runZ(t, bin, root, 0, "init", "--json")
	var initResp struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(initOut, &initResp); err != nil {
		t.Fatalf("unmarshal init: %v (%s)", err, initOut)
	}
	if initResp.Status != "created" {
		t.Fatalf("init status = %q, want created", initResp.Status)
	}

	resumeOut := runZ(t, bin, root, 0, "resume", "--json")
	var resume0 struct {
		Readiness string `json:"readiness"`
		Drift     []any  `json:"drift"`
	}
	mustUnmarshal(t, resumeOut, &resume0)
	if resume0.Readiness != "clean" {
		t.Fatalf("readiness after init = %q, want clean (no phase started yet)", resume0.Readiness)
	}
	if len(resume0.Drift) != 0 {
		t.Fatalf("drift after init = %v, want none", resume0.Drift)
	}

	// --- intake ---
	runZ(t, bin, root, 0, "intake", "--type", "new-spec", "--summary", "lifecycle smoke test", "--lane", "tiny", "--json")

	// --- SPEC.md fixture (a valid id makes validate report only the
	// already-known not_yet_implemented SPEC->PLAN gap, not missing_key) ---
	specID := ulid.Make().String()
	writeFile(t, filepath.Join(root, ".kit", "planning", "SPEC.md"), "---\nid: "+specID+"\ntype: spec\nstatus: locked\n---\n\n# SPEC\n")

	// --- story ---
	storyOut := runZ(t, bin, root, 0, "story", "--slug", "lifecycle-smoke", "--goal", "prove full CLI lifecycle", "--json")
	var storyResp struct {
		Status string `json:"status"`
	}
	mustUnmarshal(t, storyOut, &storyResp)
	if storyResp.Status != "planned" {
		t.Fatalf("story status = %q, want planned (creation is the only status this CLI surface ever writes)", storyResp.Status)
	}
	writeFile(t, filepath.Join(root, ".kit", "planning", "phases", "lifecycle-smoke", "lifecycle-smoke-PLAN.md"), "# Plan: lifecycle-smoke\n")

	// --- run registration (same two-line changeset mechanism `work` uses —
	// no dedicated "run create" command exists) ---
	runID := ulid.Make().String()
	planID := ulid.Make().String()
	runArtifact := filepath.Join(root, ".kit", "runs", "work", "lifecycle-smoke.md")
	writeFile(t, runArtifact, "---\nid: "+runID+"\ntype: run\nphase: lifecycle-smoke\nplan_id: "+planID+"\n---\n\n# COOK RUN\n")

	now := time.Now().UTC().Format(time.RFC3339)
	runChangeset := filepath.Join(root, ".kit", "changesets", ulid.Make().String()+".changeset.jsonl")
	writeFile(t, runChangeset,
		`{"op":"create","entity":"run","id":"`+runID+`","fields":{"story_slug":"lifecycle-smoke","plan_id":"`+planID+`","artifact_path":".kit/runs/work/lifecycle-smoke.md","created_at":"`+now+`"},"at":"`+now+`"}`+"\n"+
			`{"op":"update","entity":"meta","id":"meta","fields":{"latest_run_id":"`+runID+`"},"at":"`+now+`"}`+"\n")
	runZ(t, bin, root, 0, "db", "changeset", "apply", runChangeset, "--json")

	resumeOut = runZ(t, bin, root, 0, "resume", "--json")
	var resume1 struct {
		Readiness string `json:"readiness"`
		Drift     []any  `json:"drift"`
	}
	mustUnmarshal(t, resumeOut, &resume1)
	if resume1.Readiness != "clean" {
		t.Fatalf("readiness after run registration = %q, want clean (readiness only reads Position.Status, which only populates when meta.current_phase is set — and nothing in the documented run-registration changeset, work's own included, ever sets it; see backlog 01KXTBG4JZYTW528Y5XZQK8FEH)", resume1.Readiness)
	}
	if len(resume1.Drift) != 0 {
		t.Fatalf("drift after run registration = %v, want none", resume1.Drift)
	}

	// --- trace add ---
	traceOut := runZ(t, bin, root, 0, "trace", "add", "--wave", "1", "--summary", "lifecycle smoke wave 1", "--run-id", runID, "--json")
	var traceResp struct {
		ID string `json:"id"`
	}
	mustUnmarshal(t, traceOut, &traceResp)
	if traceResp.ID == "" {
		t.Fatalf("trace add returned empty id")
	}

	// --- check record ---
	proofLinks := `[{"command":"go test ./...","output_ref":"pass","artifact_path":"` + runArtifact + `"}]`
	checkOut := runZ(t, bin, root, 0, "check", "record", "--verdict", "APPROVED", "--run-id", runID, "--proof-links", proofLinks, "--json")
	var checkResp struct {
		ID      string `json:"id"`
		Verdict string `json:"verdict"`
	}
	mustUnmarshal(t, checkOut, &checkResp)
	if checkResp.Verdict != "APPROVED" {
		t.Fatalf("check record verdict = %q, want APPROVED", checkResp.Verdict)
	}
	// No live command sets latest_check_id (only legacy import) — same
	// pointer-maintenance mechanism check/work already use.
	checkChangeset := filepath.Join(root, ".kit", "changesets", ulid.Make().String()+".changeset.jsonl")
	writeFile(t, checkChangeset, `{"op":"update","entity":"meta","id":"meta","fields":{"latest_check_id":"`+checkResp.ID+`"},"at":"`+now+`"}`+"\n")
	runZ(t, bin, root, 0, "db", "changeset", "apply", checkChangeset, "--json")
	writeFile(t, filepath.Join(root, ".kit", "reports", "check", "lifecycle-smoke.md"),
		"---\nid: "+checkResp.ID+"\ntype: check\nrun_id: "+runID+"\n---\n\n# CHECK REPORT\n")

	// --- handoff record ---
	handoffOut := runZ(t, bin, root, 0, "handoff", "record", "--run-id", runID, "--check-id", checkResp.ID, "--open-items", "[]", "--json")
	var handoffResp struct {
		ID string `json:"id"`
	}
	mustUnmarshal(t, handoffOut, &handoffResp)
	if handoffResp.ID == "" {
		t.Fatalf("handoff record returned empty id")
	}

	// --- final resume: latest_handoff_id surfaces, still zero unexpected
	// drift, readiness still clean (same current_phase gap — not a bug this
	// test should paper over) ---
	resumeOut = runZ(t, bin, root, 0, "resume", "--json")
	var resumeFinal struct {
		Readiness       string `json:"readiness"`
		Drift           []any  `json:"drift"`
		LatestHandoffID string `json:"latest_handoff_id"`
	}
	mustUnmarshal(t, resumeOut, &resumeFinal)
	if resumeFinal.LatestHandoffID != handoffResp.ID {
		t.Fatalf("latest_handoff_id = %q, want %q", resumeFinal.LatestHandoffID, handoffResp.ID)
	}
	if len(resumeFinal.Drift) != 0 {
		t.Fatalf("final drift = %v, want none", resumeFinal.Drift)
	}
	if resumeFinal.Readiness != "clean" {
		t.Fatalf("final readiness = %q, want clean (current_phase/story.status transition gap, backlog 01KXTBG4JZYTW528Y5XZQK8FEH — readiness never reflects \"in-progress\" or \"checked\" through the CLI as it stands today)", resumeFinal.Readiness)
	}

	// --- validate: the fixture is built to be genuinely clean modulo the
	// one already-known, already-accepted not_yet_implemented gap ---
	validateOut := runZ(t, bin, root, 0, "validate", "--json")
	var validateResp struct {
		Valid    bool `json:"valid"`
		Findings []struct {
			Link  string `json:"link"`
			Issue string `json:"issue"`
		} `json:"findings"`
	}
	mustUnmarshal(t, validateOut, &validateResp)
	if !validateResp.Valid {
		t.Fatalf("validate valid = false, findings = %+v, want true", validateResp.Findings)
	}
	if len(validateResp.Findings) != 1 || validateResp.Findings[0].Issue != "not_yet_implemented" {
		t.Fatalf("validate findings = %+v, want exactly one not_yet_implemented (SPEC->PLAN, the known gap)", validateResp.Findings)
	}

	// --- audit: pointer_drift empty, contract_violations only the same
	// known not_yet_implemented gap, entropy reflects exactly that ---
	auditOut := runZ(t, bin, root, 0, "audit", "--json")
	var auditResp struct {
		PointerDrift       []any `json:"pointer_drift"`
		ContractViolations []struct {
			Issue string `json:"issue"`
		} `json:"contract_violations"`
		UnlinkedProofs []any `json:"unlinked_proofs"`
	}
	mustUnmarshal(t, auditOut, &auditResp)
	if len(auditResp.PointerDrift) != 0 {
		t.Fatalf("audit pointer_drift = %v, want none", auditResp.PointerDrift)
	}
	if len(auditResp.UnlinkedProofs) != 0 {
		t.Fatalf("audit unlinked_proofs = %v, want none (proof link points at a real, existing file)", auditResp.UnlinkedProofs)
	}
	if len(auditResp.ContractViolations) != 1 || auditResp.ContractViolations[0].Issue != "not_yet_implemented" {
		t.Fatalf("audit contract_violations = %+v, want exactly one not_yet_implemented", auditResp.ContractViolations)
	}
}

func mustUnmarshal(t *testing.T, data []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("unmarshal %s: %v", data, err)
	}
}
