# Context: continuity

Phase: continuity
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: high
Expected Proof: e2e (cross-machine resume)

## Goal
watzup and handoff share one CLI-backed continuity contract; resume is exact across sessions and machines; git/interview get their minimal CLI integration, completing the 8-skill scope.

## Deviation 2026-07-17 — narrow addition to `check` (Forbidden Surface)
While building Wave 2 T4's cross-machine e2e chain, found `check record` never writes `meta.latest_check_id` — the same pointer-maintenance gap class Phase 5 (skill-adapters) already fixed for `latest_run_id` by having `work`'s SKILL.md hand-author a meta changeset. `check`'s SKILL.md never got the equivalent step (it was a Forbidden Surface in both skill-adapters and this phase), and no roadmap phase currently owns it — an orphaned gap. Left unfixed, `resume --json`'s `latest_check_id` would stay `null` even after a real check was recorded, which would fail T4's own acceptance check ("recap fields == handoff entity fields", since the handoff entity does carry `check_id` explicitly via `--check-id`).

User approved a narrow, one-step addition to `check`'s SKILL.md (Step 1.6, step 5): after `check record` returns its id, author+apply a one-line `meta.latest_check_id` changeset, mirroring `work`'s already-approved pattern exactly. `check`'s `metadata.version` bumped `1.3.0` → `1.4.0`. No other part of `check` touched. Treated as continuity's own finding (this phase's whole purpose is exact cross-machine resume state) rather than a new to-plan cycle for a one-line, fully-precedented fix.

## Scope Boundary
### Allowed Surfaces
- `skills/workflow/watzup/SKILL.md` + `references/output-contract.md`, `references/artifact-recap.md`
- `skills/workflow/handoff/SKILL.md` + `references/handoff-template.md`, `references/continuity-sources.md`
- `skills/workflow/git/SKILL.md` (read-only integration: query check verdict before commit/PR steps)
- `skills/workflow/interview/SKILL.md` (gate block only)

### Forbidden Surfaces
- `cli/**` (missing capability goes back to Phase 4/6)
- brainstorm/to-plan/work/check skills

## Spec Hooks
- R18 (handoff records entity, watzup renders resume), R20 (readiness states, recovery), R17 (gate in every rewritten skill)
- Acceptance: cross-machine recap matches last handoff exactly; staled pointer → specific recovery step

## Locked Decisions
- Readiness states: `clean | in-progress | drifted | no-harness`; `no-harness` routes to install (`scripts/install-zharness.sh`) or `zharness import` for legacy projects
- handoff writes both: handoff entity via `zharness handoff record` (anchors: state, latest run/check ULIDs, open items) AND human `HANDOFF.md` per Phase 2 template — entity is canonical, markdown is the narrative
- watzup renders exclusively from `zharness resume --json`; no independent re-derivation from prose; recap sections map 1:1 to snapshot fields
- Recovery action table (from STATE.md) reproduced in watzup output-contract so drift always prints its named next step
- git skill: before commit/PR summary steps, `zharness query check --latest --json` — warn (not block) if latest check verdict is FAIL/missing

## Assumptions
- `handoff record` command shape exists from Phase 4 (`check record` pattern); if the CLI lacks it, that is a Phase 4 gap — escalate, don't shim
- Changesets committed by handoff's normal git flow carry state across machines

## Canonical Refs
- `cli/docs/CONTRACT.md`, `cli/docs/STATE.md`
- Phase 2 handoff-template frontmatter contract

## Rejected Options
- watzup keeping prose-based recap as fallback — SPEC: no fallback; `no-harness` state exists instead
- Blocking git on FAIL verdict — chain UX preserved; git warns, human decides (matches current git skill safety posture)

## Deferred Ideas
- Read-only state snapshot for GitHub glanceability (SPEC deferred)

## Escalate If
- resume snapshot lacks a field the recap needs → to-plan phase cli-domain
- Cross-machine e2e fails due to changeset semantics (not skill text) → to-plan phase cli-core
