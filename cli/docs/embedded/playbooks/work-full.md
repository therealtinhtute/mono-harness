# Playbook: work (full mode)

Loaded by `docs/playbooks/work.md` once the resolved mode is `full`. Preconditions, Memory, and Exit Conditions stay there.

## Owned Plan Sections

- the selected phase's lifecycle `status:` line in `## Phases and Verification`
- `## Log` — append-only, one line per entry: `- <UTC> — <phase>[/<task>] — <kind> — <text>`, where kind is `start|done|blocked|decision`. A `done` line names the exact check and its result plus changed surfaces; a `blocked` line names the blocker and `failure_class`; a `decision` line states a deviation, trade-off, or wrong assumption with its rationale.
- `## Current State and Next Action` — active phase, lifecycle status, blockers, open items, one exact next action

Preserve everything else: Goal, approach, phase/task definitions, prior Log entries, validation evidence. Do not add or update task-definition `status` fields. `## Log` is the sole task execution-status source. Timestamps come from `date -u`, never estimated.

## Full-Mode Execution

1. **Load state** — slice `docs/plans/active/{slug}.md` by section, never whole-file: `bash scripts/plan-slice.sh docs/plans/active/{slug}.md "{heading}"` IF the script exists; else (consumer default — `zharness install` ships no scripts):

   ```sh
   awk '/^## {heading}$/{on=1;next} on&&/^## /{exit} on' docs/plans/active/{slug}.md
   ```

   Read the target phase block under `## Phases and Verification` and the tail of `## Log`. Select the requested phase, else the first non-done phase whose dependencies are done. Stop IF Current State and the phase block disagree on status (reconcile first), or IF no non-empty active plan exists (recommend `brainstorm lock`).
2. **Check boundaries** — compare the requested diff and working tree against the phase's touched/avoided surfaces. Work already outside authority → stop `BLOCKED_CONTRACT_DRIFT`; a task without a check → stop `BLOCKED_VERIFICATION`.
3. **Start the run** — set that phase's plan status to `in-progress`, set Current State to the same phase/status, and append a `start` Log entry. Never continue while statuses disagree.
4. **Execute tasks** — in plan order (waves in order when the plan has them); parallelize only tasks the plan marks independent; read every target before editing; stay inside approved surfaces.
5. **Verify each task** — run its exact command. One targeted fix after a failure; a second failure is `BLOCKED_VERIFICATION` → append its `blocked` entry, naming the failed command as the gap, *before* any further edit.
6. **Log** — append a `done` entry per verified task; batching a wave's `done` entries into one edit is fine. A `blocked` entry is written immediately, together with any `done` entries still held. Every `blocked` entry includes `failure_class: MISSING_CONTEXT|WRONG_TOOL|BAD_OUTPUT|REPEATED_LOOP|UNSAFE_ACTION|LOST_DECISION|UNKNOWN`. Append a `decision` entry only IF execution finds a plan gap, valid trade-off, deviation, or wrong assumption; never rewrite an earlier entry.
7. **Refresh Current State** — active phase, `lifecycle_status: in-progress`, blockers, open items, exact next action. The plan stays `status: active` until final closure.
8. **Gate the phase** — re-read the phase entry; require `in-progress` in both the status line and Current State. After all tasks, run `check.md`'s durable `gate` steps on the phase diff **yourself, in this same session**, except on a `high-risk` lane, where an independent judge (a separate agent) runs them. Outside `high-risk`, never hand the gate to another model or agent: prompt cache is per model, so a switch re-reads the phase cold (F1). The complete review runs once, on the final phase, as a `handoff.md` precondition. `check.md` sets a cleanly gated non-final phase `checked`. On the initiative's final phase, keep the phase and Current State `in-progress`; on a clean verdict set the next action to an independent `check full`: `check`'s preflight rejects a `checked` phase. Never mark a phase `done`; closing `handoff` owns that transition.

## Status Routing

| Result | Log kind | Action |
|---|---|---|
| Implemented and verified | `done` | Continue |
| Verified with a surfaced concern | `done` + `decision` | Continue only with acceptance |
| Required information is absent | `blocked` (`MISSING_CONTEXT`) | Stop and obtain context |
| Cannot finish inside authority or proof boundary | `blocked` | Stop with `BLOCKED_CONTEXT`, `BLOCKED_SCOPE`, `BLOCKED_VERIFICATION`, or `BLOCKED_CONTRACT_DRIFT` |
