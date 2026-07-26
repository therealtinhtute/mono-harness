# Context: root layout

Phase: root-layout
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: high
Expected Proof: unit, integration, migration rehearsal, command-output, manual-check

## Goal
Activate one root `harness.db`, keep `.kit/changesets`, and project managed workflow docs to root safely while retaining the existing artifact chain until Phase 3.

## Scope Boundary
### Allowed Surfaces
- DB/docs path constants in interfaces/application
- `cli/internal/infrastructure/migrations.go` and changeset allowlists
- init, embedded projection, marked-block and managed-doc logic
- migrate command and tests
- `.gitignore`, root `AGENTS.md`, root docs projection
- CLI/state/schema/migration documentation

### Forbidden Surfaces
- lifecycle artifact model or deletion
- typed lifecycle table removal
- changeset rewriting/compaction
- `.harness/` creation
- CI workflows

## Spec Hooks
- Requirements 1, 8, 9, and constraints on replay/rollback.

## Locked Decisions
- Root `harness.db` is the only final DB.
- `.kit/changesets` remains the replay source.
- Managed doc hashes live in a DB table, not a manifest file.
- Migration uses temporary replay + parity + atomic activation; never copies the old DB as truth.
- AGENTS updates only a marked block.

## Assumptions
- Temporary migration files are allowed during execution but are removed/ignored after activation.
- Legacy lifecycle markdown remains in place in this phase.

## Canonical Refs
- `cli/internal/interfaces/init.go`
- `cli/internal/interfaces/db.go`
- `cli/internal/application/init.go`
- `cli/internal/infrastructure/{migrations,changeset}.go`
- `cli/internal/embedded/projection_drift_test.go`
- `cli/docs/{CONTRACT,SCHEMA,STATE}.md`

## Rejected Options
- A second core-state DB: violates the one-DB constraint.
- Force-overwrite docs: can destroy consumer authority.
- Copy old DB to root: does not prove replay continuity.

## Deferred Ideas
- Changeset compaction and self-update machinery.
- Lifecycle artifact consolidation belongs to Phase 3.

## Escalate If
- Existing changesets cannot replay into the additive schema.
- Normalized state cannot be made equal without rewriting historical changesets.
