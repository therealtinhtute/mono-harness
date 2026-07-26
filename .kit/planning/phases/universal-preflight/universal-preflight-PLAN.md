# Plan: universal preflight

Phase: universal-preflight
Status: ready
Wave Count: 3
Execution Owner: work
Updated At: 2026-07-26

## Goal
Add one pure routing contract and make all workflow skills consume it while current behavior/layout stays compatible.

## Inputs
- `.kit/planning/SPEC.md`
- `universal-preflight-CONTEXT.md`
- current `next`, `resume`, root command wiring, and eight SKILL files

## Wave 1
### T1 — Define and test the preflight matrix
- type: implementation + test
- inputs:
  - current stage/mode behavior in embedded playbooks
- touches:
  - `cli/internal/domain/preflight.go`
  - `cli/internal/application/preflight.go`
  - corresponding `_test.go` files
- avoid:
  - filesystem writes, DB writes, schema changes
- steps:
  1. Define valid stages, modes, readiness, docs/DB status, stop codes, and recovery strings.
  2. Implement a pure application function receiving observed DB/docs state.
  3. Add table-driven cases for all stage/mode combinations and missing/unreadable state.
  4. Assert reduced and blocked routing exactly matches the SPEC matrix.
- expected outputs:
  - deterministic pure preflight result
- verification:
  - `cd cli && go test ./internal/domain ./internal/application -run Preflight`
- stop if:
  - routing requires a new product decision not present in SPEC
- escalate to:
  - brainstorm refine

## Wave 2
### T2 — Expose `zharness preflight`
- type: implementation + test
- inputs:
  - Wave 1 application contract
- touches:
  - `cli/internal/interfaces/preflight.go`
  - `cli/internal/interfaces/root.go`
  - interface tests
  - `cli/docs/CONTRACT.md`
- avoid:
  - `next` semantics, DB path constants, migrations
- steps:
  1. Add Cobra command and flags.
  2. Observe current DB/docs state without opening a writable transaction or creating paths.
  3. Emit stable JSON and concise human output.
  4. Test valid stages, invalid stage/mode, reduced, blocked, and zero-write behavior.
- expected outputs:
  - live CLI command wired into help
- verification:
  - `cd cli && go test ./internal/interfaces -run Preflight && go run ./cmd/zharness preflight work --mode simple --json`
- stop if:
  - the command creates `.kit`, docs, changesets, or DB state
- escalate to:
  - check

## Wave 3
### T3 — Convert all workflow skills to common preflight
- type: docs + integration test
- inputs:
  - Wave 2 command contract
- touches:
  - `skills/workflow/{brainstorm,to-plan,work,check,handoff,watzup,git,interview}/SKILL.md`
  - `skills/workflow/README.md`
  - skill validation tests/scripts only if already present
- avoid:
  - workflow behavior unrelated to readiness routing
  - CI workflows
- steps:
  1. Add version gate + preflight to all eight active skills.
  2. Make spine skills read the returned playbook path instead of a hardcoded path.
  3. Keep git/interview behavior intact after preflight.
  4. Remove the generated-only `skills/workflow/plan/` directory with `trash` if it still contains no skill.
  5. Validate every skill and run the full Go suite.
- expected outputs:
  - one readiness policy across all skills
- verification:
  - `for d in skills/workflow/{brainstorm,to-plan,work,check,handoff,watzup,git,interview}; do bash scripts/validate-skill.sh "$d"; done && cd cli && go test ./...`
- stop if:
  - any skill requires changing its user-facing purpose
- escalate to:
  - to-plan phase universal-preflight

## Risks / Watch-fors
- Global skill copies under `~/.claude/skills` are not source files for this phase; edit repository source only.
- Human output must not become a second routing contract separate from JSON.
