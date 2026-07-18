# Workflow Harness — Concept

The `workflow/` skill chain (`watzup, brainstorm, to-plan, work, interview, check, git, handoff`) is evolving from prompt-only orchestration into a harness-backed runtime. This doc locks the mental model before any contract or code work begins.

**Workflow chain UX is preserved.** Skill order, names, and intent do not change. What changes is *underneath*: every stage that matters gets a durable, machine-recorded, replayable trail instead of relying on markdown pointers alone.

## 4-Layer Model

- **harness** — the durable state layer. SQLite (`harness.db`, gitignored) materialized by replaying committed, ULID-named JSONL changesets under `.kit/changesets/`. This is the source of truth for intake, story/phase, trace, and check history — not the markdown.
- **workflows** — the lifecycle contract itself: `Intent → Intake → Story/Plan → Trace → Proof → Handoff/Resume`. Tool-independent; describes what must happen, not how.
- **skills** — the 8 `SKILL.md` files under `skills/workflow/` that trigger the lifecycle for Claude Code and other skills.sh-compatible agents. The 6 spine skills (`brainstorm`, `to-plan`, `work`, `check`, `handoff`, `watzup`) are thin triggers (≤30 rendered lines): version-gate, ensure `.kit/docs/` is scaffolded, then read and follow the matching embedded playbook under `.kit/docs/playbooks/`. The operating logic itself lives in those playbooks — not in this file, not in `references/` — so any agent that can read a file and run a CLI can execute the same lifecycle, not just Claude Code.
- **cli** — `zharness`, the Go binary (cobra command tree, `modernc.org/sqlite`, CGO disabled) that skills call to read and write the harness layer. Every mutating command appends a changeset before touching the database.

## Lifecycle

### Intent → Intake — `brainstorm`
A raw idea, notes, or files enter through `brainstorm`. It classifies the request into a risk lane (tiny / normal / high-risk) and locks the result into `.kit/planning/SPEC.md`. `zharness intake` records the classification; the intake ID is persisted in the SPEC frontmatter.

### Story/Plan — `to-plan`
Once the spec is locked, `to-plan` derives `.kit/planning/ROADMAP.md` and per-phase `-CONTEXT.md`/`-PLAN.md` files. `zharness init` + `zharness story` run per phase and record phase pointers in the harness.

### Trace — `work`
`work` executes the active phase wave-by-wave, verifying every task. Each wave emits `zharness trace add`, linked to the run so the execution trail is queryable after the fact.

### Proof — `check`
`check` runs `zharness audit` + `zharness score-trace`, evaluates the proof matrix for the intake lane, and records a deterministic verdict with `zharness check record`. Missing required proof always fails, naming the missing evidence.

### Handoff/Resume — `handoff`, `watzup`
`handoff` records a handoff entity capturing session state. `watzup` renders `zharness resume` at the start of the next session: workflow position, latest run/check/handoff IDs, and a named recovery action for any drift.

`git` and `interview` sit outside this spine — see mapping table below.

## Version Gate

`MIN_ZHARNESS_VERSION = 0.2.0` (the first release with embedded playbooks — bumped from `0.1.0` once `cli/v0.2.0` shipped). Every one of the 6 spine skills runs `zharness --version` before anything else; a `dev` build (unreleased local build) always satisfies the gate. Otherwise, a missing binary or a version below `0.2.0` prints `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stops the skill.

## Thin-Trigger Template

Every one of the 6 spine skills follows this shape, ≤30 rendered lines including frontmatter:

```markdown
---
name: {skill-name}
description: {unchanged from before this initiative — skills.sh discovery/trigger UX is Claude-facing content, stays here}
---

Run `zharness --version`. Below MIN_ZHARNESS_VERSION (0.2.0) or missing: stop, tell the user to run `bash scripts/install-zharness.sh`. A `dev` build always passes.

Ensure docs are present: run `zharness init` if `.kit/docs/` is missing (idempotent — always safe to run).

Read `.kit/docs/playbooks/{stage}.md` and follow it exactly. That file is the operating logic; this file only triggers it.

Defer to: {one line naming the skills this stage hands off to or resumes from}
```

`references/` for a rewritten skill is pruned to only what the corresponding playbook does *not* absorb — most of it is deleted once the playbook is proven to carry the same content (diff-checked during `playbook-authoring`).

## Skill ↔ Command ↔ Entity Mapping

| Skill | Harness artifact | `zharness` command group | Entity |
| :--- | :--- | :--- | :--- |
| `brainstorm` | `SPEC.md` (intake ID in frontmatter) | `intake` | intake |
| `to-plan` | `ROADMAP.md` + phase `-CONTEXT.md`/`-PLAN.md` | `init`, `story` | story (phase) |
| `work` | `.kit/runs/work/*.md` | `trace add` | trace (run) |
| `check` | `.kit/reports/check/*.md` | `audit`, `score-trace`, `check record` | check (verdict) |
| `handoff` | `.kit/HANDOFF.md` | `handoff record` | handoff |
| `watzup` | console recap | `resume` | resume snapshot |
| `git` | commit / PR | — | minimal-integration: no dedicated harness entity |
| `interview` | feeds `brainstorm` / `to-plan` output | — | minimal-integration: no dedicated harness entity |

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

See `.kit/planning/archive/workflow-harness-2026-07-17/SPEC.md` (archived — this initiative's roadmap is complete) for the full requirement set, and `docs/workflow-harness/gap-matrix.md` for the current-state gap inventory.

## Pilot Evidence & Go/No-Go

Piloted 2026-07-17 by dogfooding this repo (`Lab/skills`) itself — real legacy `.kit/workflow-state.yml`-driven history, not a synthetic target. Full evidence: `docs/workflow-harness/pilot-evidence/2026-07-17-lab-skills-import.md`.

**Verdict: GO.**

- `zharness init && zharness import && zharness query state --json` — **pass**. Derived state matched this repo's real pre-import `workflow-state.yml` (`current_phase`, `entry_phase`) exactly, on real history, not a fixture.
- Rebuild-from-changesets (cross-machine resume mechanism) — **pass**. A scratch copy of `.kit/changesets/**` replayed through `zharness init` + `zharness db changeset apply` in ULID order produced a byte-identical `resume --json` to the original. Zero divergence.
- `zharness validate` / `zharness audit` — **2 real gaps found, both filed, neither blocking**:
  - [#24](https://github.com/therealtinhtute/skills/issues/24) — `resume.go`'s drift `Recovery` strings don't match `cli/docs/STATE.md`'s documented text (escalated from continuity's `check` gate)
  - [#25](https://github.com/therealtinhtute/skills/issues/25) — phases 1-6's RUN/CHECK/HANDOFF artifacts predate the harness and fail `validate`'s ULID cross-link checks (`entropy_score: 100`, zero DB-level `pointer_drift` — the gap is markdown frontmatter, not the harness itself)

Neither gap breaks the chain's core promise: state derivation from legacy `.kit/` is correct, and the changeset-rebuild mechanism is proven byte-exact on this repo's own real history. Both gaps are scoped, filed, and routed to future planning cycles rather than hotfixed mid-pilot, per this phase's own rule.
