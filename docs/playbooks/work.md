# Playbook: work

## Purpose

Execute the next approved work from `docs/plans/active/{slug}.md`. Full mode records run/trace rows and appends execution state to the same plan. Bounded mode changes only the requested product files and produces no durable lifecycle bookkeeping.

## Preconditions and Modes

1. Resolve mode:
   - `full [phase {stable-phase-slug}]` — durable initiative execution from an active plan.
   - `bounded` (alias: `simple`) — known subsystem, bounded files, direct success criterion.
2. Run `zharness preflight work --mode {full|bounded} --json`. Missing binary: print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop. Otherwise check its `version` field — a `dev` build satisfies the gate; below MIN_ZHARNESS_VERSION (`0.8.1` — see `skills/workflow/README.md`), print the same message and stop. Then follow its stop/recovery result exactly. In full mode its `context` field is the source of Step 1's phase list and lifecycle position below — do not call `query state`/`query phases` again to obtain it.
3. If this session's context was compacted or summarized since the last `preflight` call — including mid-phase, across waves — re-run it before trusting any earlier-read `context` packet or lifecycle ID; a summarized turn cannot be assumed to have carried exact DB state forward.

**Zero-write rule:** bounded/simple mode creates no lifecycle rows, plans, reports, changesets, or markdown artifacts. It does not edit an existing active plan. The Git diff plus captured executable/observable proof are its durable evidence.

Reject bounded/simple mode when scope is unclear, crosses an unfamiliar subsystem, exceeds five files or roughly 100 changed lines, needs multi-phase coordination, or lacks a known verification path. Route that work through `brainstorm` and `to-plan`.

## Owned Plan Sections

In full mode, append or update only:

- the selected phase's lifecycle `status` field in `## Phases and Verification`, solely to mirror the public DB transition
- `## Progress` — append-only wave/task entries, written by `trace add` itself (task status, run ID, and a one-line summary; fold changed surfaces and the exact verification command/result into that summary text since the CLI's own entry has no separate fields for them)
- `## Decisions` — append-only execution decisions or deviations with rationale and affected phase/task, written by `decision add` itself
- `## Current State and Next Action` — active phase, lifecycle status, latest IDs, blockers, open items, and one exact next action (still hand-written; `run create` does not write markdown)

Preserve the initiative definition, planned approach, every phase/task definition, prior progress, decisions, and validation evidence. Do not add or update task-definition `status` fields. Append-only `## Progress` is the sole task execution-status source.

## Full-Mode Execution

1. **Load state** — read `context.phases` and `context.position` from the same `preflight work --json` response (Preconditions step 2) instead of separately calling `query state`/`query phases`. Select the requested phase or the first non-done phase whose dependencies are done. For that phase, call `zharness query plan --section phase --phase {stable-phase-slug} --json` for its waves/tasks/checks definition and `zharness query traces --phase {stable-phase-slug} --json` for what's already recorded against it, instead of reading the whole active plan (the ceremony audit's P3 proposal). If `query plan` reports `degraded: true`, read the plan file directly for this phase's definition. Treat pre-existing disagreement between DB status (`context.phases`) and plan status as a stop requiring reconciliation.
   - If `query plan` (or `preflight`) instead fails with the shared active-plan resolver's Stop contract (R2, `docs/audit/consumer-adoption-audit.md` D1), branch on its `code` — do not guess which plan is live and do not read a plan body to decide:
     - `code: "ambiguous"` — more than one non-empty plan exists under `docs/plans/active/`. The error's `candidates` list (R3's bounded Tier 1/Tier 2 packet — frontmatter `updated:` or, per R4, last commit date when frontmatter is missing/unparseable) names each one newest-first; it never contains plan-body content. Report the candidates to the user and stop; do not resume until the user runs `zharness plan complete` or `zharness plan abandon` on all but one.
     - `code: "none"` — no non-empty plan exists under `docs/plans/active/`. Stop and recommend `brainstorm lock` to create one; there is no phase to execute.
