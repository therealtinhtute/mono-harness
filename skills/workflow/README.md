# Workflow Harness — Concept

The `workflow/` skill chain (`watzup, brainstorm, to-plan, work, interview, check, git, handoff`) is a harness-backed runtime: the 6 spine skills are thin triggers that defer to canonical playbooks embedded in the `zharness` binary. This doc locks the mental model.

**Workflow chain UX is preserved.** Skill order, names, and intent do not change. What changes is *underneath*: every stage that matters gets a durable, machine-recorded, replayable trail instead of relying on markdown pointers alone.

## 4-Layer Model

- **harness** — the durable state layer. SQLite (`harness.db`, gitignored, repo root) materialized by replaying local, ULID-named JSONL changesets under `.kit/changesets/` (also gitignored — per-machine state, not committed). This is the source of truth for intake, story/phase, trace, and check history — not the markdown.
- **workflows** — the lifecycle contract itself: `Intent → Intake → Story/Plan → Trace → Proof → Handoff/Resume`. Tool-independent; describes what must happen, not how.
- **skills** — the 8 `SKILL.md` files under `skills/workflow/` that trigger the lifecycle for Claude Code and other skills.sh-compatible agents. Every skill version-gates and calls `zharness preflight` for one shared readiness/rail-guard decision. The 6 spine skills (`brainstorm`, `to-plan`, `work`, `check`, `handoff`, `watzup`) then follow the playbook path returned by preflight; operating logic lives in those playbooks, not in the trigger.
- **cli** — `zharness`, the Go binary (cobra command tree, `modernc.org/sqlite`, CGO disabled) that routes every workflow stage and reads/writes the harness layer. `preflight` is read-only; every mutating command appends a changeset before touching the database.

## Lifecycle

### Intent → Intake — `brainstorm`
A raw idea, notes, or files enter through `brainstorm`. It classifies the request into a risk lane (tiny / normal / high-risk) and locks the result into one evolving plan at `docs/plans/active/{slug}.md`, owning that plan's Outcome, Authority and Requirements, and Non-goals sections. `zharness intake` records the classification; the intake ID is persisted in the plan's frontmatter.

### Story/Plan — `to-plan`
Once the plan is locked, `to-plan` writes its Approach and Risks plus Phases and Verification (waves, tasks, checks) into that same file — no separate roadmap or per-phase context/plan files. `zharness story` records one story row per stable phase.

### Trace — `work`
`work` executes the active phase wave-by-wave, verifying every task, and appends execution state to the plan's append-only Progress/Decisions sections. Each wave emits `zharness trace add`, linked to the run so the execution trail is queryable after the fact.

### Proof — `check`
`check` runs the automated gate, audits durable lifecycle links with `zharness audit`, evaluates the required-proof matrix for the intake's risk lane (tiny/normal/high-risk), and records a deterministic verdict with `zharness check record` appended to the plan's Validation section. Missing required proof always fails, naming the missing evidence.

### Handoff/Resume — `handoff`, `watzup`
`handoff` records a handoff entity and updates the plan's Current State and Next Action; a final clean closure moves the plan from `docs/plans/active/{slug}.md` to `docs/plans/completed/{slug}.md`. `watzup` renders `zharness resume` at the start of the next session: workflow position, latest run/check/handoff IDs, and a named recovery action for any drift.

`git` and `interview` sit outside this spine — see mapping table below.

## Version Gate

