# Playbook: brainstorm (lock and refine)

Loaded by `docs/playbooks/brainstorm.md` for `lock` and `refine`, together with `docs/playbooks/brainstorm-grill.md`. Steps 1–3 stay in `brainstorm.md`.

## Plan Skeleton

A lock writes exactly this shape. The pre-commit guard parses the headings `## Phases and Verification`, `## Validation`, and `## Current State and Next Action` and the fields `lane:`, `phase_slug:`, `status:`, and `lifecycle_status:`; never rename them.

```markdown
---
status: active
lane: normal
---
# <Title>

## Goal
- outcome: <observable result>
- success_signal: <measurable check that proves the outcome>
- actors: <who uses or is affected by the change>
- authority: <owner decision, spec, or source files>
- requirements:
  - R1: <falsifiable requirement> | acceptance: <how it is verified> | source: <authority>
- non-goals:
  - NG1: <explicit exclusion>

## Phases and Verification
- approach: not-planned

## Log

## Validation

## Current State and Next Action
- active_phase: none
- lifecycle_status: planned
- blockers: none
- open_items: none
- exact_next_action: to-plan
```

`to-plan` writes `## Phases and Verification`; `work` appends to `## Log`; `check` appends to `## Validation`; `work`, `check`, and `handoff` refresh `## Current State and Next Action`. `## Log` and `## Validation` are append-only, and `## Log` is the sole task execution-status source: task definitions never carry status fields. Once `to-plan` has defined phases and tasks, those definitions are immutable: a refinement never alters them, replaces the plan, or resets lifecycle history.

## Steps

4. **Clarify the boundary** — run the Grill loop in `docs/playbooks/brainstorm-grill.md` over whatever is missing (the whole Goal in `grill`; only the gaps in `lock`/`refine`). Require a concrete outcome, a measurable `success_signal:`, the affected actors, constraints, accepted requirements each with an `acceptance:` check, and non-goals. Stop rather than invent an unresolved product decision.
5. **Choose the slug** — short and stable. The canonical active path is `docs/plans/active/{slug}.md`; never create a second durable initiative markdown for the same work.
6. **Answer project identity (the stage's single forced write)** — IF `docs/PROJECT.md` is absent → copy `cli/docs/embedded/templates/project.identity.md`; IF that template is also absent (consumer repo) → run `zharness install`. Fill every identity question inline. The lock never completes while any `<...>` question remains: halt and name them. Only the owner-facing scope decision may justify pausing here.
7. **Create a new lock** — confirm no non-empty plan exists under `docs/plans/active/`; IF one exists → stop and name it; the owner must complete or move it aside first. Create `docs/plans/active/{slug}.md` from the skeleton with the lane set, `## Goal` filled, and the bootstrap values shown everywhere else.
8. **Refine an existing lock** — read the plan; preserve the lane (unless reclassification is explicitly approved) and every section except `## Goal`; update Goal in place. IF the refinement changes what the project is or how it is verified → update the affected `docs/PROJECT.md` answers in the same pass.
9. **Self-review** — confirm: no `<...>` placeholder remains in the plan or `docs/PROJECT.md`; requirements are numbered and falsifiable, and each has an `acceptance:`; `success_signal:` is checkable; outcome, requirements, and non-goals agree; rejected alternatives were surfaced; bootstrap state is honest; no second markdown exists.
10. **Review gate** — show the plan path, the answered `docs/PROJECT.md`, and a concise decision summary. Explicit execution intent may satisfy the gate only when scope is bounded and no unresolved product decision, destructive action, or outward-facing action remains; otherwise wait for approval before routing to `to-plan`.
