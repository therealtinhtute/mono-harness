# Context: single-source-playbooks

Phase: single-source-playbooks
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: low
Expected Proof: unit, command-output, manual-check

## Goal
Make the Go embed (`cli/docs/embedded/playbooks/`) the single source of truth for playbooks and `.kit/docs/playbooks/` a pure projection of it, enforced by a test that fails on divergence. Removes the standing two-copies-must-stay-in-sync liability (audit #2; drift bug #24 already happened).

## Scope Boundary
### Allowed Surfaces
- `cli/internal/application/init.go` — projection behavior (already scaffolds; make the contract explicit)
- `cli/internal/embedded/` — embed loading
- a new test (or `zharness playbooks verify` command) comparing scaffold vs embed
- `cli/docs/CONTRACT.md`, root `README.md`, `docs/workflow-harness/*` — document "edit the embed, never `.kit/docs/`"

### Forbidden Surfaces
- playbook *content* (Phases 1 & 3 finalize it; this phase only locks the projection + guard)
- changeset format, schema, scoring, entities

## Spec Hooks
- Requirement R4 (single-source playbooks)
- Acceptance: a test asserts `.kit/docs/playbooks/*` == embed byte-for-byte.
- Depends on Phases 1 & 3 having finalized the embed text.

## Locked Decisions
- **Embed is canonical; `.kit/docs/` is generated output**, never hand-edited. The test compares the on-disk scaffold to the embed and fails on any diff.
- Prefer a **Go test** (`embedded_test.go` already exists as a home) over a new user-facing command, to keep the CLI surface from growing while we're subtracting. A `zharness playbooks verify` is optional if a runtime check is wanted too.

## Assumptions
- `zharness init --refresh-docs` already rewrites `.kit/docs/` from the embed (confirmed in migration.md) — this phase adds the *guard*, not the projection itself.
- The six playbooks are the only generated docs under `.kit/docs/playbooks/`.

## Canonical Refs
- `cli/internal/embedded/embedded_test.go` (existing embed test home)
- `cli/internal/application/init.go` (scaffolding)
- `docs/workflow-harness/migration.md` (`.kit/docs/` is a generated scaffold; issue #24 drift precedent)

## Rejected Options
- Committing `.kit/docs/` and treating it as canonical — invites the exact drift this phase removes.
- A git pre-commit hook instead of a test — weaker (bypassable, not in CI); a Go test travels with the module.

## Deferred Ideas
- Shrinking the playbooks to target lengths (audit #5) — separate initiative; this phase only locks the source-of-truth, not the length.

## Escalate If
- Phases 1 or 3 aren't finalized when this starts (embed still changing) → wait / route back; locking a projection over unfinished text is pointless.
