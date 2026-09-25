# Playbook: to-plan

## Purpose

Turn the locked `## Goal` of `docs/plans/active/{slug}.md` into an executable `## Phases and Verification` in that same file. Never executes work or creates another markdown artifact.

## Preconditions

Require the active plan with `status: active` and a complete `## Goal`. Exactly one non-empty plan may exist under `docs/plans/active/`. IF the initiative is ambiguous → route to `brainstorm refine`; never guess.

## Arguments

- `full` — plan the complete approach, phases, tasks, and checks.
- `phase {stable-phase-slug}` — plan one more phase without replacing existing phase content.

## Owned Plan Section

`## Phases and Verification`: the approach and its risks first, then the phases. A `normal` plan has one phase and a flat task list; a `high-risk` plan adds `constraints:` and `recovery:`, and each phase carries its goal, dependencies, surfaces, `escalate_when:`, waves, and a phase check:

```markdown
## Phases and Verification
- approach: <chosen path and why; the rejected alternative>
- risks: <risk → mitigation; stop and recovery condition>
- seams: <public interfaces where behavior is tested; existing ones first>
- phase_slug: `<slug>`
  status: planned
  - goal: R1, R2 | depends_on: none
  - surfaces: <touched> | avoided: <untouched>
  - escalate_when: <condition that stops the phase and asks the owner>
  - wave 1:
    - T1 <task> — output: <expected result> — check: `<command>` — stop_if: <condition>
  - phase check: `<command>`
```

Write `status:` on its own line directly under `phase_slug:`, indented two spaces with no bullet: the guard's completed-plan check reads only that form. Status values: `planned|in-progress|checked|done`.

Invariants:

- Preserve every other section and all append-only history. Keep frontmatter `status: active`.
- Scope comes from the plan; planning never adds product behavior. Every task traces to a requirement.
- After a phase/task definition is written, it is immutable, slugs included; work/check/handoff may change only that phase's lifecycle status in this file. Do not add task status fields.
- `## Log` is the sole task execution-status source; never write parallel task/phase state anywhere else.

## Steps

1. **Read the plan** — extract outcome, success signal, actors, authority, requirements with their acceptance checks, non-goals, lane, constraints, and the validation expectations they set (`success_signal:` and each `acceptance:`).
2. **Choose the smallest viable approach** — write the path, why it is preferred, the rejected alternative, risks, mitigations, and stop/recovery conditions at the head of the section.
3. **Define phases** — `normal`: one phase. `high-risk`: split only where dependency, risk, or independent verification warrants; order by real dependency and risk reduction, not size; each phase leaves the system usable if the next never lands.
4. **Build tasks** — per phase: tasks (waves only in `high-risk`; same wave only for tasks that can proceed independently), touched and avoided surfaces, and on each task its `output:` and `stop_if:`. `high-risk`: each phase names `escalate_when:`. Slice every task as a **tracer bullet**:
   - Vertical: a thin but complete path through every layer the change touches (data, logic, interface, test), verifiable on its own — never one horizontal layer.
   - Sized to finish in one fresh context window.
   - Prefactor first: a task that makes the change easy precedes the task that makes the change.
   - Exception: a wide mechanical refactor (one rename or retype fanning across many call sites) is one task of its own, checked by the build or type check.
5. **Pick the seams** — list in `seams:` the public interfaces where each task's behavior is observed and tested. Prefer existing seams, then the highest new one; fewer seams is better. IF a seam is new (no test yet exercises that interface) → confirm it with the user before writing the section.
6. **Write checks before execution** — every meaningful task gets an observable command or inspection, run at one of its seams. Missing verification is a planning blocker; `work` may not invent it later.
7. **Update the file** — replace `approach: not-planned` with the section; set Current State `exact_next_action: work full phase {first-phase-slug}`.
8. **Verify coherence** — unique phase slugs, acyclic dependencies, statuses coherent with `## Log`, no second initiative markdown anywhere in the tree. Traceability: every requirement appears in some phase `goal:`, and every requirement's `acceptance:` maps to at least one task `check:` or phase check; a gap is a planning blocker.

## Exit Conditions

Complete only when the plan holds a decision-complete approach, explicit risks and recovery, confirmed seams, tracer-bullet tasks, stable phases with coherent statuses, executable tasks, and exact verification commands, and the next action names the first executable phase for `work full`.
