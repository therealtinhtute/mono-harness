---
title: Harness Durability & Knowledge Contract
status: accepted — Q1–Q3 resolved 2026-08-15
interviewed: 2026-08-15
---

> **Accepted.** The four open corrections raised after the interview have been resolved
> and folded into this file — see [`open-questions.md`](open-questions.md) for the
> reasoning behind each. The resident-cost gate is now two benchmarked gates (Q1), ADR
> promotion filters against explicit triggers instead of promoting everything (Q2), and
> **M8 work-shape gating leads the program as P0** (Q3). Q4's two mechanisms are folded
> into P1 and P3 rather than deferred.

## Outcome

Ceremony scales with work shape, so a bounded task costs no markdown at all. Nothing an
agent or human needs to remember survives only in a gitignored path, and a cold agent can
orient in a consumer repo without loading the repo's documentation. `zharness` owns one
directory (`.harness/`), project knowledge owns another (`docs/`), and the derived half of
`docs/` is generated rather than maintained.

## Success Condition

Three hard gates, all mechanically checked:

1. **Work shape** — the virtualizer-fix scenario (single-file change, resumable from its
   diff) runs through `preflight` and produces **0 markdown files**, down from 142 lines
   across a run doc and a check report. The walter-theme migration still classifies
   `durable` and keeps its plan intact.

2. **Durability** — a test clones the repo fresh (no `.harness/state/`), runs
   `zharness db rebuild`, and asserts every decision, plan, product doc, and ADR is
   present. Today this test fails: decisions exist only in `harness.db` and
   `.kit/changesets/`, both gitignored.

3. **Resident cost** — two budgets, not one, both benchmarked against
   `repository-harness`'s measured numbers rather than an aspirational figure:

   | Gate | Budget | Upstream reference |
   |---|---:|---:|
   | Entrypoint pair — `AGENTS.md` managed block + `docs/README.md` | ≤ 900 | 808 |
   | Full resident path — the above + `.harness/WORKFLOW.md` | ≤ 2,400 | 2,199 |

   The split is diagnostic: it separates "the map got bloated" from "the procedure got
   bloated." Still ~2.7x lighter than onedrive-cloud's measured 6.6k. Principles
   (`docs/HARNESS.md`-equivalent) stay **out** of the resident path, loaded on demand —
   that separation is what holds a procedure doc under budget upstream.

Field reference: a cold agent in onedrive-cloud today spends ~17k tokens on process and
docs before reading any `src/`. See [`audit-onedrive-cloud.md`](audit-onedrive-cloud.md).

## Scope

**May change**

- `cli/internal/domain/preflight.go`, `cli/internal/application/preflight.go`,
  `cli/internal/interfaces/preflight.go` — work-shape classification and its JSON contract
- `cli/docs/embedded/templates/{run,check}.md` — conditional section rendering
- `cli/internal/application/{story_create,run_create,check_record}.go` — record collapse
  for bounded shape
- `cli/internal/interfaces/paths.go`, `preflight.go`, `next.go`, `db.go` — path constants
  and globs
- `cli/internal/application/{init,layout_migration,managed_docs,audit,decision}.go`
- `cli/internal/interfaces/{decision,audit}.go`
- `cli/docs/embedded/**` — WORKFLOW, AGENTS block, playbooks, templates, new scaffolds
- New: `cli/internal/{application,interfaces}/wiki*.go`
- `skills/workflow/**` playbook triggers
- `docs/plans/active/` in this repo

**Must not change**

- `/Users/tinhtute/Personal/onedrive-cloud` — **read-only, zero writes.** Evidence only.
- `/Users/tinhtute/Lab/harness-experimental`, `hoangnb24/repository-harness`,
  `Houseofmvps/codesight` — read for structure, never adopted as dependencies
- R1–R9 behavior: `query plan --section`, batched `trace add`, preflight `context`
  parity, JSONL logging contract (the path moves; the contract does not)
- Command shapes locked in `cli/docs/CONTRACT.md` for existing commands
- DB schema, except the migration itself
- **Plan quality.** P0 gates *which records get created*, never how a plan is written.
  The 186-line walter-theme plan is the counter-evidence: planning is not the weight.

## Context to Read First

- [`work-shape.md`](work-shape.md) — the P0 evidence and the authority-chain model that
  answers "how do docs, tooling, context management, and CLI combine"
- `cli/internal/domain/preflight.go` — where shape classification lands
- `cli/internal/interfaces/paths.go` — the 7-constant block this work turns on
- `cli/internal/application/layout_migration.go:42` — `MigrateLayout`, the existing
  `.kit/harness.db` → `harness.db` migration to extend
- `cli/internal/application/managed_docs.go:42` — `SyncManagedDocs`, origin of the drift
- `cli/internal/application/init.go:13,29` — `gitignoreEntries`, `ScaffoldDocs`
- `docs/audit/sdlc-token-cache-audit.md` and `docs/audit/cost-model/` — existing
  measurement methodology to reuse
