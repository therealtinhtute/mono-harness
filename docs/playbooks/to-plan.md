# Playbook: to-plan

## Purpose

Turn the locked initiative definition in `docs/plans/active/{slug}.md` into an executable approach in that same file. This stage owns Approach and Risks plus Phases and Verification. It creates story rows for stable phases but does not execute work or create another markdown artifact.

## Preconditions

1. Run `zharness preflight to-plan --mode full --json`. Missing binary: print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop. Otherwise check its `version` field — a `dev` build satisfies the gate; below MIN_ZHARNESS_VERSION (`0.8.1` — see `skills/workflow/README.md`), print the same message and stop. Then follow its stop/recovery result. Durable planning requires an initialized readable/writable database and ready managed docs.
2. Require the intended active plan at `docs/plans/active/{slug}.md` with `status: active`, stable `id`/`intake_id`, and complete Outcome, Authority and Requirements, and Non-goals sections. If the initiative is ambiguous, route to `brainstorm refine` rather than guessing.

## Arguments

- `full` — define the complete approach, phases, waves, tasks, and checks that are still `not-planned`.
- `phase {stable-phase-slug}` — define one not-yet-planned phase in the same plan without replacing existing phase content.

## Owned Plan Sections

- `## Approach and Risks` — chosen implementation approach, constraints, dependencies, rejected alternatives, risks, mitigations, and recovery.
- `## Phases and Verification` — stable phase slugs and story IDs, dependency order, waves, tasks, expected outputs, checks, and current lifecycle status.

Preserve every other section. After a phase/task definition is written, it is immutable; only that phase's lifecycle `status` may later change to mirror DB transitions. Keep frontmatter `status: active` and refresh only `updated`.

## Steps

1. **Read and normalize the active plan** — extract outcome, authority, requirements, non-goals, lane, constraints, and validation expectations. Every planned task must trace to an accepted requirement.
2. **Choose the smallest viable approach** — state the implementation path, why it is preferred, rejected alternatives, primary risks, mitigations, and stop/recovery conditions in Approach and Risks.
3. **Define stable phases** — split only where dependency, risk, or independent verification warrants it. Each phase needs a stable slug, goal, dependencies, allowed/avoided surfaces, and lifecycle status from `planned|in-progress|checked|done`. Treat every definition field except lifecycle status as immutable once written.
4. **Create story rows once** — for each new phase run `zharness story --slug {stable-phase-slug} --goal "{phase goal}" [--depends-on {slug}] --json`. Write the returned story ID beside that phase and set its plan status to `planned`, matching the new DB row. On repeated invocation, preserve existing story IDs, current statuses, and definitions; do not recreate rows, rewrite definitions, or reset progress.
5. **Build executable waves** — under each new phase list waves, tasks, dependencies, touched/avoided surfaces, expected outputs, verification commands, stop conditions, and escalation route. Put tasks in the same wave only when they can proceed independently. Do not add task status fields; append-only `## Progress` is the sole task execution-status source.
6. **Write checks before execution** — every meaningful task gets an observable command or inspection. Missing verification is a planning blocker, not something `work` may invent later.
7. **Update the same file** — replace only the two owned sections, preserve the plan identity and all append-only lifecycle history, remove the `not-planned` bootstrap values from those sections, and refresh `updated`.
8. **Verify durable state** — run `zharness query phases --json`; confirm one story per listed phase, matching slugs, IDs, dependencies, and `planned` statuses for new phases. Confirm the plan phase statuses match the DB and no second initiative markdown was created.
9. **Handoff** — name the first executable phase and route to `work full`.

## Planning Rules

- Scope comes from the active plan; planning cannot add product behavior.
- Phase slugs and story IDs are stable after creation.
- Phase/task definitions are immutable after planning; work/check/handoff may change only phase lifecycle status to mirror their DB transitions.
- Append-only `## Progress` is the sole task execution-status source; task definitions contain no status fields.
- Phases are ordered by real dependency and risk reduction, not arbitrary size.
- Waves expose executable coordination; tasks expose exact proof.
- Do not hand-author database changes or lifecycle changesets; use the public story command.

## Command Reference

- `zharness preflight to-plan --mode full --json`
- `zharness story --slug {stable-phase-slug} --goal "..." [--depends-on {slug}] --json`
- `zharness query phases --json`

## Exit Conditions

Complete only when the same active plan contains a decision-complete approach, explicit risks/recovery, stable phases with story IDs and DB-matching statuses, executable waves/tasks, and exact verification commands. The next action must name the phase that `work full` should start.
