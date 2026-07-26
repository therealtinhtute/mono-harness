# Plan: root layout

Phase: root-layout
Status: ready
Wave Count: 4
Execution Owner: work
Updated At: 2026-07-26

## Goal
Ship the final storage/docs layout with safe migration while keeping pre-v3 artifact behavior operational.

## Inputs
- completed universal-preflight phase
- root-layout-CONTEXT.md
- 43 existing changesets and current DB snapshot

## Wave 1
### T1 — Add root path contract and managed-doc schema
- type: implementation + migration + test
- touches:
  - `cli/internal/interfaces/{init,db,migrate,import,validate}.go`
  - `cli/internal/application/init.go`
  - `cli/internal/infrastructure/{migrations,changeset}.go`
  - schema/path tests
- avoid:
  - lifecycle artifact path removal
  - historical changeset edits
- steps:
  1. Centralize root DB, `.kit`, changeset, docs, and conflict paths.
  2. Add `managed_docs(path, installed_sha256, docs_version, updated_at)` additively.
  3. Extend changeset application for managed-doc rows.
  4. Update ignore entries for DB/WAL/SHM/conflicts.
  5. Prove old changesets replay under the new schema.
- expected outputs:
  - one path contract and additive schema
- verification:
  - `cd cli && go test ./internal/infrastructure ./internal/application ./internal/interfaces`
- stop if:
  - replay requires modifying an existing changeset
- escalate to:
  - check

## Wave 2
### T2 — Implement safe root docs projection
- type: implementation + test
- touches:
  - `cli/internal/application/init.go`
  - embedded manifest/projection code and tests
  - root `AGENTS.md` managed-block logic
  - `cli/docs/embedded/**`
- avoid:
  - workflow content rewrite beyond path compatibility
- steps:
  1. Project WORKFLOW/playbooks to root docs.
  2. Implement marked-block insert/update preserving outside bytes.
  3. Implement hash cases: missing, untouched, already-current, local-only, both-changed conflict.
  4. Stage only upstream conflict content under ignored `.kit/conflicts/`.
  5. Keep explicit force as the only overwrite path.
- expected outputs:
  - safe managed docs refresh without manifest file
- verification:
  - `cd cli && go test ./internal/application ./internal/embedded -run 'Scaffold|Managed|Projection|Agents'`
- stop if:
  - any default path overwrites a locally changed file
- escalate to:
  - check

## Wave 3
### T3 — Add reusable layout migration
- type: implementation + integration test
- touches:
  - `cli/internal/interfaces/migrate.go`
  - new application/infrastructure migration helpers
  - migration integration tests
  - `cli/docs/CONTRACT.md`
- avoid:
  - deleting legacy lifecycle markdown
- steps:
  1. Preserve standalone schema migration behavior.
  2. Add `migrate layout --to v2 --dry-run` command shape.
  3. Replay changesets into a temporary root DB.
  4. Compare normalized resume/query views.
  5. Activate atomically only after parity and docs projection succeed.
  6. Leave old state untouched on every tested failure point.
- expected outputs:
  - dry-run and apply migration with rollback guarantees
- verification:
  - `cd cli && go test ./... -run 'LayoutMigration|Replay|Rollback'`
- stop if:
  - migration can leave two active DBs or partially activated docs
- escalate to:
  - check

## Wave 4
### T4 — Dogfood root layout
- type: migration + docs
- touches:
  - root `.gitignore`, `AGENTS.md`, `docs/WORKFLOW.md`, `docs/playbooks/`
  - `.kit/harness.db` tracking state
  - `.kit/changesets` only through normal migration metadata writes
- avoid:
  - `.kit/planning`, `.kit/runs`, `.kit/reports`, `.kit/HANDOFF.md` deletion
- steps:
  1. Capture pre-migration resume/query JSON.
  2. Build the current CLI and run migration dry-run.
  3. Apply migration and compare normalized state.
  4. Remove `.kit/harness.db` from Git tracking only after parity.
  5. Run full tests and projection checks.
- expected outputs:
  - this repository operates on the final storage/docs layout
- verification:
  - `cd cli && go test ./... && go build -o ../bin/zharness ./cmd/zharness`
- stop if:
  - any parity field, drift entry, or changeset status differs unexpectedly
- escalate to:
  - check

## Risks / Watch-fors
- During dogfood, use the newly built binary explicitly; the globally installed 0.5.0 binary still expects old paths.
- Do not delete the old DB until replay parity is recorded.
