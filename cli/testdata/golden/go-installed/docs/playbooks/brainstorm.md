# Playbook: brainstorm

## Purpose

Explore a decision in the response, or lock/refine the durable initiative at `docs/plans/active/{slug}.md`. Owns only the plan's `## Goal`; never designs phases or executes implementation.

## Modes

Choose from the request shape; ask only when mode or scope is genuinely ambiguous.

| Mode | Input | Durable effect |
|---|---|---|
| `explore` | Trade-off or recommendation request without lock intent | None; answer in the response only |
| `lock-from-idea` | Raw idea or bounded initiative | Create one active plan |
| `lock-from-files` | Authoritative source files | Create one active plan |
| `refine` | Existing active plan needs scope changes | Update the same plan |

**Zero-write rule:** explore creates no plans, reports, changesets, or markdown artifacts. IF exploration reaches lock intent → switch modes explicitly before writing.

## Lanes

Pick the lane from risk, not size, and record it in frontmatter `lane:`.

| Lane | When | Plan |
|---|---|---|
| `tiny` | Known files, a direct success check, nothing hard to reverse | None: route to `work bounded` |
| `normal` | One reviewable change with a known verification path | One phase with `phase_slug:`, `status:`, and a flat task list |
| `high-risk` | User-data migration, public contract, security, or dependent steps | Phases, waves, and tasks; every gate needs an independent judge |

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

1. **Classify** — mode, lane, risk flags, affected surfaces. `tiny` → stop and route to `work bounded`.
2. **Gather minimum authority** — read named sources and repository instructions. Check prior lessons first: `grep -ri "<topic keywords>" docs/memory/` (plain committed files, never a database). Discovery may clarify scope, never expand it.
3. **Compare options** — 2–3 viable paths, or 1–2 alternatives rejected by authoritative sources. State recommendation and trade-offs before locking.
4. **Clarify the boundary** — require a concrete outcome, a measurable `success_signal:`, the affected actors, constraints, accepted requirements each with an `acceptance:` check, and non-goals. Stop rather than invent an unresolved product decision.
5. **Choose the slug** — short and stable. The canonical active path is `docs/plans/active/{slug}.md`; never create a second durable initiative markdown for the same work.
6. **Answer project identity (the stage's single forced write)** — IF `docs/PROJECT.md` is absent → copy `cli/docs/embedded/templates/project.identity.md`; IF that template is also absent (consumer repo) → run `zharness install`. Fill every identity question inline. The lock never completes while any `<...>` question remains: halt and name them. Only the owner-facing scope decision may justify pausing here.
7. **Create a new lock** — confirm no non-empty plan exists under `docs/plans/active/`; IF one exists → stop and name it; the owner must complete or move it aside first. Create `docs/plans/active/{slug}.md` from the skeleton with the lane set, `## Goal` filled, and the bootstrap values shown everywhere else.
8. **Refine an existing lock** — read the plan; preserve the lane (unless reclassification is explicitly approved) and every section except `## Goal`; update Goal in place. IF the refinement changes what the project is or how it is verified → update the affected `docs/PROJECT.md` answers in the same pass.
9. **Self-review** — confirm: no `<...>` placeholder remains in the plan or `docs/PROJECT.md`; requirements are numbered and falsifiable, and each has an `acceptance:`; `success_signal:` is checkable; outcome, requirements, and non-goals agree; rejected alternatives were surfaced; bootstrap state is honest; no second markdown exists.
10. **Review gate** — show the plan path, the answered `docs/PROJECT.md`, and a concise decision summary. Explicit execution intent may satisfy the gate only when scope is bounded and no unresolved product decision, destructive action, or outward-facing action remains; otherwise wait for approval before routing to `to-plan`.

## Exit Conditions

- Explore: recommendation, rationale, and rejected alternatives in the response; zero durable writes.
- Lock: exactly one active plan satisfying steps 6–9; next action `to-plan`, which updates the same file.
- Refine: same plan; later-stage sections preserved.
