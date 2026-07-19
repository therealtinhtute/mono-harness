# Plan: harness-mode-parity

Phase: harness-mode-parity
Status: ready
Wave Count: 3
Execution Owner: work
Updated At: 2026-07-19

## Goal
Make `zharness validate --json` return `valid:true` on a simple-mode-produced chain by (a) stopping `work`/`check` simple mode from attempting DB registration that structurally cannot succeed, and (b) teaching `validate` mode-aware carve-outs — without weakening full-mode validation. Ship the fix as a new CLI release.

## Inputs
- `harness-mode-parity-CONTEXT.md` (Locked Decisions, Scope Boundary)
- `cli/internal/application/validate.go`, `validate_test.go`, `cli/testdata/chain-valid`, `cli/testdata/chain-broken`
- `cli/internal/infrastructure/migrations.go` (schema — read-only reference, not touched)
- `cli/docs/embedded/playbooks/work.md`, `check.md`
- `cli/docs/CONTRACT.md`
- GitHub #38, backlog `01KXWH4YNC9RRFR1VPE6DK8P14`

## Wave 1
### T1 — validate.go mode-awareness
- type: implementation
- inputs:
  - `cli/internal/application/validate.go`, `validate_test.go`
- touches:
  - `cli/internal/application/validate.go`
  - `cli/internal/application/validate_test.go`
  - new fixture: `cli/testdata/chain-simple-mode/` (RUN with `mode: simple`, `phase: none`, `plan_id: none`; a CHECK with `mode: simple`, no `run_id` cross-link; mirrors the pilot's actual chain shape)
- avoid:
  - `cli/internal/infrastructure/migrations.go` (no schema migration — Rejected Options)
- steps:
  1. In the RUN loop, read `fields["mode"]`. When `== "simple"`: skip the `phase`-existence broken_link check, skip requiring `plan_id` to be a ULID (still fine if absent/`"none"`), skip the DB `stale_pointer` check for that run's `id`. `id` itself must still pass `requireULID` unconditionally.
  2. Same for the CHECK loop: `mode == "simple"` skips the `run_id`-cross-link-to-known-RUN check and the DB `stale_pointer` check on `id`. `id` still required as a ULID.
  3. Missing/absent `mode`, or `mode != "simple"`, keeps every existing check exactly as today (default = full-mode strictness — Locked Decision).
  4. Add `cli/testdata/chain-simple-mode/` fixture matching the pilot's real artifact shape.
  5. Add tests: `TestValidateSimpleModeRunSkipsFullModeChecks` (fixture → `valid:true`), `TestValidateSimpleModeCheckSkipsRunLink` (same), `TestValidateFullModeRegressionUnaffected` (existing `chain-valid`/`chain-broken` fixtures still produce identical results — no behavior change for `mode: full` or no-mode artifacts).
- expected outputs:
  - `Validate()` returns `valid:true` for the new fixture; existing fixtures' results unchanged
- verification:
  - `cd cli && go test ./internal/application/... -run TestValidate -v`
- stop if:
  - a full-mode existing test's expected result changes (regression) — do not proceed, re-check the mode-gate condition
- escalate to:
  - to-plan phase harness-mode-parity

### T2 — work.md: mode-aware run registration
- type: docs
- inputs:
  - `cli/docs/embedded/playbooks/work.md`
- touches:
  - `cli/docs/embedded/playbooks/work.md` (Artifacts § RUN frontmatter template; Execution Loop Step 2)
- avoid:
  - any other playbook file; no SKILL.md changes (R4/R7 boundary)
- steps:
  1. Add `mode: {full|simple}` to the RUN artifact frontmatter template (currently only in the body's `Mode:` line, which `validate.go`'s frontmatter parser cannot see).
  2. Rewrite Execution Loop Step 2: branch explicitly — **full mode** keeps today's two-line changeset registration (`story_slug` FK) unchanged; **simple mode** does NOT author or apply a run changeset — write the run artifact with `mode: simple`, `phase: none`, `plan_id: none`, and a note explaining DB registration is skipped by design (no story to satisfy `runs.story_slug`).
  3. Resolve the section header self-contradiction that caused #38: "## Execution Loop (per phase, full mode)" documents a step that Step 2's own prose already applies to simple mode too — retitle or add an explicit sub-heading so the mode branch is unambiguous, not implied.
- expected outputs:
  - `work.md` no longer instructs simple mode to attempt a registration that always fails FK
- verification:
  - manual re-read: grep `mode:` and `story_slug` in the file, confirm the simple-mode branch is present and the header no longer contradicts it
- stop if:
  - the full-mode branch's wording changes in any way (must stay byte-identical in behavior — no regression to the working full-mode path)
- escalate to:
  - to-plan phase harness-mode-parity

### T3 — check.md: mode-aware check registration
- type: docs
- inputs:
  - `cli/docs/embedded/playbooks/check.md`
- touches:
  - `cli/docs/embedded/playbooks/check.md` (Artifacts § persisted report frontmatter template; Step 4 Harness Gate Flow)
- avoid:
  - any other playbook file; no SKILL.md changes
- steps:
  1. Add `mode: {full|simple}` to the persisted CHECK report frontmatter template, inherited from the RUN artifact it gates (read the RUN's own `mode` field).
  2. In Step 4, before the `zharness check record` call: if the gated RUN's `mode` is `simple`, skip `zharness check record` entirely (no run row exists to satisfy `checks.run_id`'s FK) — write the report with `mode: simple` and a note ("harness check registration: skipped — simple-mode run has no DB row"), same shape as the RUN-side skip in T2. Full-mode gated runs keep the existing `check record` call unchanged.
  3. This resolves the check-side twin already backlogged (`01KXWH4YNC9RRFR1VPE6DK8P14`) and GitHub #30's root cause (check.md Step 4 had no defined behavior for simple-mode-originated diffs) via the same mechanism as T2, not a separate design.
- expected outputs:
  - `check.md` no longer instructs a `check record` call that always fails `unknown_run_id` for simple-mode-gated diffs
- verification:
  - manual re-read: grep `mode:` and `check record` in the file, confirm the simple-mode skip branch is present
- stop if:
  - the full-mode branch's wording or the `check record` invocation for full-mode changes in any way
- escalate to:
  - to-plan phase harness-mode-parity

## Wave 2
### T4 — CONTRACT.md: document the fix
- type: docs
- inputs:
  - T1, T2, T3 outputs (final field names/behavior)
- touches:
  - `cli/docs/CONTRACT.md` (`validate` entry)
- avoid:
  - any entry other than `validate`'s
- steps:
  1. Add `not_yet_implemented` to the documented issue enum (`{"link": ..., "issue": "missing_key"|"broken_link"|"stale_pointer"|"not_yet_implemented", ...}`) — already emitted by code (line 67), was undocumented before this phase.
  2. Add a short paragraph documenting the mode-aware carve-out: RUN/CHECK artifacts with `mode: simple` are exempt from phase-existence, `plan_id`, and DB stale-pointer checks; `id` remains required; missing/`full` mode is unaffected.
- expected outputs:
  - CONTRACT.md's `validate` section matches T1's actual implemented behavior exactly
- verification:
  - diff CONTRACT.md's issue enum against `grep -n "Issue:" cli/internal/application/validate.go` — every emitted issue string appears in both
- stop if:
  - a discrepancy is found between documented and emitted behavior anywhere in the `validate` section (not just the new addition)
- escalate to:
  - to-plan phase harness-mode-parity

### T5 — Scratch-dir integration proof
- type: test
- inputs:
  - T1–T4 landed (built binary reflects all changes)
- touches:
  - none (test runs against a scratch directory outside the repo, per this phase's standard proof convention — same pattern as `cli-stale-drift`'s integration suite)
- avoid:
  - this repo's own live `.kit/` (do not run the simulated simple-mode work/check flow against Lab/skills' own state)
