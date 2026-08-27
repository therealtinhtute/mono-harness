# Workflow Harness — Concept

The `workflow/` skill chain (`watzup, brainstorm, to-plan, work, interview, check, git, handoff`) is a harness-backed runtime: the 6 spine skills are thin triggers that defer to canonical playbooks embedded in the `zharness` binary. This doc locks the mental model.

**Workflow chain UX is preserved.** Skill order, names, and intent do not change. What changes is *underneath*: every stage that matters gets a durable, machine-recorded, replayable trail instead of relying on markdown pointers alone.

## 4-Layer Model

- **harness** — the optional ledger layer (`harness.db`, gitignored, repo root, present only while the binary exists). Markdown plus git stays the system of record — `db rebuild` regenerates the index from committed plans alone (`docs/plans/completed/harness-markdown-truth.md`).
- **workflows** — the lifecycle contract itself: `Intent → Plan → Trace → Proof → Handoff/Resume`. Tool-independent; describes what must happen, not how.
- **skills** — the 8 `SKILL.md` files under `skills/workflow/` that trigger the lifecycle for Claude Code and other skills.sh-compatible agents. Every skill attempts `zharness preflight` where available for one shared readiness/rail-guard decision; without the binary each degrades to its markdown-first playbook instead of failing. The 6 spine skills (`brainstorm`, `to-plan`, `work`, `check`, `handoff`, `watzup`) then follow the playbook path returned by preflight — or `docs/playbooks/{stage}.md` directly when it is absent; operating logic lives in those playbooks, not in the trigger.
- **cli** — `zharness`, the Go binary (cobra command tree; `modernc.org/sqlite` remains only while the binary exists and is deleted at v0.15). Where present it reconciles the ledger after markdown writes; nothing waits on it.

## Lifecycle

### Intent → Intake — `brainstorm`
A raw idea, notes, or files enter through `brainstorm`. It classifies the request into a risk lane (tiny / normal / high-risk) persisted in the plan's frontmatter `lane:` field and locks the result into one evolving plan at `docs/plans/active/{slug}.md`, owning that plan's Outcome, Authority and Requirements, and Non-goals sections; `zharness intake` mirrors the classification into the ledger whenever the binary exists.

### Story/Plan — `to-plan`
Once the plan is locked, `to-plan` writes its Approach and Risks plus Phases and Verification (waves, tasks, checks) into that same file — no separate roadmap or per-phase context/plan files — assigning one stable story identity per stable phase; `zharness story` mirrors the row while the ledger exists.

### Trace — `work`
`work` executes the active phase wave-by-wave, verifying every task, and appends execution state to the plan's append-only Progress/Decisions sections; `zharness trace add` mirrors each flushed wave into the queryable trail whenever the binary exists.

### Proof — `check`
`check` runs the automated gate, evaluates the required-proof matrix for the plan's lane (tiny/normal/high-risk), and appends a deterministic verdict with nested proof-command sub-bullets to the plan's Validation section; while the binary exists `zharness audit` checks lifecycle links and `check record` re-executes every cited proof command itself. Missing required proof always fails, naming the missing evidence.

### Handoff/Resume — `handoff`, `watzup`
`handoff` updates the plan's Current State and Next Action directly and, on final clean closure, moves the plan from `docs/plans/active/{slug}.md` to `docs/plans/completed/{slug}.md`; `zharness handoff record` mirrors each closure while the ledger exists. `watzup` renders a session-start recap from Git state plus the plan itself, with an optional `zharness resume` position packet.

`git` and `interview` sit outside this spine — see mapping table below.

## SDLC Stage Coverage

The workflow chain covers plan → code → verify → commit/PR. Deployment, release management, and production monitoring are explicitly **out of scope** for this chain: `check`'s automated gate and manual review end at a clean local verdict — nothing in `work`, `check`, or `git` ships a build artifact, runs a release, or watches one in production. This is a declared non-goal, not an ambient gap (G1 of the SDLC gap analysis, deleted by `655c6ac` — see `docs/decisions/0004-docs-directory-deletion-655c6ac.md`): revisit only if a real deploy target materializes, at which point the extension point is a future `ship` skill (G2 of the same analysis), not a retrofit onto `work` or `check`.

## Version Gate

