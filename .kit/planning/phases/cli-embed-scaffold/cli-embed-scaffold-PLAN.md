# Plan: cli-embed-scaffold

Phase: cli-embed-scaffold
Status: ready
Wave Count: 3
Execution Owner: work
Updated At: 2026-07-18

## Goal
Embed the doc set in the binary; make `init` scaffold `.kit/` + docs + shim + `.gitignore`; stamp `docs_version`; implement `--refresh-docs`.

## Inputs
- `cli/docs/embedded/**` (Phase 1 output, checked complete)
- `cli/docs/SCHEMA.md` (meta storage shape), `cli/internal/interfaces/init.go`, `cli/internal/application/**`

## Wave 1
### T1 — Embed package with integrity tests
- type: implementation
- inputs: `cli/docs/embedded/**`
- touches: new `cli/internal/embedded/` (or equivalent) package, `go:embed` directives
- avoid: init command changes (T3), any doc content rewrites
- steps:
  1. Create the embed package exposing the doc set as an fs.FS + a manifest (path list + docs version)
  2. Unit tests: every manifest path present and non-empty; playbook count == 6; shim/CONTEXT_RULES/AUTHORITY present; docs version exposed
- expected outputs: embed package + passing tests
- verification: `cd cli && go test ./internal/embedded/...`
- stop if: embed of the directory fails due to path constraints
- escalate to: to-plan phase

### T2 — Decide + implement docs_version storage
- type: implementation
- inputs: `cli/docs/SCHEMA.md`, existing meta read/write code
- touches: meta layer, `cli/docs/SCHEMA.md` + `cli/docs/STATE.md` docs
- avoid: breaking schema changes without a versioned migration
- steps:
  1. Read the meta storage shape; if key-value, add `docs_version` as a key; else add a versioned migration
  2. Wire read/write through the existing changeset-first mutation path
  3. Document the field in SCHEMA.md/STATE.md
- expected outputs: docs_version readable/writable via the application layer
- verification: `cd cli && go test ./...`; manual: apply a changeset setting docs_version, `zharness query state --json` (or equivalent) reflects it
- stop if: requires editing any committed changeset or breaking replay
- escalate to: to-plan phase

## Wave 2
### T3 — init scaffolding + idempotency matrix
- type: implementation
- inputs: T1 embed package, T2 storage
- touches: `cli/internal/interfaces/init.go`, application layer
- avoid: resume/drift logic
- steps:
  1. `init`: create `.kit/` if missing (kills the `db_not_writable` footgun), create db as today, write `.kit/docs/**` from the embed FS, write root `AGENTS.md` only if absent (notice otherwise), ensure `.gitignore` entries (`.kit/harness.db`, `.kit/cache/`) by append
  2. Stamp docs_version via changeset in the same init flow
  3. Implement the idempotency matrix from CONTEXT (fresh / db-only / docs-present)
- expected outputs: one-command scaffold
- verification: unit tests per matrix cell; manual scratch dir: `zharness init --json` twice — second run reports exists, no file churn (`git status` clean in a git-inited scratch)
- stop if: any matrix cell would overwrite user content
- escalate to: user clarification

### T4 — `init --refresh-docs`
- type: implementation
- inputs: T3
- touches: init command flags + application layer
- avoid: touching db state other than docs_version
- steps:
  1. `--refresh-docs`: rewrite `.kit/docs/**` from embed FS, update docs_version stamp, leave stories/runs/checks/meta pointers untouched
  2. Unit test: refresh on a project with modified docs restores canonical content and bumps stamp; `resume` output otherwise unchanged
- expected outputs: working refresh path (the recovery cli-stale-drift will name)
- verification: `cd cli && go test ./...`
- stop if: refresh semantics need to merge user edits (it must not — canonical overwrite is the contract)
- escalate to: brainstorm refine

## Wave 3
### T5 — Scratch-dir integration + docs sync check
- type: test
- inputs: T1–T4
- touches: Go integration tests (or a `scripts/` test shell), no production code except bug fixes
- avoid: release pipeline
- steps:
  1. Integration test: fresh scratch dir → `init` → assert `.kit/docs/` tree, AGENTS.md, .gitignore entries, docs_version stamped, `resume --json` readiness clean
  2. Assert embedded manifest matches `cli/docs/embedded/**` on disk (guard against a doc added to the repo but not the manifest)
- expected outputs: green integration suite
- verification: `cd cli && go test ./...` + run the scratch scenario via test
- stop if: init behavior differs across the matrix from T3's spec
- escalate to: check

## Risks / Watch-fors
- `.gitignore` append logic must be conservative — never reorder/rewrite existing content
- AGENTS.md at repo root is shared surface with consumer projects — the never-overwrite rule is load-bearing (public-contract risk flag)
