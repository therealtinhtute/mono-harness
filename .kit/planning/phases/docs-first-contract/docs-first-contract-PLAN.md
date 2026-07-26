# Plan: docs first contract

Phase: docs-first-contract
Status: ready
Wave Count: 3
Execution Owner: work
Updated At: 2026-07-26

## Goal
Finalize concise docs-first guidance, remove stale narrative, and prove the complete v3 harness against its public contract.

## Inputs
- completed one-plan-lifecycle phase
- docs-first-contract-CONTEXT.md
- final Cobra help and root layout

## Wave 1
### T1 — Rewrite entrypoint and workflow authority
- type: docs + test
- touches:
  - embedded/root `AGENTS.md`
  - embedded/root `docs/WORKFLOW.md`
  - projection and word-budget tests
- avoid:
  - new workflow capabilities
- steps:
  1. Write compact repository map and source hierarchy.
  2. Encode read-only, bounded, durable, and human-judgment shapes.
  3. Encode authority-before-edit and honest proof rules.
  4. State the local CLI/DB divergence explicitly.
  5. Enforce combined ≤1,000-word budget.
- expected outputs:
  - small stable entrypoint and canonical workflow map
- verification:
  - `cd cli && go test ./internal/embedded ./internal/application -run 'Projection|Word|Agents|Workflow'`
- stop if:
  - a required global rule exists only in a stage playbook
- escalate to:
  - check

## Wave 2
### T2 — Slim stage playbooks and public contracts
- type: docs + contract test
- touches:
  - six embedded/root playbooks
  - `skills/workflow/README.md`
  - `cli/docs/{CONTRACT,SCHEMA,STATE}.md`
  - root README/CLAUDE path references where required
- avoid:
  - CI workflows and new commands
- steps:
  1. Remove duplicated global rules from playbooks.
  2. Keep exact stage commands, artifacts, stops, and exits only.
  3. Remove dead command/path/artifact references.
  4. Reframe DB as ledger, not authority.
  5. Cross-check every documented command against live help.
- expected outputs:
  - progressive disclosure with no stale command contract
- verification:
  - `cd cli && go test ./... && go run ./cmd/zharness --help`
- stop if:
  - docs require a command not delivered by Phases 1–3
- escalate to:
  - to-plan phase docs-first-contract

## Wave 3
### T3 — Remove stale docs and run final acceptance
- type: cleanup + verification
- touches:
  - `docs/SKILLS_HARNESS_ARCHITECTURE.md`
  - `docs/workflow-state-dogfood.md`
  - final repository docs and evidence
- avoid:
  - `docs/workflow-harness/` provenance
  - CI workflows
- steps:
  1. Read and confirm both stale docs remain superseded.
  2. Remove them with `trash`.
  3. Run full Go tests, skill validation, live preflight/init/migrate/resume/query/validate smoke checks, and projection checks.
  4. Inspect final tree for `.harness`, duplicate DB, legacy artifact paths, dead commands, and untracked managed output.
  5. Run the phase `check` gate.
- expected outputs:
  - v3 acceptance evidence and no stale parallel model
- verification:
  - `cd cli && go test ./... && for d in ../skills/workflow/{brainstorm,to-plan,work,check,handoff,watzup,git,interview}; do bash ../scripts/validate-skill.sh "$d"; done`
- stop if:
  - any acceptance criterion lacks observable evidence
- escalate to:
  - check

## Risks / Watch-fors
- Keep provenance docs even when their old paths are historical.
- Do not claim application operability without a consumer-owned runbook and real commands.
