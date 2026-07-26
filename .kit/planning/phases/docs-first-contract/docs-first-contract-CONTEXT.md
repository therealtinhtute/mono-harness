# Context: docs first contract

Phase: docs-first-contract
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: medium
Expected Proof: docs contract, unit, integration, command-output, manual-check

## Goal
Make the final repository-facing guidance match upstream’s docs-first authority and work-shape model while preserving the approved mandatory CLI preflight divergence.

## Scope Boundary
### Allowed Surfaces
- root `AGENTS.md`, `docs/WORKFLOW.md`, `docs/playbooks/`, plan/decision templates
- `skills/workflow/README.md`, root README/CLAUDE references
- `cli/docs/{CONTRACT,SCHEMA,STATE}.md`, embedded docs, projection/contract tests
- stale architecture/dogfood docs approved for removal

### Forbidden Surfaces
- new runtime features or schema changes
- CI workflows
- compaction/updater/service additions
- invented application commands or runbook

## Spec Hooks
- Requirements 2, 11, 12, and 14.

## Locked Decisions
- Repository truth outranks DB lifecycle commentary.
- Work shape and human judgment are independent decisions.
- CLI preflight is always required; DB is required only for durable mutation.
- AGENTS block + WORKFLOW ≤1,000 words.
- Stage playbooks contain only stage-specific procedure.

## Assumptions
- Phase 3 has already established final paths and one-plan commands.

## Canonical Refs
- upstream `AGENTS.md`, `docs/WORKFLOW.md`, `docs/HARNESS.md`
- root `AGENTS.md`
- `skills/workflow/README.md`
- final CLI help output

## Rejected Options
- Exact upstream optional CLI: weakens the requested guard.
- Copy upstream compatibility docs wholesale: imports unused surface.
- Add generic runbook/improve skill: speculative files and scope.

## Deferred Ideas
- CI enforcement and consumer-specific runbooks.

## Escalate If
- A documented command is absent from the final Cobra tree.
- The context budget cannot be met without removing a required authority rule.
