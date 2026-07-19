# Playbook: work

## Purpose

Execution conductor: read planning artifacts, execute the next incomplete phase wave-by-wave, verify each task, route into `check` as the quality gate, and surface clean handoffs. Own the HOW of execution; never redesign scope.

Defer elsewhere: `brainstorm` — spec missing or weak and the task exceeds simple-mode scope. `to-plan` — spec exists but no roadmap/phase files. `check` — gate-only or code-review-only request without execution. `git` — pure commit/push request. A bug with unknown root cause needs investigation first, outside this playbook's scope.

## Preconditions

- **Version gate**: run `zharness --version` before anything else. A `dev` build always satisfies the gate. Otherwise, if the binary is missing or below `0.1.0` (`MIN_ZHARNESS_VERSION`), print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop.
- **Full mode hard gate**: confirm `.kit/planning/SPEC.md` is locked AND the target phase has both `-CONTEXT.md` and `-PLAN.md`. If either is missing or visibly stale (placeholders, contradictions, undated decisions), stop and route to the correct upstream playbook. Never silently expand scope mid-flight.
- **Simple mode**: no `.kit/planning/` artifacts required. The hard gate is replaced by the scope guard (see Simple Mode below) — if research reveals > 5 files, > 100 lines, or an unknown subsystem, stop and route to `brainstorm` → `to-plan` → `work full`.

## Arguments

- `auto` (default) — resolve mode automatically from available artifacts
- `full` — strict pipeline requiring `.kit/planning/` artifacts; starts at the first incomplete phase
- `full phase <slug>` / `phase <slug>` — strict pipeline for one named phase (`phase <slug>` is an alias for `full phase <slug>`)
- `simple [@file?]` — lightweight execution from a prompt or brainstorm explore file; no `.kit/planning/` required
- `--notes` — opt-in flag (any mode): append decisions, deviations, and tradeoffs to `.kit/implementation-notes.md` during execution. Default off.

## Mode Resolution (run first, every invocation)

| Argument | Mode |
|---|---|
| `simple` or `simple @file` | `simple` |
| `full`, `full phase <slug>`, `phase <slug>` | `full` |
| No argument | auto-detect |

**Auto-detect (no argument):**

| Observed state | Resolved mode |
|---|---|
| `.kit/planning/SPEC.md` + `ROADMAP.md` + target phase artifacts all present | `full` |
| `.kit/reports/brainstorm/*.md` present or referenced, no SPEC.md | `simple` |
| Only a direct prompt, no planning or brainstorm artifacts | `simple` (after prompt-quality check) |
| No argument, no artifacts, no meaningful prompt | Stop → route to `brainstorm` |
| Ambiguous (stale SPEC + new prompt, or SPEC exists but a brainstorm file is also referenced) | Ask the user a short structured question to clarify |

## Full Mode Detection Table

| State | Signal | Action |
|---|---|---|
| `no-spec` | `.kit/planning/SPEC.md` missing or empty | Stop, route to `brainstorm` |
| `no-plan` | `.kit/planning/ROADMAP.md` missing | Stop, route to `to-plan` |
| `no-phase` / `no-context` | `{slug}-PLAN.md` or `{slug}-CONTEXT.md` missing for the selected phase | Stop, route to `to-plan phase {slug}` |
| `stale-plan` | Phase plan references files/symbols that no longer exist | Stop, route to `to-plan phase {slug}` to refresh |
| `placeholder-plan` | Phase plan contains `TBD`, `TODO`, "similar to", "implement later" | Stop, route to `to-plan phase {slug}` |
| `contract-drift` | Working tree or requested scope already touches files outside `Allowed Surfaces`/task `touches`, or conflicts with `Forbidden Surfaces`/task `avoid` | Stop, route to `to-plan phase {slug}` or `brainstorm refine` |
| `multiple-incomplete` | More than one phase has incomplete waves | Ask the user which phase to run (mark the first incomplete phase by roadmap order as recommended) |
| `ready` | All required files present and concrete | Proceed to the execution loop |

