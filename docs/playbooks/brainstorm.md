# Playbook: brainstorm

## Purpose

Explore a decision in the response, or lock/refine the durable initiative at `docs/plans/active/{slug}.md`. For durable work, this stage owns the plan's Outcome, Authority and Requirements, and Non-goals sections. It does not design phases or execute implementation.

## Preconditions

1. Choose `explore`, `lock-from-idea`, `lock-from-files`, or `refine` from the request shape. Ask only when mode or scope is genuinely ambiguous.
2. Run `zharness preflight brainstorm --mode {explore|lock} --json`. Missing binary: print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop. Otherwise check its `version` field — a `dev` build satisfies the gate; below MIN_ZHARNESS_VERSION (`0.8.1` — see `skills/workflow/README.md`), print the same message and stop. Then follow its status and recovery exactly. Explore may continue in reduced mode; lock/refine requires durable readiness.

## Modes

| Mode | Input | Durable effect |
|---|---|---|
| `explore` | Trade-off or recommendation request without lock intent | None; answer in the response only |
| `lock-from-idea` | Raw idea or bounded initiative | Create one active plan and one intake row |
| `lock-from-files` | Authoritative source files | Create one active plan and one intake row |
| `refine` | Existing active plan needs scope changes | Update the same plan; preserve its IDs |

**Zero-write rule:** explore creates no lifecycle rows, plans, reports, changesets, or markdown artifacts. If exploration leads to lock intent, switch modes explicitly and rerun preflight in lock mode before writing anything.

## Owned Plan Sections

Brainstorm writes or refines only these initiative-definition sections:

- `## Outcome` — the observable result and success conditions.
- `## Authority and Requirements` — authoritative sources plus numbered, falsifiable requirements.
- `## Non-goals` — explicit exclusions and deferred scope.

Preserve all later-stage content already present in the plan. Once `to-plan` has defined phases and tasks, those definitions are immutable; a refinement must not alter them, replace the plan, mint a new plan ID, or create another intake row. Append-only `## Progress` is the sole task execution-status source; task definitions never contain status fields.

## Steps

1. **Resolve intent and classification** — choose input type (`new-spec`, `spec-slice`, `change-request`, `new-initiative`, `maintenance`, or `harness-improvement`), lane (`tiny`, `normal`, or `high-risk`), applicable risk flags, and affected surfaces.
2. **Gather minimum authority** — read named source files and repository instructions. Discovery may clarify scope but must not expand it.
3. **Compare options** — evaluate 2–3 viable paths, or identify 1–2 alternatives rejected by authoritative source files. State the recommendation and trade-offs before locking.
4. **Clarify the boundary** — require a concrete outcome, actors, constraints, accepted requirements, non-goals, and checkable success conditions. Stop instead of inventing an unresolved product decision.
5. **Choose the stable slug** — use a short initiative slug. The canonical active path is `docs/plans/active/{slug}.md`; do not create a second durable initiative markdown for the same work.
6. **Create a new lock**:
   - Run `zharness id --json` and keep the returned ID as the plan's own `id`.
   - Run `zharness scaffold plan --path docs/plans/active/{slug}.md --json`.
   - Run `zharness intake --type {input-type} --summary "{one-line summary}" --lane {lane} --plan-path docs/plans/active/{slug}.md --json`; keep the returned ID as `intake_id`.
   - Fill frontmatter with both IDs, `status: active`, lane, and dates, then fill the three owned sections.
   - Replace every unowned template placeholder with honest bootstrap state: `approach: not-planned`; `planning_status: not-planned`; phases, Progress, Decisions, and Validation as `none`; Current State IDs/blockers as `none`; and `exact_next_action: to-plan`.
7. **Refine an existing lock** — read the active plan, preserve `id`, `intake_id`, lane unless reclassification is explicitly approved, and all non-owned sections; update the three owned sections in place and refresh `updated`.
8. **Self-review** — remove all literal scaffold placeholders; confirm requirements are numbered and falsifiable; confirm Outcome, requirements, and Non-goals do not contradict; confirm rejected alternatives were surfaced; confirm the bootstrap state is honest; confirm no second markdown was created.
9. **Review gate and handoff** — show the active plan path and concise decision summary. Explicit execution intent may satisfy the procedural gate when scope is bounded and no unresolved product decision, destructive action, or outward-facing action remains. Otherwise wait for approval before routing to `to-plan`.

## Command Reference

- `zharness preflight brainstorm --mode {explore|lock} --json`
- `zharness id --json`
- `zharness scaffold plan --path docs/plans/active/{slug}.md --json`
- `zharness intake --type {input-type} --summary "..." --lane {lane} --plan-path docs/plans/active/{slug}.md --json`

## Exit Conditions

- Explore: one recommendation, rationale, and rejected alternatives in the response; zero durable writes.
- Lock: exactly one active plan exists with valid plan/intake IDs, complete owned sections, no literal fake lifecycle placeholders, honest `not-planned`/`none` bootstrap state, and exact next action `to-plan`.
- Refine: the same active plan and IDs remain, with later-stage sections preserved.
- Durable next step: `to-plan` updates the same file.
