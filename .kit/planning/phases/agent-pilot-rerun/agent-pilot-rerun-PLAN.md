# Plan: agent-pilot-rerun

Phase: agent-pilot-rerun
Status: ready
Wave Count: 2
Execution Owner: work
Updated At: 2026-07-19

## Goal
Re-run the second-agent pilot against the `harness-mode-parity` release; confirm `zharness validate --json` returns `valid:true` on the produced chain, closing R9 with a GO verdict or routing a new distinct finding.

## Inputs
- Released `zharness` from `harness-mode-parity` (Phase 7) installed on the pilot machine
- Codex CLI runtime (same as `agent-pilot`, unless unavailable — then report BLOCKED)
- Evidence-format precedent: `docs/workflow-harness/pilot-evidence/2026-07-19-second-agent-pilot.md`

## Wave 1
### T1 — Pilot setup
- type: implementation
- inputs: `harness-mode-parity` release installed
- touches: fresh scratch project dir (outside this repo)
- avoid: this repo's `.kit/`; any coaching material beyond `AGENTS.md`; reusing the first pilot's scratch dir
- steps:
  1. Install the `harness-mode-parity` release; confirm `zharness --version` shows the new version (not 0.2.0)
  2. Fresh scratch dir: `git init`, `zharness init` — confirm scaffold complete (docs, shim, `.gitignore`)
  3. Prepare a new small-but-real sample task statement (different from `textutils.py`/`slugify`) and the pilot protocol prompt (pointer to `AGENTS.md` + task only)
- expected outputs: ready pilot environment + recorded protocol
- verification: `zharness --version` shows the new release; `zharness resume --json` clean on the scratch; task statement contains zero harness-mechanics hints (inspection)
- stop if: no second-agent runtime available, or the new release isn't actually installed
- escalate to: user clarification (phase is blocked, not skippable)

### T2 — Pilot run
- type: test
- inputs: T1 environment
- touches: scratch project only
- avoid: answering harness-mechanics questions (log as findings instead); touching the agent's session beyond the protocol; touching `cli/**` or `skills/workflow/**` even if a gap surfaces
- steps:
  1. Second agent executes the task via `work simple` (or equivalent per its own reading of `AGENTS.md`): expected path is read `AGENTS.md` → `AUTHORITY`/`CONTEXT_RULES` → `work.md` playbook → implement → verify
  2. Capture the full transcript; log every hesitation, wrong turn, and doc gap live
  3. Run `zharness validate --json` on the produced chain — this is the literal R9 bar
- expected outputs: completed lifecycle on the scratch, with a `validate --json` result captured verbatim
- verification: `zharness validate --json` shows `valid:true`; transcript confirms zero SKILL.md reads and zero mechanics coaching
- stop if: `validate` still returns `valid:false` for any reason, or the agent hard-stuck on a docs gap after reasonable self-recovery
- escalate to: to-plan phase (further fix cycle), then re-run — do not patch `cli/**` mid-pilot

## Wave 2
### T3 — Evidence doc + R9 closure
- type: docs
- inputs: T2 transcript + `validate`/`resume` output
- touches: `docs/workflow-harness/pilot-evidence/{YYYY-MM-DD}-agent-pilot-rerun.md`
- avoid: editing playbooks/CLI in this phase
- steps:
  1. Write the evidence doc in the established format: setup, protocol, verdict (GO/NO-GO against R9, this time citing the literal `validate --json` output), findings (if any) with severity + routing, raw command outputs
  2. If GO: state R9 formally closed, cross-reference GitHub #38 as resolved
  3. If NO-GO: file the new distinct finding as a GitHub issue, same duplicate-avoidance convention (`gh issue list --search`) as prior findings this initiative
- expected outputs: published evidence doc; R9 status resolved one way or the other
- verification: doc exists, `validate --json` output is quoted verbatim showing `valid:true` (GO case) or the specific failing finding (NO-GO case)
- stop if: n/a — this task always produces a durable verdict
- escalate to: to-plan phase (if NO-GO, another fix cycle) or initiative close-out (if GO)

## Risks / Watch-fors
- Same contamination risk as the first pilot: any mechanics hint in the prompt invalidates the run — treat the protocol as a hard contract
- If GO, this is the point the 8-phase initiative can actually close — don't let phase completion get conflated with initiative close-out; still needs an explicit final wrap (git/handoff) per this session's standing practice
