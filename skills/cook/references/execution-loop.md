# Execution Loop — Wave Dispatch, Subagents, Status Routing

Use this after the routing table returns `ready`. The loop runs once per phase.

## Wave Dispatch

A wave is a group of tasks that can run together. Phase `-PLAN.md` declares wave structure.

```text
phase = [wave_1, wave_2, wave_3, ...]
wave  = { tasks: [...], parallel: bool, dependencies: [prior_wave_id?] }
```

### Per-wave procedure

1. Load wave tasks from `-PLAN.md`
2. Confirm dependencies on prior waves are satisfied (status `DONE` or `DONE_WITH_CONCERNS` user-acked)
3. If `parallel: true` and 2+ tasks present → dispatch in the same response (multiple tool calls in one turn for inline, or multiple subagent calls for heavy)
4. If `parallel: false` → run tasks sequentially in declared order
5. After every task, capture status; do not advance to the next wave until the current wave is fully `DONE`/`DONE_WITH_CONCERNS`

## Inline vs Subagent — Choose per Task

| Use inline | Use subagent |
|------------|--------------|
| 1-3 line edits in a single file | Cross-file refactor (3+ files) |
| Read-and-summarize | Heavy research / repo grep across many paths |
| Running a single shell command | Generating a new file from a template |
| Trivial config tweaks | Migration sweeps, codemod-style changes |

The subagent's job is to keep the main session's context clean. If the task touches one file and one concept, do it inline.

## Subagent Dispatch Prompt (template)

When dispatching, the prompt MUST be self-contained — no "see prior context", no "based on earlier discussion".

```text
Task: {one-sentence task goal from -PLAN.md}

Context:
- Phase: {phase-slug}
- Spec section: {relevant numbered requirement or "see SPEC.md §X"}
- Files in scope: {comma-separated paths}
- Forbidden scope: {comma-separated paths or "everything not listed above"}

Verification command: {exact command from -PLAN.md, e.g. `pnpm test path/to/file.test.ts`}
Expected output signal: {pass/fail criterion}

Report back with:
- Status: DONE | DONE_WITH_CONCERNS | NEEDS_CONTEXT | BLOCKED
- Files changed (paths only)
- Verification output (last 20 lines if long)
- Any concerns or blockers (one paragraph max)

Hard limits:
- Do not edit files outside the scope list
- Do not edit `.planning/SPEC.md` or `.planning/ROADMAP.md`
- Do not skip the verification command
```

## Status Enum Routing

Every task returns one of four statuses. Cook routes them like this:

| Status | Meaning | Cook's response |
|--------|---------|-----------------|
| `DONE` | Implemented, verified, no surprises | Continue to next task in wave |
| `DONE_WITH_CONCERNS` | Implemented and verified, but flagged a side observation | Surface the concern in the wave summary; user decides whether to halt |
| `NEEDS_CONTEXT` | Subagent lacked information not in the dispatch prompt | Provide the missing context, redispatch — do not invent the answer |
| `BLOCKED` | Task cannot complete inside the spec/plan boundary | Stop the wave, surface the blocker, suggest `plan phase {slug}` or `brainstorm refine` |

## Verification Discipline

- Every task MUST have a verification command in `-PLAN.md`. Missing verification → escalate to `BLOCKED`, do not invent one.
- Run the verification command after the task, not before claiming DONE.
- If the command fails: try at most one targeted fix in the same wave. A second failure on the same task → `BLOCKED`.
- Capture the verification output verbatim in the wave summary; never summarize "tests passed" without proof.

## Phase Gate

After all waves return `DONE` or user-acked `DONE_WITH_CONCERNS`:

1. Build a phase diff (`git diff` against the phase start commit, or against `main`/`master` if no checkpoint)
2. Invoke `/check full` and feed it the phase diff
3. Wait for the gate verdict
4. On clean gate: surface "phase {slug} clean" and offer `/git cm`, `/handoff`, or `/watzup` as next moves
5. On non-clean gate: do NOT mark the phase complete; surface findings and pause for user direction

## What Cook Never Does in the Loop

- Never reorder waves the plan declared dependent
- Never add a task that wasn't in `-PLAN.md` — capture in `-CONTEXT.md` as a deferred idea instead
- Never accept "done" without a verification command output
- Never auto-commit or auto-push between phases
