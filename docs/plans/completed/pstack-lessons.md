---
status: completed
lane: normal
---
# Adopt pstack lessons: entry router, handback proof, guard fixtures

## Goal
- outcome: a user-invoked `/mono-mode` entry skill routes a task to the right stage, `work` records skipped steps with a reason, worker handbacks without source identity or proof output are rejected, an installed pre-commit wrapper detects when it is stale against its template, and the doc-link and skill validators each have a fixture that must trip them.
- success_signal: `bash scripts/test-guards.sh` exits 0 with new failing-fixture cases for the stale wrapper, `verify-doc-links.sh` and `validate-skill.sh` that each fail on their bad fixture; `bash scripts/validate-skill.sh skills/workflow/mono-mode/SKILL.md` exits 0; the four gate commands in AGENTS.md exit 0.
- actors: owner, lead coding agents and worker sub-agents on skill-capable hosts.
- authority: owner request 2026-10-22 to combine the `/mono-mode` port and the pivot audit into an improvement plan; `docs/research/pstack-concepts-audit.md` (roadmap items 1–2 and findings); `docs/evals/failures.md` row "2026-09-16 — an installed pre-commit wrapper can go stale against its own template" (`uncovered`); `docs/research/2026-10-12-lead-peer-contract.md` (handback must match the brief); `skills/workflow/work/references/work-full.md` step 4.
- requirements:
  - R1: `skills/workflow/mono-mode/SKILL.md` is a user-invoked, self-contained router (no hook, no SessionStart injection, no required read of `docs/WORKFLOW.md`): a task → playbook table, and the rule to open a todo list whose first items are the matched playbook's steps copied verbatim, each skipped step kept as `skip: <reason>`. | acceptance: `bash scripts/validate-skill.sh skills/workflow/mono-mode/SKILL.md` exit 0; `grep -c 'skip: <reason>' skills/workflow/mono-mode/SKILL.md` ≥ 1; `git grep -n -e SessionStart -e WORKFLOW.md -- skills/workflow/mono-mode` empty | source: owner request and answers 2026-10-22
  - R1b: four playbooks under `skills/workflow/mono-mode/playbooks/` — `feature.md`, `bug-fix.md`, `investigation.md`, `session-pickup.md` — each a numbered step list that composes existing skills by name (as pstack's playbooks call `how`/`architect`/`interrogate`) and adds only task-type judgment plus a `Reply:` line; stage procedures stay in their own skills. feature: brainstorm lock → work full → check full → git (on request) → handoff. bug-fix: hunt reproduces first → failing test before the fix → work → check. investigation: read-only, think/hunt, cited answer, writes nothing. session-pickup: watzup → reconcile plan with Git → route the rest to a matching playbook without redoing logged `done` work. | acceptance: each file exists and names its composed skills; `git grep -n -E '^[0-9]+\. ' -- skills/workflow/mono-mode/playbooks` lists steps in all four; every skill named resolves to a directory under `skills/workflow/`; `bash scripts/verify-doc-links.sh` exit 0 | source: owner answers 2026-10-22, pstack `poteto-mode/playbooks/`
  - R2: `work-full.md` requires a skipped procedure step or check to be logged as `skip: <reason>`; a silent skip is not allowed. | acceptance: `grep -n 'skip: <reason>' skills/workflow/work/references/work-full.md` matches; validate-skill on work exit 0 | source: owner request 2026-10-22, audit finding on enforcement
  - R3: `work-full.md` step 4 treats a worker handback as a claim and rejects it unless it names the tested source (commit SHA or dirty-diff hash) and the acceptance command output; a rejected handback is re-run or logged `blocked`. | acceptance: `grep -n 'reject' skills/workflow/work/references/work-full.md` shows the rule naming SHA and proof output; validate-skill on work exit 0 | source: lead-peer contract C2, audit finding on multi-agent coordination
  - R4: the installed pre-commit hook compares itself with the template in `scripts/install-git-hooks.sh` and fails with a re-install instruction when they differ. | acceptance: a new `scripts/test-guards.sh` case installs the hook, edits the installed copy, and expects a non-zero commit with the re-install message; `bash scripts/test-guards.sh` exit 0; the 2026-09-16 ledger row's coverage changes from `uncovered` to that case | source: `docs/evals/failures.md` 2026-09-16 row
  - R5: `verify-doc-links.sh` and `validate-skill.sh` each have a committed bad fixture that must make them exit non-zero, run from `scripts/test-guards.sh`. | acceptance: `bash scripts/test-guards.sh` exit 0 including both cases; removing the check that each fixture targets makes its case fail | source: audit roadmap item 2 (pstack invariants-must-trip pattern)
- non-goals:
  - NG1: no Chief-of-Staff coordinator, event-driven orchestration, task queue, SessionStart or other auto-injection hook.
  - NG2: no principle library, no runtime ports or adapters for other harnesses, no new CLI command (lean-sdlc NG1 still applies).
  - NG3: no change to plan grammar, lanes or guard schema beyond the R4 staleness check.
  - NG4: no re-measure of routing cases H01/H04 in this plan (audit roadmap item 3 stays separate).

## Phases and Verification
- approach: five small additions on existing surfaces (one new skill, two `work-full.md` rules, one guard, two fixtures); rejected: porting pstack's hook-injected router and orchestration scripts, which add runtime surface without a demonstrated failure here.
- phase_slug: `main`
  status: done
  - goal: R1, R1b, R2, R3, R4, R5 | depends_on: none
  - surfaces: `skills/workflow/mono-mode/`, `skills/workflow/README.md`, `skills/workflow/work/references/work-full.md`, `scripts/install-git-hooks.sh`, `scripts/test-guards.sh`, guard fixtures, `docs/evals/failures.md` coverage cell | avoided: `cli/` crate, plan grammar, other skills' procedures
  - phase check: `bash scripts/test-guards.sh`; `bash scripts/validate-skill.sh skills/workflow/mono-mode/SKILL.md`; `bash scripts/validate-skill.sh skills/workflow/work/SKILL.md`; `bash scripts/verify-doc-links.sh`; `cd cli && cargo test --test projection_parity`

## Log
- 2026-10-22T08:21Z — main — decision — independent judge 14e6d960 returned REQUEST_CHANGES: TRIP doc-links passed only because `find` failed on missing dirs, and the fixture link had no `/` so it was never a claim. Fixed: the fixture links `../docs/missing-target.md`, the case adds a clean control run and asserts exit 1 plus "broken doc cross-reference", and the validate-skill case asserts "they must match". test-guards 74 passed 0 failed; disabling the findings exit or the name-match message turns its TRIP case red (71 passed, 1 failed each).
- 2026-10-22T08:22Z — closure — decision — absorb: guard scripts/test-guards.sh

## Validation

- 2026-10-22T08:22:00Z — phase `main` — verdict: APPROVED — mode: full
  - `bash scripts/test-guards.sh` — exit 0; guards: 74 passed, 0 failed (STALE, TRIP included)
  - `bash scripts/validate-skill.sh skills/workflow/mono-mode/SKILL.md` — exit 0
  - `bash scripts/validate-skill.sh skills/workflow/work/SKILL.md` — exit 0
  - `bash scripts/verify-doc-links.sh` — exit 0; 0 findings
  - `cd cli && cargo test --test projection_parity` — exit 0; 3 passed
  - requirements: R1 met | R1b met | R2 met | R3 met | R4 met | R5 met
  - scope: judge mutation-tested STALE and both TRIP cases on a temp copy (each turns red); NG1–NG4 respected; prior judge 14e6d960 REQUEST_CHANGES on R5 fixed per Log decision
  - judge: independent
  - judge_model: claude-opus-5-5

## Current State and Next Action
- active_phase: none
- lifecycle_status: completed
- blockers: none
- open_items: none
- exact_next_action: none — initiative closed; H01/H04 re-measure (audit roadmap item 3) needs its own plan
