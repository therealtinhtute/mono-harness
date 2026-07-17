# Context: cli-domain

Phase: cli-domain
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: medium
Expected Proof: unit + integration (fixtures in CI)

## Goal
All daily-driver commands work: 7 ported domain commands + workflow additions (`resume`, `check record`, `validate`), every mutation changeset-first, `--json` everywhere.

## Reopened 2026-07-17 — handoff record gap
Phase was already implemented, committed, and gated APPROVED. Reopened narrowly for one addition: `zharness handoff record`. Root cause: SPEC R6 locks exactly 19 commands and does not list `handoff`, yet R18 ("handoff records entity") requires one; `cli/docs/CONTRACT.md`'s own escalation note (top of file) and `continuity-CONTEXT.md`'s Assumption both name this exact route (`to-plan phase cli-domain`) as the resolution path. `domain.Handoff`/`HandoffAnchors` and the `handoffs` table already exist (T1 built them ahead of the gap being hit) — only the write-path command is missing. User approved reopening for this narrow addition rather than a full `brainstorm refine` of R6's count.

## Reopened 2026-07-17 (second, then reverted) — run creation gap, false alarm
Discovered while executing `continuity-PLAN.md` Wave 2 T4 (cross-machine resume e2e): appeared that no CLI command writes a `runs` row. Reopened cli-domain (Wave 5/T6) and implemented a dedicated `run` command — but before committing, discovered Phase 5 (`skill-adapters`, already gated **APPROVED**, see `.kit/reports/check/20260717-1800-skill-adapters.md`) had already solved this exact gap differently and deliberately: `work`'s SKILL.md Step 2 hand-authors a two-line changeset (`create run` + `update meta.latest_run_id`) and applies it via the already-generic `zharness db changeset apply` — explicitly choosing **not** to add a dedicated command ("no dedicated 'run create' CLI command exists" is a direct quote from `work`'s own SKILL.md, confirmed still current). That gate report calls this mechanism intended design, confirmed against SCHEMA.md's producer-mapping table, and it also correctly updates `meta.latest_run_id` — something the new dedicated command did not do.

User chose to discard the new `run` command rather than duplicate an already-approved mechanism. All Wave 5/T6 code (`run.go`, `run_create.go`, `run_create_test.go`, `root.go` registration, CONTRACT.md entries) was reverted; `cli-domain-PLAN.md`'s Wave 5 section removed. **Lesson for future investigation**: before diagnosing "no command creates X," check the consuming skill's SKILL.md for a hand-authored-changeset pattern, not just `cli/internal/interfaces/root.go`'s registered commands — Phase 5 established that pattern as the intended way for skills (not just dedicated commands) to close CLI gaps.

Continuity's Wave 2 T4 (cross-machine resume e2e) should exercise `work`'s existing run-registration changeset (already proven end-to-end in skill-adapters' own dry-run: `query state --json` showed `latest_run_id` correctly populated) — not a new command.

## Scope Boundary
### Allowed Surfaces
- `cli/internal/**`, `cli/cmd/**`, `cli/testdata/**`
- `cli/docs/CONTRACT.md` (reopened scope, add-only: document the new `handoff record` command and resolve the existing escalation note — CONTRACT.md is the single source of truth for flags, it cannot stay stale once the gap it names is closed)

### Forbidden Surfaces
- `skills/workflow/**` (Phase 5+)
- Research commands score-trace/score-context/audit/propose (Phase 6)
- Schema migrations that break Phase 3 changesets (append-only migration only)

## Spec Hooks
- R6 (surface + --json), R7 (changeset-first), R9 (ULIDs), R11 (validate semantics), R12 (resume snapshot)
- Acceptance: validate fails broken fixture / passes fixed one; chain of IDs cross-references

## Locked Decisions
- `validate` reads artifact frontmatter from `.kit/planning/**` and run/report/handoff files, walks links SPEC→PLAN→RUN→CHECK→HANDOFF, checks: required keys, link targets exist, ULID formats, pointer freshness vs DB state
- `resume` output (JSON + human): position, phase status, latest run/check/handoff IDs, drift findings each with `recovery` string, readiness state `clean|in-progress|drifted|no-harness`
- `check record` writes verdict + proof links as a check entity; deterministic — no free-text-only verdicts
- Intake lanes fixed: `tiny|normal|high-risk` (upstream)

## Assumptions
- CONTRACT.md is the single source for flags — no undocumented flags added while porting
- Upstream domain validation rules (domain.rs) port as behavior tests first, then implementation

## Canonical Refs
- `cli/docs/CONTRACT.md`, `SCHEMA.md`, `STATE.md`
- `~/Lab/harness-experimental/crates/harness-cli/src/domain.rs` (validation ground truth)

## Rejected Options
- Implementing validate as a repo script instead of a CLI command — enforcement must ship with the binary skills gate on
- Free-form verdict text in check record — determinism requirement (R19) needs structured verdicts

## Deferred Ideas
- `query` graph views across stories/decisions — post-pilot if needed

## Escalate If
- CONTRACT.md shape proves unimplementable or ambiguous mid-port → to-plan phase harness-contracts
- validate needs a frontmatter key the Phase 2 templates don't define → to-plan phase harness-contracts
