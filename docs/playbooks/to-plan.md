# Playbook: to-plan

## Purpose

Turn the locked initiative definition in `docs/plans/active/{slug}.md` into an executable approach in that same file. This stage owns Approach and Risks plus Phases and Verification. It does not execute work or create another markdown artifact.

## Preconditions

1. Require the intended active plan at `docs/plans/active/{slug}.md` with `status: active`, stable `id`/`intake_id`, and complete Outcome, Authority and Requirements, and Non-goals sections. Exactly one non-empty plan may exist under `docs/plans/active/`. If the initiative is ambiguous, route to `brainstorm refine` rather than guessing.

## Arguments

- `full` — define the complete approach, phases, waves, tasks, and checks that are still `not-planned`.
- `phase {stable-phase-slug}` — define one not-yet-planned phase in the same plan without replacing existing phase content.

## Owned Plan Sections

- `## Approach and Risks` — chosen implementation approach, constraints, dependencies, rejected alternatives, risks, mitigations, and recovery.
- `## Phases and Verification` — stable phase slugs and story IDs, dependency order, waves, tasks, expected outputs, checks, and current lifecycle status.

Preserve every other section. After a phase/task definition is written, it is immutable; only that phase's lifecycle `status` may later change to mirror execution transitions. Keep frontmatter `status: active` and refresh only `updated`.

## Steps

1. **Read and normalize the active plan** — extract outcome, authority, requirements, non-goals, lane, constraints, and validation expectations. Every planned task must trace to an accepted requirement.
2. **Choose the smallest viable approach** — state the implementation path, why it is preferred, rejected alternatives, primary risks, mitigations, and stop/recovery conditions in Approach and Risks.
3. **Define stable phases** — split only where dependency, risk, or independent verification warrants it. Each phase needs a stable slug, goal, dependencies, allowed/avoided surfaces, and lifecycle status from `planned|in-progress|checked|done`. Treat every definition field except lifecycle status as immutable once written.
4. **Assign phase identities** — for each new phase mint a stable `story_id`: a unique token (timestamp-suffixed slug is acceptable) written beside the phase with its plan status set to `planned`. On repeated invocation preserve existing story IDs, statuses, and definitions; do not recreate identities, rewrite definitions, or reset progress.
5. **Build executable waves** — under each new phase list waves, tasks, dependencies, touched/avoided surfaces, expected outputs, verification commands, stop conditions, and escalation route. Put tasks in the same wave only when they can proceed independently. Do not add task status fields; append-only `## Progress` is the sole task execution-status source.
6. **Write checks before execution** — every meaningful task gets an observable command or inspection. Missing verification is a planning blocker, not something `work` may invent later.
7. **Update the same file** — replace only the two owned sections, preserve the plan identity and all append-only lifecycle history, remove the `not-planned` bootstrap values from those sections, and refresh `updated`.
8. **Verify coherence** — confirm one `story_id` per listed phase, acyclic dependencies, and statuses coherent with the append-only history; confirm no second initiative markdown was created anywhere in the tree.
9. **Handoff** — name the first executable phase and route to `work full`.

## Planning Rules

- Scope comes from the active plan; planning cannot add product behavior.
- Phase slugs and story IDs are stable after creation.
- Phase/task definitions are immutable after planning; work/check/handoff may change only that phase's lifecycle status in this file.
- Append-only `## Progress` is the sole task execution-status source; task definitions contain no status fields.
- Phases are ordered by real dependency and risk reduction, not arbitrary size.
- Waves expose executable coordination; tasks expose exact proof.
- Lifecycle lives in the plan file itself; never write parallel task/phase state anywhere else.

## Exit Conditions

Complete only when the same active plan contains a decision-complete approach, explicit risks/recovery, stable phases with story IDs and coherent statuses, executable waves/tasks, and exact verification commands. The next action must name the phase that `work full` should start.
