# Playbook: work

## Purpose

Execute the next approved work from `docs/plans/active/{slug}.md`. Full mode appends execution state to the same plan — append-only Progress and Decisions are the durable record. Bounded mode changes only the requested product files and produces no lifecycle bookkeeping.

## Preconditions and Modes

1. Resolve mode:
   - `full [phase {stable-phase-slug}]` — durable initiative execution from an active plan.
   - `bounded` (alias: `simple`) — known subsystem, bounded files, direct success criterion.
2. Confirm exactly one non-empty plan exists under `docs/plans/active/`; with several, report every candidate and stop rather than guessing.
3. If this session's context was compacted or summarized since you last read the plan, re-read before trusting any earlier-read anchor.

**Zero-write rule:** bounded/simple mode creates no lifecycle rows, plans, reports, changesets, or markdown artifacts. It does not edit an existing active plan. The Git diff plus captured executable/observable proof are its durable evidence.

Reject bounded/simple mode when scope is unclear, crosses an unfamiliar subsystem, exceeds five files or roughly 100 changed lines, needs multi-phase coordination, or lacks a known verification path. Route that work through `brainstorm` and `to-plan`.

## Owned Plan Sections

In full mode, append or update only:

- the selected phase's lifecycle `status` field in `## Phases and Verification`
- `## Progress` — append-only entries recording timestamp, phase, wave, task, task status, exact verification/result, and changed surfaces or blocker
- `## Decisions` — append-only execution decisions or deviations with rationale and affected phase/task
- `## Current State and Next Action` — active phase, lifecycle status, latest anchors taken from your own entries, blockers, open items, and one exact next action

Preserve the initiative definition, planned approach, every phase/task definition, prior progress, decisions, and validation evidence. Do not add or update task-definition `status` fields. Append-only `## Progress` is the sole task execution-status source.

## Full-Mode Execution

