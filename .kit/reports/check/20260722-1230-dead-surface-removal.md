---
id: 01KY42FNJ8GR9EZN6RXGC858NV
type: check
phase: dead-surface-removal
lane: high-risk
mode: full
run_id: 01KY41Q3D4FCZMY2SS8AH6PD9R
proof_links: [{"command":"go build ./...","output_ref":"exit 0, no output","artifact_path":"cli/"},{"command":"go vet ./...","output_ref":"exit 0, no output","artifact_path":"cli/"},{"command":"go test ./...","output_ref":"ok across cmd/zharness, internal/application, internal/domain, internal/embedded, internal/infrastructure, internal/interfaces","artifact_path":"cli/"},{"command":"go test ./internal/infrastructure/... -run Migrat -v","output_ref":"PASS TestMigrate — applied [0001_init 0002_meta_docs_version 0003_drop_dead_surface], schema_version 3, table set matches trimmed 8-table list","artifact_path":"cli/internal/infrastructure/migrations_test.go"},{"command":"go test ./internal/application/... -run TestRunCreateReplaySafety -v","output_ref":"PASS — incremental vs. replayed Resume views byte-identical on fresh v3-schema DB","artifact_path":"cli/internal/application/run_create_replay_test.go"},{"command":"go test ./internal/application/... -run TestImport -v","output_ref":"PASS TestImportRoundTrip against cli/testdata/legacy-kit","artifact_path":"cli/internal/application/import_test.go"},{"command":"zharness --help | grep -E \"decision|backlog|tool|propose|score-context\"","output_ref":"empty (grep exit 1)","artifact_path":"cli/internal/interfaces/root.go"},{"command":"gofmt -l .","output_ref":"empty after fix (was: internal/application/score.go, internal/interfaces/audit.go, internal/interfaces/score.go — trailing blank lines from sub-agent deletions, fixed via gofmt -w)","artifact_path":"cli/"},{"command":"zharness audit --json","output_ref":"one pointer_drift (latest_check stale vs. latest_run_id — expected, resolves via this report's check record call); remaining contract_violations/unlinked_proofs are pre-existing debt predating this phase, unrelated to the dead-surface-removal diff","artifact_path":"n/a"},{"command":"zharness score-trace {3 wave trace ids} --json","output_ref":"all 3 traces scored tier=detailed","artifact_path":".kit/runs/work/20260722-1135-dead-surface-removal.md"}]
created: 2026-07-22
updated: 2026-07-22
---

# CHECK REPORT

Run ID: check-20260722-1230-dead-surface-removal
Scope: full
Artifact Alignment: aligned
Review Verdict: APPROVED
Phase: dead-surface-removal
Spec: .kit/planning/SPEC.md
Plan: .kit/planning/phases/dead-surface-removal/dead-surface-removal-PLAN.md
Cook Run: .kit/runs/work/20260722-1135-dead-surface-removal.md
Created At: 2026-07-22 12:30

## Gate Evidence
- tests: `go test ./...` → pass (7 packages, 1 no-test-files)
- types: n/a (Go — covered by build)
- lint: `go vet ./...` → pass; `gofmt -l .` → pass (after fixing 3 files' trailing blank lines left by Wave 1's sub-agent deletions)
- build: `go build ./...` → pass

## Artifact Alignment
- status: aligned
- notes:
  - **Spec coverage**: implements SPEC's "cut dead surface" requirement in full — `decision`/`backlog`/`tool`/`propose`/`score-context` commands, their application/domain logic, their DB tables, and their changeset-engine entries are all removed; CONTRACT.md/SCHEMA.md updated to match (in Allowed Surfaces, previously undone until this gate — fixed now, see Minor findings below for the process note).
  - **Boundary compliance**: changed files stayed within Allowed Surfaces (`cli/internal/interfaces/`, `cli/internal/application/`, `cli/internal/infrastructure/changeset.go`, `cli/internal/infrastructure/migrations.go`, `cli/docs/CONTRACT.md`/`SCHEMA.md`, corresponding `*_test.go`). One interpretation call, flagged transparently in the run artifact rather than made silently: `cli/internal/domain/{decision,backlog,tool}.go` were also deleted even though CONTEXT.md's Allowed-Surfaces bullets technically omit `cli/internal/domain/` — justified by the phase Goal's own "entities + commands + tables" wording, which requires the domain layer to be in scope for the removal to be complete (leaving orphaned domain types with no application/interface consumer would itself be dead code).
  - **Proof trail**: every materially changed behavior has a matching verification command, and the RUN artifact (`.kit/runs/work/20260722-1135-dead-surface-removal.md`) shows each of the 3 waves' commands actually ran, with a `detailed`-tier trace recorded per wave (see `proof_links` above and `score-trace` output).
  - Bookkeeping files outside this phase's Allowed Surfaces (`.kit/HANDOFF.md`, `.kit/planning/ROADMAP.md`, `.kit/planning/phases/write-boundary/write-boundary-{CONTEXT,PLAN}.md`, `.kit/harness.db`) are also present in the uncommitted diff — these are corrections for a **separate**, already-resolved discovery earlier this session (write-boundary was already done at commit `32cb60c` but its status text was stale) and an explicit user-approved action ("Fix bookkeeping, then move to Phase 2"), not scope drift from this phase's plan. Noted here for transparency; recommend committing them as a separate commit from the dead-surface-removal diff (see `git` skill).

## Findings
### Critical
- none

### Major
- none

### Minor / Suggestions
- 🟡 `cli/docs/embedded/AUTHORITY.md` still lists `backlog`/`decision`/`tool`/`score-context` in its "Forbidden" command list even though these commands no longer exist. `AUTHORITY.md` is **not** in this phase's Allowed Surfaces (per CONTEXT.md), so left untouched rather than silently expanded — worth a follow-up doc pass (could fold into a future doc-hygiene task or Phase 4 `single-source-playbooks`).
- 🟡 `cli/docs/SCHEMA.md`'s Verification section already undercounted `handoff record` before this phase (its entity-producing command list excludes `handoff record` even though CONTRACT.md documents it as the 20th command). Pre-existing gap, unrelated to dead-surface-removal — the post-removal counts in this edit (14 commands, 8 tables) preserve that same pre-existing baseline methodology rather than silently fixing an unrelated discrepancy; flagged here instead.
- 💡 `zharness audit --json`'s `contract_violations`/`unlinked_proofs` lists carry a long tail of pre-existing entries from earlier phases (readme-workflow-refresh phase-file mismatch, several checks with unlinked proof artifacts, non-ULID `plan_id`/`run_id`/`id` values in older reports) — none touch this diff's files; not this phase's responsibility to clean up.

## Next Action
- ready for PR — suggest `git` (split into 2 commits: dead-surface-removal proper, and the separate write-boundary bookkeeping fix) then `handoff`
