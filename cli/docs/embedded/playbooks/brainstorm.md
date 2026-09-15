# Playbook: brainstorm

## Purpose

Explore a decision in the response, or lock/refine the durable initiative at `docs/plans/active/{slug}.md`. Owns only the initiative definition; never designs phases or executes implementation.

## Modes

Choose from the request shape; ask only when mode or scope is genuinely ambiguous.

| Mode | Input | Durable effect |
|---|---|---|
| `explore` | Trade-off or recommendation request without lock intent | None; answer in the response only |
| `lock-from-idea` | Raw idea or bounded initiative | Create one active plan |
| `lock-from-files` | Authoritative source files | Create one active plan |
| `refine` | Existing active plan needs scope changes | Update the same plan; preserve its IDs |

**Zero-write rule:** explore creates no plans, reports, changesets, or markdown artifacts. IF exploration reaches lock intent → switch modes explicitly before writing.

## Owned Plan Sections

- `## Outcome` — observable result and success conditions.
- `## Authority and Requirements` — authoritative sources plus numbered, falsifiable requirements.
- `## Non-goals` — explicit exclusions and deferred scope.

Preserve all later-stage content. Once `to-plan` has defined phases and tasks, those definitions are immutable: a refinement never alters them, replaces the plan, mints a new plan ID, or resets lifecycle history. Append-only `## Progress` is the sole task execution-status source; task definitions never contain status fields.

## Steps

1. **Classify** — input type (`new-spec`, `spec-slice`, `change-request`, `new-initiative`, `maintenance`, `harness-improvement`), lane (`tiny`, `normal`, `high-risk`), risk flags, affected surfaces.
2. **Gather minimum authority** — read named sources and repository instructions. Check prior lessons first: `grep -ri "<topic keywords>" docs/memory/` (plain committed files, never a database). Discovery may clarify scope, never expand it.
3. **Compare options** — 2–3 viable paths, or 1–2 alternatives rejected by authoritative sources. State recommendation and trade-offs before locking.
4. **Clarify the boundary** — require a concrete outcome, actors, constraints, accepted requirements, non-goals, and checkable success conditions. Stop rather than invent an unresolved product decision.
5. **Choose the slug** — short and stable. The canonical active path is `docs/plans/active/{slug}.md`; never create a second durable initiative markdown for the same work.
6. **Answer project identity (the stage's single forced write)** — IF `docs/PROJECT.md` is absent → copy `cli/docs/embedded/templates/project.identity.md`; IF that template is also absent (consumer repo) → run `zharness install`. Fill every identity question inline. The lock never completes while any `<...>` question remains: halt and name them. Only the owner-facing scope decision may justify pausing here.
7. **Create a new lock**:
   - Confirm no non-empty plan exists under `docs/plans/active/`; IF one exists → stop and name it; the owner must complete or move it aside first.
   - Mint two unique identifier tokens locally (timestamp-suffixed is fine): the plan `id` and `intake_id`.
   - Create `docs/plans/active/{slug}.md`; fill frontmatter (both IDs, `status: active`, lane, dates) and the three owned sections.
   - Replace every unowned placeholder with honest bootstrap state: `approach: not-planned`; `planning_status: not-planned`; phases, Progress, Decisions, and Validation as `none`; Current State IDs/blockers as `none`; `exact_next_action: to-plan`.
8. **Refine an existing lock** — read the plan; preserve `id`, `intake_id`, lane (unless reclassification is explicitly approved), and all non-owned sections; update owned sections in place; refresh `updated`. IF the refinement changes what the project is or how it is verified → update the affected `docs/PROJECT.md` answers in the same pass.
9. **Self-review** — confirm: no literal fake lifecycle placeholders remain; `docs/PROJECT.md` has no `<...>` question or template marker; requirements are numbered and falsifiable; Outcome, requirements, and Non-goals agree; rejected alternatives were surfaced; bootstrap state is honest; no second markdown exists.
10. **Review gate** — show the plan path, the answered `docs/PROJECT.md`, and a concise decision summary. Explicit execution intent may satisfy the gate only when scope is bounded and no unresolved product decision, destructive action, or outward-facing action remains; otherwise wait for approval before routing to `to-plan`.

## Exit Conditions

- Explore: recommendation, rationale, and rejected alternatives in the response; zero durable writes.
- Lock: exactly one active plan satisfying steps 6–9; next action `to-plan`, which updates the same file.
- Refine: same plan and IDs; later-stage sections preserved.
