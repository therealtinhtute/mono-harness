# Plan: cli-domain

Phase: cli-domain
Status: ready
Wave Count: 4
Execution Owner: work
Updated At: 2026-07-17 (reopened — Wave 4/T5 handoff record added; a Wave 5/T6 run-create addition was drafted and then reverted, see cli-domain-CONTEXT.md's "false alarm" note — Phase 5 already solved that gap differently)

## Goal
intake/story/decision/backlog/tool/intervention/trace + resume/check-record/validate implemented, changeset-first, fixture-proven.

## Inputs
- Phase 3 binary + changeset engine
- `cli/docs/CONTRACT.md` command schemas
- `cli/testdata/` (extend with sample-chain and validate fixtures)

## Wave 1
### T1 — Domain entities + validation rules
- type: implementation
- inputs:
  - SCHEMA.md entities; upstream domain.rs rules
- touches:
  - `cli/internal/domain/`
- avoid:
  - I/O in domain package (layering rule)
- steps:
  1. Port validation rules per entity (intake lane enum, story fields, trace event shape, check verdict enum) as pure functions with table-driven tests
  2. Add workflow entities: phase (status enum), run, check, handoff
- expected outputs:
  - Domain package with tests covering accept/reject per entity
- verification:
  - `cd cli && go test ./internal/domain/`
- stop if:
  - upstream rule contradicts CONTRACT.md
- escalate to:
  - user clarification

## Wave 2
### T2 — Ported domain commands
- type: implementation
- inputs:
  - T1 entities
- touches:
  - `cli/internal/{interfaces,application}/` (intake, story, decision, backlog, tool, intervention, trace)
- avoid:
  - research commands; undocumented flags
- steps:
  1. Implement each command: parse → domain validate → changeset append → DB apply → `--json` response
  2. Integration test per command asserting a changeset file was written before DB rows appear
- expected outputs:
  - 7 commands live per CONTRACT.md
- verification:
  - `cd cli && go test ./... -run TestCommand` and `zharness intake --type feature --summary x --lane tiny --json | jq .id` returns a ULID on a scratch project
- stop if:
  - a command needs schema changes
- escalate to:
  - to-plan phase harness-contracts

### T3 — resume + check record
- type: implementation
- inputs:
  - T1; STATE.md drift rules
- touches:
  - `cli/internal/{interfaces,application}/`
- avoid:
  - watzup/handoff skill edits (Phase 7)
- steps:
  1. `check record`: structured verdict + proof links → check entity + changeset
  2. `resume`: assemble snapshot (position, status, latest IDs, drift findings with `recovery`, readiness state)
  3. Drift detection per STATE.md: missing file, unknown slug, out-of-order IDs
- expected outputs:
  - Both commands emit CONTRACT.md-shaped JSON
- verification:
  - `cd cli && go test ./... -run 'Test(Resume|CheckRecord)'` — includes a staled-pointer fixture asserting a non-empty `recovery` per finding
- stop if:
  - a drift case has no defined recovery action in STATE.md
- escalate to:
  - to-plan phase harness-contracts

## Wave 3
### T4 — validate + chain fixtures
- type: implementation
- inputs:
  - T2, T3; Phase 2 frontmatter contracts
- touches:
  - `cli/internal/**` (validate), `cli/testdata/chain-valid/`, `cli/testdata/chain-broken/`
- avoid:
  - relaxing a contract to make a fixture pass
- steps:
  1. Implement `validate`: walk SPEC→PLAN→RUN→CHECK→HANDOFF by frontmatter ULIDs; check required keys, link targets, formats, freshness vs DB
  2. Build `chain-valid` fixture (complete linked chain) and `chain-broken` (one broken cross-link) under testdata
  3. Exit non-zero with machine-readable findings on violation
- expected outputs:
  - `zharness validate --json` passing/failing correctly on the two fixtures
- verification:
  - `cd cli && go test ./... -run TestValidate` — broken fixture exits 1 naming the broken link; valid fixture exits 0
- stop if:
  - frontmatter contract lacks a key validate needs
- escalate to:
  - to-plan phase harness-contracts

## Wave 4 (reopened 2026-07-17 — narrow addition, see cli-domain-CONTEXT.md's "Reopened" note)
### T5 — `handoff record` command
- type: implementation
- inputs:
  - existing `domain.Handoff`/`HandoffAnchors` (T1, already committed), `handoffs` table (already committed), `check_record.go`/`check.go` as the exact pattern to mirror
- touches:
  - `cli/internal/domain/handoff.go` (Validate() only — struct/table unchanged)
  - `cli/internal/application/handoff.go` (new, mirrors `check_record.go`)
  - `cli/internal/interfaces/handoff.go` (new, mirrors `check.go`)
  - `cli/internal/interfaces/root.go` (register `newHandoffCmd()`)
  - `cli/docs/CONTRACT.md` (add `handoff record` entry under "Domain — workflow additions"; resolve the top-of-file escalation note; update the Workflow-Step→CLI-Action mapping row for "handoff: close-out")
- avoid:
  - touching `check`/`audit`/`score`/`validate` code or any other already-gated command
  - relaxing SPEC R6's other 19 commands' shapes
- steps:
  1. `handoff record --run-id {ulid} --check-id {ulid} --open-items '["...","..."]'` — both IDs optional (anchors are `*string`), `open-items` defaults `[]`
  2. Validate: malformed `--open-items` JSON → `invalid_open_items`; non-empty `--run-id`/`--check-id` must reference existing rows → `unknown_run_id`/`unknown_check_id` (DB-lookup, same pattern as `trace add`'s optional `--run-id`)
  3. Changeset-first write via `AppendAndApply`, entity `handoff`, ULID id
  4. `--json`: `{"id": "ulid"}`
  5. Update CONTRACT.md's escalation note and mapping table row to point at the new command instead of describing the gap
- expected outputs:
  - `zharness handoff record --json` returns a ULID on a scratch db; `resume`'s `latest_handoff_id` picks it up unchanged (no `resume.go` edit needed — `latestHandoffID` already queries `handoffs`)
- verification:
  - `cd cli && go test ./... -run TestHandoffRecord`
  - `cd cli && go build ./... && go vet ./...`
- stop if:
  - CONTRACT.md's flag shape proves ambiguous once implemented
- escalate to:
  - user clarification (this is a reopened, already-gated phase — do not silently expand further)

## Risks / Watch-fors
- Fixtures must mirror real templates from Phase 2 — regenerate them from the actual reference templates, don't hand-craft divergent copies
- Changeset-first ordering is the invariant most tempting to shortcut under test pressure — keep the integration assertion in every command test
