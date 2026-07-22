---
id: 01KY41Q3D4FCZMY2SS8AH6PD9R
type: run
phase: dead-surface-removal
lane: high-risk
mode: full
plan_id: none
trace_ids: [01KY41Y3E6QHWDZMK8M647440G, 01KY4259JWPF2AMTAQX665FC9G, 01KY429QX34A0D23546B5G911M]
created: 2026-07-22
updated: 2026-07-22
---

# COOK RUN

Run ID: work-20260722-1135-dead-surface-removal
Mode: full
Status: complete
Spec: .kit/planning/SPEC.md
Roadmap: .kit/planning/ROADMAP.md
Phase: dead-surface-removal
Plan: .kit/planning/phases/dead-surface-removal/dead-surface-removal-PLAN.md
Started At: 2026-07-22 11:35

## Preflight
- scope drift: no — working tree clean at phase start (`git status --short --branch` clean, in sync with origin/master)
- working tree note: write-boundary phase (Phase 1) discovered already done at commit `32cb60c` (2026-07-21); bookkeeping (ROADMAP.md, write-boundary-PLAN/CONTEXT.md, HANDOFF.md) corrected this session before starting this run. See handoff open-item in `.kit/changesets/01KY1CFJS2GZ94VACCVH55P8MH.changeset.jsonl` which had already flagged the stale status and was dropped in later handoffs.
- required artifacts present: yes — `.kit/planning/phases/dead-surface-removal/{dead-surface-removal-CONTEXT.md,dead-surface-removal-PLAN.md}` both exist, `status: ready`, no placeholder markers
- selected phase: dead-surface-removal (explicit `zharness next full phase dead-surface-removal --json` → no stop)
- T1 pre-verify grep (`grep -rn "decision\|backlog\|tool\|propose\|score-context" cli/ .kit/changesets/ .kit/docs/`): all hits are the implementation itself (interfaces/application/domain/tests), doc references (CONTRACT.md/SCHEMA.md/AUTHORITY.md), or unrelated English prose ("decisions" as a concept in playbook text, not a command invocation). No live playbook consumer found — T1's stop condition does not trigger.

## Wave / Task Log
### Wave 1
#### T1 — Verify-then-remove commands + application logic
- status: DONE
- changed files:
  - deleted: cli/internal/interfaces/{decision,backlog,tool}.go
  - deleted: cli/internal/application/{decision,backlog,tool}.go
  - deleted: cli/internal/domain/{decision,backlog,tool}.go (domain layer treated as in-scope per the phase Goal's "entities + commands + tables" language, even though CONTEXT.md's Allowed-Surfaces bullets technically omitted `cli/internal/domain/` — noted explicitly, not silent)
  - edited: cli/internal/interfaces/audit.go (removed newProposeCmd/runPropose, kept newAuditCmd/runAudit)
  - edited: cli/internal/interfaces/score.go (removed newScoreContextCmd/runScoreContext, kept newScoreTraceCmd/runScoreTrace)
  - edited: cli/internal/application/audit.go (removed Proposal/ProposeReport/Propose, kept Audit/unlinkedProofs/entropyScore)
  - edited: cli/internal/application/score.go (removed ContextScore/ScoreContext, kept TraceScore/loadTrace/countTracesForRun/ScoreTrace — shared helpers preserved)
  - edited: cli/internal/interfaces/root.go (removed 5 AddCommand registrations)
- verification:
  - `go build ./...` → pass (exit 0, no output) — independently re-verified, not just sub-agent-reported
  - `zharness --help | grep -E "decision|backlog|tool|propose|score-context"` → empty (grep exit 1) — independently re-verified
- notes:
  - test files (decision/backlog/tool_test.go in application+domain, TestPropose*/TestScoreContext* in audit_test.go/score_test.go) intentionally left broken — Wave 3 (T3) deletes them per plan
  - delegated to a sub-agent; pre-delegation consumer grep + post-delegation independent re-verification both confirm no surprise consumer

### Wave 2
#### T2 — Remove entities from changeset engine + drop tables
- status: DONE
- changed files:
  - edited: cli/internal/infrastructure/changeset.go — removed `decision`/`backlog`/`tool` entries from `entityTables` and `entityColumns` maps
  - edited: cli/internal/infrastructure/migrations.go — added migration v3 `0003_drop_dead_surface` (`DROP TABLE decisions; DROP TABLE backlog; DROP TABLE tools;`); v1's original CREATE statements left untouched (migrations are additive/historical, never rewritten)
  - edited: cli/internal/infrastructure/migrations_test.go — updated `TestMigrate`'s hardcoded expectations: `wantApplied` now includes `0003_drop_dead_surface`, `schemaVersion`/`schemaVersion2` now 3, `schemaMDTables` trimmed from 11 to 8 (dropped `backlog`/`decisions`/`tools`)