- `docs/plans/completed/sdlc-token-optimization.md` — R1–R9, so none of it is re-litigated
- [`decisions.md`](decisions.md) — what was chosen and rejected, with rationale

## Key Decisions

1. **Ceremony follows work shape** — `preflight` classifies read-only / bounded / durable
   and the playbook path follows. Because the harness records ceremony over knowledge at
   23 : 1, and a one-file fix acquired a wave log, a four-commit split plan, and a
   three-severity finding taxonomy. See D13.
2. **Two directories, by owner** — `.harness/` = workflow (CLI/skills), `docs/` = project
   knowledge. Because mixing them produced 68KB of `docs/` where product architecture sits
   beside stage playbooks.
3. **`.kit/` is absorbed, not kept** — `.harness/state/` holds db, changesets, conflicts,
   cache, log; everything else in `.harness/` is tracked. Because `.kit` being 100%
   gitignored stranded real product authority one `rm -rf` from gone.
4. **Ship as a release with migration** — extend `MigrateLayout`. Because `legacyDBPath`
   already sets the precedent and consumers must not break on upgrade.
   Depends on: decision 3.
5. **Decisions promote to `docs/decisions/` ADRs at handoff, filtered by trigger** — the
   gate asserts promotion was *considered*, not that everything was promoted. Because 78
   changeset entries yielded 1 decision; promoting all of them is the noise ADRs exist to
   prevent. See D14. Depends on: decision 6.
6. **`docs/` splits generated from written** — `docs/map/` generated, `docs/product/` and
   `docs/decisions/` written. Because hand-maintaining derived facts is what rotted
   `codebase-summary.md` for 8 months.
7. **`zharness wiki` is deterministic** — file walk plus regex, no AST, no LLM, 0 tokens.
   Reference codesight's method, not codesight as a dependency. Verified tractable: 15
   routes, 11 env vars, and coverage JSON are all derivable without a TypeScript parser.
   Depends on: decision 6.
8. **A CLI upgrade must never modify a tracked file** — stage to
   `.harness/state/conflicts/*.upstream` and report. Because it silently rewrote six
   playbooks in a live repo, byte-identical to the embedded versions.
9. **Authority flows one direction** — owner decision → accepted rule → mechanical check →
   observed fact → the work, and nothing invents authority from the layer below it.
   Conventions, tests, and tool defaults show behavior; they do not authorize a rule.

## Phases

Each phase is independently mergeable: after it ships, the system is usable even if the
next never lands.

### P0 — Work-shape gating (M8)

The smallest change in the program, it needs none of P1–P4 to land, and it is the pain
actually reported. Full evidence in [`work-shape.md`](work-shape.md).

- `preflight` returns a **shape** — `read-only` | `bounded` | `durable` — and the playbook
  path follows from it:

  | Shape | Trigger | Records created |
  |---|---|---|
  | read-only | answer, review, diagnose, status | none |
  | bounded | single session, resumable from its diff | trace only |
  | durable | spans sessions, has dependencies, needs recovery, or cannot resume from its diff | one plan + trace + decisions |

- The escalation predicate is **"can this resume from its diff?"** — mechanical enough to
  ask, and already the upstream test
- A shape may **escalate mid-task, never de-escalate**. Bounded work that turns out to
  span sessions promotes to durable and writes its plan then — cheap, because nothing was
  written before
- **Templates emit no empty sections.** `### Critical` appears only with a critical
  finding; a conditional commit is written only when its condition holds. Template
  rendering, not policy
- `story`, `run`, and `check` records collapse into `trace` for bounded work. `story`'s 23
  entries bought nothing a trace does not

*Ships alone: the harness becomes cheap to run on small work, which is what gets P1–P4
exercised at all.*

### P1 — Durability core (M1 + M6)

- `cli/internal/interfaces/paths.go`: `dbPath` → `.harness/state/harness.db`,
  `changesetDir` → `.harness/state/changesets`, `conflictDir` →
  `.harness/state/conflicts`; add `harnessDir = ".harness"`; `docsDir` stays `docs`
- `preflight.go:24-30` playbook map → `.harness/playbooks/*.md`;
  plan globs in `preflight.go:20`, `next.go:33`, `db.go:20` → `.harness/plans/active/*.md`
- Extend `MigrateLayout` with a v2→v3 path (`.kit/` + root `harness.db` → `.harness/`),
  reusing the existing snapshot, parity-check, and rollback machinery
- `init.go:13` `gitignoreEntries` → `.harness/state/`
- `audit` detects legacy generations (`.kit/workflow-state.yml`, `.kit/planning/`,
  `.kit/runs/`, `.kit/reports/`) and reports each as a named finding with recovery text
- **`audit` hardening** (Q4): findings distinguish rungs on the enforcement ladder —
  local → optional hook → checked-in CI → branch protection — and **none proves another**,
  so "a check exists" is never reported as "a check ran and passed." Every finding uses
  the diagnostic standard: violating item, broken rule, **authority pointer**, next
  action. Never bare `validation failed`