`MIN_ZHARNESS_VERSION = 0.8.1` (bumped from `0.4.1` after the `harness-memory-ceremony-convergence` initiative shipped as `cli/v0.8.1`, whose plan record was deleted by `655c6ac` — see `docs/decisions/0004-docs-directory-deletion-655c6ac.md`: schema 6 to 9, `decision add`/`query decisions`, task-granularity `trace add`, `query checks`, atomic CLI-owned markdown writes for `trace`/`decision`/`check`/`handoff`, and the stage-shaped `context` packet in `preflight` that folds `--version`/`resume`/`query phases` into one call. A pre-`0.8.1` binary predates all of it — playbooks written against this version would silently degrade to manual bookkeeping the older CLI can't back). Every one of the 6 spine skills runs `zharness preflight {stage} --json` as its first readiness call when the binary is present — but `MIN_ZHARNESS_VERSION` is documentation, not a blocking gate: the response's own `version` field (`preflight`'s payload — no separate `zharness --version` round trip, F3 of the ceremony audit, deleted by `655c6ac`; see `docs/decisions/0004-docs-directory-deletion-655c6ac.md`) only decides fresh vs degraded behavior. A missing binary or a version below `0.8.1` degrades instead of halting: print one fallback line and follow the stage's markdown-first playbook (`docs/playbooks/{stage}.md`) directly from repo-local state alone; a `dev` build (unreleased local build) simply behaves fresh.

### Non-spine skills degrade instead of stopping

A skill that owns no harness entity must not hard-stop on the harness. Of the 8 workflow skills, exactly two have no dedicated entity in the mapping table below: `git` and `interview`. Neither writes to the harness, so a missing, stale, or unreadable `zharness` is never a reason to refuse their actual work — staging and committing, or grilling an intent. Each prints one line noting harness enrichment is unavailable (the `git` gate-verdict warning, or nothing at all for `interview`) and proceeds regardless. The 6 spine skills now degrade the same way when the harness is absent: markdown plus git stays the system of record, so a missing binary costs bookkeeping convenience, never the ability to do the work.

## Thin-Trigger Template

Every one of the 6 spine skills follows this shape, ≤30 rendered lines including frontmatter:

```markdown
---
name: {skill-name}
description: {unchanged from before this initiative — skills.sh discovery/trigger UX is Claude-facing content, stays here}
---

Run `zharness preflight {stage} [--mode {mode}] --json`. Missing binary or `version` below MIN_ZHARNESS_VERSION (0.8.1): degrade, don't halt — print one fallback line and follow the returned `playbook` path directly, or `docs/playbooks/{stage}.md` when nothing was returned. If it returns `stop`, state the message and follow the exact recovery. Reduced mode must remain read-only.

Defer to: {one line naming the skills this stage hands off to or resumes from}
```

`references/` for a rewritten skill is pruned to only what the corresponding playbook does *not* absorb — most of it is deleted once the playbook is proven to carry the same content (diff-checked during `playbook-authoring`).

## Skill ↔ Command ↔ Entity Mapping

| Skill | Plan section owned | `zharness` command group | Entity |
| :--- | :--- | :--- | :--- |
| `brainstorm` | Outcome, Authority and Requirements, Non-goals (intake ID in frontmatter) | `intake` | intake |
| `to-plan` | Approach and Risks, Phases and Verification | `init`, `story` | story (phase) |
| `work` | Progress, Decisions (append-only) | `run create`, `trace add` | run + trace |
| `check` | Validation (append-only) | `audit`, `check record` | check (verdict) |
| `handoff` | Current State and Next Action | `handoff record` | handoff |
| `watzup` | console recap | `resume` | resume snapshot |
| `git` | commit / PR | `preflight`, `query check --latest` | read-only: no dedicated harness entity |
| `interview` | feeds `brainstorm` / `to-plan` output | `preflight` | read-only: no dedicated harness entity |

## Scope

### In Scope (this initiative)
- `cli/` Go module, release pipeline, install script
- Rewrites of all 8 `skills/workflow/*` `SKILL.md` files + their references
- Artifact template contracts (brainstorm/work/check/handoff references)
- `docs/workflow-harness/`, this doc, root `README.md`, `CLAUDE.md` updates
- Legacy `.kit/` import path and migration guide
- Pilot run + evidence

### Out of Scope
- Repo restructure beyond adding `cli/` and `docs/workflow-harness/` — skills stay in place
- Any skill outside `skills/workflow/` (`craft/`, `shipping/` untouched)
- Markdown fallback / CLI-optional compatibility mode — explicitly rejected by that initiative; superseded 2026-08: the chain is now fail-open and degrades to markdown-first playbooks when the binary is absent (see Version Gate above)

This initiative's own roadmap is complete. Its consolidated history and the one-plan/one-DB convergence work that followed were both recorded in completed plans that `655c6ac` deleted; see `docs/decisions/0004-docs-directory-deletion-655c6ac.md` for what was removed and how to retrieve it.

## Pilot Evidence & Go/No-Go

Piloted 2026-07-17 by dogfooding this repo itself — real legacy `.kit/workflow-state.yml`-driven history, not a synthetic target. The full evidence log was deleted by `655c6ac` and remains retrievable at `655c6ac^`; see `docs/decisions/0004-docs-directory-deletion-655c6ac.md`.

**Verdict: GO.**

- `zharness init && zharness import && zharness query state --json` — **pass**. Derived state matched this repo's real pre-import `workflow-state.yml` (`current_phase`, `entry_phase`) exactly, on real history, not a fixture.
- Rebuild-from-changesets (cross-machine resume mechanism) — **pass**. A scratch copy of `.kit/changesets/**` replayed through `zharness init` + `zharness db changeset apply` in ULID order produced a byte-identical `resume --json` to the original. Zero divergence.
- `zharness validate` / `zharness audit` — **2 real gaps found, both filed, neither blocking**:
  - [#24](https://github.com/therealtinhtute/skills/issues/24) — `resume.go`'s drift `Recovery` strings don't match `cli/docs/STATE.md`'s documented text (escalated from continuity's `check` gate)
  - [#25](https://github.com/therealtinhtute/skills/issues/25) — phases 1-6's RUN/CHECK/HANDOFF artifacts predate the harness and fail `validate`'s ULID cross-link checks (`entropy_score: 100`, zero DB-level `pointer_drift` — the gap is markdown frontmatter, not the harness itself)

Neither gap breaks the chain's core promise: state derivation from legacy `.kit/` is correct, and the changeset-rebuild mechanism is proven byte-exact on this repo's own real history. Both gaps are scoped, filed, and routed to future planning cycles rather than hotfixed mid-pilot, per this phase's own rule.
