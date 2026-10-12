# Playbook: work (full mode)

Loaded by `docs/playbooks/work.md` for `full`.

## Owned Plan Sections

- the selected phase's `status:` line
- `## Log` — append-only: `- <UTC> — <phase>[/<task>] — <kind> — <text>`, kind `start|done|blocked|decision`. `done` names the check, its result, changed surfaces, and for a behavior task the test that went red first; `blocked` names the blocker and `failure_class`; `decision` states a deviation or trade-off and why.
- `## Current State and Next Action` — active phase, lifecycle status, blockers, open items, one exact next action.

Preserve everything else. Do not add or update task-definition `status` fields. `## Log` is the sole task execution-status source. Timestamps come from `date -u`.

## Full-Mode Execution

1. **Load state** — slice `docs/plans/active/{slug}.md` by section: `bash scripts/plan-slice.sh <plan> "{heading}"` IF it exists, else `awk '/^## {heading}$/{on=1;next} on&&/^## /{exit} on' <plan>`. Read the phase block and the `## Log` tail. Pick the requested phase, else the first non-done phase whose dependencies are done. Stop IF Current State and the phase disagree, or no plan exists (→ `brainstorm lock`).
2. **Check boundaries** — work already outside the phase surfaces → stop `BLOCKED_CONTRACT_DRIFT`; a task without a check → `BLOCKED_VERIFICATION`.
3. **Start** — set that phase's plan status to `in-progress`, mirror it in Current State, append `start`; never continue while statuses disagree. Baseline: run the phase's test command once; log any failure that predates your work as a `decision` and do not count it against your tasks.
4. **Execute tasks** — in plan order; parallel only where the plan marks tasks independent, each parallel writer in its own worktree or branch. Read before editing; stay inside the surfaces. A behavior change runs **red → green**: one failing test at the planned seam (failing for the named reason), then the least code that passes; an unplanned seam is a plan gap → `decision` and ask. Never delete, skip, or weaken a test to get green. No-behavior tasks (docs, config, mechanical refactor) just run their check. A phase with no task list: work each requirement in turn, and each `done` names the requirement and its acceptance check.
   **Sub-agent brief:** pointers, not copies — (1) plan path + phase/task id and files to read, (2) output path, (3) the check that proves it, (4) a time budget and stop rule, (5) write scope. Shared exploration runs once into a scratch file outside the repo.
5. **Verify each task** — run its exact command; one targeted fix; a second failure → `blocked` (`BLOCKED_VERIFICATION`) before any further edit.
6. **Log** — a `done` per verified task (a wave may batch); a verified task with a concern adds a `decision` and continues only with acceptance. Missing information → `blocked` with `failure_class: MISSING_CONTEXT`; cannot finish inside the authority or proof boundary → `BLOCKED_CONTEXT`, `BLOCKED_SCOPE`, `BLOCKED_VERIFICATION`, or `BLOCKED_CONTRACT_DRIFT`. A `blocked` entry is written immediately, with any held `done` entries, and carries `failure_class: MISSING_CONTEXT|WRONG_TOOL|BAD_OUTPUT|REPEATED_LOOP|UNSAFE_ACTION|LOST_DECISION|UNKNOWN`. Never rewrite an entry.
7. **Refresh Current State** — `lifecycle_status: in-progress`, blockers, open items, next action.
8. **Gate the phase** — re-read the phase entry; with both statuses `in-progress`, run `check.md`'s durable `gate` on the phase diff yourself, in this session (`high-risk`: an independent agent runs it). Outside `high-risk`, never hand the gate to another model or agent. A clean non-final gate leaves the phase `checked`. On the final phase keep `in-progress` and set the next action to an independent `check full`; a single-phase plan skips this gate and goes straight there. Never mark a phase `done`; closing `handoff` does.
