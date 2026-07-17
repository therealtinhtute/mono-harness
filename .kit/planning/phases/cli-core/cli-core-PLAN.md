# Plan: cli-core

Phase: cli-core
Status: ready
Wave Count: 4
Execution Owner: work
Updated At: 2026-07-17

## Goal
Working `zharness` core: init→import→query round-trip on a legacy fixture, changeset idempotency proven in CI, binaries releasable, install script functional.

## Inputs
- `cli/docs/CONTRACT.md`, `SCHEMA.md`, `STATE.md` (Phase 2)
- A real legacy `.kit/` sample (copy one from an existing project into `cli/testdata/legacy-kit/`)

## Wave 1
### T1 — Scaffold Go module + layer skeleton
- type: implementation
- inputs:
  - CONTRACT.md command list
- touches:
  - `cli/go.mod`, `cli/cmd/zharness/main.go`, `cli/internal/{interfaces,application,domain,infrastructure}/` package stubs
- avoid:
  - implementing command logic in this task
- steps:
  1. `go mod init github.com/therealtinhtute/skills/cli`; add cobra, modernc.org/sqlite, oklog/ulid
  2. Wire cobra root + subcommand registration in `internal/interfaces`; version flag from build ldflags
  3. Enforce layering: domain imports nothing internal; interfaces→application→domain/infrastructure
- expected outputs:
  - `go build ./...` produces a `zharness` binary answering `--version` and `--help` with all core commands listed
- verification:
  - `cd cli && CGO_ENABLED=0 go build ./... && go vet ./...`
- stop if:
  - any dependency pulls cgo (`go list -deps` shows C imports)
- escalate to:
  - to-plan phase cli-core

## Wave 2
### T2 — SQLite store + migrations (infrastructure)
- type: implementation
- inputs:
  - SCHEMA.md tables
- touches:
  - `cli/internal/infrastructure/` (store, migrations), `cli/internal/domain/` (entity structs)
- avoid:
  - tables for domain commands not yet needed? No — create the FULL schema now (SCHEMA.md is frozen); avoid only query helpers beyond core needs
- steps:
  1. Implement `init` (create db) + `migrate` (versioned migrations) per SCHEMA.md
  2. Entity structs in domain with validation stubs
- expected outputs:
  - `zharness init && zharness migrate` creates all SCHEMA.md tables
- verification:
  - `cd cli && go test ./internal/infrastructure/ -run TestMigrate` — asserts table list matches SCHEMA.md
- stop if:
  - SCHEMA.md ambiguity forces invented columns
- escalate to:
  - to-plan phase harness-contracts

### T3 — Changeset engine (append/apply/replay)
- type: implementation
- inputs:
  - SCHEMA.md changeset line shape + naming
- touches:
  - `cli/internal/infrastructure/` (changeset writer/reader), `cli/internal/application/` (append-then-apply orchestration)
- avoid:
  - editing/compacting past changesets (append-only invariant)
- steps:
  1. Implement ULID-named `{ulid}.changeset.jsonl` writer under `.kit/changesets/`
  2. Implement `db changeset apply <path>` (idempotent) + `db changeset status`
  3. Implement full replay: fresh db + all changesets in ULID order = materialized state
- expected outputs:
  - Append→apply→replay pipeline with idempotency
- verification:
  - `cd cli && go test ./... -run 'TestChangeset(Idempotent|Replay)'` — double-apply no-op; replay equals incremental state
- stop if:
  - an operation cannot be expressed as an append-only changeset
- escalate to:
  - brainstorm refine

## Wave 3
### T4 — import (legacy .kit seeding) + query
- type: implementation
- inputs:
  - STATE.md legacy mapping; `cli/testdata/legacy-kit/` fixture
- touches:
  - `cli/internal/application/`, `cli/internal/interfaces/` (import, query commands)
- avoid:
  - mutating the legacy files being imported
- steps:
  1. Implement `import`: parse legacy `workflow-state.yml` + planning markdown → changesets → DB, per STATE.md mapping; second run = no-op
  2. Implement `query` views (at minimum: `state`, `phases`, `artifacts`) with `--json`
- expected outputs:
  - Round-trip works on the fixture
- verification:
  - `cd cli && go test ./... -run TestImportRoundTrip` — init→import→`query state --json` matches golden file; re-import produces zero new changesets
- stop if:
  - a legacy field's mapping is missing from STATE.md
- escalate to:
  - to-plan phase harness-contracts

## Wave 4
### T5 — Release pipeline + install script
- type: implementation
- inputs:
  - T1–T4 green
- touches:
  - `.goreleaser.yaml` (in cli/ or root), `.github/workflows/cli-release.yml`, `.github/workflows/cli-ci.yml`, `scripts/install-zharness.sh`
- avoid:
  - publishing a release tag before CI is green; raw public URLs (private repo)
- steps:
  1. CI workflow: CGO_ENABLED=0 build + `go test ./...` on push affecting `cli/**`
  2. goreleaser config: darwin/linux × amd64/arm64; release on `cli/v*` tags
  3. `install-zharness.sh`: detect OS/arch → `gh release download` → install to `~/.local/bin` → verify `zharness --version`
- expected outputs:
  - Tagged pre-release `cli/v0.1.0` with 4 binaries; script installs on this machine
- verification:
  - `bash scripts/install-zharness.sh && zharness --version` on a clean PATH; `gh release view cli/v0.1.0` lists 4 assets
- stop if:
  - gh auth insufficient to create releases
- escalate to:
  - user clarification

## Risks / Watch-fors
- Port behavior, not Rust structure — resist line-porting unused upstream code
- Keep the full schema from T2 even though only core commands use it yet; Phase 4 must not need migrations rework