- steps:
  1. `cd cli && go build -o /tmp/zharness-parity-test ./...` (or equivalent scratch build)
  2. In a scratch dir: `zharness init`, then hand-write a RUN artifact + apply the `work.md` T2 flow's simple-mode branch (no changeset), then hand-write a CHECK report + apply the `check.md` T3 flow's simple-mode branch (no `check record` call)
  3. Run `zharness validate --json` in that scratch dir — confirm `valid: true`
  4. As a regression check, run the same scratch-dir flow in full mode (with a real story/plan) and confirm `valid` behavior is unchanged from pre-phase baseline
- expected outputs:
  - `{"valid":true,...}` for the simple-mode scratch chain; unchanged full-mode result
- verification:
  - captured `zharness validate --json` output for both runs
- stop if:
  - the simple-mode scratch chain does not reach `valid:true` — re-open T1–T3, do not ship
- escalate to:
  - to-plan phase harness-mode-parity

## Wave 3
### T6 — Release cli/v0.2.1 (or next appropriate version)
- type: implementation
- inputs:
  - Wave 2 clean (T4 doc-parity, T5 scratch proof both pass)
- touches:
  - git tags, GitHub release (reuse `cli-release`'s proven flow verbatim — do not improvise tag naming)
- avoid:
  - workflow/CI edits unless the pipeline itself fails
- steps:
  1. `cd cli && go build ./... && go test ./...` clean at the release commit
  2. Push the `cli/vX.Y.Z` trigger tag per the documented flow (CI creates the bare-semver tag for goreleaser) — pick the version per semver given this is a bug fix + additive doc/behavior change (patch or minor, not major)
  3. Watch the release workflow; on failure, fix-forward minimally and re-tag per the pipeline's documented recovery (never force-move a broken tag — a failed `vX.Y.Z` becomes `vX.Y.(Z+1)`)
  4. Verify release assets exist for the full platform matrix
- expected outputs:
  - published GitHub release with the fix
- verification:
  - `gh release view` shows assets; download one asset and run `--version`
- stop if:
  - pipeline needs structural rework beyond a fix-forward
- escalate to:
  - to-plan phase

### T7 — MIN_ZHARNESS_VERSION bump (conditional)
- type: docs
- inputs:
  - T6's published version number
- touches:
  - `skills/workflow/README.md`
- avoid:
  - the 6 thin-trigger SKILL.md files themselves (they reference `README.md`'s constant, not a hardcoded version)
- steps:
  1. Decide: does an older CLI silently misbehave (crash on simple-mode `work`, or falsely report `valid:false`/`valid:true`) if used against this phase's updated playbook docs? If yes, bump `MIN_ZHARNESS_VERSION` to the new release; if no, leave it and record why in the run artifact.
  2. If bumped, update the constant and its one documented reference point.
- expected outputs:
  - `MIN_ZHARNESS_VERSION` current and correct, or an explicit recorded reason it wasn't bumped
- verification:
  - `grep -rn "MIN_ZHARNESS_VERSION" skills/workflow/README.md`
- stop if:
  - n/a — always resolvable within this task
- escalate to:
  - none

## Risks / Watch-fors
- T1 is the load-bearing task — T2/T3's doc changes and T5's integration proof both depend on its exact field name/values (`mode: full|simple`) being final before they're written against it; if T1's design changes mid-wave, T2/T3 need a quick re-check even though Wave 1 is marked parallel-safe
- Do not let this phase expand into fixing #36 (`plan_id` `query phases --json` gap) — that's a distinct root cause, already filed, out of this phase's scope
- Release versioning: confirm the exact next semver against the CHANGELOG/tag history at execution time, don't hardcode a guess into the plan
