# Workflow Harness — Concept

The `workflow/` skill chain (`watzup, brainstorm, to-plan, work, interview, check, git, handoff`) is a harness-backed runtime: the 6 spine skills are thin triggers that defer to canonical playbooks embedded in the `zharness` binary. This doc locks the mental model.

**Workflow chain UX is preserved.** Skill order, names, and intent do not change. What changes is *underneath*: every stage that matters gets a durable, machine-recorded, replayable trail instead of relying on markdown pointers alone.

## 4-Layer Model

- **harness** — gone since v0.15. Markdown plus git is the system of record; the archive decision trail lives in `docs/plans/completed/harness-markdown-truth.md` and the root CHANGELOG.
- **workflows** — the lifecycle contract itself: `Intent → Plan → Trace → Proof → Handoff/Resume`. Tool-independent; describes what must happen, not how.
- **skills** — the 8 `SKILL.md` files under `skills/workflow/` that trigger the lifecycle for Claude Code and other skills.sh-compatible agents. The 6 spine skills (`brainstorm`, `to-plan`, `work`, `check`, `handoff`, `watzup`) route straight to `docs/playbooks/{stage}.md`; operating logic lives in those playbooks, not in the trigger. No binary sits between the skill and its playbook — `zharness` installs and updates the managed doc set and plays no part in running a stage.
- **cli** — `zharness`, the Go binary being reduced to install / update / uninstall for the managed doc set (`docs/plans/**`, playbooks, the AGENTS block). Lifecycle enforcement moved to repo scripts plus the pre-commit hook; nothing waits on the binary.

## Lifecycle

### Intent → Intake — `brainstorm`
A raw idea, notes, or files enter through `brainstorm`. It classifies the request into a risk lane (tiny / normal / high-risk) persisted in the plan's frontmatter `lane:` field and locks the result into one evolving plan at `docs/plans/active/{slug}.md`, owning that plan's Outcome, Authority and Requirements, and Non-goals sections; The lane lives nowhere else — no separate store exists.

### Story/Plan — `to-plan`
Once the plan is locked, `to-plan` writes its Approach and Risks plus Phases and Verification (waves, tasks, checks) into that same file — no separate roadmap or per-phase context/plan files — assigning one stable story identity per stable phase (a plain unique token written beside the phase).

### Trace — `work`
`work` executes the active phase wave-by-wave, verifying every task, and appends execution state to the plan's append-only Progress/Decisions sections in one editing pass per wave.

### Proof — `check`
`check` runs the automated gate, evaluates the required-proof matrix for the plan's lane (tiny/normal/high-risk), and appends a deterministic verdict with nested proof-command sub-bullets to the plan's Validation section; the pre-commit hook re-executes every cited proof itself at commit time and rejects false claims. Missing required proof always fails, naming the missing evidence.

### Handoff/Resume — `handoff`, `watzup`
`handoff` updates the plan's Current State and Next Action directly and, on final clean closure, moves the plan from `docs/plans/active/{slug}.md` to `docs/plans/completed/{slug}.md`. `watzup` renders a session-start recap from Git state plus the plan alone.

`git` and `interview` sit outside this spine — see mapping table below.

## SDLC Stage Coverage

The workflow chain covers plan → code → verify → commit/PR. Deployment, release management, and production monitoring are explicitly **out of scope** for this chain: `check`'s automated gate and manual review end at a clean local verdict — nothing in `work`, `check`, or `git` ships a build artifact, runs a release, or watches one in production. This is a declared non-goal, not an ambient gap (G1 of the SDLC gap analysis, deleted by `655c6ac` — see `docs/decisions/0004-docs-directory-deletion-655c6ac.md`): revisit only if a real deploy target materializes, at which point the extension point is a future `ship` skill (G2 of the same analysis), not a retrofit onto `work` or `check`.

## No Version Gate

There is no binary in the execution path, so there is nothing to version-gate. Every spine skill reads `docs/playbooks/{stage}.md` from the repository it is running in, and that playbook is committed — a fresh clone with no `zharness` installed executes the same lifecycle as a fully provisioned machine.

The `preflight`-based readiness call and its `MIN_ZHARNESS_VERSION` documentation gate (last set to `0.8.1`) were deleted in v0.15 along with the rest of the lifecycle command surface; see [`docs/ARCHITECTURE.md`](../../docs/ARCHITECTURE.md). What `zharness` still does — `install` / `update` / `uninstall` — is scaffold and three-way-update the managed docs, which is a setup concern, not a per-stage one.

### Non-spine skills degrade instead of stopping

A skill that owns no harness entity must not hard-stop on the harness. Of the 8 workflow skills, exactly two have no dedicated entity in the mapping table below: `git` and `interview`. Neither writes to the harness, so a missing, stale, or unreadable `zharness` is never a reason to refuse their actual work — staging and committing, or grilling an intent. Each prints one line noting harness enrichment is unavailable (the `git` gate-verdict warning, or nothing at all for `interview`) and proceeds regardless. The 6 spine skills now degrade the same way when the harness is absent: markdown plus git stays the system of record, so a missing binary costs bookkeeping convenience, never the ability to do the work.

## Thin-Trigger Template

Every one of the 6 spine skills follows this shape, ≤30 rendered lines including frontmatter:

```markdown
---
name: {skill-name}
description: {unchanged from before this initiative — skills.sh discovery/trigger UX is Claude-facing content, stays here}
---

{Resolve the invocation mode when the stage has one, then} follow `docs/playbooks/{stage}.md` — it holds this stage's operating logic. Read `docs/WORKFLOW.md` first if the routing is unclear. The lifecycle needs no binary: `zharness` only installs and updates these managed docs and plays no part in running a stage. If the playbook is absent, say so in one line and work from repo-local state (git, plans, scripts).

Defer to: {one line naming the skills this stage hands off to or resumes from}
```

`references/` for a rewritten skill is pruned to only what the corresponding playbook does *not* absorb — most of it is deleted once the playbook is proven to carry the same content (diff-checked during `playbook-authoring`).

## Skill ↔ Plan-Section Mapping

| Skill | Owns | Mechanism |
| :--- | :--- | :--- |
| `brainstorm` | Outcome, Authority and Requirements, Non-goals (lane in frontmatter) | playbook + hand-edited markdown |
| `to-plan` | Approach and Risks, Phases and Verification (story tokens) | playbook + hand-edited markdown |
| `work` | Progress, Decisions (append-only) | playbook + hand-edited markdown |
| `check` | Validation (append-only) | playbook + nested proof sub-bullets |
| `handoff` | Current State and Next Action, phase closure | playbook + `git mv` on completion |
| `watzup` | console recap | git + plan reads only |
| `git` / `interview` | no plan sections | enrichment optional, never blocking |

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
  - [#24](https://github.com/therealtinhtute/skills/issues/24) — pilot-era gap #24 (archived with tag `v0.14.0`; `resume.go` and its STATE doc were removed in v0.15)
  - [#25](https://github.com/therealtinhtute/skills/issues/25) — phases 1-6's RUN/CHECK/HANDOFF artifacts predate the harness and fail `validate`'s ULID cross-link checks (`entropy_score: 100`, zero DB-level `pointer_drift` — the gap is markdown frontmatter, not the harness itself)

Neither gap breaks the chain's core promise: state derivation from legacy `.kit/` is correct, and the changeset-rebuild mechanism is proven byte-exact on this repo's own real history. Both gaps are scoped, filed, and routed to future planning cycles rather than hotfixed mid-pilot, per this phase's own rule.
