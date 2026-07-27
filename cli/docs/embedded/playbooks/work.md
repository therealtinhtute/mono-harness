# Playbook: work

## Purpose

Execute the next approved work from `docs/plans/active/{slug}.md`. Full mode records run/trace rows and appends execution state to the same plan. Bounded mode changes only the requested product files and produces no durable lifecycle bookkeeping.

## Preconditions and Modes

1. Run `zharness --version`. A `dev` build satisfies the gate; otherwise require version `0.1.0` or newer. If unavailable or stale, print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop.
2. Resolve mode:
   - `full [phase {stable-phase-slug}]` — durable initiative execution from an active plan.
   - `bounded` (alias: `simple`) — known subsystem, bounded files, direct success criterion.
3. Run `zharness preflight work --mode {full|bounded} --json` and follow its stop/recovery result exactly.

**Zero-write rule:** bounded/simple mode creates no lifecycle rows, plans, reports, changesets, or markdown artifacts. It does not edit an existing active plan. The Git diff plus captured executable/observable proof are its durable evidence.

Reject bounded/simple mode when scope is unclear, crosses an unfamiliar subsystem, exceeds five files or roughly 100 changed lines, needs multi-phase coordination, or lacks a known verification path. Route that work through `brainstorm` and `to-plan`.

## Owned Plan Sections

In full mode, append or update only:

- the selected phase's lifecycle `status` field in `## Phases and Verification`, solely to mirror the public DB transition
- `## Progress` — append-only phase/wave/task entries with task execution status, run ID, trace ID, changed surfaces, and verification result
- `## Decisions` — append-only execution decisions or deviations with rationale and affected phase/task
- `## Current State and Next Action` — active phase, lifecycle status, latest IDs, blockers, open items, and one exact next action

Preserve the initiative definition, planned approach, every phase/task definition, prior progress, decisions, and validation evidence. Do not add or update task-definition `status` fields. Append-only `## Progress` is the sole task execution-status source.

## Full-Mode Execution

1. **Load state** — read the active plan, then run `zharness query state --json` and `zharness query phases --json`. Select the requested phase or the first non-done phase whose dependencies are done. Treat pre-existing disagreement between DB status and plan status as a stop requiring reconciliation.
2. **Check boundaries** — compare the requested diff and working tree against phase/task touched and avoided surfaces. Stop with `BLOCKED_CONTRACT_DRIFT` if work is already outside authority. Stop with `BLOCKED_VERIFICATION` if a task lacks a check.
3. **Create the run row and synchronize the plan** — run `zharness run create --slug {stable-phase-slug} --plan-id {plan frontmatter id} --json`. Do not pass an artifact path. Immediately after success, save the returned run ID, set that phase's plan status to `in-progress`, update Current State to the same phase/status/run ID, and append the phase-start Progress entry with `task_status=in-progress`. Do not mutate the task definition, and do not continue while the DB says `in-progress` and the plan phase still says `planned`.
4. **Confirm the wave** — restate the phase goal, selected wave, tasks, and checks. Ask only when the plan does not identify the next incomplete wave unambiguously.
5. **Execute tasks** — follow wave dependencies and task order. Parallelize only tasks explicitly marked parallel-safe. Read every target before editing and stay inside approved surfaces.
6. **Verify each task** — run its exact command after implementation. One targeted fix is allowed after a failure; a second failure is `BLOCKED_VERIFICATION`. Record `DONE`, `DONE_WITH_CONCERNS`, `NEEDS_CONTEXT`, or `BLOCKED` only as the task execution status in Progress.
7. **Append Progress** — immediately append one structured entry per attempted task with timestamp, phase, wave, task, task status, run ID, changed surfaces, exact verification command/result, and concern/blocker if any.
8. **Record decisions** — append only when execution discovers a plan gap, valid trade-off, deviation, or wrong assumption. Include the decision, rationale, and affected phase/task; never rewrite an earlier decision.
9. **Complete the wave** — after every task is `DONE` or explicitly accepted `DONE_WITH_CONCERNS`, run `zharness trace add --wave {N} --summary "{one-line outcome}" --run-id {run-id} --json`. Append the returned trace ID to that wave's Progress entry.
10. **Refresh current state** — update active phase, `lifecycle_status: in-progress`, latest run/trace IDs, blockers, open items, and exact next action. Keep the plan `status: active` until final closure.
11. **Verify synchronization and gate the phase** — rerun `zharness query phases --json`; require the selected phase to be `in-progress` in both DB and plan. After all phase waves complete, invoke `check full` on the phase diff. Do not mark the phase checked or done; durable `check` and closing `handoff` own those transitions.

## Bounded/Simple Execution

1. Read at most the directly named files and nearest conventions.
2. State the surgical change and success criterion.
3. Apply only the bounded edit.
4. Run the narrowest real verification command and capture its output in the response.
5. Stop after stating changed files and proof; do not create or update lifecycle markdown or DB state.

## Status Routing

| Status | Meaning | Action |
|---|---|---|
| `DONE` | Implemented and verified | Continue |
| `DONE_WITH_CONCERNS` | Verified with a surfaced concern | Record concern; continue only with acceptance |
| `NEEDS_CONTEXT` | Required information is absent | Stop and obtain context |
| `BLOCKED` | Cannot finish inside authority or proof boundary | Stop with `BLOCKED_CONTEXT`, `BLOCKED_SCOPE`, `BLOCKED_VERIFICATION`, or `BLOCKED_CONTRACT_DRIFT` |

## Command Reference

- `zharness --version`
- `zharness preflight work --mode {full|bounded} --json`
- `zharness query state --json`
- `zharness query phases --json`
- `zharness run create --slug {stable-phase-slug} --plan-id {plan-id} --json`
- `zharness trace add --wave {N} --summary "..." --run-id {run-id} --json`

## Exit Conditions

- Full mode: the selected phase is `in-progress` in both DB and plan, every attempted task is appended to Progress, material decisions are appended with rationale, each completed wave has a trace ID, Current State is resumable, and completed implementation is routed to `check full`.
- Bounded/simple mode: requested code and proof are shown in the response with zero lifecycle or markdown writes.