2. **Check boundaries** — compare the requested diff and working tree against phase/task touched and avoided surfaces. Stop with `BLOCKED_CONTRACT_DRIFT` if work is already outside authority. Stop with `BLOCKED_VERIFICATION` if a task lacks a check.
3. **Create the run row and synchronize the plan** — run `zharness run create --slug {stable-phase-slug} --plan-id {plan frontmatter id} --json`. Do not pass an artifact path. Immediately after success, save the returned run ID, set that phase's plan status to `in-progress`, update Current State to the same phase/status/run ID, and append the phase-start Progress entry with `task_status=in-progress`. Do not mutate the task definition, and do not continue while the DB says `in-progress` and the plan phase still says `planned`.
4. **Confirm the wave** — restate the phase goal, selected wave, tasks, and checks. Ask only when the plan does not identify the next incomplete wave unambiguously.
5. **Execute tasks** — follow wave dependencies and task order. Parallelize only tasks explicitly marked parallel-safe. Read every target before editing and stay inside approved surfaces.
6. **Verify each task** — run its exact command after implementation. One targeted fix is allowed after a failure; a second failure is `BLOCKED_VERIFICATION`. Determine `DONE`, `DONE_WITH_CONCERNS`, `NEEDS_CONTEXT`, or `BLOCKED` as the task execution status.
7. **Record task progress** — after each task, add `{"task":"{task}","task_status":"{status}","summary":"{one-line result}"}` to this wave's pending list instead of calling the CLI right away. A `DONE` entry stays pending for one batched flush at wave end (step 9) — R5 of the SDLC token-cache audit: most tasks land clean, so paying a round trip per task was ceremony, not signal. `BLOCKED`, `NEEDS_CONTEXT`, or `DONE_WITH_CONCERNS` flushes the whole pending list immediately instead: run `zharness trace add --wave {N} --tasks '[...]' --run-id {run-id} --json` right away, since work is stopping (or a concern needs recording) now, not at a wave end that may not arrive. The CLI appends one matching `## Progress` line per element itself, in the same call — do not also hand-write one (P3, "CLI owns the pen"; G1's task-granularity trace).
8. **Record decisions** — only when execution discovers a plan gap, valid trade-off, deviation, or wrong assumption, run `zharness decision add --decisions '[{"decision":"...","rationale":"...","phase":"{stable-phase-slug}","task":"{task}"}, ...]' --run-id {run-id} --json`. The CLI appends the matching `## Decisions` entry itself; never hand-edit or rewrite an earlier decision.
9. **Complete the wave** — after every task is `DONE` or explicitly accepted `DONE_WITH_CONCERNS`, flush any pending task entries step 7 hasn't already flushed in one `zharness trace add --wave {N} --tasks '[...]' --run-id {run-id} --json` call (skip this if step 7 already flushed everything for this wave). Then run `zharness trace add --wave {N} --summary "{one-line outcome}" --run-id {run-id} --json` (wave-level: omit `--task`/`--task-status`/`--tasks`). This appends its own `## Progress` summary line; there is no trace ID to splice into an earlier entry.
10. **Refresh current state** — update active phase, `lifecycle_status: in-progress`, latest run/trace IDs, blockers, open items, and exact next action. Keep the plan `status: active` until final closure.
11. **Verify synchronization and gate the phase** — rerun `zharness query phases --json`; require the selected phase to be `in-progress` in both DB and plan. After all phase waves complete, perform `check.md`'s durable `gate` steps (Review and Gate Steps 1-4 and 6-11, skipping step 5's complete manual review) **yourself, in this same session**, on the phase diff — do not dispatch to the separate `/check` skill for this: its own frontmatter pins `model: opus`, and prompt caches are model-scoped, so routing through it would force exactly the cold-cache switch this step exists to avoid (F1 of the SDLC token-cache audit — a per-phase `check full` was costing $0.275/phase, 63% of the gate's own cost, for a review most phases don't need). The complete manual review (`check.md` full mode) runs exactly once, on the initiative's final phase, as a `handoff` closure precondition (`handoff.md` step 6) — that single review may switch model, since it happens once per initiative rather than once per phase. Do not mark the phase checked or done; durable `check` and closing `handoff` own those transitions.

## Command Reference

- `zharness preflight work --mode {full|bounded} --json`
- `zharness query plan --section phase --phase {stable-phase-slug} --json` (step 1 — the phase's own definition)
- `zharness query traces --phase {stable-phase-slug} --json` (step 1 — what's already recorded for it)
- `zharness query phases --json` (step 11 only — post-mutation re-verification; Step 1 reads `context.phases` from preflight instead)
- `zharness run create --slug {stable-phase-slug} --plan-id {plan-id} --json`
- `zharness trace add --wave {N} --tasks '[{"task":"...","task_status":"...","summary":"..."}, ...]' --run-id {run-id} --json` (step 7/9 — batched task-entry flush, 1-20 entries)
- `zharness trace add --wave {N} --summary "..." --run-id {run-id} --json` (step 9 — wave-level summary; also the single-entry form for a lone task via `[--task "..." --task-status {status}]`)
- `zharness decision add --decisions '[...]' --run-id {run-id} --json` (step 8 only, when execution surfaces a decision)

## Exit Conditions

- Full mode: the selected phase is `in-progress` in both DB and plan, every attempted task is appended to Progress, material decisions are appended with rationale, each completed wave has a trace ID, Current State is resumable, and completed implementation is gated in-session per step 11 (`check.md`'s `full` mode, via the separate `/check` skill, applies only when the selected phase is the initiative's final phase — see `handoff.md` step 6).
- Bounded/simple mode: requested code and proof are shown in the response with zero lifecycle or markdown writes.

## Bounded/Simple Execution

The rarer branch — routed here only when Preconditions step 1 resolved `bounded`/`simple`; a full-mode run never needs this section.

1. Read at most the directly named files and nearest conventions.
2. State the surgical change and success criterion.
3. Apply only the bounded edit.
4. Run the narrowest real verification command and capture its output in the response.
5. Stop after stating changed files and proof; do not create or update lifecycle markdown or DB state.

## Status Routing

Referenced by Full-Mode Execution steps 6-7 when recording a task's execution status.

| Status | Meaning | Action |
|---|---|---|
| `DONE` | Implemented and verified | Continue |
| `DONE_WITH_CONCERNS` | Verified with a surfaced concern | Record concern; continue only with acceptance |
| `NEEDS_CONTEXT` | Required information is absent | Stop and obtain context |
| `BLOCKED` | Cannot finish inside authority or proof boundary | Stop with `BLOCKED_CONTEXT`, `BLOCKED_SCOPE`, `BLOCKED_VERIFICATION`, or `BLOCKED_CONTRACT_DRIFT` |
