---
id: 01KYF6VS9TCHEZH41X91ZKH9DK
type: check
phase: universal-preflight
lane: high-risk
mode: full
run_id: 01KYF4Z95MQJG7PEZPF00JKC9H
proof_links:
  - command: go -C cli test ./...
    output_ref: inline — full suite passed
    artifact_path: .kit/runs/work/20260726-1204-universal-preflight.md
  - command: go -C cli vet ./... && go -C cli build ./...
    output_ref: inline — exit 0
    artifact_path: .kit/runs/work/20260726-1204-universal-preflight.md
  - command: phase binary preflight SHA/count comparison
    output_ref: inline — DB SHA unchanged; changesets=52
    artifact_path: .kit/runs/work/20260726-1204-universal-preflight.md
created: 2026-07-26
updated: 2026-07-26
---

# CHECK REPORT

Run ID: 01KYF6VS9TCHEZH41X91ZKH9DK
Scope: full
Artifact Alignment: drift
Review Verdict: REQUEST CHANGES
Phase: universal-preflight
Spec: .kit/planning/SPEC.md
Plan: .kit/planning/phases/universal-preflight/universal-preflight-PLAN.md
Cook Run: .kit/runs/work/20260726-1204-universal-preflight.md
Created At: 2026-07-26 12:36 UTC

## Gate Evidence
- tests: `go -C cli test ./...` → pass
- types: Go compiler through test/build → pass
- lint: `go -C cli vet ./...` → pass
- build: `go -C cli build ./...` → pass
- command-output: live phase binary returned ready on this repo; DB SHA and changeset count remained unchanged
- audit: historical contract violations remain; current out-of-order pointer is expected until this gate is recorded

## Artifact Alignment
- status: drift
- notes:
  - implementation maps to universal-preflight scope and no forbidden schema/layout behavior was added
  - zero-write proof is present
  - two routing edge cases contradict the locked reduced/durable contract and block approval
  - generic skill validator mismatch is pre-existing and does not block the phase-specific thin-trigger checks

## Findings
### Critical
- none

### Major
- `work --mode auto` resolves to reduced unconditionally. When a durable plan exists but the DB is missing, preflight permits reduced continuation instead of returning `harness_required`; this weakens the mandatory durable guard.
- When managed docs are missing, preflight still returns `.kit/docs/playbooks/{stage}.md`. Reduced-mode skills can attempt to read a path that preflight already knows does not exist rather than falling back to repository-native guidance.

### Minor / Suggestions
- `OpenReadOnly` is a justified one-file surface deformation because the existing writable opener enables WAL; keep the rationale in the run evidence.

## Next Action
- rerun `work` on universal-preflight to resolve both major routing findings, then repeat the focused/full gates
