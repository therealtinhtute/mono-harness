# Context: harness-contracts

Phase: harness-contracts
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: medium
Expected Proof: docs inspection + template lint

## Goal
Freeze all machine contracts before Go code: state semantics, artifact frontmatter, 19 command schemas, SQLite schema, changeset format.

## Scope Boundary
### Allowed Surfaces
- `cli/docs/CONTRACT.md`, `cli/docs/SCHEMA.md`, `cli/docs/STATE.md` (new; creating `cli/docs/` is fine, no Go files)
- `skills/workflow/brainstorm/references/spec-template.md`
- `skills/workflow/work/references/run-artifact-template.md`
- `skills/workflow/check/references/report-template.md`
- `skills/workflow/handoff/references/handoff-template.md`

### Forbidden Surfaces
- Any `SKILL.md` (behavior changes are Phases 5–7)
- Go source files; `go.mod`
- `to-plan/references/workflow-state-template.yml` (removed in Phase 8, after pilot)

## Spec Hooks
- R4 (contracts live in cli/docs/), R6 (19 commands, --json), R7–R9 (changeset/ULID invariants), R10 (legacy mapping), R12–R16 (state, schema, frontmatter, cross-links)
- Constraint: append-only changesets — no contract may require editing a past changeset

## Locked Decisions
- Frontmatter required keys: `id` (ULID), `type`, `phase` (story slug), `lane`, `created`, `updated`
- Cross-links: PLAN→SPEC id; RUN→PLAN id + trace ids; CHECK→RUN id + proof links; HANDOFF→latest RUN/CHECK ids
- Phase status enum: `planned | in-progress | checked | done`
- Changeset line shape: one JSON object per line — `{op, entity, id, fields, at}`; file `{ulid}.changeset.jsonl`
- Error codes in CONTRACT.md: numeric exit codes + stable machine-readable `code` strings in JSON errors
- Story↔phase mapping confirmed by Phase 1 (`docs/workflow-harness/gap-matrix.md`): one `zharness` story per `to-plan` phase, story slug = phase slug. No longer an open assumption — the `phase` frontmatter key above is this story slug.

## Assumptions
- Upstream command semantics (intake lanes, decision/backlog fields, trace shape) port as-is unless CONTRACT.md notes a deviation

## Canonical Refs
- `.kit/planning/SPEC.md`
- `docs/workflow-harness/gap-matrix.md` (Phase 1 output)
- `~/Lab/harness-experimental/crates/harness-cli/src/interface.rs` (command args ground truth)
- `~/Lab/harness-experimental/docs/TRACE_SPEC.md`, `docs/TEST_MATRIX.md`

## Rejected Options
- JSON Schema files instead of markdown contracts — heavier to maintain solo; markdown + fixtures in CI gives equivalent enforcement (SPEC decision: contracts in cli/docs/)
- Splitting state contract into docs/contracts/ — rejected in SPEC (single home with the code)

## Deferred Ideas
- Versioned contract negotiation (`zharness` serving multiple contract versions) — schema_version field is enough for v1

## Escalate If
- A ported command's semantics cannot be determined from upstream source → user clarification
- Frontmatter keys conflict with existing template fields in a way that breaks current users mid-migration → brainstorm refine
