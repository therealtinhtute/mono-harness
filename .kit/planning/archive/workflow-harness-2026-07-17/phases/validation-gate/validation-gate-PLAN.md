# Plan: validation-gate

Phase: validation-gate
Status: ready
Wave Count: 3
Execution Owner: work
Updated At: 2026-07-17

## Goal
Research commands implemented; check rewritten to gate on the matrix; determinism proven by fixtures.

## Inputs
- Phase 4/5 outputs (domain commands, adapter-produced artifacts)
- `cli/docs/CONTRACT.md` research schemas; upstream TRACE_SPEC.md, TEST_MATRIX.md

## Wave 1
### T1 — Implement score-trace + score-context
- type: implementation
- inputs:
  - Upstream tier definitions; recorded traces from Phase 5 sample
- touches:
  - `cli/internal/**`, `cli/testdata/**`
- avoid:
  - inventing new tiers; skill files
- steps:
  1. Port trace quality tiers; `score-trace <id> --json` returns tier + reasons
  2. Port context-read scoring; `score-context <trace-id> --json`
- expected outputs:
  - Both commands per CONTRACT.md
- verification:
  - `cd cli && go test ./... -run 'TestScore(Trace|Context)'` with golden outputs
- stop if:
  - tier definitions don't fit trace shape recorded by Phase 4
- escalate to:
  - to-plan phase harness-contracts

### T2 — Implement audit + propose
- type: implementation
- inputs:
  - STATE.md drift rules; validate findings shape (Phase 4)
- touches:
  - `cli/internal/**`, `cli/testdata/**`
- avoid:
  - duplicating validate logic — audit composes it and adds scoring
- steps:
  1. `audit --json`: pointer drift + contract violations + unlinked proofs + entropy score, stable ordering
  2. `propose --json`: improvement proposals from observed patterns (port upstream behavior, mark reserved)
- expected outputs:
  - Gate-consumable audit report; reserved propose command
- verification:
  - `cd cli && go test ./... -run TestAudit` — staled-pointer fixture changes the score and lists the finding; same input twice = identical output
- stop if:
  - entropy scoring needs data the schema doesn't capture
- escalate to:
  - to-plan phase harness-contracts

## Wave 2
### T3 — Rewrite check skill with validation matrix
- type: implementation
- inputs:
  - T1, T2; existing check references
- touches:
  - `skills/workflow/check/SKILL.md`, `references/gate-checklist.md`, `references/artifact-alignment.md`
- avoid:
  - dropping existing review dimensions — map them into the matrix
- steps:
  1. Add gate block (same standard text as Phase 5)
  2. Write matrix table (lane × proof class, every cell required/optional/n-a) into gate-checklist.md
  3. Rewrite gate flow: `zharness audit` + `score-trace` inline → matrix evaluation → `zharness check record --verdict … --json`
  4. Document fail conditions + escalation (human override via `zharness intervention`)
- expected outputs:
  - Deterministic, CLI-first check skill
- verification:
  - `grep -n 'check record\|audit' skills/workflow/check/SKILL.md`; matrix table has zero empty cells (inspection)
- stop if:
  - a dimension resists required/optional classification
- escalate to:
  - brainstorm refine

## Wave 3
### T4 — Determinism fixtures
- type: test
- inputs:
  - T3; Phase 5 sample project
- touches:
  - `cli/testdata/gate-missing-proof/`, scratch project
- avoid:
  - fixtures that pass by luck — pin exact expected verdict JSON
- steps:
  1. Fixture: sample chain minus one required proof → run gate flow → expect FAIL naming that proof
  2. Re-run gate on identical inputs → identical verdict (byte-stable JSON)
- expected outputs:
  - CI-runnable determinism proof
- verification:
  - `cd cli && go test ./... -run TestGateDeterminism`
- stop if:
  - verdict varies across runs (ordering/timestamp leak)
- escalate to:
  - check

## Risks / Watch-fors
- Timestamps and map ordering are the classic determinism leaks — canonicalize JSON output
- Keep audit and validate complementary, not overlapping — validate = contracts, audit = drift + scoring
