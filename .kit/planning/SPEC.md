---
id: 01KY1ACKSV6F2TSV46EYED8JRN
type: spec
phase: none
lane: high-risk
intake_id: 01KY1AG58T7HEV3JYKGCBWTQMY
created: 2026-07-21
updated: 2026-07-21
---

# SPEC: Harness Subtraction Pass — close the write-boundary, cut dead surface, drop fake scoring

Status: locked
Input Type: harness-improvement
Lane: high-risk
Risk Flags: public-contract, existing-behavior, data-model, multi-domain
Affected Surfaces: provider, docs
Downstream: to-plan full
Updated At: 2026-07-21

## Source Mode
files

## Source Inputs
- `.kit/reports/audit/20260721-harness-architecture-audit.md` — the architecture audit (source of truth for the problems)
- User scoping decisions (brainstorm, 2026-07-21): Scope **A + remove scoring**; SQLite **kept, add write-commands**; scoring layer **removed in favor of the proof matrix**
- Code read this session: `cli/internal/infrastructure/changeset.go`, `application/{score,audit}.go`, `.kit/docs/playbooks/{work,check}.md`

## Scenario
refine existing spec — a subtraction/refactor initiative on the harness itself

## Goal
Make the CLI own 100% of harness writes, remove built-but-unused surface, and delete the "deterministic verdict" scoring that measures string length rather than evidence — so the harness is leaner, its playbooks shorter, and no non-deterministic component (the LLM) ever hand-authors durable state. Keep every guarantee that matters: replay, traceability, and the lane×proof gate.

## Users / Actors
The single owner-operator of this repo (the author), via the `work`/`check` skills and the `zharness` CLI. No second user; no multi-team concern.

## Requirements
1. **Close the write-boundary.** Add CLI commands so no playbook hand-authors changeset JSONL:
   - `zharness run create` — registers the run row AND sets `meta.latest_run_id` in one transaction (replaces `work.md` step 2 full-mode hand-authored two-line changeset).
   - `check record` sets `meta.latest_check_id` itself (via flag or by default), replacing the hand-authored meta changeset in `check.md` step 4.
   - Audit every playbook for any remaining "hand-author a changeset" instruction; each must map to a real command.
2. **Delete dead surface** (built, tested, cobra-wired, zero playbook consumers — grep-verified): `decision`, `backlog`, `tool` subcommands + their entities/tables/columns; `propose` and `score-context` commands. Remove their tests. Bump `schema_version` if tables are dropped from `migrations.go`.
3. **Remove the scoring layer, keep the proof matrix.** Delete `score-trace` tier logic and `entropy_score` from `audit --json`; remove the `score-trace` loop from `check.md` Step 4. The lane×proof-class matrix remains the verdict. Update `CONTRACT.md` for the changed `audit --json` shape.
4. **Single-source the playbooks.** Make `.kit/docs/playbooks/*` a pure projection of the Go embed (`cli/docs/embedded/playbooks/`). Add a test (or `zharness playbooks verify`) that fails if a scaffolded copy diverges from the embed. Humans edit only the embed.
5. **Update the playbooks** (`work.md`, `check.md`) to call the new commands and drop the removed steps — net line reduction is a success signal.

## Boundaries
### In Scope
- `cli/internal/**` command additions (run create, check meta pointer), deletions (dead entities/commands), scoring removal
- `cli/internal/infrastructure/migrations.go` schema change for dropped tables (+ schema_version bump)
- `cli/docs/CONTRACT.md`, `SCHEMA.md` updates for changed/removed command shapes
- Embedded playbooks (`cli/docs/embedded/playbooks/{work,check}.md`) rewrites; `.kit/docs/` re-scaffold
- A playbook drift-check test
- Tests for every new command; removal of tests for deleted surface

