# Playbook: handoff

Persist resumable state in the active plan's `## Current State and Next Action`, close every cleanly checked phase, and at final closure compact the plan and move `docs/plans/active/{slug}.md` to `docs/plans/completed/{slug}.md`.

## Owned Plan State

The closing phase's `status:` line (only to `done`), `## Current State and Next Action`, and the step 6 compaction. Preserve every phase/task definition. `## Log` is the sole task execution-status source. Re-read the plan after any compaction.

## Steps

1. **Repository state** — `git status --short --branch`, recent `git log`, diff stats.
2. **Plan state** — Current State and the `## Log`/`## Validation` tails; stop on any status disagreement.
3. **Continuity facts** — done, in progress, blockers with unblock conditions, proof gaps.
4. **Incomplete handoff** — update Current State: honest status, `blockers:`, `open_items:`, one exact next action. The plan stays active.
5. Close every cleanly checked phase:
   - Require: phase `checked`, prior phases `done`, latest Validation `APPROVED`/`APPROVE_WITH_REQUESTS` for this phase's work, no open blocker or proof gap.
   - Set the phase and Current State to `done`; re-read to confirm.
   - IF a dependent phase remains → keep frontmatter `status: active` and the same active path; next action `work full phase {next-slug}`.
6. **Complete the initiative**:
   - Before closing the final phase, require every prior phase to be `done` and the final phase's latest Validation entry to carry mode `full` and `judge: independent`. No other path.
   - Close it via step 5. Only after the plan shows every phase `done`, set `active_phase: none`, `lifecycle_status: completed`, `blockers: none`, `open_items: none`.
   - **Absorb** — append exactly one `decision` entry: `absorb: none`, or any of `absorb: adr <path>`, `absorb: guard <path>`, `absorb: memory <id>`, each naming an existing artifact. A hard-to-reverse decision or failure class with no ADR or guard → stop and write it first.
   - **Compact** — in `## Log`, drop `start` and `done` entries (and legacy `task_status=` lines); keep `blocked` and `decision` byte-identical. In `## Validation`, keep the last entry per phase byte-identical and drop the rest.
   - Set frontmatter `status: completed`, then `git mv docs/plans/active/{slug}.md docs/plans/completed/{slug}.md`. Never copy: exactly one file may represent the initiative afterwards. Commit through the hook; it must pass. Cite the ADR or guard, never the completed plan; never delete it in this close.
7. **Verify** — statuses match, blockers are specific, one exact next action ("continue work" is not one), no secrets, no second continuity markdown.
