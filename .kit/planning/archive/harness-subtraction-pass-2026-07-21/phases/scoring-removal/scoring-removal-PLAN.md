# Plan: scoring-removal

Phase: scoring-removal
Status: ready
Wave Count: 2
Execution Owner: work
Updated At: 2026-07-21

## Goal
Remove `ScoreTrace`/`score-trace` + `entropy_score`; keep the lane×proof matrix; prove the gate outcome is unchanged.

## Inputs
- `score.go`, `audit.go`, `interfaces/score.go`
- `check.md` embed (Step 4 loop + Validation Matrix)

## Wave 1
### T1 — Remove scoring from the CLI
- type: refactor
- inputs:
  - `cli/internal/application/score.go`, `audit.go`, `cli/internal/interfaces/score.go`
- touches:
  - those three files + their tests
- avoid:
  - matrix logic (lives in the playbook, not the CLI), changeset format, dropped entities
- steps:
  1. Delete `ScoreTrace` + the `score-trace` cobra command; delete `TraceScore` if now unused.
  2. Remove `entropyScore` and the `EntropyScore` field from `AuditReport` in `audit.go`.
  3. Update/delete `score_test.go` and any audit test asserting `entropy_score`.
- expected outputs:
  - `zharness --help` has no `score-trace`; `audit --json` output has no `entropy_score` key
- verification:
  - `cd cli && go build ./... && go test ./internal/application/ -run Audit` → pass; `zharness audit --json | grep entropy_score` → empty
- stop if:
  - removing `TraceScore` breaks an unrelated caller
- escalate to:
  - to-plan phase scoring-removal

## Wave 2
### T2 — Update check playbook + contract docs, prove gate unchanged
- type: docs
- inputs:
  - `cli/docs/embedded/playbooks/check.md` (Step 4), `CONTRACT.md`, `SCHEMA.md`
- touches:
  - those docs; re-scaffold `.kit/docs/`
- avoid:
  - the Validation Matrix section (keep it verbatim — it is the verdict)
- steps:
  1. In `check.md` Step 4, delete the `score-trace` loop bullet; keep the matrix evaluation and `check record`.
  2. Update `CONTRACT.md`/`SCHEMA.md`: remove `score-trace`; update `audit --json` shape (no `entropy_score`).
  3. Re-scaffold `.kit/docs/` from the embed.
  4. Prove the gate still works: run `check` (or a scripted matrix eval) on a fixture with a missing required proof cell → still FAILs naming the missing class.
- expected outputs:
  - `check.md` gates purely on the matrix; a missing required proof still FAILs
- verification:
  - `grep -c score-trace cli/docs/embedded/playbooks/check.md` → 0; manual: run a high-risk fixture missing `integration` proof → gate FAIL (capture output)
- stop if:
  - the matrix stops FAILing correctly without score-trace
- escalate to:
  - to-plan phase scoring-removal

## Risks / Watch-fors
- The invariant "gate pass/fail unchanged" is the whole point — the T2 fixture proof is mandatory.
- Do Phase 2 (dead-surface) first: both edit `score.go`/`audit.go`; sequential avoids conflicts.
