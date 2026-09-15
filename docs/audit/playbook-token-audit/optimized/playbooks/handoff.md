# Playbook: handoff

## Purpose

Persist resumable state for a durable initiative in its active plan's `## Current State and Next Action`. Close every cleanly checked phase in that file; only final clean closure marks the plan completed and moves it from `docs/plans/active/{slug}.md` to `docs/plans/completed/{slug}.md`.

## Preconditions

1. Read the plan's `## Current State and Next Action` before changing it.
2. IF context was compacted or summarized since the last full plan read → re-read before trusting any earlier anchor.

## Owned Plan State

- the closing phase's lifecycle `status` in `## Phases and Verification`, only to mark `done`
- `## Current State and Next Action` — active phase, lifecycle status, latest anchors from the plan's own Progress/Validation entries, completed work, open items, blockers with unblock conditions, one exact next action

Refresh frontmatter `updated`. Keep `status: active` and the active path for incomplete initiatives and non-final closures. Preserve plan identity, initiative definition, append-only history, and validation evidence. Preserve every phase/task definition; only the closing phase's lifecycle status may change. Append-only `## Progress` is the sole task execution-status source.

## Steps

1. **Capture repository state** — `git status --short --branch`, recent `git log`, diff stats; summarize branch, uncommitted scope, checkpoint state.
2. **Load plan state** — tails of `## Progress`, `## Decisions`, `## Validation`, plus `## Current State and Next Action`; take the latest run/check/handoff anchors from those entries. Stop on any existing phase-status disagreement.
3. **Identify continuity facts** — completed, in progress, blockers and unblock conditions, proof gaps, exact next action. Preserve known blocker taxonomy.
4. **Incomplete handoff** (phase not closable) — update Current State by hand: honest execution status, `blockers:`/`open_items:`, one exact next action; the plan stays active. Closing requires a matching clean check entry recorded per `check.md`.
5. **Close every cleanly checked phase**:
   - Require: phase `checked`; every prerequisite phase `done`; latest Validation entry `APPROVED` or `APPROVE_WITH_REQUESTS` and gating this phase's work; no unresolved blocker or required proof gap. A `gate`-depth verdict suffices for any non-final phase.
   - Set the phase status and Current State lifecycle status to `done`, record final anchors, and re-read to confirm agreement.
   - IF a dependent phase remains → keep frontmatter `status: active` and the same active path; next action `work full phase {next-phase-slug}`; never complete or move the initiative.
6. **Complete the initiative only after final phase closure**:
   - Before closing the final phase, require every prior phase to be `done` and the final phase's latest Validation verdict to carry mode `full` and `judge: independent` — the initiative's one complete Security, Performance, Architecture, and Code Quality review (R4, SDLC token-cache audit). A gate-depth verdict never satisfies this. No alternative or early-completion path exists.
   - Close the final phase via step 5. Only after the plan shows every phase `done`, set Current State: final anchors, `active_phase: none`, `blockers: none`, `open_items: none`, exact closure next action.
   - **Absorb before leaving `active/`** — append exactly one line to `## Decisions`: `absorb: none` (the next session needs no new rule, ADR, or guard), or `absorb: adr <path>` and/or `absorb: guard <path>` and/or `absorb: memory <id>` naming artifacts that already exist. IF a class-of-failure or expensive-to-reverse decision has no ADR or guard → **stop**, no `git mv`; write the ADR or encode the guard first. Write `docs/memory/` only IF `work.md`'s memory triggers already fired.
   - Set frontmatter `status: completed`, refresh `updated`, and `git mv docs/plans/active/{slug}.md docs/plans/completed/{slug}.md`. Never copy: exactly one file may represent the initiative afterwards, carrying the same plan ID.
   - A completed plan is a run log, not project knowledge: cite the ADR or guard, never the completed plan. It may be deleted later once nobody needs its waves, commands, or dead ends (if unsure, keep); never delete it in this close.
7. **Verify continuity** — branch captured; anchors and phase statuses match the plan's own entries; blockers specific; one exact next action ("continue work" is not one); no sensitive data; no second continuity markdown; exactly one plan file represents the initiative.

## Exit Conditions

- Incomplete: the active plan holds honest execution status, anchors, blockers/open items, and one exact next action.
- Non-final closure: that phase is `done`; the plan stays active; the next action names the unlocked dependent phase.
- Final: every phase `done` after a clean `full` check; statuses match; the plan exists only at `docs/plans/completed/{slug}.md` with `status: completed`.
