# Playbook: handoff

## Purpose

Persist resumable state for a durable initiative by updating `## Current State and Next Action` in its active plan. Every cleanly checked phase is closed in that file; only final clean closure marks the same plan completed and moves it from `docs/plans/active/{slug}.md` to `docs/plans/completed/{slug}.md`.

## Preconditions

1. Read the active plan's `## Current State and Next Action` before changing it. Preserve its identity, initiative definition, planned work, append-only history, and validation evidence.
2. If this session's context was compacted or summarized since the last full read of the plan, re-read before trusting any earlier-read anchor — a summarized turn cannot be assumed to have carried exact state forward.

## Owned Plan State

Handoff updates:

- the closing phase's lifecycle `status` field in `## Phases and Verification`, solely to mark transition to `done`
- `## Current State and Next Action` with active phase, lifecycle status, latest anchors known from the plan's own Progress/Validation entries, completed work, remaining open items, blockers with unblock conditions, and one exact next action

Refresh frontmatter `updated`. Keep `status: active` and the active path for incomplete initiatives and non-final phase closures. Preserve every phase/task definition; only the closing phase lifecycle status may change. Append-only `## Progress` is the sole task execution-status source.

## Steps

1. **Capture repository state** — run `git status --short --branch`, recent `git log`, and diff stats. Summarize branch, uncommitted scope, and checkpoint state.
2. **Load lifecycle state from the plan** — read the tails of append-only `## Progress`, `## Decisions`, and `## Validation` plus `## Current State and Next Action`; take latest run/check/handoff anchors from what those entries themselves record. Stop on any pre-existing phase-status disagreement inside the plan.
3. **Identify continuity facts** — what completed, what remains in progress, blockers and unblock conditions, proof gaps, and the exact next action. Preserve known blocker taxonomy.
4. **Record an incomplete handoff when the phase is not closable** — update `## Current State and Next Action` by hand with honest execution status, unresolved blockers as `blockers:`/`open_items:`, and one exact next action; keep the plan active. Closing a phase requires a matching clean check entry recorded per `check.md`.
5. **Close every cleanly checked phase**:
   - Require the phase to be `checked` in the plan, every prerequisite phase to be `done`, the phase's latest Validation entry to be `APPROVED` or `APPROVE_WITH_REQUESTS` and to gate the phase's work, and no unresolved blocker or required proof gap. A `gate`-depth verdict suffices for any non-final phase.
   - Set that phase's plan status and Current State lifecycle status to `done`, record final anchors, and re-read the plan to confirm internal agreement.
   - If a dependent phase remains, keep frontmatter `status: active` and the same active path. Set the exact next action to `work full phase {next-phase-slug}`; do not complete or move the initiative.
6. **Complete the initiative only after final phase closure**:
   - Before closing the final phase, require every prior phase to be `done` and the final phase's latest Validation verdict to carry mode `full` and `judge: independent` — the initiative's one required complete Security, Performance, Architecture, and Code Quality review (R4 of the SDLC token-cache audit); a gate-depth verdict alone does not satisfy this, no matter how clean. There is no alternative or early-completion condition.
   - Close the final phase through Step 5. Only after the plan shows every phase `done`, update Current State with final anchors, `active_phase: none`, `blockers: none`, `open_items: none`, and an exact closure next action.
   - **Absorb before leaving `active/`** — append to `## Decisions` exactly one absorb line:
     - `absorb: none` when the next session does not need a new rule, ADR, or guard from this run; or
     - `absorb: adr <path>` and/or `absorb: guard <path>` and/or `absorb: memory <id>` naming artifacts that already exist.
     If a class-of-failure or expensive-to-reverse decision exists and no ADR or guard yet records it, **stop**. Do not `git mv`. Write the ADR or encode the guard first. This step does not write `docs/memory/` unless `work.md`'s three memory triggers already fired.
   - Move the plan out of active state yourself: set frontmatter `status: completed`, refresh `updated`, and `git mv docs/plans/active/{slug}.md docs/plans/completed/{slug}.md`. Do not copy — exactly one file may represent the initiative afterwards, and the completed path must contain the same plan ID.
   - A completed plan is a run log, not project knowledge: cite the ADR or guard, never `docs/plans/completed/{slug}.md`. After absorb, it may later be deleted when no one still needs the waves, verify commands, or dead ends; if unsure, keep. Do not delete it as part of this close.
7. **Verify continuity quality** — branch captured, anchors match the plan's own entries, phase statuses match the plan, blockers are specific, one exact next action exists, sensitive data is absent, and exactly one plan file represents the initiative.

## Exit Conditions

- Incomplete phase: the same active plan contains honest execution status, anchors, blockers/open items, and one exact next action.
- Closed non-final phase: a clean closing record changed that phase to `done` in the plan; the same plan remains active and names the unlocked dependent phase as the next action.
- Final: every prior phase was already `done`, the final clean full-mode check and closing record finished the final phase, all statuses match, and the same plan exists only at `docs/plans/completed/{slug}.md` with `status: completed`.

## Anti-Patterns

- Writing a second continuity markdown instead of updating the plan.
- Saying "continue work" without naming the exact next action.
- Closing with a failed/mismatched check or unresolved proof gap.
- Completing the initiative before every phase is `done`.
- Copying the plan to completed while leaving an active duplicate.
- `git mv` to completed without an `absorb:` line in `## Decisions`.
