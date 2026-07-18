# Plan: cli-stale-drift

Phase: cli-stale-drift
Status: ready
Wave Count: 2
Execution Owner: work
Updated At: 2026-07-18

## Goal
Ship `stale_docs` drift detection in `resume` with its named recovery, prove clearing semantics, and land the scratch-dir lifecycle integration suite.

## Inputs
- cli-embed-scaffold output: docs_version in meta, `--refresh-docs` working
- `cli/internal/interfaces/resume.go`, `cli/docs/STATE.md`

## Wave 1
### T1 — stale_docs drift in resume
- type: implementation
- inputs: docs_version storage (Phase 2)
- touches: `cli/internal/interfaces/resume.go` + application layer, one Go constant for the recovery string
- avoid: any mutation from resume; new readiness states
- steps:
  1. Read meta.docs_version during resume; apply the firing rule (exists ∧ differs ∧ neither side `dev`)
  2. Append drift entry `{type: "stale_docs", detail: "<written> vs <cli>", recovery: <constant>}`; readiness follows the existing non-empty-drift rule
  3. Missing stamp → informational `docs: unversioned` field (or omit), never drift
- expected outputs: resume reporting stale_docs correctly in all four cases (match, differ, dev, missing)
- verification: unit tests per case: `cd cli && go test ./...`
- stop if: drift plumbing requires structural changes to the drift array shape
- escalate to: to-plan phase

### T2 — STATE.md drift-table row
- type: docs
- inputs: T1 constant
- touches: `cli/docs/STATE.md`
- avoid: paraphrasing the recovery string
- steps:
  1. Add the `stale_docs` row to the stale-pointer rules table, quoting the Go constant verbatim and noting the single-source rule
- expected outputs: STATE.md row matching code exactly
- verification: `grep` the recovery string in both files — byte-identical
- stop if: —
- escalate to: check

## Wave 2
### T3 — Clearing-semantics test
- type: test
- inputs: T1, `--refresh-docs` (Phase 2)
- touches: Go tests
- avoid: production changes except bugs found
- steps:
  1. Test: project stamped with an older docs_version → resume fires stale_docs → run `init --refresh-docs` → resume clean; assert only `.kit/docs/**` + stamp changed (stories/runs/meta pointers byte-stable)
- expected outputs: proven acceptance criterion (SPEC bullet 2)
- verification: `cd cli && go test ./...`
- stop if: refresh touches state beyond docs + stamp
- escalate to: check (finding against Phase 2)

### T4 — Scratch-dir lifecycle integration suite
- type: test
- inputs: full CLI
- touches: Go integration tests (build-tagged if slow)
- avoid: network, release assets
- steps:
  1. Scripted pass on a temp dir: `init` → `intake` → `story` → run-registration changeset → `trace add` → `check record` → `handoff record` → `resume --json`
  2. Assert: readiness transitions (clean → in-progress → checked), zero drift, `validate --json` and `audit --json` clean on the produced chain
- expected outputs: the SPEC's integration validation expectation, reusable by cli-release as a smoke test
- verification: `cd cli && go test ./...` (suite green)
- stop if: any lifecycle command's real behavior contradicts its playbook description from Phase 1
- escalate to: check (playbook fix routed back — doc bug, not code bug, unless STATE.md disagrees too)

## Risks / Watch-fors
- The `dev`-never-fires rule must be tested explicitly — this repo dogfoods dev builds and silent drift spam would poison every watzup recap
- T4 doubles as the release smoke: keep it runnable against an installed binary, not only `go test`
