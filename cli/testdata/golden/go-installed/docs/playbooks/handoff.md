# Playbook: handoff

## Purpose

Persist resumable state for a durable initiative in its active plan's `## Current State and Next Action`. Close every cleanly checked phase in that file; only final clean closure compacts the plan, marks it completed, and moves it from `docs/plans/active/{slug}.md` to `docs/plans/completed/{slug}.md`.

## Preconditions

1. Read the plan's `## Current State and Next Action` before changing it.
2. IF context was compacted or summarized since the last full plan read → re-read before trusting any earlier anchor.

## Owned Plan State

- the closing phase's lifecycle `status:` line in `## Phases and Verification`, only to mark `done`
- `## Current State and Next Action` — active phase, lifecycle status, completed work, open items, blockers with unblock conditions, one exact next action
- at final close only: the compaction in step 6

Keep frontmatter `status: active` and the active path for incomplete initiatives and non-final closures. Preserve every phase/task definition; only the closing phase's lifecycle status may change. `## Log` is the sole task execution-status source.

## Steps

1. **Capture repository state** — `git status --short --branch`, recent `git log`, diff stats; summarize branch, uncommitted scope, checkpoint state.
2. **Load plan state** — tails of `## Log` and `## Validation`, plus `## Current State and Next Action`. Stop on any existing phase-status disagreement.
3. **Identify continuity facts** — completed, in progress, blockers and unblock conditions, proof gaps, exact next action.
4. **Incomplete handoff** (phase not closable) — update Current State by hand: honest execution status, `blockers:`/`open_items:`, one exact next action; the plan stays active. Closing requires a matching clean check entry recorded per `check.md`.
5. **Close every cleanly checked phase**:
   - Require: phase `checked`; every prerequisite phase `done`; latest Validation entry `APPROVED` or `APPROVE_WITH_REQUESTS` and gating this phase's work; no unresolved blocker or required proof gap. A `gate`-depth verdict suffices for any non-final phase.
   - Set the phase status and Current State lifecycle status to `done`, and re-read to confirm agreement.
   - IF a dependent phase remains → keep frontmatter `status: active` and the same active path; next action `work full phase {next-phase-slug}`; never complete or move the initiative.
6. **Complete the initiative only after final phase closure**:
   - Before closing the final phase, require every prior phase to be `done` and the final phase's latest Validation verdict to carry mode `full` and `judge: independent` — the initiative's one complete Security, Performance, Architecture, and Code Quality review. A gate-depth verdict never satisfies this. No alternative or early-completion path exists.
   - Close the final phase via step 5. Only after the plan shows every phase `done`, set Current State: `active_phase: none`, `lifecycle_status: completed`, `blockers: none`, `open_items: none`, exact closure next action.
   - **Absorb before leaving `active/`** — append exactly one `decision` Log entry: `absorb: none` (the next session needs no new rule, ADR, or guard), or `absorb: adr <path>` and/or `absorb: guard <path>` and/or `absorb: memory <id>` naming artifacts that already exist. IF a class-of-failure or expensive-to-reverse decision has no ADR or guard → **stop**, no `git mv`; write the ADR or encode the guard first. Write `docs/memory/` only IF `work.md`'s memory triggers already fired.
   - **Compact** — in `## Log`, drop `start` and `done` entries (in a plan migrated from the older format, `task_status=in-progress` and `task_status=DONE` lines); keep `blocked` and `decision` entries byte-identical. In `## Validation`, keep the last entry per phase byte-identical and drop earlier ones. Change nothing else.
   - Set frontmatter `status: completed` and `git mv docs/plans/active/{slug}.md docs/plans/completed/{slug}.md`. Never copy: exactly one file may represent the initiative afterwards. Commit through the hook; it must pass.
   - A completed plan is a run log, not project knowledge: cite the ADR or guard, never the completed plan. It may be deleted later once nobody needs it (if unsure, keep); never delete it in this close.
7. **Verify continuity** — branch captured; phase statuses match the plan's own entries; blockers specific; one exact next action ("continue work" is not one); no sensitive data; no second continuity markdown; exactly one plan file represents the initiative.

## Exit Conditions

- Incomplete: the active plan holds honest execution status, blockers/open items, and one exact next action.
- Non-final closure: that phase is `done`; the plan stays active; the next action names the unlocked dependent phase.
- Final: every phase `done` after a clean `full` check; the plan is compacted and exists only at `docs/plans/completed/{slug}.md` with `status: completed`.