### Selecting the active phase (auto mode)
1. Parse `.kit/planning/ROADMAP.md` for the ordered phase list.
2. For each phase, check whether its `-PLAN.md` shows all waves complete (status section, completion markers; if absent, assume incomplete).
3. The first incomplete phase is the candidate.
4. If two phases are partially done (rare — indicates a previous handoff), trigger `multiple-incomplete` and ask.

### Stop message shapes
State the blocker plainly, then the exact recovery command:
- no-spec: "No `.kit/planning/SPEC.md` found. Lock the problem first. Run `brainstorm` with the idea, notes, or file refs."
- no-plan: "SPEC exists, no plan. Generate the roadmap first. Run `to-plan full`."
- no-phase/no-context: "Phase artifacts missing for `{slug}`. Run `to-plan phase {slug}`."
- stale-plan/placeholder-plan: name the specific file/symbol or placeholder text and line; run `to-plan phase {slug}` to refresh, then re-invoke `work`.
- contract-drift: name the working-tree file or requested-scope item and why it's outside/inside the forbidden boundary; run `to-plan phase {slug}` if the contract should refresh, or `brainstorm refine` if the spec boundary itself changed.

## Execution Loop (per phase; Step 2 branches by mode, everything else is full-mode)

1. **Load context** — run `zharness query state --json` and `zharness query phases --json` first when a db exists, then verify against `.kit/planning/SPEC.md`, the phase `-CONTEXT.md`, and the phase `-PLAN.md`. Note open assumptions; treat a `current_phase` mismatch as a plan-refresh signal, not as truth.
2. **Create the run artifact** — first run `zharness id --json`; save the returned `id` as the **RUN id** and use it exactly in the artifact's `id:` frontmatter and every later `--run-id`/run-row reference. Never invent a placeholder ULID. Write `.kit/runs/work/{YYYYMMDD-HHmm}-{slug}.md` (see Artifacts below), with `mode: full` or `mode: simple` in the frontmatter matching the resolved mode. In simple mode, the slug comes from the prompt or brainstorm file, and `phase`/`plan_id` are both `none`. Run `zharness init` if no db exists yet (idempotent).
   - **Full mode**: register the run in the harness. Run `zharness id --json` again and save this second, distinct ID as the **changeset id**; author the two-line changeset at `.kit/changesets/{changeset-id}.changeset.jsonl` (never reuse the RUN id as the filename) — line 1 creates the run row (`{"op":"create","entity":"run","id":"{RUN id}","fields":{"story_slug":"{phase-slug}","artifact_path":"{run file path}","created_at":"{RFC3339 now}"},"at":"{RFC3339 now}"}`), line 2 points `meta.latest_run_id` at it (`{"op":"update","entity":"meta","id":"meta","fields":{"latest_run_id":"{RUN id}"},"at":"{RFC3339 now}"}`) — and apply the file in one shot with `zharness db changeset apply {path} --json` (no dedicated "run create" command exists; `db changeset apply` is the same generic command used to rebuild state from changesets). Both lines land in the same transaction, so `latest_run_id` never lags the run it points to.
   - **Simple mode**: do NOT author or apply a run changeset. Simple mode has no phase and therefore no story, and `runs.story_slug` is a `NOT NULL` foreign key into `stories(slug)` — there is nothing for it to reference, and forcing a write here always fails with a FOREIGN KEY constraint error. Skip DB registration entirely; the run artifact file itself (with `mode: simple`) is the durable record, and `validate` treats `mode: simple` RUN artifacts as exempt from phase/plan/DB-registration checks by design (see `CONTRACT.md`). Note the skip inline in the run artifact.
