# Playbook: handoff

## Purpose

Persist resumable state for a durable initiative by updating `## Current State and Next Action` in its active plan and recording a DB handoff row. Every cleanly checked phase is closed with `--close-phase`; only final clean closure marks the same plan completed and moves it from `docs/plans/active/{slug}.md` to `docs/plans/completed/{slug}.md`.

## Preconditions

1. Run `zharness --version`. A `dev` build satisfies the gate; otherwise require version `0.1.0` or newer. If unavailable or stale, print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop.
2. Run `zharness preflight handoff --json` and follow its stop/recovery result. Handoff has one durable mode and does not accept `--mode full`.
3. Read the active plan before changing it. Preserve its identity, initiative definition, planned work, append-only history, and validation evidence.

## Owned Plan State

Handoff updates:

- the closing phase's lifecycle `status` field in `## Phases and Verification`, solely to mirror the public DB transition to `done`
- `## Current State and Next Action` with active phase, lifecycle status, latest run/trace/check/handoff IDs, completed work, remaining open items, blockers with unblock conditions, and one exact next action

Refresh frontmatter `updated`. Keep `status: active` and the active path for incomplete initiatives and non-final phase closures. Preserve every phase/task definition; only the closing phase lifecycle status may change. Append-only `## Progress` is the sole task execution-status source.

## Steps

1. **Capture repository state** — run `git status --short --branch`, recent `git log`, and diff stats. Summarize branch, uncommitted scope, and checkpoint state.
2. **Load lifecycle state** — run `zharness resume --json` and `zharness query phases --json`, then read the active plan's Progress, Decisions, Validation, and Current State. Use DB IDs as lifecycle anchors; stop on pre-existing plan/DB phase-status disagreement.
3. **Identify continuity facts** — what completed, what remains in progress, blockers and unblock conditions, proof gaps, and the exact next action. Preserve known blocker taxonomy.
4. **Record an incomplete handoff when the phase is not closable** — create a concise JSON list from unresolved blockers and next steps, then run `zharness handoff record [--run-id {run-id}] [--check-id {check-id}] --open-items '["...", ...]' --json`. Save the returned handoff ID, update Current State with honest execution status and that ID, and keep the plan active. Do not use `--close-phase` without a matching clean check.
5. **Close every cleanly checked phase**:
   - Require the phase to be `checked` in both DB and plan, every prerequisite phase to be `done`, the latest supplied check to gate the supplied run with `APPROVED` or `APPROVE_WITH_REQUESTS`, and no unresolved phase blocker or required proof gap.
   - Run `zharness handoff record --run-id {run-id} --check-id {check-id} --open-items '[]' --close-phase --json` and save the returned handoff ID.
   - Immediately set that phase's plan status and Current State lifecycle status to `done`, record the final IDs, and rerun `zharness query phases --json` to require DB/plan agreement.
   - If a dependent phase remains, keep frontmatter `status: active` and the same active path. Set the exact next action to `work full phase {next-phase-slug}`; do not complete or move the initiative.
6. **Complete the initiative only after final phase closure**:
   - Before closing the final phase, require every prior phase to be `done` and the final phase to have a clean check for its latest run. There is no alternative or early-completion condition.
   - Close the final phase through Step 5. Only after `zharness query phases --json` shows every phase `done`, update Current State with final IDs, `active_phase: none`, `blockers: none`, `open_items: none`, and an exact closure next action.
   - Set frontmatter `status: completed` and refresh `updated`.
   - Move the same file, without copying or rewriting its history, to `docs/plans/completed/{slug}.md`. Confirm the active path no longer exists and the completed path contains the same plan ID and content.
7. **Verify continuity quality** — branch captured, IDs match DB state, phase statuses match the plan, blockers are specific, one exact next action exists, sensitive data is absent, and exactly one plan file represents the initiative.

## Command Reference

- `zharness --version`
- `zharness preflight handoff --json`
- `zharness resume --json`
- `zharness query phases --json`
- `zharness handoff record [--run-id {run-id}] [--check-id {check-id}] --open-items '[...]' --json`
- `zharness handoff record --run-id {run-id} --check-id {check-id} --open-items '[]' --close-phase --json`

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