### Out of Scope
- **Dropping SQLite / in-memory fold** (audit #7) — deferred; DB is kept as-is
- **Memory unification** (audit #4) — deferred to a later initiative
- **Playbook length reduction beyond what these changes naturally remove** (audit #5) — deferred
- **Folding `interview` into `brainstorm`, `zharness next` front door** (audit #8, #9) — deferred
- **Making scoring "real" by enriching the trace schema** — explicitly rejected (chose removal)
- Any change to the changeset JSONL on-disk format or ULID/fence/replay mechanics

## Constraints
- Changesets are append-only and replayed from empty: the existing 10 committed changesets under `.kit/changesets/` MUST still replay byte-clean after the schema change (none reference dropped entities — verify).
- `harness.db` is gitignored/rebuildable; the schema_version bump must not break `init` on a fresh clone or `import` on a legacy `.kit/`.
- No behavior change to the lane×proof gate's pass/fail outcome (only the vestigial score output is removed).
- Solo/local-first; no network, no new heavy dependencies.

## Acceptance Criteria
- `zharness --help` no longer lists `decision`, `backlog`, `tool`, `propose`, `score-context`; `go build ./...` green; `go test ./...` green with dead-surface tests removed.
- `grep` across `cli/docs/embedded/playbooks/**` and `.kit/docs/playbooks/**` finds **no** instruction to hand-author a `.changeset.jsonl` file (changeset literals appear only inside `db changeset apply` examples, not as author-this steps).
- `zharness run create ...` creates the run row and sets `latest_run_id` atomically (verified by `query state --json` + a replay-from-scratch check).
- `check record` sets `latest_check_id` with no separate hand-authored meta changeset.
- `zharness audit --json` output no longer contains an `entropy_score` field; `check.md` no longer calls `score-trace`; a real `check` gate run still blocks on a missing required proof cell.
- A test asserts `.kit/docs/playbooks/*` == embed byte-for-byte.
- Replay of the existing committed changesets reproduces the same `resume --json` as before (no regression from the schema change).

## Validation Expectations
- **unit** (required): Go unit tests for `run create`, `check record` meta-pointer, migration/schema change, scoring removal.
- **integration** (required): init → run create → replay-from-scratch produces identical state; import of a legacy `.kit/` still works after schema bump.
- **command-output** (required): real `zharness` binary smoke test of the new/changed commands, captured verbatim.
- **manual-check** (required, high-risk): review pass over the diff (Security/Arch/Quality), especially migration safety and playbook correctness.
- **e2e** (optional): a full `work → check` dry pass on a scratch phase using the new commands.

## Dependencies / Assumptions
- `decision`/`backlog`/`tool`/`propose`/`score-context` are truly unconsumed — verified by grep this session; re-verify at plan time before deleting.
- The proof matrix in `check.md` is self-sufficient as the verdict without any trace tier input — confirmed by reading `check.md` Step 4 (matrix FAIL is already the hard stop; score-trace only gates whether a trace "counts" as evidence, which the command-output/manual-check classes already cover).
- goreleaser/install path unaffected (no release-format change in scope).

## Key Decisions
- **Scope A (subtraction slice) + scoring removal**, not full core rework — chosen for low risk, no cross-dependencies, fastest leverage. Rejected: Option C (drop SQLite + unify memory + shrink playbooks) — high effort, touches replay, deferred to a later initiative.
- **Keep SQLite, add write-commands** — chosen as the conservative boundary fix. Rejected: dropping the DB and folding the changeset log in memory (audit Option B) — larger blast radius on query/resume/audit, deferred.
- **Remove scoring, keep the matrix** — the lane×proof matrix is the real, meaningful verdict; `score-trace`/`entropy_score` are deterministic-but-meaningless (measure string length / finding counts). Rejected: enriching the trace schema to make scoring valid — heavier, schema-touching, out of this slice.

## Open Questions
- Does `check record`'s meta-pointer become default behavior or an opt-in `--set-latest` flag? (Lean: default, since the playbook always wants it; confirm at plan time.)
- Drop the dead tables in `migrations.go` (schema_version bump) vs leave the tables but remove the commands? (Lean: drop tables for a true subtraction; must prove replay safety. Decide in to-plan.)

## Deferred Ideas
- Drop SQLite / in-memory state fold (audit #7)
- Unify the two memory systems into the `decisions` store (audit #4, #7)
- Shrink playbooks to target lengths / DRY the version-gate boilerplate (audit #5, #10)
- Fold `interview` into `brainstorm --grill`; add `zharness next` front door (audit #8, #9)
- Push procedural stages to Sonnet (audit #12)

## Ambiguity Report
- **Goal clarity**: high — three concrete, independent changes with a shared theme (subtraction).
- **Scope clarity**: high — In/Out scope explicit; the two big deferrals (SQLite, memory) named.
- **Constraints clarity**: high — replay-safety and no-gate-behavior-change are the hard constraints.
- **Acceptance clarity**: high — each criterion is grep-able, test-able, or a captured command output. Two open questions (defaults) are plan-time details, not scope gaps.
