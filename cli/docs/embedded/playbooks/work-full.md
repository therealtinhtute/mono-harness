# Playbook: work (full mode)

Loaded by `docs/playbooks/work.md` once the resolved mode is `full`. Preconditions, Memory, and Exit Conditions stay there.

## Owned Plan Sections

- the selected phase's lifecycle `status` in `## Phases and Verification`
- `## Progress` — append-only: timestamp, phase, wave, task, task status, exact verification/result, changed surfaces or blocker
- `## Decisions` — append-only execution decisions or deviations, with rationale and affected phase/task
- `## Current State and Next Action` — active phase, lifecycle status, latest anchors from your own entries, blockers, open items, one exact next action

Preserve everything else: initiative definition, approach, phase/task definitions, prior progress, decisions, validation evidence. Do not add or update task-definition `status` fields. Append-only `## Progress` is the sole task execution-status source.

## Full-Mode Execution

1. **Load state** — slice `docs/plans/active/{slug}.md` by section, never whole-file: `bash scripts/plan-slice.sh docs/plans/active/{slug}.md "{heading}"` IF the script exists; else (consumer default — `zharness install` ships no scripts):

   ```sh
   awk '/^## {heading}$/{on=1;next} on&&/^## /{exit} on' docs/plans/active/{slug}.md
   ```

   Read the target phase block under `## Phases and Verification` and the tails of `## Progress`/`## Decisions`. Select the requested phase, else the first non-done phase whose dependencies are done. Stop IF Current State and the phase blocks disagree on status (reconcile first), or IF no non-empty active plan exists (recommend `brainstorm lock`).
2. **Check boundaries** — compare the requested diff and working tree against task touched/avoided surfaces. Work already outside authority → stop `BLOCKED_CONTRACT_DRIFT`; a task without a check → stop `BLOCKED_VERIFICATION`.
3. **Start the run** — set that phase's plan status to `in-progress`, set Current State to the same phase/status, note the start timestamp as the run anchor, and record the phase start in your first wave-1 `## Progress` entry with `task_status=in-progress`. Never continue while statuses disagree.
4. **Confirm the wave** — restate phase goal, wave, tasks, and checks. Ask only IF the plan does not identify the next incomplete wave unambiguously.
5. **Execute tasks** — follow wave dependencies and task order; parallelize only tasks marked parallel-safe; read every target before editing; stay inside approved surfaces.
6. **Verify each task** — run its exact command. One targeted fix after a failure; a second failure is `BLOCKED_VERIFICATION` → append its Progress line, naming the failed command as the gap, *before* any further edit. Assign a status from Status Routing.
7. **Record task progress** — hold `{"task":"{task}","task_status":"{status}","summary":"{one-line result}"}` in the wave's pending list. `DONE` stays pending until step 9. `BLOCKED`, `NEEDS_CONTEXT`, or `DONE_WITH_CONCERNS` flushes the whole pending list immediately as full `## Progress` lines (fields as in Owned Plan Sections). Every `BLOCKED_*` line includes `failure_class: MISSING_CONTEXT|WRONG_TOOL|BAD_OUTPUT|REPEATED_LOOP|UNSAFE_ACTION|LOST_DECISION|UNKNOWN`.
8. **Record decisions** — only IF execution finds a plan gap, valid trade-off, deviation, or wrong assumption: append `## Decisions` entries (date, phase/task, decision, rationale). Never rewrite an earlier decision.
9. **Complete the wave** — flush every pending entry in one edit: one timestamped `## Progress` line per task plus one wave-summary line.
10. **Refresh Current State** — active phase, `lifecycle_status: in-progress`, latest anchors from your own entries, blockers, open items, exact next action. The plan stays `status: active` until final closure.
11. **Gate the phase** — re-read the Phases entry; require `in-progress` in both the status field and Current State. After all waves, run `check.md`'s durable `gate` steps 1-4 and 6-11 (skip step 5's complete review) **yourself, in this same session**, on the phase diff. Never dispatch to `/check`: its `model: opus` pin forces a cold prompt-cache model switch (F1, SDLC token-cache audit). The complete review runs once, on the final phase, as a `handoff.md` step 6 precondition. `check.md` step 10 sets a cleanly gated non-final phase `checked`. On the initiative's final phase, keep the phase and Current State `in-progress`; on a clean verdict set the next action to an independent `check full`: `check`'s preflight rejects a `checked` phase. Never mark a phase `done`; closing `handoff` owns that transition.

## Status Routing

| Status | Meaning | Action |
|---|---|---|
| `DONE` | Implemented and verified | Continue |
| `DONE_WITH_CONCERNS` | Verified with a surfaced concern | Record concern; continue only with acceptance |
| `NEEDS_CONTEXT` | Required information is absent | Stop and obtain context |
| `BLOCKED` | Cannot finish inside authority or proof boundary | Stop with `BLOCKED_CONTEXT`, `BLOCKED_SCOPE`, `BLOCKED_VERIFICATION`, or `BLOCKED_CONTRACT_DRIFT` |
