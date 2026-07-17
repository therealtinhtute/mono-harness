# Workflow Harness — Concept

The `workflow/` skill chain (`watzup, brainstorm, to-plan, work, interview, check, git, handoff`) is evolving from prompt-only orchestration into a harness-backed runtime. This doc locks the mental model before any contract or code work begins.

**Workflow chain UX is preserved.** Skill order, names, and intent do not change. What changes is *underneath*: every stage that matters gets a durable, machine-recorded, replayable trail instead of relying on markdown pointers alone.

## 4-Layer Model

- **harness** — the durable state layer. SQLite (`harness.db`, gitignored) materialized by replaying committed, ULID-named JSONL changesets under `.kit/changesets/`. This is the source of truth for intake, story/phase, trace, and check history — not the markdown.
- **workflows** — the lifecycle contract itself: `Intent → Intake → Story/Plan → Trace → Proof → Handoff/Resume`. Tool-independent; describes what must happen, not how.
- **skills** — the 8 `SKILL.md` files under `skills/workflow/` that implement the lifecycle for Claude Code and other skills.sh-compatible agents. Each is rewritten CLI-first: `zharness` calls are inline in the instructions, not tucked into references.
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

`MIN_ZHARNESS_VERSION = 0.1.0`. Every rewritten skill (`brainstorm`, `to-plan`, `work`) runs `zharness --version` before anything else; a `dev` build (unreleased local build) always satisfies the gate, since no tagged release exists yet during this initiative's own dogfooding. Otherwise, a missing binary or a version below `0.1.0` prints `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stops the skill.

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

See `.kit/planning/SPEC.md` for the full requirement set and `docs/workflow-harness/gap-matrix.md` for the current-state gap inventory.