3. **Preflight drift check** — compare the phase boundary (Allowed/Forbidden Surfaces, task `touches`/`avoid`) against the current working tree and requested scope. If files already changed outside the boundary, stop with `BLOCKED_CONTRACT_DRIFT`.
4. **Confirm scope** — restate the phase goal and wave list in one block; ask the user only if the plan is ambiguous about which wave is next.
5. **Run waves** — for each wave, execute tasks in order; parallelize only when the `-PLAN.md` marks the wave as parallel-safe.
6. **Per-task discipline** — for heavy or isolated tasks (file generation, refactor across many files, research), delegate to an isolated sub-task if the runtime supports it; otherwise perform it directly. For trivial edits (1-3 lines, single file), always do it directly.
7. **Verify per task** — run the task's verification command; capture output. Failed verification = task not done; do not advance the wave.
8. **Status enums** — after each task, mark `DONE`, `DONE_WITH_CONCERNS`, `NEEDS_CONTEXT`, or `BLOCKED`. Continue on `DONE`; surface the rest before moving on. Always append task results to the run artifact.
9. **Wave completion trace** — once a wave reaches `DONE`/`DONE_WITH_CONCERNS`, run `zharness trace add --wave {N} --summary "{one-line wave outcome}" --run-id {this run's id} --json`; append the returned `id` to the run artifact frontmatter's `trace_ids` list.
10. **State check** — after run creation and after any terminal status (`BLOCKED`, `NEEDS_CONTEXT`, clean phase completion), run `zharness query state --json` and confirm `current_phase` still matches; the harness rows written in steps 2 and 9 are the durable record — there is no separate index file to refresh.
11. **Phase gate** — when all waves complete, invoke the `check` playbook on the phase diff. Do not advance to the next phase on a non-clean gate.
12. **Handoff suggestion** — on a clean gate, suggest committing the change or running the `handoff` playbook, whichever is natural; never run either automatically.

### Wave dispatch details
A wave is a group of tasks that can run together: `phase = [wave_1, wave_2, ...]`, `wave = { tasks: [...], parallel: bool, dependencies: [prior_wave?] }`. Per wave: load its tasks from `-PLAN.md`; confirm prior-wave dependencies are `DONE` or user-acked `DONE_WITH_CONCERNS`; confirm the working tree still fits the phase boundary and the wave's `touches`/`avoid`; if `parallel: true` and 2+ tasks, dispatch together; if `parallel: false`, run sequentially in declared order; after every task, capture status and append to the run artifact — do not advance until the wave is fully `DONE`/`DONE_WITH_CONCERNS`.

### Inline vs. delegated execution
Do the work directly for: 1-3 line edits in a single file, read-and-summarize, running a single shell command, trivial config tweaks. Delegate to an isolated sub-task (if the runtime supports it) for: cross-file refactors (3+ files), heavy research or repo-wide grep, generating a new file from a template, migration/codemod-style sweeps. The point of delegating is to keep the main execution context clean — if a task touches one file and one concept, do it directly.

A delegated sub-task's instructions must be self-contained (no "see prior context"): task goal, phase slug, relevant spec section, files in scope, forbidden scope, the exact verification command from `-PLAN.md`, and the expected pass/fail signal. It must report back: status (one of the four enums), files changed, verification output (last ~20 lines if long), and any concerns/blockers in one paragraph. Hard limits for any delegated sub-task: never edit files outside the scope list; never edit `.kit/planning/SPEC.md` or `.kit/planning/ROADMAP.md`; never skip the verification command.

### Status enum routing
| Status | Meaning | Response |
|---|---|---|
| `DONE` | Implemented, verified, no surprises | Continue to next task in wave |
| `DONE_WITH_CONCERNS` | Implemented and verified, flagged a side observation | Surface the concern in the wave summary; user decides whether to halt |
| `NEEDS_CONTEXT` | Lacked information not in the dispatch instructions | Provide the missing context, redo — do not invent the answer |
| `BLOCKED` | Cannot complete inside the spec/plan boundary | Stop the wave, surface the blocker with a primary code: `BLOCKED_CONTEXT`, `BLOCKED_SCOPE`, `BLOCKED_VERIFICATION`, or `BLOCKED_CONTRACT_DRIFT` |

### Verification discipline
Every task MUST have a verification command in `-PLAN.md`. Missing verification → escalate to `BLOCKED_VERIFICATION`, do not invent one. Run the verification command after the task, not before claiming `DONE`. If it fails: try at most one targeted fix in the same wave; a second failure on the same task → `BLOCKED_VERIFICATION`. Capture the verification output verbatim in the wave summary and run artifact; never summarize "tests passed" without proof.

