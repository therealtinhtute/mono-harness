# ROADMAP: workflow-harness — zharness runtime

## Planning Basis
- source spec: `.kit/planning/SPEC.md` (locked 2026-07-16)
- planning mode: `full`
- entry phase: `harness-concept`
- execution mode: sequential phases; parallel waves inside phases where marked

## Phase 1: harness-concept
**Goal:** Lock the shared mental model and gap inventory in docs before any contract or code changes.

**Deliverables:**
- `skills/workflow/README.md` — lifecycle (Intent→Intake→Story/Plan→Trace→Proof→Handoff/Resume), 4-layer model, skill↔command↔entity mapping table (SPEC R22)
- `docs/workflow-harness/gap-matrix.md` — 6-group gap matrix + story↔phase mapping decision (SPEC Open Question 1)
- Root `README.md` links updated

**Dependencies:** none

**Risks / Watch-fors:**
- Concept doc drifting into implementation detail that belongs in contracts
- Story↔phase mapping left vague — it blocks SCHEMA.md in phase 2

## Phase 2: harness-contracts
**Goal:** Freeze every machine contract: workflow state, artifact frontmatter, CLI commands, DB schema, changeset format.

**Deliverables:**
- `cli/docs/STATE.md` — state contract v1 + complete legacy `workflow-state.yml` field mapping (R12, R14)
- Artifact frontmatter contracts in 4 templates: brainstorm spec, work run, check report, handoff (R15, R16)
- `cli/docs/CONTRACT.md` — 19 command schemas with `--json` shapes and error codes (R6)
- `cli/docs/SCHEMA.md` — SQLite tables + changeset entity types (R13)

**Dependencies:** Phase 1 (story↔phase mapping)

**Risks / Watch-fors:**
- A command without a documented consumer (skill or "reserved") — CONTRACT.md acceptance
- Any yml field without a destination or explicit "dropped" note

## Phase 3: cli-core
**Goal:** Working `zharness` binary with the durable core (init/migrate/import/db/query) plus release pipeline and install script.

**Deliverables:**
- `cli/` Go module: cobra skeleton, `internal/{interfaces,application,domain,infrastructure}` (R2, R5)
- SQLite store + migrations; ULID changeset append/apply with idempotency (R7, R8, R9)
- `init`, `migrate`, `import` (legacy `.kit` seeding, R10), `db changeset apply|status`, `query --json`
- goreleaser + GitHub Actions releases (darwin/linux × amd64/arm64, CGO=0) + `scripts/install-zharness.sh` via `gh release download` (R21)

**Dependencies:** Phase 2 (CONTRACT.md, SCHEMA.md, STATE.md)

**Risks / Watch-fors:**
- Port scale (~6k lines upstream infrastructure) — stay contract-driven, do not line-port unused Rust
- cgo sneaking in via a transitive dep — CI must build with CGO_ENABLED=0

## Phase 4: cli-domain
**Goal:** The commands the adapters call daily: intake/story/decision/backlog/tool/intervention/trace + workflow additions resume/check-record/validate.

**Deliverables:**
- 7 ported domain commands, each appending a changeset before DB writes (R6, R7)
- `resume` continuity snapshot with drift findings + recovery actions (R12)
- `check record` verdict writes; `validate` walking SPEC→PLAN→RUN→CHECK→HANDOFF by ULIDs with machine-readable findings (R11)
- Pass/fail fixtures for `validate` in CI

**Dependencies:** Phase 3

**Risks / Watch-fors:**
- `validate` semantics drifting from the frontmatter contracts written in Phase 2
- Fixtures too synthetic — base them on a realistic `.kit/` sample

## Phase 5: skill-adapters
**Goal:** brainstorm/to-plan/work rewritten CLI-first: explicit `zharness` calls inline in the flow, mandatory version gate, no `workflow-state.yml` writes.

**Deliverables:**
- Rewritten `skills/workflow/{brainstorm,to-plan,work}/SKILL.md` + touched references (R17, R18)
- Standard version-gate block (binary check → install instructions on failure)
- Sample-project chain run whose IDs cross-reference under `zharness validate`

**Dependencies:** Phase 4

**Risks / Watch-fors:**
- UX drift — skill order and intent must stay identical (SPEC goal)
- Residual yml writes hiding in references/ files

## Phase 6: validation-gate
**Goal:** Research commands land and `check` gates deterministically on evidence.

**Deliverables:**
- `score-trace`, `score-context`, `audit`, `propose` commands (R6); `audit` emits a gate-consumable drift/violation report
- Rewritten `skills/workflow/check/SKILL.md` + `gate-checklist.md` + `artifact-alignment.md`: validation matrix (lane × proof class), verdict via `check record` (R19)
- Determinism fixtures: missing required proof ⇒ FAIL naming the proof

**Dependencies:** Phase 5

**Risks / Watch-fors:**
- Trace tier definitions (SPEC Open Question 2) — default to upstream tiers, note deviations
- Matrix cells left judgment-based — every cell must be required/optional/n-a

## Phase 7: continuity
**Goal:** watzup and handoff share one CLI-backed continuity contract; resume is exact across sessions and machines.

**Deliverables:**
- Rewritten `skills/workflow/{watzup,handoff}/SKILL.md` + `output-contract.md` + `handoff-template.md` (R18, R20)
- Unified readiness states `clean | in-progress | drifted | no-harness` with recovery routing
- Minimal CLI-awareness updates to `git` (reads check verdict via `query`) and `interview` (version gate) — completes the 8-skill scope
- Cross-machine e2e: clone → install → `resume` recap matches last handoff

**Dependencies:** Phase 6

**Risks / Watch-fors:**
- Drift surfaced as generic warnings — every finding needs its named recovery step
- handoff writing prose anchors that `resume` cannot parse

## Phase 8: pilot-migration
**Goal:** Prove the chain end-to-end on a real task, then publish migration/adoption docs and purge legacy semantics.

**Deliverables:**
- Pilot: full chain incl. cross-machine resume; evidence committed; gap issues filed; go/no-go verdict (R24)
- `docs/workflow-harness/migration.md` (checklist, rollback notes, contributor playbook) + root README quickstart (R22)
- `CLAUDE.md` + workflow references purged of `workflow-state.yml`; template file removed (R14, R23)

**Dependencies:** Phase 7

**Risks / Watch-fors:**
- Pilot on a toy task proves nothing — pick a real change in a real repo
- Purge breaking this repo's own planning flow mid-migration — sequence purge after pilot go