*Ships alone: state is consolidated, any legacy generation becomes visible, and `audit`
stops overclaiming.*

### P2 — Decision durability (M4)

- `cli/docs/embedded/templates/decision.md` — ADR template: context, decision,
  alternatives rejected with rationale, consequences, plus **`## Status`**
  (Proposed | Accepted | Superseded | Rejected) and **`## Follow-Up`**. `Status` is what
  makes supersession work without deletion
- `zharness decision promote` — presents candidates against the five durability triggers
  and the human selects; writes `docs/decisions/NNNN-{slug}.md` from the selected DB rows,
  allocating the next number and updating `docs/decisions/README.md`'s index

  Triggers: a lasting product or architecture choice changes; public compatibility or data
  ownership changes; security or recovery policy changes; validation is materially added,
  removed, or weakened; or the source-of-truth hierarchy changes.
  Exclusion, stated in one line: **task-local choices stay in the active plan.**

- `zharness decision add --durable` — marks intent at the call site when it is already
  known, so the candidate list is pre-sorted rather than re-derived
- `handoff` playbook gate: a plan may not move to `completed/` until promotion has been
  **considered** — a non-empty candidate review, not a promote-everything requirement
- `zharness init` scaffolds an **empty** decision index plus the trigger criteria. Never
  mono-harness's own ADRs

*Ships alone: decisions survive a fresh clone even if nothing else lands.*

### P3 — Docs contract (M3 + M5 + M7)

- Scaffold `docs/README.md` (router), `docs/product/README.md`, `docs/decisions/README.md`
- `SyncManagedDocs` never overwrites a tracked file whose content differs from the
  previously installed version without `--force`; stages to
  `.harness/state/conflicts/*.upstream` and reports
- Emit `CLAUDE.md` containing `@AGENTS.md` when absent
- `audit` flags instruction files >4KB not referenced by `AGENTS.md`
- **Pruning as a named act** (Q4): `docs/README.md` and `docs/decisions/README.md` each
  carry a `## History` section recording what was *removed* and why. Superseded material
  is pruned from the tree so retrieval returns current authority; git history and
  immutable tags are the provenance. Deliberate absence is a retrieval feature, not
  data loss — but only when it is explained

*Ships alone: navigation and upgrade drift are fixed.*

### P4 — Knowledge map (M2)

- `zharness wiki` → `docs/map/{index,routes,config,graph,coverage}.md` and
  `docs/map/subsystems/*.md`
- Next.js App Router + TypeScript adapter first; adapter-shaped for other stacks
- Every generated file carries: DO-NOT-EDIT banner, freshness stamp, regeneration
  command, and an explicit negative-space section naming what it does not cover
- Every fact carries its source path; `index.md` prices each navigation path in tokens

*Ships alone: the derived tier exists and stops rotting by construction.*

## Validation Loop

**During work**

- `cd cli && go test ./...` after each change
- `bash scripts/verify-doc-links.sh`
- For P1: `zharness migrate layout --dry-run` against a fixture before any real move

**Final proof**

- Work-shape gate: the virtualizer-fix fixture produces 0 markdown files; the
  walter-theme fixture still classifies `durable` and keeps its plan
- No template renders an empty section — asserted per template, both directions
- Fresh-clone durability test passes (new test, added in P2)
- Resident-token measurement: entrypoint pair ≤ 900 **and** full resident path ≤ 2,400,
  using the existing `docs/audit/cost-model/` method
- `zharness wiki` run twice produces byte-identical output
- A CLI version bump against a fixture repo with committed playbooks leaves
  `git status` clean
- The embedded-docs projection-drift test stays green

**Both-directions proof.** A passing repository with no exercised violation does not prove
the guard can detect recurrence. Every new gate ships with a fixture that trips it.

## Stop / Pause

**Done when** all five phases are merged, all three hard gates pass, and
`zharness migrate layout` moves a v2 fixture to v3 with parity verified and rollback
proven.

**Pause if**

- shape classification cannot be made mechanical — if "can this resume from its diff?"
  needs a judgment call the CLI cannot make, P0 needs a different predicate, not a
  heuristic that guesses
- the migration cannot guarantee parity on a repo with uncommitted changesets
- either resident budget forces the router below usefulness — report the real number
  rather than shipping a useless map
- `zharness wiki` cannot reach determinism on a real repo without an AST, which means P4
  needs re-scoping rather than forcing

## Risk Note

P0 is the smallest build and the highest-felt payoff; it carries one real risk, which is
mis-classifying durable work as bounded. Escalate-only (never de-escalate) is the
mitigation: the cost of a wrong `bounded` call is writing the plan later, not losing it.

P4 is the largest build and the only phase with genuine unknowns (stack adapters).
P0–P3 fix everything that has already caused data loss, drift, or daily friction; P4
prevents future rot. If scope must be cut, P4 is the clean cut line.
