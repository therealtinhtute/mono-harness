# Playbook: brainstorm

Owns only the plan's `## Goal` at `docs/plans/active/{slug}.md`; a `high-risk` plan's phases belong to `to-plan`, code to `work`.

## Modes

A leading subcommand sets the mode; else infer it (grill/interview/stress-test → `grill`) and ask only when truly ambiguous.

| Mode | Use for | Durable effect |
|---|---|---|
| `explore` (default) | Trade-off or recommendation | None |
| `grill` | Fuzzy idea, or a plan to validate | None until the user approves a lock or refine |
| `raw` | Fast first pass | None |
| `lock` | Create the initiative | One new active plan |
| `refine` | Change an active plan's scope | Same plan, Goal only |

`grill`/`raw`: read [brainstorm-grill.md](brainstorm-grill.md) now. `lock`/`refine`: read [brainstorm-lock.md](brainstorm-lock.md) and [brainstorm-grill.md](brainstorm-grill.md) now.

**Zero-write rule:** explore creates no plans, reports, changesets, or markdown artifacts; neither do `raw` and an unapproved `grill`. Switch modes explicitly before writing.

## Lanes

Pick from uncertainty, impact, dependency and recoverability, never from file or line counts; record it as frontmatter `lane:`.

| Lane | When | Plan |
|---|---|---|
| `tiny` | Known files, direct check, nothing hard to reverse | None |
| `normal` | One reviewable change, known verification | Goal + one phase (lock writes it) |
| `high-risk` | Data migration, public contract, security, dependent steps | Phases and waves via `to-plan`; every gate needs an independent judge |

`normal`: Goal + one phase `main`, no `to-plan`.

## Steps

1. **Classify** — mode, lane, risk, surfaces. `tiny` → stop and route to `work bounded`.
2. **Gather authority** — named sources, repo instructions, `grep -ri "<keywords>" docs/memory/`. Discovery clarifies scope, never expands it.
3. **Compare options** — 2–3 paths with trade-offs and a recommendation.
4. **Clarify the boundary** — `brainstorm-grill.md`; `lock`/`refine` continue with steps 5–10 in `brainstorm-lock.md`.
