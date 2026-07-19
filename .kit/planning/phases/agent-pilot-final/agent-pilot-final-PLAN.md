# Plan: agent-pilot-final

Phase: agent-pilot-final
Status: ready
Wave Count: 2
Execution Owner: work
Updated At: 2026-07-19

## Goal
Produce the final uncontaminated R9 proof against the autonomous-entry-parity release.

## Inputs
- Installed cli/v0.4.1
- Fresh isolated HOME/CODEX_HOME containing auth.json only
- Task-only protocol requiring exactly six product/test files (inventory package + two test modules), with no harness/mode/threshold hint
- Prior evidence doc format

## Wave 1
### T1 — Establish uncontaminated target and runtime
- type: test
- inputs: v0.4.1 release
- touches: fresh scratch target + isolated temporary HOME only
- avoid: target-local prompt/transcript files, global Codex skills/config/rules/memories
- steps:
  1. Fresh target: `git init`, `zharness init`; confirm only scaffold output exists.
  2. Create auth-only HOME/CODEX_HOME outside target; use `--ignore-user-config --ignore-rules --ephemeral`.
  3. Keep prompt and transcript outside target.
  4. Verify `zharness --version` is 0.4.1 and `resume --json` is clean.
- expected outputs: cold target + isolated runtime controls recorded
- verification: file listing, version, resume output
- stop if: runtime cannot authenticate in isolated HOME
- escalate to: user clarification (BLOCKED, do not fabricate)

### T2 — Execute and grade pilot
- type: test
- inputs: T1
- touches: scratch target only
- avoid: any human response or mechanics coaching after launch
- steps:
  1. Run Codex once with root-AGENTS + task-only protocol. Task requires: `inventory/__init__.py`, `inventory/models.py` (frozen Item dataclass with Decimal price), `inventory/store.py` (replace-by-SKU + adjust/unknown-SKU error + deterministic items), `inventory/report.py` (Decimal total + low-stock sorted by SKU), `test_store.py`, `test_report.py`.
  2. Wait for natural process exit; do not touch target mid-run.
  3. Run task tests, `zharness validate --json`, `resume --json`, `audit --json`, and `query phases/artifacts/check --latest` independently.
  4. Grep transcript for all global/local SKILL.md paths and procedural questions.
  5. Confirm lifecycle evidence: intake exists, story exists, RUN is `mode: full` and registered, trace IDs are linked, CHECK is registered and latest pointer matches.
- expected outputs: six requested files + tests pass; full lifecycle present; `valid:true`; zero pointer drift/unlinked proofs; zero SKILL reads; no mechanics question
- verification: verbatim outputs + transcript command trail
- stop if: any conjunct fails
- escalate to: further finding cycle (no hotfix)

## Wave 2
### T3 — Publish final evidence and close R9
- type: docs
- inputs: T2 transcript/outputs + excluded-attempt summaries
- touches: `docs/workflow-harness/pilot-evidence/2026-07-19-agent-pilot-final.md`
- avoid: cli/skills/.kit edits in this phase
- steps:
  1. Document isolation controls and unchanged protocol.
  2. List excluded attempts: account SKILL contamination; flags-only SKILL contamination; isolated pre-v0.4.0 procedural stop (#39).
  3. Preserve final command order and quote validation/test outputs.
  4. State literal GO/NO-GO against all R9 clauses: intake/story/trace/check lifecycle, zero SKILL reads, and `validate:true`.
- expected outputs: durable final evidence doc with a quotable R9 verdict
- verification: path exists; all claimed commands appear in transcript or this session's output; linked issues #38/#39/#40 exist
- stop if: evidence cannot support every verdict clause
- escalate to: to-plan phase if NO-GO; initiative close-out if GO

## Risks / Watch-fors
- Do not count a dry pilot from Phase 9 as final evidence; Phase 10 needs a new target against the published release.
- Transcript grep must distinguish textual mentions from actual commands opening SKILL.md; report both honestly.
