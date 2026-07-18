# Plan: agent-pilot

Phase: agent-pilot
Status: ready
Wave Count: 2
Execution Owner: work
Updated At: 2026-07-18

## Goal
Run and document the second-agent pilot proving R9; publish evidence with findings routed.

## Inputs
- Released `zharness` 0.2.0 installed on the pilot machine
- A working non-Claude agent runtime (Codex CLI or Cursor — pick at start)
- Evidence-format precedent: `docs/workflow-harness/pilot-evidence/2026-07-17-lab-skills-import.md`

## Wave 1
### T1 — Pilot setup
- type: implementation
- inputs: installed 0.2.0
- touches: scratch project dir (outside this repo)
- avoid: this repo's `.kit/`; any coaching material beyond AGENTS.md
- steps:
  1. Pick the runtime (record which and why in the run artifact)
  2. Scratch dir: git init, `zharness init` — confirm scaffold complete (docs, shim, .gitignore)
  3. Prepare the sample task statement (small real feature with a testable behavior) and the pilot protocol prompt (pointer to AGENTS.md + task only)
- expected outputs: ready pilot environment + recorded protocol
- verification: `zharness resume --json` clean on the scratch; task statement contains zero harness-mechanics hints (inspection)
- stop if: no second-agent runtime available
- escalate to: user clarification (phase is blocked, not skippable)

### T2 — Pilot run
- type: test
- inputs: T1 environment
- touches: scratch project only
- avoid: answering harness-mechanics questions (log them as findings instead); touching the agent's session beyond the protocol
- steps:
  1. Second agent executes the task: expected path is read AGENTS.md → AUTHORITY/CONTEXT_RULES → brainstorm/to-plan playbooks as needed → `intake` → `story` → implement → `trace add` → `check record`
  2. Capture the full transcript; log every hesitation, wrong turn, and doc gap live
- expected outputs: completed lifecycle on the scratch (or a documented failure point)
- verification: `zharness validate --json` + `resume --json` on the scratch show a complete, drift-free chain; transcript confirms zero SKILL.md reads and zero mechanics coaching
- stop if: agent hard-stuck on a docs gap after reasonable self-recovery
- escalate to: to-plan phase (playbook-fix mini-phase), then re-run

## Wave 2
### T3 — Evidence doc + findings routing
- type: docs
- inputs: T2 transcript + validate/audit output
- touches: `docs/workflow-harness/pilot-evidence/{YYYY-MM-DD}-second-agent-pilot.md`
- avoid: editing playbooks/CLI in this phase
- steps:
  1. Write the evidence doc in the established format: setup, protocol, verdict (GO/NO-GO against R9), per-finding severity + routing (playbook / CLI / authority-doc), raw command outputs
  2. File GitHub issues for routed findings; link them from the doc
- expected outputs: published evidence, issues filed
- verification: doc exists, linked issues exist, R9 acceptance bullet quotable from the doc
- stop if: verdict is NO-GO
- escalate to: to-plan phase (fix cycle) — the initiative does not close on NO-GO

## Risks / Watch-fors
- Contamination is the main threat to evidence value: any mechanics hint in the prompt invalidates the run — treat the protocol as a hard contract
- A pass with heavy product-side hand-holding is still a pass for R9 (harness-mechanics autonomy is what's under test), but note it in the evidence