### Phase gate
After all waves return `DONE` or user-acked `DONE_WITH_CONCERNS`: build a phase diff (against the phase start commit, or the main/default branch if no checkpoint); invoke the `check` playbook on that diff; wait for the verdict. Clean gate: surface "phase `{slug}` clean" and suggest committing, the `handoff` playbook, or the `watzup` playbook as next moves. Non-clean gate: do NOT mark the phase complete; surface findings and pause for user direction.

## Simple Mode — Lightweight Execution

Activate when mode resolves to `simple`. Do NOT touch `.kit/planning/`.

**Appropriate when**: scope is known upfront (clear file(s), clear change, clear success criterion); input is a direct prompt or a brainstorm explore file; work doesn't need multi-phase coordination or schema migrations; root cause is already known.

**Reject simple mode when**: task clearly touches > 5 files, > 100 lines, or an unfamiliar subsystem (even if the user says "just do it"); root cause is unknown (investigate first); the task is a scope-defining question (route to `brainstorm`); the task is a complete multi-phase feature (use `work full`).

### 7-Step workflow

1. **Prompt quality check** — a good prompt answers what to change, where, and what "done" looks like. If thin (< 1 sentence or missing 2+ of those), ask 1-2 targeted questions. If still thin afterward, stop and tell the user to add concrete detail (file, change, success criterion) before re-attempting simple mode.
   - Rubric: file path + change + success criterion → proceed. File path + change, no criterion → proceed, infer criterion. Subsystem name + vague change → ask one question about the specific change. "Fix the bug"/"make it better" → ask which file, what behavior is wrong.
2. **Quick research** — minimal read pass: read each explicitly referenced file, 1-2 targeted greps for the symbol/pattern being changed, skim the nearest convention file for style rules, find one similar existing implementation as a pattern reference if adding new behavior. Hard limit: 5 file reads maximum — more likely means the task exceeds simple-mode scope.
3. **State approach** — 2-3 sentences before touching code: what changes, where it lives (file:line if known), why this approach. No spec, no design doc, no options list.
4. **Scope guard (hard stop)** — after research, count the impact: ≤5 files AND ≤100 lines AND known subsystem → proceed. Otherwise → hard stop, route to `brainstorm` → `to-plan full` → `work full`. This does not negotiate, even if the user says "just do it" — they must explicitly invoke `work full` to proceed past this gate.
5. **Execute** — prefer direct edits for 1-3 line changes in a single file; delegate only for one heavy isolated task (e.g. generating a new file from a clear template), never for exploration. No waves, phases, or task lists — single-shot execution. Write nothing outside the files identified in steps 2-3. If a new assumption surfaces mid-execution that would expand scope past the guard, stop and surface it.
6. **Light verify** — run the narrowest check that proves the change works: the most relevant test file(s), lint on changed files if fast lint exists. If no test exists for the changed code, state that explicitly — never claim "tests pass" vacuously. Capture verification output verbatim.
7. **Handoff suggest** — state what changed (files, line delta), state the verification result (command + output), suggest one natural next move: a review pass if the change is non-trivial or touches a security boundary, committing if the change is clean and ready, the `handoff` playbook if stopping for the day, or nothing if it's a one-liner the user will handle themselves. Never auto-run these — suggest, then stop.

### Simple mode output log (optional)
If `.kit/reports/` exists, write a one-line summary to `.kit/reports/work/{YYYYMMDD}-{slug}.md`:
```yaml
---
date: YYYY-MM-DD
mode: simple
input: {one-line description of prompt or file ref}
files_changed: [path1, path2]
lines_delta: +N -N
verification: {command} → pass / fail
---
```
Optional, but helps `watzup` capture simple-mode work alongside full-pipeline phases.

## Notes Mode (`--notes` flag)

Output file: `.kit/implementation-notes.md` — append-only; create on first entry if absent, with header `# Implementation Notes` / `work --notes — {date}. Spec: {SPEC.md title or prompt slug}.`

Write one entry when a task involves: a decision the spec didn't cover, a change from what the plan specified, a tradeoff between two valid options, a wrong assumption discovered mid-execution, or status `DONE_WITH_CONCERNS`/`NEEDS_CONTEXT`. No entry for tasks that ran exactly as planned.

