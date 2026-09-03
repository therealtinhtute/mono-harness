# Workflow Harness — Concept

The `workflow/` skill chain (`watzup, brainstorm, to-plan, work, interview, check, git, handoff`) is a markdown-first lifecycle. The 6 spine skills are thin triggers that defer to playbooks in the repository. This doc locks the mental model.

## 4-Layer Model

- **harness** — gone as a runtime since v0.15. Markdown plus git is the system of record; the archive trail lives in `docs/plans/completed/harness-markdown-truth.md` and the root CHANGELOG.
- **workflows** — the lifecycle contract itself: `Intent → Plan → Trace → Proof → Handoff/Resume`. Tool-independent; describes what must happen, not how.
- **skills** — the 8 `SKILL.md` files under `skills/workflow/`. The 6 spine skills (`brainstorm`, `to-plan`, `work`, `check`, `handoff`, `watzup`) route straight to `docs/playbooks/{stage}.md`. No binary sits between the skill and its playbook — `zharness` installs and updates the managed doc set and plays no part in running a stage.
- **cli** — `zharness`, the Go binary reduced to install / update / uninstall for the managed doc set. Lifecycle enforcement lives in repo scripts plus the pre-commit hook.

## Lifecycle

### Intent → Intake — `brainstorm`
A raw idea, notes, or files enter through `brainstorm`. It classifies the request into a risk lane (tiny / normal / high-risk) persisted in the plan's frontmatter `lane:` field and locks the result into one evolving plan at `docs/plans/active/{slug}.md`, owning that plan's Outcome, Authority and Requirements, and Non-goals sections. The lane lives nowhere else — no separate store exists.

### Story/Plan — `to-plan`
Once the plan is locked, `to-plan` writes its Approach and Risks plus Phases and Verification (waves, tasks, checks) into that same file — no separate roadmap or per-phase context/plan files — assigning one stable story identity per stable phase (a plain unique token written beside the phase).

### Trace — `work`
`work` executes the active phase wave-by-wave, verifying every task, and appends execution state to the plan's append-only Progress/Decisions sections in one editing pass per wave.

### Proof — `check`
`check` runs the automated gate, evaluates the required-proof matrix for the plan's lane (tiny/normal/high-risk), and appends a deterministic verdict with nested proof-command sub-bullets to the plan's Validation section; the pre-commit hook re-executes every cited proof itself at commit time and rejects false claims. Missing required proof always fails, naming the missing evidence.

### Handoff/Resume — `handoff`, `watzup`
`handoff` updates the plan's Current State and Next Action directly and, on final clean closure, records an `absorb:` line then moves the plan from `docs/plans/active/{slug}.md` to `docs/plans/completed/{slug}.md`. `watzup` renders a session-start recap from Git state plus the plan alone.

`git` and `interview` sit outside this spine — see mapping table below.

## SDLC Stage Coverage

The workflow chain covers plan → code → verify → commit/PR. Deployment, release management, and production monitoring are explicitly **out of scope** for this chain: `check`'s automated gate and manual review end at a clean local verdict — nothing in `work`, `check`, or `git` ships a build artifact, runs a release, or watches one in production. This is a declared non-goal, not an ambient gap (G1 of the SDLC gap analysis; see `docs/decisions/0004-docs-directory-deletion-655c6ac.md`): revisit only if a real deploy target materializes, at which point the extension point is a future `ship` skill, not a retrofit onto `work` or `check`.

## No Version Gate

There is no binary in the execution path, so there is nothing to version-gate. Every spine skill reads `docs/playbooks/{stage}.md` from the repository it is running in, and that playbook is committed — a fresh clone with no `zharness` installed executes the same lifecycle as a fully provisioned machine.

The old `preflight` readiness call and its `MIN_ZHARNESS_VERSION` documentation gate were deleted in v0.15. What `zharness` still does — `install` / `update` / `uninstall` — is scaffold the managed docs, fresh-overwriting playbooks/WORKFLOW.md and three-way-merging PROJECT.md and the AGENTS.md block, which is a setup concern, not a per-stage one.

### Non-spine skills do not stop on a missing binary

`git` and `interview` own no plan sections. A missing `zharness` binary is never a reason to refuse staging, committing, or grilling an intent. The 6 spine skills are the same: markdown plus git is the system of record.

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

## Skill ↔ Plan-Section Mapping

| Skill | Owns | Mechanism |
| :--- | :--- | :--- |
| `brainstorm` | Outcome, Authority and Requirements, Non-goals (lane in frontmatter) | playbook + hand-edited markdown |
| `to-plan` | Approach and Risks, Phases and Verification (story tokens) | playbook + hand-edited markdown |
| `work` | Progress, Decisions (append-only) | playbook + hand-edited markdown |
| `check` | Validation (append-only) | playbook + nested proof sub-bullets |
| `handoff` | Current State and Next Action, phase closure | playbook + absorb line + `git mv` on completion |
| `watzup` | console recap | git + plan reads only |
| `git` / `interview` | no plan sections | enrichment optional, never blocking |
| `encode-invariant` | no plan sections | non-spine; pattern `docs/patterns/encoding-invariants.md`; never blocking on a missing binary |
| `improve-harness` | no plan sections | non-spine; template `docs/templates/harness-improvement.md`; never blocking on a missing binary |