`MIN_ZHARNESS_VERSION = 0.8.1` (bumped from `0.4.1` after `harness-memory-ceremony-convergence` (`cli/v0.8.1`, `docs/plans/active/harness-memory-ceremony-convergence.md`): schema 6 to 9, `decision add`/`query decisions`, task-granularity `trace add`, `query checks`, atomic CLI-owned markdown writes for `trace`/`decision`/`check`/`handoff`, and the stage-shaped `context` packet in `preflight` that folds `--version`/`resume`/`query phases` into one call. A pre-`0.8.1` binary predates all of it — playbooks written against this version would silently degrade to manual bookkeeping the older CLI can't back). Every one of the 6 spine skills runs `zharness preflight {stage} --json` as its first and only readiness call; a missing binary fails that shell invocation directly. Otherwise the skill checks the response's own `version` field (`preflight`'s payload — no separate `zharness --version` round trip, F3, `docs/audit/workflow-harness-ceremony-audit.md`): a `dev` build (unreleased local build) always satisfies the gate, and a version below `0.8.1` prints `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stops the skill.

### Non-spine skills degrade instead of stopping

A skill that owns no harness entity must not hard-stop on the harness. Of the 8 workflow skills, exactly two have no dedicated entity in the mapping table below: `git` and `interview`. Neither writes to the harness, so a missing, stale, or unreadable `zharness` is never a reason to refuse their actual work — staging and committing, or grilling an intent. Each prints one line noting harness enrichment is unavailable (the `git` gate-verdict warning, or nothing at all for `interview`) and proceeds regardless. This does not weaken the 6 spine skills' hard stop above: a durable write with no harness to write the durable record to is a real blocker, not a degraded feature, precisely because those 6 each own the entity they'd be writing.

## Thin-Trigger Template

Every one of the 6 spine skills follows this shape, ≤30 rendered lines including frontmatter:

```markdown
---
name: {skill-name}
description: {unchanged from before this initiative — skills.sh discovery/trigger UX is Claude-facing content, stays here}
---

Run `zharness preflight {stage} [--mode {mode}] --json`. Missing binary: stop, tell the user to run `bash scripts/install-zharness.sh`. Otherwise check its `version` field — below MIN_ZHARNESS_VERSION (0.8.1): stop with the same message; a `dev` build always passes. If it returns `stop`, state the message and follow the exact recovery. Otherwise read and follow its returned `playbook` path when present; reduced mode must remain read-only.

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
- Markdown fallback / CLI-optional compatibility mode — explicitly rejected; `zharness` is mandatory

This initiative's own roadmap is complete; see `docs/plans/completed/workflow-harness-history-2026-07.md` for its consolidated history, and `docs/plans/completed/harness-convergence-pass-v3.md` for the one-plan/one-DB convergence work that followed it.

## Pilot Evidence & Go/No-Go

Piloted 2026-07-17 by dogfooding this repo (`Lab/skills`) itself — real legacy `.kit/workflow-state.yml`-driven history, not a synthetic target. Full evidence: `docs/workflow-harness/pilot-evidence/2026-07-17-lab-skills-import.md`.

**Verdict: GO.**

- `zharness init && zharness import && zharness query state --json` — **pass**. Derived state matched this repo's real pre-import `workflow-state.yml` (`current_phase`, `entry_phase`) exactly, on real history, not a fixture.
- Rebuild-from-changesets (cross-machine resume mechanism) — **pass**. A scratch copy of `.kit/changesets/**` replayed through `zharness init` + `zharness db changeset apply` in ULID order produced a byte-identical `resume --json` to the original. Zero divergence.
- `zharness validate` / `zharness audit` — **2 real gaps found, both filed, neither blocking**:
  - [#24](https://github.com/therealtinhtute/skills/issues/24) — `resume.go`'s drift `Recovery` strings don't match `cli/docs/STATE.md`'s documented text (escalated from continuity's `check` gate)
  - [#25](https://github.com/therealtinhtute/skills/issues/25) — phases 1-6's RUN/CHECK/HANDOFF artifacts predate the harness and fail `validate`'s ULID cross-link checks (`entropy_score: 100`, zero DB-level `pointer_drift` — the gap is markdown frontmatter, not the harness itself)

Neither gap breaks the chain's core promise: state derivation from legacy `.kit/` is correct, and the changeset-rebuild mechanism is proven byte-exact on this repo's own real history. Both gaps are scoped, filed, and routed to future planning cycles rather than hotfixed mid-pilot, per this phase's own rule.