Entry format:
```
## {timestamp} {phase/task}
**Decision**: what you chose and why (1-2 sentences).
**Spec gap**: what the spec or plan didn't cover.
**Tradeoff**: what you gave up.
**Risk**: one sentence.
```

Rules: write the entry immediately after the task, not batched at phase end; append only, never edit or delete prior entries; when the flag is absent, stay silent with zero overhead.

## Artifacts

### Run artifact — `.kit/runs/work/{YYYYMMDD-HHmm}-{slug}.md`

```markdown
---
id: {ULID}
type: run
phase: {phase-slug} | none
lane: {tiny|normal|high-risk}
mode: {full|simple}
plan_id: {ULID of the phase PLAN this run executes} | none
trace_ids: [{ULID}, ...]
created: {YYYY-MM-DD}
updated: {YYYY-MM-DD}
---

# COOK RUN

Run ID: work-YYYYMMDD-HHmm-{slug}
Mode: full | simple
Status: running | blocked | passed | aborted
Spec: .kit/planning/SPEC.md | none
Roadmap: .kit/planning/ROADMAP.md | none
Phase: {phase-slug} | none
Plan: .kit/planning/phases/{phase-slug}/{phase-slug}-PLAN.md | none
Started At: YYYY-MM-DD HH:mm

## Preflight
- scope drift: yes | no
- working tree note
- required artifacts present: yes | no
- selected phase / source prompt

## Wave / Task Log
### Wave 1
#### T1 — Task name
- status: DONE | DONE_WITH_CONCERNS | NEEDS_CONTEXT | BLOCKED
- changed files:
  - path
- verification:
  - command → pass | fail
- notes:
  - concern or blocker detail

## Summary
- passed tasks
- blocked tasks
- unresolved concerns

## Next Recommended Action
- `check full`
- `to-plan phase {slug}`
- `brainstorm refine`
- `handoff`
```

Rules: one new run file per invocation, never overwrite an older one; every task attempted appears in the log; blocker reasons map to the stop taxonomy; after each wave reaches `DONE`/`DONE_WITH_CONCERNS`, run `zharness trace add ...` and append the returned `id` to the frontmatter `trace_ids` list; `plan_id` links to the phase PLAN this run executes (or the phase's story ULID if no dedicated plan ULID exists), `trace_ids` accumulates one ULID per `trace add` call; `mode` must match the resolved mode from Mode Resolution above — it is what `validate` reads to apply full-mode strictness or the simple-mode carve-out.

## Command Reference

- `zharness --version` — version gate
- `zharness id --json` — mint one fresh ULID without DB state; call separately for the RUN id and each manually-authored changeset filename
- `zharness init` — idempotent; run if no db exists yet
- `zharness query state --json` / `zharness query phases --json` — first lookup index for context and post-run/wave state checks; always verify the pointed files before acting
- `zharness db changeset apply {path} --json` — registers the run (two-line changeset) and any other harness-durable pointer updates
- `zharness trace add --wave {N} --summary "{one-line wave outcome}" --run-id {run id} --json` — fired once per completed wave

## Exit / Handoff Conditions

- State routing produced `ready` or a clean stop with a routed-to playbook
- A run artifact was created and updated with preflight + task status
- Selected phase: every wave executed, every task verified, `check` returned a clean gate
- Handoff suggestion stated; no auto-commit, no auto-wrap
- If stopped early: blocker named with stop taxonomy, next action obvious

## Anti-Patterns

- Walking the roadmap when the spec is stale — re-locks scope without authority
- Running `check` only at the end of the roadmap instead of per phase — bundles unrelated risk
- Delegating a one-line edit to a sub-task — context overhead with no benefit
- Skipping verification because "it obviously works" — every task carries a check command for a reason
- Never invent missing artifacts to "unblock" execution
- Never edit `.kit/planning/SPEC.md` or `.kit/planning/ROADMAP.md` from inside `work` — route back to `brainstorm` or `to-plan` instead
- Never proceed past a `BLOCKED` status without user input
- Never bypass the simple-mode scope guard — not even when the user says "just do it"
- Surface every `BLOCKED` or `DONE_WITH_CONCERNS` status to the user before continuing
