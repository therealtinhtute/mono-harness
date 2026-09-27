# Playbook: brainstorm

## Purpose

Explore a decision in the response, grill intent until the Goal is concrete, or lock/refine the durable initiative at `docs/plans/active/{slug}.md`. Owns only the plan's `## Goal`; phases belong to `to-plan`, implementation to `work`.

## Modes

A leading subcommand (`explore`, `grill`, `raw`, `lock`, `refine`) sets the mode. Without one, choose from the request shape — a request to grill, interview, or stress-test an idea or plan is `grill`; ask only when mode or scope is genuinely ambiguous.

| Mode | Input | Durable effect |
|---|---|---|
| `explore` (default) | Trade-off or recommendation request without lock intent | None; answer in the response only |
| `grill` | Fuzzy idea, or an existing plan to validate before `work` | None until the user approves a lock or refine |
| `raw` | Idea that needs a fast first pass | None; raw summary in the response |
| `lock` | Raw idea, bounded initiative, or authoritative source files | Create one active plan |
| `refine` | Existing active plan needs scope changes | Update the same plan |

`grill` and `raw`: read `docs/playbooks/brainstorm-grill.md` now. `lock` and `refine`: read `docs/playbooks/brainstorm-lock.md` and `docs/playbooks/brainstorm-grill.md` now. `explore` needs neither.

**Zero-write rule:** explore creates no plans, reports, changesets, or markdown artifacts; raw and an unapproved grill follow the same rule. IF a session reaches lock intent → switch modes explicitly before writing.

## Lanes

Pick the lane from risk, not size, and record it in frontmatter `lane:`.

| Lane | When | Plan |
|---|---|---|
| `tiny` | Known files, a direct success check, nothing hard to reverse | None: route to `work bounded` |
| `normal` | One reviewable change with a known verification path | One phase with `phase_slug:`, `status:`, and a flat task list |
| `high-risk` | User-data migration, public contract, security, or dependent steps | Phases, waves, and tasks; every gate needs an independent judge |

## Steps

1. **Classify** — mode, lane, risk flags, affected surfaces. `tiny` → stop and route to `work bounded`.
2. **Gather minimum authority** — read named sources and repository instructions. Check prior lessons first: `grep -ri "<topic keywords>" docs/memory/` (plain committed files, never a database). Discovery may clarify scope, never expand it.
3. **Compare options** — 2–3 viable paths, or 1–2 alternatives rejected by authoritative sources. State recommendation and trade-offs before locking.
4. **Clarify, then lock** (`grill`/`raw`/`lock`/`refine`) — steps 4–10 in the companions named under Modes.

## Exit Conditions

- Explore: recommendation, rationale, and rejected alternatives in the response; zero durable writes.
- Grill: the settled tree and open questions in the response, then a lock or refine only on approval.
- Raw: decided / assumed / open lists in the response; zero durable writes.
- Lock: exactly one active plan satisfying steps 6–9; next action `to-plan`, which updates the same file.
- Refine: same plan; later-stage sections preserved.