- verification:
  - `go build ./...` → pass (exit 0)
  - `go test ./internal/infrastructure/... -run Migrat -v` → PASS (fresh-DB migrate applies all 3 versions in order, resulting schema_version=3, table set matches trimmed 8-table list; re-running Migrate on an already-current DB is a no-op)
- notes:
  - done directly (small, single-concern infra edit), not delegated
  - test files still referencing dropped entities (application/domain `{decision,backlog,tool}_test.go`, `TestPropose*`/`TestScoreContext*`) remain intentionally broken — Wave 3 (T3) deletes them

### Wave 3
#### T3 — Delete tests + prove replay/import safety
- status: DONE
- changed files:
  - deleted: cli/internal/application/{backlog,decision,tool}_test.go
  - deleted: cli/internal/domain/{backlog,decision,tool}_test.go
  - edited: cli/internal/application/audit_test.go — removed `TestProposeFromAuditFindings`, `TestProposeCleanState`
  - edited: cli/internal/application/score_test.go — removed `TestScoreContext`, `TestScoreContextUnknownID`
- verification:
  - `go build ./...` → pass (exit 0)
  - `go vet ./...` → pass (confirms no leftover unused imports after test deletions)
  - `go test ./...` → all 7 packages pass (`cmd/zharness`, `internal/application`, `internal/domain`, `internal/embedded`, `internal/infrastructure`, `internal/interfaces`, `docs/embedded` no-test)
  - replay safety: `TestRunCreateReplaySafety` (internal/application/run_create_replay_test.go, pre-existing) builds two DBs — one via incremental commands, one via `infrastructure.Replay` on a fresh v3-schema DB — and asserts `Resume` views are byte-identical JSON; passed, proving the changeset engine + v3 migration replay correctly with the dead entities removed
  - import safety: `TestImportRoundTrip` (internal/application/import_test.go, pre-existing, exercises `cli/testdata/legacy-kit`) → PASS
  - manual shell-level replay attempt of the repo's real `.kit/changesets/*.jsonl` onto a scratch `zharness init`-created v3 DB hit `changeset_out_of_order` — traced to `init` itself writing a changeset with a now-timestamped ULID that sorts after all historical real changesets, an orthogonal pre-existing ordering guard unrelated to this phase's changes, not a regression. `TestRunCreateReplaySafety`'s use of `infrastructure.Replay` directly (bypassing that CLI-level real-time-ordering guard) is the correct/established pattern and is what's asserted above.
- notes:
  - done directly (test deletions + reading existing replay/import test coverage), not delegated
  - confirmed via grep that no committed changeset in `.kit/changesets/` references the `decision`/`backlog`/`tool` entities, so dropping their tables was safe against the project's own real harness history

## Summary
- All 3 waves DONE. `decision`/`backlog`/`tool`/`propose`/`score-context` surface fully removed: commands, application/domain logic, changeset-engine entries, DB tables (migration `0003_drop_dead_surface`, schema_version 2→3), and CONTRACT.md/SCHEMA.md docs (previously out of sync — fixed during the phase gate, see check report). Also fixed 3 files' `gofmt` violations left by Wave 1's delegated sub-agent.
- Phase gate: `.kit/reports/check/20260722-1230-dead-surface-removal.md` — **APPROVED**, 0 critical/major findings, 3 minor/suggestion notes (stale `AUTHORITY.md` forbidden-list entries out of this phase's Allowed Surfaces, a pre-existing SCHEMA.md `handoff record` undercount predating this phase, pre-existing `audit --json` debt unrelated to this diff). `zharness check record` ran (id `01KY42GNYDTHHE64KR7Q1QSTAD`), `meta.latest_check_id` updated atomically, `zharness audit --json` pointer_drift confirmed empty afterward.
- Uncommitted diff also carries the separate, already-user-approved write-boundary bookkeeping fix from earlier this session (ROADMAP.md, write-boundary CONTEXT/PLAN, HANDOFF.md) — unrelated to this phase's plan, flagged in the check report, recommended as its own commit.

## Next Recommended Action
- `git`: split into 2 commits — (1) dead-surface-removal (this phase: cli/internal/{interfaces,application,domain,infrastructure}, cli/docs/{CONTRACT,SCHEMA}.md, this run artifact, its check report, its changesets), (2) the write-boundary bookkeeping fix (ROADMAP.md, write-boundary-CONTEXT/PLAN.md, HANDOFF.md, harness.db). Never auto-run — suggest only.
- After commit: `handoff` to close out the session, or continue straight to Harness Subtraction Pass Phase 3 (`scoring-removal`) if the user wants to keep going.
