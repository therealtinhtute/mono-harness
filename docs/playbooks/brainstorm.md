# Playbook: brainstorm

## Purpose

Explore a decision in the response, or lock/refine the durable initiative at `docs/plans/active/{slug}.md`. For durable work, this stage owns the plan's Outcome, Authority and Requirements, and Non-goals sections. It does not design phases or execute implementation.

## Preconditions

1. Choose `explore`, `lock-from-idea`, `lock-from-files`, or `refine` from the request shape. Ask only when mode or scope is genuinely ambiguous.

## Modes

| Mode | Input | Durable effect |
|---|---|---|
| `explore` | Trade-off or recommendation request without lock intent | None; answer in the response only |
| `lock-from-idea` | Raw idea or bounded initiative | Create one active plan |
| `lock-from-files` | Authoritative source files | Create one active plan |
| `refine` | Existing active plan needs scope changes | Update the same plan; preserve its IDs |

**Zero-write rule:** explore creates no lifecycle rows, plans, reports, changesets, or markdown artifacts. If exploration leads to lock intent, switch modes explicitly before writing anything.

## Owned Plan Sections

Brainstorm writes or refines only these initiative-definition sections:

- `## Outcome` — the observable result and success conditions.
- `## Authority and Requirements` — authoritative sources plus numbered, falsifiable requirements.
- `## Non-goals` — explicit exclusions and deferred scope.

Preserve all later-stage content already present in the plan. Once `to-plan` has defined phases and tasks, those definitions are immutable; a refinement must not alter them, replace the plan, mint a new plan ID, or reset lifecycle history. Append-only `## Progress` is the sole task execution-status source; task definitions never contain status fields.

## Steps

1. **Resolve intent and classification** — choose input type (`new-spec`, `spec-slice`, `change-request`, `new-initiative`, `maintenance`, or `harness-improvement`), lane (`tiny`, `normal`, or `high-risk`), applicable risk flags, and affected surfaces.
2. **Gather minimum authority** — read named source files and repository instructions. Check prior durable lessons first: `grep -ri "<topic keywords>" docs/memory/` — memory is plain committed files, retrieved by direct grep, never a database. Discovery may clarify scope but must not expand it.
3. **Compare options** — evaluate 2–3 viable paths, or identify 1–2 alternatives rejected by authoritative source files. State the recommendation and trade-offs before locking.
4. **Clarify the boundary** — require a concrete outcome, actors, constraints, accepted requirements, non-goals, and checkable success conditions. Stop instead of inventing an unresolved product decision.
5. **Choose the stable slug** — use a short initiative slug. The canonical active path is `docs/plans/active/{slug}.md`; do not create a second durable initiative markdown for the same work.
6. **Force the identity write (the stage's single forced write step)** — ensure `docs/PROJECT.md` exists: if absent, copy `cli/docs/embedded/templates/project.identity.md` into place. Fill every identity question inline; the lock does not complete while any question remains in unanswered `<...>` form — halt and name the unanswered questions. Only the owner-facing scope decision may justify pausing here; an unanswered PROJECT.md never locks.
7. **Create a new lock**:
   - Confirm at most one non-empty plan exists under `docs/plans/active/`; if one exists, stop and name it — it must be completed or moved aside by the owner first.
   - Mint two unique identifier tokens locally (timestamp-suffixed tokens are acceptable): one as the plan's own `id`, one as `intake_id`.
   - Create `docs/plans/active/{slug}.md`; fill frontmatter with both IDs, `status: active`, lane, and dates, then fill the three owned sections.
   - Replace every unowned template placeholder with honest bootstrap state: `approach: not-planned`; `planning_status: not-planned`; phases, Progress, Decisions, and Validation as `none`; Current State IDs/blockers as `none`; and `exact_next_action: to-plan`.
8. **Refine an existing lock** — read the active plan, preserve `id`, `intake_id`, lane unless reclassification is explicitly approved, and all non-owned sections; update the three owned sections in place and refresh `updated`. Re-check `docs/PROJECT.md`: if the refinement changes what the project is or how it is verified, update the affected identity answers in the same pass.
9. **Self-review** — remove all literal template placeholders so no literal fake lifecycle placeholders remain; confirm `docs/PROJECT.md` carries no unanswered `<...>` question and no template marker; confirm requirements are numbered and falsifiable; confirm Outcome, requirements, and Non-goals do not contradict; confirm rejected alternatives were surfaced; confirm the bootstrap state is honest; confirm no second markdown was created.
10. **Review gate and handoff** — show the active plan path, the answered `docs/PROJECT.md`, and a concise decision summary. Explicit execution intent may satisfy the procedural gate when scope is bounded and no unresolved product decision, destructive action, or outward-facing action remains. Otherwise wait for approval before routing to `to-plan`.

## Exit Conditions

- Explore: one recommendation, rationale, and rejected alternatives in the response; zero durable writes.
- Lock: exactly one active plan exists with unique plan/intake identifiers, complete owned sections, an answered `docs/PROJECT.md` (no unanswered questions, no template marker), honest `not-planned`/`none` bootstrap state, and exact next action `to-plan`.
- Refine: the same active plan and IDs remain, with later-stage sections preserved.
- Durable next step: `to-plan` updates the same file.
