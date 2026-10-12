# Playbook: to-plan

Turn the locked `## Goal` of `docs/plans/active/{slug}.md` into `## Phases and Verification` in that same file. Requires exactly one non-empty active plan with a complete Goal; ambiguous → `brainstorm refine`. Arguments: `full` (whole plan) or `phase {slug}` (add one phase).

## Section Shape

`normal`: one phase, flat task list. `high-risk`: add `constraints:` and `recovery:`, and waves per phase.

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

Write `status:` directly under `phase_slug:`, indented two spaces with no bullet (the guard reads only that form). Values: `planned|in-progress|checked|done`.

After a phase/task definition is written, it is immutable; work/check/handoff may change only that phase's lifecycle status in this file. Do not add task status fields. `## Log` is the sole task execution-status source; never write parallel task/phase state anywhere else.

## Steps

1. **Read the Goal** — requirements, non-goals, lane, and the validation expectations (`success_signal:`, each `acceptance:`).
2. **Approach** — the smallest viable path, the rejected alternative, risks and recovery.
3. **Phases** — split only for dependency, risk, or independent verification; each phase leaves the system usable.
4. **Tasks** — vertical tracer bullets (one thin path through every touched layer, one context window each); prefactor first; a wide mechanical refactor is its own task.
5. **Seams** — prefer existing interfaces; confirm a new seam with the user.
6. **Checks** — every task gets a command at a seam; a missing check is a planning blocker.
7. **Write** — replace `approach: not-planned`; set `exact_next_action: work full phase {first-slug}`.
8. **Verify** — unique slugs, acyclic dependencies, no second plan file; every requirement appears in some phase `goal:` and every `acceptance:` maps to a check.
