# Playbook: brainstorm (lock and refine)

Loaded by `docs/playbooks/brainstorm.md` for `lock` and `refine`, with `brainstorm-grill.md`.

## Plan Skeleton

A lock writes exactly this shape; the pre-commit guard parses its headings and `lane:`/`phase_slug:`/`status:`/`lifecycle_status:`, so never rename them.

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

Each later section belongs to its stage. `## Log` is the sole task execution-status source. Once `to-plan` defines phases and tasks, those definitions are immutable: a refine never alters them or resets history.

## Steps

5. **Slug** — short and stable. The canonical active path is `docs/plans/active/{slug}.md`; never a second file for the same work.
6. **Project identity** — IF `docs/PROJECT.md` is absent → copy `cli/docs/embedded/templates/project.identity.md` (consumer repo: `zharness install`). Answer every question; the lock never completes while a `<...>` remains.
7. **Lock** — confirm no non-empty plan exists under `docs/plans/active/`; IF one exists → stop and name it. Write the skeleton with the lane and `## Goal` filled.
8. **Refine** — edit only `## Goal` (lane changes need approval); update `docs/PROJECT.md` IF what the project is or how it is verified changed.
9. **Self-review** — no `<...>` left; requirements numbered, falsifiable, each with `acceptance:`; `success_signal:` checkable; Goal parts agree.
10. **Review gate** — show the plan path, the answered `docs/PROJECT.md`, and a decision summary; wait for approval before `to-plan`, unless the user already asked to proceed, scope is bounded, and nothing destructive, outward-facing, or undecided remains.
