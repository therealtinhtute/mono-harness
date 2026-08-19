# Playbook: handoff

## Purpose

Persist resumable state for a durable initiative by updating `## Current State and Next Action` in its active plan and recording a DB handoff row. Every cleanly checked phase is closed with `--close-phase`; only final clean closure marks the same plan completed and moves it from `docs/plans/active/{slug}.md` to `docs/plans/completed/{slug}.md`.

## Preconditions

1. Run `zharness preflight handoff --json`. Missing binary: print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop. Otherwise check its `version` field — a `dev` build satisfies the gate; below MIN_ZHARNESS_VERSION (`0.8.1` — see `skills/workflow/README.md`), print the same message and stop. Then follow its stop/recovery result. Handoff has one durable mode and does not accept `--mode full`. Its `context` field is the source of Step 2's lifecycle position and phase list below — do not call `resume`/`query phases` again to obtain it.
2. If this session's context was compacted or summarized since the last `preflight` call, re-run it before trusting any earlier-read `context` packet or lifecycle ID — a summarized turn cannot be assumed to have carried exact DB state forward.
3. Read the active plan's `## Current State and Next Action` (`zharness query plan --section current-state --json`; if it reports `degraded: true`, read the whole file instead) before changing it. Preserve its identity, initiative definition, planned work, append-only history, and validation evidence.

## Owned Plan State

Handoff updates:

- the closing phase's lifecycle `status` field in `## Phases and Verification`, solely to mirror the public DB transition to `done`
- `## Current State and Next Action` with active phase, lifecycle status, latest run/trace/check/handoff IDs, completed work, remaining open items, blockers with unblock conditions, and one exact next action

Refresh frontmatter `updated`. Keep `status: active` and the active path for incomplete initiatives and non-final phase closures. Preserve every phase/task definition; only the closing phase lifecycle status may change. Append-only `## Progress` is the sole task execution-status source.

## Steps

1. **Capture repository state** — run `git status --short --branch`, recent `git log`, and diff stats. Summarize branch, uncommitted scope, and checkpoint state.
2. **Load lifecycle state** — read `context.position`, `context.phases`, and latest run/check/handoff IDs from the same `preflight handoff --json` response (Preconditions step 1) instead of separately calling `resume`/`query phases`. For the plan's own record, call `zharness query traces --tail 10 --json`, `zharness query decisions --tail 10 --json`, and `zharness query checks --tail 3 --json` for recent Progress/Decisions/Validation, and reuse Preconditions step 3's `query plan --section current-state` result for Current State — do not re-read the whole file for any of this (P3 of the ceremony audit). Use those IDs as lifecycle anchors; stop on pre-existing plan/DB phase-status disagreement.
3. **Identify continuity facts** — what completed, what remains in progress, blockers and unblock conditions, proof gaps, and the exact next action. Preserve known blocker taxonomy.
4. **Record an incomplete handoff when the phase is not closable** — create a concise JSON list from unresolved blockers and next steps, then run `zharness handoff record [--run-id {run-id}] [--check-id {check-id}] --open-items '["...", ...]' --json`. Save the returned handoff ID, update Current State with honest execution status and that ID, and keep the plan active. Do not use `--close-phase` without a matching clean check.
5. **Close every cleanly checked phase**:
   - Require the phase to be `checked` in both DB and plan, every prerequisite phase to be `done`, the latest supplied check to gate the supplied run with `APPROVED` or `APPROVE_WITH_REQUESTS`, and no unresolved phase blocker or required proof gap. A `check gate` verdict is sufficient here for any non-final phase — the complete manual review is not required until the initiative's final phase (step 6).
   - Run `zharness handoff record --run-id {run-id} --check-id {check-id} --open-items '[]' --close-phase --json` and save the returned handoff ID.
   - Immediately set that phase's plan status and Current State lifecycle status to `done`, record the final IDs, and rerun `zharness query phases --json` to require DB/plan agreement.
   - If a dependent phase remains, keep frontmatter `status: active` and the same active path. Set the exact next action to `work full phase {next-phase-slug}`; do not complete or move the initiative.
6. **Complete the initiative only after final phase closure**:
   - Before closing the final phase, require every prior phase to be `done` and the final phase's latest run to have a clean `check full` verdict — the initiative's one required complete Security, Performance, Architecture, and Code Quality review (R4 of the SDLC token-cache audit); a `check gate` verdict alone does not satisfy this, no matter how clean. There is no alternative or early-completion condition. `check record` does not yet persist which mode produced a verdict, so confirm this from the check's own `mode: full` output line (`check.md`'s Output Format) or the session's own record of having invoked `check full`; do not infer it from the verdict alone. If the final phase's latest check was only `gate`, invoke `check full` on it before closing.
   - Close the final phase through Step 5. Only after `zharness query phases --json` shows every phase `done`, update Current State with final IDs, `active_phase: none`, `blockers: none`, `open_items: none`, and an exact closure next action.
   - Run `zharness plan complete --json` (R5, `docs/audit/consumer-adoption-audit.md` D1) — it refuses with `open_phase` unless every phase_slug the plan defines is a done story, otherwise it sets frontmatter `status: completed`, refreshes `updated`, records the transition, and moves the file to `docs/plans/completed/{slug}.md` itself. Do not set `status:` or move the file by hand. Confirm the returned `path` names `docs/plans/completed/{slug}.md`, the active path no longer exists, and the completed path contains the same plan ID.
7. **Verify continuity quality** — branch captured, IDs match DB state, phase statuses match the plan, blockers are specific, one exact next action exists, sensitive data is absent, and exactly one plan file represents the initiative.

## Command Reference

- `zharness preflight handoff --json`
- `zharness query plan --section current-state --json` (Preconditions step 3 / step 2 — the plan's free-text snapshot; not table-backed, so this is the only markdown read this playbook needs)
- `zharness query traces --tail {N} --json`, `zharness query decisions --tail {N} --json`, `zharness query checks --tail {N} --json` (step 2 — recent Progress/Decisions/Validation, in place of reading the file)
- `zharness query phases --json` (steps 5/6 only — post-mutation re-verification; step 2 reads `context.position`/`context.phases` from preflight instead)
- `zharness handoff record [--run-id {run-id}] [--check-id {check-id}] --open-items '[...]' --json`
- `zharness handoff record --run-id {run-id} --check-id {check-id} --open-items '[]' --close-phase --json`
- `zharness plan complete --json` (step 6 only — final phase closure; refuses with `open_phase` unless every phase is done)

## Exit Conditions

- Incomplete phase: one DB handoff row recorded; the same active plan contains honest execution status, anchors, blockers/open items, and one exact next action.
- Closed non-final phase: a clean closing handoff changed that phase to `done` in DB and plan; the same plan remains active and names the unlocked dependent phase as the next action.
- Final: every prior phase was already `done`, the final clean check and closing handoff changed the final phase to `done`, all DB/plan statuses match, and the same plan exists only at `docs/plans/completed/{slug}.md` with `status: completed`.

## Anti-Patterns

- Writing a second continuity markdown instead of updating the plan.
- Saying "continue work" without naming the exact next action.
- Closing with a failed/mismatched check or unresolved proof gap.
- Completing the initiative before every phase is `done`.
- Copying the plan to completed while leaving an active duplicate.
