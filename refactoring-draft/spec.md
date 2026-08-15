---
title: Harness Durability & Knowledge Contract
status: draft
interviewed: 2026-08-15
---

## Outcome

Nothing an agent or human needs to remember survives only in a gitignored path, and a
cold agent can orient in a consumer repo without loading the repo's documentation.
`zharness` owns one directory (`.harness/`), project knowledge owns another (`docs/`),
and the derived half of `docs/` is generated rather than maintained.

## Success Condition

Two hard gates, both mechanically checked:

1. **Durability** — a test clones the repo fresh (no `.harness/state/`), runs
   `zharness db rebuild`, and asserts every decision, plan, product doc, and ADR is
   present. Today this test fails: decisions exist only in `harness.db` and
   `.kit/changesets/`, both gitignored.

2. **Resident cost** — `AGENTS.md` managed block + `.harness/WORKFLOW.md` +
   `docs/README.md` ≤ **1,000 tokens combined**. The current zharness-owned surface is
   ~750 tokens; this budget is what constrains the new router to stay a map instead of
   growing into a document.

Field reference: a cold agent in onedrive-cloud today spends ~17k tokens on process and
docs before reading any `src/`. See [`audit-onedrive-cloud.md`](audit-onedrive-cloud.md).

## Scope

**May change**

- `cli/internal/interfaces/paths.go`, `preflight.go`, `next.go`, `db.go` — path constants and globs
- `cli/internal/application/{init,layout_migration,managed_docs}.go`
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

## Context to Read First

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

1. **Two directories, by owner** — `.harness/` = workflow (CLI/skills), `docs/` = project
   knowledge. Because mixing them produced 68KB of `docs/` where product architecture sits
   beside stage playbooks.
2. **`.kit/` is absorbed, not kept** — `.harness/state/` holds db, changesets, conflicts,
   cache, log; everything else in `.harness/` is tracked. Because `.kit` being 100%
   gitignored stranded real product authority one `rm -rf` from gone.
3. **Ship as a release with migration** — extend `MigrateLayout`. Because `legacyDBPath`
   already sets the precedent and consumers must not break on upgrade.
   Depends on: decision 2.
4. **Decisions promote to `docs/decisions/` ADRs at handoff** — plans stay disposable,
   the knowledge base grows. Depends on: decision 5.
5. **`docs/` splits generated from written** — `docs/map/` generated, `docs/product/` and
   `docs/decisions/` written. Because hand-maintaining derived facts is what rotted
   `codebase-summary.md` for 8 months.
6. **`zharness wiki` is deterministic** — file walk plus regex, no AST, no LLM, 0 tokens.
   Reference codesight's method, not codesight as a dependency. Verified tractable: 15
   routes, 11 env vars, and coverage JSON are all derivable without a TypeScript parser.
   Depends on: decision 5.
7. **A CLI upgrade must never modify a tracked file** — stage to
   `.harness/state/conflicts/*.upstream` and report. Because it silently rewrote six
   playbooks in a live repo, byte-identical to the embedded versions.

## Phases

Each phase is independently mergeable: after it ships, the system is usable even if the
next never lands.

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

*Ships alone: state is consolidated and any legacy generation becomes visible.*

### P2 — Decision durability (M4)

- `cli/docs/embedded/templates/decision.md` — ADR template (context, decision,
  alternatives rejected with rationale, consequences)
- `zharness decision promote` — writes `docs/decisions/NNNN-{slug}.md` from DB rows,
  allocating the next number and updating `docs/decisions/README.md`'s index
- `handoff` playbook gate: promotion required before a plan moves to `completed/`

*Ships alone: decisions survive a fresh clone even if nothing else lands.*

### P3 — Docs contract (M3 + M5 + M7)

- Scaffold `docs/README.md` (router), `docs/product/README.md`, `docs/decisions/README.md`
- `SyncManagedDocs` never overwrites a tracked file whose content differs from the
  previously installed version without `--force`; stages to
  `.harness/state/conflicts/*.upstream` and reports
- Emit `CLAUDE.md` containing `@AGENTS.md` when absent
- `audit` flags instruction files >4KB not referenced by `AGENTS.md`

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

- Fresh-clone durability test passes (new test, added in P2)
- Resident-token measurement ≤ 1,000, using the existing `docs/audit/cost-model/` method
- `zharness wiki` run twice produces byte-identical output
- A CLI version bump against a fixture repo with committed playbooks leaves
  `git status` clean
- The embedded-docs projection-drift test stays green

## Stop / Pause

**Done when** all four phases are merged, both hard gates pass, and
`zharness migrate layout` moves a v2 fixture to v3 with parity verified and rollback
proven.

**Pause if**

- the migration cannot guarantee parity on a repo with uncommitted changesets
- the resident budget forces the router below usefulness — report the real number rather
  than shipping a useless map
- `zharness wiki` cannot reach determinism on a real repo without an AST, which means P4
  needs re-scoping rather than forcing

## Risk Note

P4 is the largest build and the only phase with genuine unknowns (stack adapters).
P1–P3 fix everything that has already caused data loss or drift; P4 prevents future rot.
If scope must be cut, P4 is the clean cut line.
