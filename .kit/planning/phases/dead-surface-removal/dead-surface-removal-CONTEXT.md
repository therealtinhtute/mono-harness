# Context: dead-surface-removal

Phase: dead-surface-removal
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: high
Expected Proof: unit, integration, command-output, manual-check

## Goal
Remove five built-but-unused surfaces — `decision`, `backlog`, `tool` (entities + commands + tables) and `propose`, `score-context` (commands) — plus their tests, and drop the unused tables from the schema with a `schema_version` bump. Zero behavior change for anything a playbook actually uses.

## Scope Boundary
### Allowed Surfaces
- `cli/internal/interfaces/` — remove the five subcommands' wiring
- `cli/internal/application/` — remove `Propose`, `ScoreContext`, and decision/backlog/tool logic
- `cli/internal/infrastructure/changeset.go` — remove `decision`/`backlog`/`tool` from `entityTables`/`entityColumns`
- `cli/internal/infrastructure/migrations.go` — drop the three tables; bump `schema_version` (2 → 3)
- `cli/docs/CONTRACT.md`, `SCHEMA.md` — remove the deleted command/entity docs
- corresponding `*_test.go` deletions

### Forbidden Surfaces
- `score.go`'s `ScoreTrace` and `audit.go`'s `entropyScore` (Phase 3 owns those)
- changeset on-disk format, ULID/fence/replay engine
- the playbooks (no playbook references these — verify, then leave untouched)

## Spec Hooks
- Requirement R2 (delete dead surface)
- Constraint: existing committed changesets MUST still replay after the schema change; `import` of a legacy `.kit/` must still work.
- Acceptance: `zharness --help` no longer lists the five; build + tests green.

## Locked Decisions
- **Drop the unused tables** (not just the commands) for a true subtraction — chosen over "leave tables, remove commands" because dead tables are exactly the mental-skip weight the audit flagged. Guarded by a replay-safety test.
- `schema_version` 2 → 3; `migrations.go` gets a forward migration that drops `decisions`, `backlog`, `tools` (safe: no rows in practice, no changeset references them).

## Assumptions
- **Re-verify at execution start**: `grep -rn` for `decision|backlog|tool|propose|score-context` across `.kit/changesets/**` and all playbooks returns no *consumer* (only the code being deleted). If any real consumer appears, STOP.
- No external script or release step calls these commands (install/goreleaser unaffected).

## Canonical Refs
- `cli/internal/infrastructure/changeset.go` (`entityTables`, `entityColumns`)
- `cli/internal/infrastructure/migrations.go` (CREATE TABLE list, schema_version)
- `cli/internal/application/{audit.go (Propose), score.go (ScoreContext)}`
- `cli/internal/interfaces/*` command registrations (from `Use:` grep)

## Rejected Options
- Keeping the commands "in case they're needed later" — YAGNI; they've been unconsumed since inception and are recoverable from git.
- Leaving tables, deleting only commands — half-measure that keeps schema weight.

## Deferred Ideas
- Reviving `decisions` as the unified memory store (audit #4) — a *later* initiative may re-add a decision entity with real consumers; that is not this deletion's concern.

## Escalate If
- A grep finds any changeset or playbook that actually references a deleted entity/command → STOP, route to `brainstorm refine` (scope was wrong).
- Dropping a table breaks replay of committed changesets or legacy `import` → STOP, route to `to-plan phase dead-surface-removal` (keep tables, delete commands only).
