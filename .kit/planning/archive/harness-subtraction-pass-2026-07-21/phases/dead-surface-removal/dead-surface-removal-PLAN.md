# Plan: dead-surface-removal

Phase: dead-surface-removal
Status: ready
Wave Count: 3
Execution Owner: work
Updated At: 2026-07-21

## Goal
Delete `decision`/`backlog`/`tool`/`propose`/`score-context` surface + tests + tables; bump schema_version; prove replay + import still work.

## Inputs
- grep inventory of the five names across `cli/`, `.kit/changesets/`, playbooks
- `changeset.go`, `migrations.go`, `audit.go`, `score.go`, interfaces/*

## Wave 1
### T1 — Verify-then-remove commands + application logic
- type: refactor
- inputs:
  - `grep -rn "decision\|backlog\|tool\|propose\|score-context" cli/ .kit/changesets/ .kit/docs/`
- touches:
  - `cli/internal/interfaces/*` (command registration), `cli/internal/application/{audit.go,score.go, decision/backlog/tool files}`
- avoid:
  - `ScoreTrace`, `entropyScore` (Phase 3); changeset format
- steps:
  1. Run the grep; confirm no consumer outside the code being deleted. If any consumer exists → STOP (escalate).
  2. Remove the five cobra subcommands and their handlers.
  3. Remove `Propose` (audit.go) and `ScoreContext` (score.go) functions.
  4. Remove decision/backlog/tool application logic.
- expected outputs:
  - `zharness --help` lists none of the five
- verification:
  - `cd cli && go build ./... 2>&1` → compiles; `zharness --help | grep -E "decision|backlog|tool|propose|score-context"` → empty
- stop if:
  - grep finds a real consumer
- escalate to:
  - brainstorm refine

## Wave 2
### T2 — Remove entities from changeset engine + drop tables
- type: migration
- inputs:
  - `changeset.go` (`entityTables`, `entityColumns`), `migrations.go`
- touches:
  - `cli/internal/infrastructure/changeset.go`, `cli/internal/infrastructure/migrations.go`
- avoid:
  - trace/run/check/handoff/intake/intervention/story entities (all live)
- steps:
  1. Remove `decision`/`backlog`/`tool` keys from `entityTables` and `entityColumns`.
  2. Add a forward migration dropping `decisions`, `backlog`, `tools`; bump `schema_version` 2 → 3.
  3. Ensure `Migrate` applies cleanly on a fresh DB and on an existing v2 DB.
- expected outputs:
  - new DB has no dropped tables; schema_version = 3
- verification:
  - `go test ./cli/internal/infrastructure/ -run Migrat` → pass
- stop if:
  - migration can't drop tables safely on an existing v2 db
- escalate to:
  - to-plan phase dead-surface-removal

## Wave 3
### T3 — Delete tests + prove replay/import safety
- type: test
- inputs:
  - existing `*_test.go` for removed surface; committed changesets under `.kit/changesets/`
- touches:
  - test files; new replay-safety assertion
- avoid:
  - weakening live-entity tests
- steps:
  1. Delete tests that only covered removed commands/entities.
  2. Add/confirm a test: replay all committed changesets on a fresh v3 DB → `resume --json` unchanged vs a pre-change baseline.
  3. Confirm `import` of `cli/testdata/legacy-kit` still passes.
  4. Full `go test ./...` + `go build ./...`.
- expected outputs:
  - green suite; replay + import unaffected
- verification:
  - `cd cli && go test ./... && go build ./...` → pass (capture output)
- stop if:
  - replay or import regresses
- escalate to:
  - to-plan phase dead-surface-removal

## Risks / Watch-fors
- The grep-verify gate in T1 is a hard stop — never delete before confirming zero consumers.
- Schema bump is the highest-risk change in the whole initiative; the replay + import proofs are mandatory, not optional.