1. **Load state from the plan file** — slice `docs/plans/active/{slug}.md` by section, never whole-file (`bash scripts/plan-slice.sh docs/plans/active/{slug}.md "{heading}"`): read the target phase's block under `## Phases and Verification` for its waves/tasks/checks definition, then read the tails of `## Progress`/`## Decisions` to see what already happened in that phase. Select the requested phase or the first non-done phase whose dependencies are done. Treat a phase-status disagreement between Current State and the phase blocks as a stop requiring reconciliation. With no non-empty active plan at all, stop and recommend `brainstorm lock`.
2. **Check boundaries** — compare the requested diff and working tree against phase/task touched and avoided surfaces. Stop with `BLOCKED_CONTRACT_DRIFT` if work is already outside authority. Stop with `BLOCKED_VERIFICATION` if a task lacks a check.
3. **Start the run** — set that phase's plan status to `in-progress`, update Current State to the same phase/status, note the start timestamp as the run anchor, and keep the phase-start fact inside your first wave-1 `## Progress` entry with `task_status=in-progress`. Do not continue while statuses disagree.
4. **Confirm the wave** — restate the phase goal, selected wave, tasks, and checks. Ask only when the plan does not identify the next incomplete wave unambiguously.
5. **Execute tasks** — follow wave dependencies and task order. Parallelize only tasks explicitly marked parallel-safe. Read every target before editing and stay inside approved surfaces.
6. **Verify each task** — run its exact command after implementation. One targeted fix is allowed after a failure; a second failure is `BLOCKED_VERIFICATION`. On that second failure, append the Progress line *before* any further edit, naming the failed command as the gap. Determine `DONE`, `DONE_WITH_CONCERNS`, `NEEDS_CONTEXT`, or `BLOCKED` as the task execution status.
7. **Record task progress** — after each task, hold `{"task":"{task}","task_status":"{status}","summary":"{one-line result}"}` in this wave's pending list instead of writing immediately. A `DONE` entry stays pending for one batched flush at wave end (step 9). `BLOCKED`, `NEEDS_CONTEXT`, or `DONE_WITH_CONCERNS` flushes the whole pending list immediately: append the matching `## Progress` lines right away with timestamp, phase, wave, task, task_status, exact verification/result, and changed surfaces or blocker. Every `BLOCKED_*` line includes `failure_class: MISSING_CONTEXT|WRONG_TOOL|BAD_OUTPUT|REPEATED_LOOP|UNSAFE_ACTION|LOST_DECISION|UNKNOWN`.
8. **Record decisions** — only when execution discovers a plan gap, valid trade-off, deviation, or wrong assumption, append matching `## Decisions` entries with date, phase/task, decision, and rationale. Never rewrite an earlier decision.
9. **Complete the wave** — flush every pending task entry from step 7 in one editing pass: one `## Progress` line per task plus one final wave-summary line naming the wave outcome. Each entry carries its own timestamp so the ordering survives.
10. **Refresh current state** — update active phase, `lifecycle_status: in-progress`, latest anchors found in your own appended entries, blockers, open items, and exact next action. Keep the plan `status: active` until final closure.
11. **Verify synchronization and gate the phase** — re-read the plan's Phases entry; require the selected phase to show `in-progress` in both the status field and Current State. After all phase waves complete, perform `check.md`'s durable `gate` steps (Review and Gate Steps 1-4 and 6-11, skipping step 5's complete manual review) **yourself, in this same session**, on the phase diff — do not dispatch to the separate `/check` skill for this: its own frontmatter pins `model: opus`, and prompt caches are model-scoped, so routing through it would force exactly the cold-cache switch this step exists to avoid (F1 of the SDLC token-cache audit). The complete manual review runs exactly once, on the initiative's final phase, as a `handoff` closure precondition (`handoff.md` step 6) — that single review may switch model, since it happens once per initiative rather than once per phase. Do not mark the phase checked or done; durable `check` and closing `handoff` own those transitions.

## Memory conventions

Record durable memory under `docs/memory/{id}.md` — plain committed files, edited directly — only when one of three triggers fires:

- **Fact correction** — an earlier memory is wrong or has been superseded; write the corrected entry, then set the old file's frontmatter `superseded_by` to the new id with the date so lineage remains greppable.
- **Durable lesson** — a cross-session learning, architecture decision, or hard-won gotcha that would otherwise be rediscovered.
- **Owner preference** — an explicit owner instruction about style, process, or scope that should persist beyond the current session.

**Redaction rule** — never store credentials, secrets, API keys, or token values in a memory body; bodies are committed markdown and must not contain sensitive values. Record only that a secret exists, its scope, and where to fetch it.

Retrieval is grep-first: search `docs/memory/*.md` by keyword directly; a file whose frontmatter says superseded stays visible for lineage but should be discounted in answers.

## Exit Conditions

- Full mode: the selected phase is `in-progress` in both the status field and Current State, every attempted task has a `## Progress` entry, material decisions are recorded with rationale, each completed wave ends with a summary line, Current State is resumable, and completed implementation is gated in-session per step 11 (`check.md`'s `full` mode applies only when the selected phase is the initiative's final phase — see `handoff.md` step 6).
- Bounded/simple mode: requested code and proof are shown in the response with zero lifecycle or markdown writes.

## Status Routing

Referenced by Full-Mode Execution steps 6-7 when recording a task's execution status.

| Status | Meaning | Action |
|---|---|---|
| `DONE` | Implemented and verified | Continue |
| `DONE_WITH_CONCERNS` | Verified with a surfaced concern | Record concern; continue only with acceptance |
| `NEEDS_CONTEXT` | Required information is absent | Stop and obtain context |
| `BLOCKED` | Cannot finish inside authority or proof boundary | Stop with `BLOCKED_CONTEXT`, `BLOCKED_SCOPE`, `BLOCKED_VERIFICATION`, or `BLOCKED_CONTRACT_DRIFT` |

## escalate_when

Ask the owner and stop. Do not invent.

- locked schema or requirements would change
- the same verification command failed twice
- a product rule conflicts
