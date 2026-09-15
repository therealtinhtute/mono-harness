# Playbook: to-plan

## Purpose

Turn the locked initiative in `docs/plans/active/{slug}.md` into an executable approach in that same file. Never executes work or creates another markdown artifact.

## Preconditions

Require the active plan with `status: active`, stable `id`/`intake_id`, and complete Outcome, Authority and Requirements, and Non-goals sections. Exactly one non-empty plan may exist under `docs/plans/active/`. IF the initiative is ambiguous → route to `brainstorm refine`; never guess.

## Arguments

- `full` — plan the complete approach, phases, waves, tasks, and checks still `not-planned`.
- `phase {stable-phase-slug}` — plan one not-yet-planned phase without replacing existing phase content.

## Owned Plan Sections

- `## Approach and Risks` — approach, constraints, dependencies, rejected alternatives, risks, mitigations, recovery.
- `## Phases and Verification` — stable phase slugs and story IDs, dependency order, waves, tasks, expected outputs, checks, lifecycle status.

Invariants:

- Preserve every other section, the plan identity, and all append-only history. Keep frontmatter `status: active`; refresh only `updated`.
- Scope comes from the plan; planning never adds product behavior. Every task traces to an accepted requirement.
- After a phase/task definition is written, it is immutable, slugs and story IDs included; work/check/handoff may change only that phase's lifecycle status in this file.
- Append-only `## Progress` is the sole task execution-status source. Do not add task status fields.
- Lifecycle lives in the plan file; never write parallel task/phase state anywhere else.

## Steps

1. **Read the plan** — extract outcome, authority, requirements, non-goals, lane, constraints, and validation expectations.
2. **Choose the smallest viable approach** — write the path, why it is preferred, rejected alternatives, risks, mitigations, and stop/recovery conditions into Approach and Risks.
3. **Define phases** — split only where dependency, risk, or independent verification warrants; order by real dependency and risk reduction, not size. Each phase: stable slug, goal, dependencies, allowed/avoided surfaces, lifecycle status from `planned|in-progress|checked|done`.
4. **Assign identities** — for each new phase mint a stable `story_id` (unique token; a timestamp-suffixed slug is fine) beside the phase, status `planned`. On repeat invocation preserve existing IDs, statuses, definitions, and progress.
5. **Build waves** — per new phase: waves, tasks, dependencies, touched/avoided surfaces, expected outputs, verification commands, stop conditions, escalation route. Same wave only for tasks that can proceed independently.
6. **Write checks before execution** — every meaningful task gets an observable command or inspection. Missing verification is a planning blocker; `work` may not invent it later.
7. **Update the file** — replace only the two owned sections, removing their `not-planned` bootstrap values.
8. **Verify coherence** — one `story_id` per listed phase, acyclic dependencies, statuses coherent with append-only history, no second initiative markdown anywhere in the tree.

## Exit Conditions

Complete only when the plan holds a decision-complete approach, explicit risks/recovery, stable phases with story IDs and coherent statuses, executable waves/tasks, and exact verification commands, and the next action names the first executable phase for `work full`.
