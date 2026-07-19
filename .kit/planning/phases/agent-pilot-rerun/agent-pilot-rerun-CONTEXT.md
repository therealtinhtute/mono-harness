# Context: agent-pilot-rerun

Phase: agent-pilot-rerun
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: low
Expected Proof: e2e, manual-check

## Goal
Re-run the second-agent pilot against the `harness-mode-parity` release to confirm R9's literal acceptance bar is now met: a cold, non-Claude agent completes the lifecycle using only written docs + CLI, and `zharness validate --json` returns `valid:true` on the produced chain — the exact clause that failed in the first pilot.

## Scope Boundary
### Allowed Surfaces
- A fresh scratch target project (outside this repo) for the pilot run
- `docs/workflow-harness/pilot-evidence/` (new dated evidence doc)
- Read-only on everything else

### Forbidden Surfaces
- `cli/**`, `skills/workflow/**` — pilot failures are findings, never hotfixes (same roadmap rule as `agent-pilot`)
- This repo's live `.kit/` — the pilot runs on a scratch target, not on Lab/skills' own state

## Spec Hooks
- R9 acceptance criterion, verbatim: "the non-Claude agent completed intake → story → trace → check record on a sample task without reading any SKILL.md, and `zharness validate --json` passes on the produced chain" — this phase's entire purpose is re-testing the second clause against the `harness-mode-parity` fix

## Locked Decisions
- **Same runtime, same protocol** as `agent-pilot` (Codex CLI, `codex exec`, non-interactive; protocol prompt = pointer to `AGENTS.md` + task statement only, no coaching) — keeps the two pilots comparable.
- **Fresh scratch dir + a different sample task** than the first pilot (not `textutils.py`/`slugify`) — avoids any question of the runtime carrying residual familiarity from the first run; task must still be small-but-real (a genuine feature with testable behavior), per the original pilot's own locked decision.
- **Install the `harness-mode-parity` release** (not v0.2.0) before starting — `zharness --version` must show the new version before the pilot begins.
- **Success is graded on R9's literal text**: cold completion (no SKILL.md reads, no mechanics coaching) AND `zharness validate --json` returns `valid:true`. Both clauses required — same conjunctive bar as the first pilot, no loosening.

## Assumptions
- The `harness-mode-parity` fix is sufficient for the specific chain shape a fresh simple-mode pilot run produces. If this pilot's chain hits a `validate` finding class not covered by that phase's fix, that is a new, distinct finding — route it, do not patch mid-pilot (Forbidden Surfaces).
- Same runtime availability assumption as `agent-pilot`: if Codex CLI (or another second-agent runtime) is unavailable at execution time, report BLOCKED rather than fabricate evidence.

## Canonical Refs
- `docs/workflow-harness/pilot-evidence/2026-07-19-second-agent-pilot.md` (first pilot — protocol, format, and NO-GO verdict this phase re-tests against)
- `.kit/planning/phases/harness-mode-parity/harness-mode-parity-{CONTEXT,PLAN}.md` (the fix this phase verifies)
- GitHub #38 (closed by `harness-mode-parity`, verify closure here)

## Rejected Options
- **Reusing the first pilot's scratch dir/task**: rejected — a fresh dir + new task keeps the second pass genuinely cold, consistent with the original pilot's own rejection of anything that "contaminates" evidence value.
- **Skipping the re-pilot and trusting `harness-mode-parity`'s own unit/integration proof alone**: rejected — R9 is explicitly a pilot-checked requirement (SPEC's Key Decisions: "agent-agnosticism is the initiative's whole point; without a live pass it stays a claim"); a fix's own tests proving the mechanism works is not the same evidence class as a live second-agent pass.

## Deferred Ideas
- Testing a third agent runtime (Cursor) for broader agent-agnosticism confidence — out of this phase, R9 only requires "at least one" second agent

## Escalate If
- The pilot fails on a `validate` finding this phase didn't anticipate → route as a new finding, do not hotfix mid-pilot; loop back to a further `to-plan` mini-phase
- The pilot passes cold-discovery-wise but `validate` still returns `valid:false` for any reason → treat as FAIL for R9 exactly as the first pilot did, per `agent-pilot`'s own Escalate-If rule (still binding here)
