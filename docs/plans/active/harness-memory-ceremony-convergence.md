---
id: {mint with `zharness id` at intake time}
type: plan
intake_id: {mint with `zharness intake` at intake time}
lane: high-risk
status: active
created: 2026-08-07
updated: 2026-08-07
---

# Plan: Harness Memory & Ceremony Convergence

> **Pre-harness note.** This plan was authored before `zharness` was installed in the
> working environment, so its `id` and `intake_id` are unminted. The implementer's first
> action is `zharness id` + `zharness intake --type harness-improvement --lane high-risk
> --plan-path docs/plans/active/harness-memory-ceremony-convergence.md`, then filling both
> IDs into the frontmatter above. Everything below is verified against the codebase at
> commit `13ce88e`.

## Outcome

- result: the harness becomes what `cli/docs/CONTRACT.md` already assumes it is — a
  compressed, queryable index of the plan markdown, sufficient for an agent to resume
  and execute without reading the full plan.
- success_signals:
  - one tiny change (one file, ~20 lines) through the full durable path costs ≤ 35
    mandated operations, down from a measured 62
  - `watzup` cold start costs 2 operations, down from 5
  - the same lifecycle against a plan carrying 1 phase of history and one carrying 8
    costs within 20% per stage; today the ratio is 6.4x
  - every append-only markdown section (`Progress`, `Decisions`, `Validation`) has a
    query that returns its compressed form — no section is write-only
  - after a simulated two-machine interleaved changeset merge, either every row is
    present after `db rebuild`, or `db status` names exactly what is missing

## Authority and Requirements

- authority:
  - `docs/audit/workflow-harness-ceremony-audit.md` — measured baseline (F1–F3)
  - `skills/workflow/README.md` — 4-layer model, thin-trigger template, version gate
  - `cli/docs/CONTRACT.md` — locked command shapes; `:131-137` resume, `:76` changesets
  - `cli/docs/SCHEMA.md` — tables, changeset format, replay rules
  - `docs/evals/failures.md` — the graduation rule this work must not violate
  - Owner decision, this session: "the DB is not a copy of the markdown; it is the
    compressed index of it"

- requirements:
  - R1 [accepted]: every append-only markdown section has an index row at the same
    granularity the markdown records it | source: owner's locked mental model
  - R2 [accepted]: any step needing "what happened" reads the index; any step verifying
    correctness reads markdown | source: owner's locked mental model
  - R3 [accepted]: the CLI writes the history half; the agent writes the definition half
    | source: index and markdown can only be guaranteed consistent by one writer
  - R4 [accepted]: `preflight <stage>` returns one stage-shaped memory packet, not a menu
    of query commands | source: a choice of queries invites over-reading
  - R5 [accepted]: any bounded packet declares what it omitted and how to fetch it
    | source: an agent that does not know it is missing something will open the file
  - R6 [accepted]: no change may depend on a Claude Code-exclusive feature
    | source: portability constraint locked at interview
  - R7 [accepted]: a changeset written under the old schema replays cleanly through the
    new one | source: `docs/workflow-harness/migration.md:49`

## Non-goals

- NG1: splitting the plan into multiple files. It stays one human-readable,
  git-diffable markdown file; only the read path changes.
- NG2: optimizing `check`'s full-plan read. That step is audit by definition (R2).
- NG3: deciding the trust model — whether `check` verifies proof rather than accepting
  a declared string. Trades ceremony against integrity; owner's call. See Decisions D5.
- NG4: changing `.gitignore` for `.kit/changesets/`. Separate owner decision.
- NG5: `changeset compact`/`prune`. `docs/workflow-harness/migration.md:49` forbids
  editing or deleting a committed changeset.
- NG6: any change under `skills/craft/`, `skills/shipping/`, or `rules/`.

## Approach and Risks

- approach: complete the compressed index first (Phases 1–2), make one writer own both
  sides of it (Phase 3), then expose it as a stage-shaped packet (Phase 4) and finally
  rewrite the playbook read steps to use it (Phase 5). Phases 1–3 deliver the ceremony
  reduction; Phases 4–5 deliver the token reduction and are release-coupled.

- constraints:
  - `docs/playbooks/*.md` must stay byte-identical to `cli/docs/embedded/playbooks/*.md`
    (`cli/internal/embedded/projection_drift_test.go:11-45`)
  - `docs_version` is the CLI's own version string, so any release marks existing repos
    `stale` and requires `zharness init --refresh-docs`
    (`cli/internal/interfaces/preflight.go:110`, `cli/internal/application/preflight.go:82-88`)

- risks:
  - R-A: the markdown section appender (Phase 3) corrupts a hand-edited plan. A writer
    that damages the record is worse than the 19 manual edits it replaces. Mitigated by
    following the existing parse precedent and by a malformed-plan test. **Pause on failure.**
  - R-B: replaying an old changeset through the new schema loses or reorders a row.
    `db rebuild` exists to recover data and must never destroy it. **Pause on failure.**
  - R-C: re-adding `decisions` reverses an approved subtraction (see D2). Justified by
    R1, but the justification must hold, not merely be asserted.
  - R-D: **ordering hazard.** Bumping `MIN_ZHARNESS_VERSION` before the binary is
    released points every skill's version gate at a version nobody can install, breaking
    the chain. Releases are triggered by pushing a `cli/vX.Y.Z` tag; `install-zharness.sh`
    resolves the latest published release. The bump and the release are therefore one
    ordered unit, not two independent edits. See phase `p5b-release`.
  - R-E: **P2 is a point of no return.** `migrations.go` is forward-only with no `Down`,
    and once a changeset carries the `decision` entity it cannot replay on schema 6. The
    remedy for a mistake in 0007 is migration 0008, not a rollback. Land P2 only when its
    replay test is green.

## Verified Findings — read before implementing

These were confirmed against the codebase during plan verification. Four corrected
earlier assumptions.

**V1 — `decisions` was deliberately dropped, and was never written to.**
`cli/internal/infrastructure/migrations.go` migration `0003_drop_dead_surface` drops
`decisions`, `backlog`, and `tools`. Commit `73bc50c`, "Harness Subtraction Pass",
gated APPROVED. `git log -S "decision add"` returns nothing: the table never had a
writer. So re-adding it is not reversing a rejection — it is completing a surface that
was speculative before and is now derivable from R1. Two consequences: migration 0007
must **re-create** a table that once existed, and no historical changeset references
the `decision` entity, so replay is unaffected.

**V2 — G2 (gate `--judge independent` for `high-risk`) cannot land in Phase 1.**
There is no join from a check back to `intakes.lane`. `runs.plan_id` is a free-text ULID
(`cli/internal/application/run_create.go:27` validates shape only); `intakes.plan_path`
is a file path (migration `0005_intake_plan_path`). They share no key. The task therefore
needs either an `intakes.plan_id` column or resolution through the plan file's
frontmatter — both schema or parse work. **Moved from Phase 1 to Phase 2.**

**V3 — Phase 3 is not greenfield.** `cli/internal/application/next.go:127-170` already
globs `docs/plans/active/*.md`, reads each file, and parses it
(`parseActivePlanPhaseOrder`). The section appender should follow this existing
precedent rather than introducing a second parsing style.

**V4 — migration mechanics are simple.** `migrations.go:18` is a flat
`[]migration{Version, Name, SQL}` slice; `CurrentSchemaVersion()` takes the max. Adding
0007 is additive. Current version is 6 (`0006_check_judge`).

**V5 — test conventions.** 32 colocated `*_test.go` files across
`internal/{application,domain,infrastructure,interfaces}` plus `cmd/zharness`.
`t.TempDir()` is the established isolation pattern.

**V6 — the CLI already degrades for `git`; only the skill file blocks.** In a repository
with no database, `zharness preflight git --json` returns
`{"stage":"git","mode":"reduced","db":"missing","docs":"missing","readiness":"reduced"}`
with exit 0 and no `stop`. The hard stop comes from `skills/workflow/git/SKILL.md:13`,
which halts when the binary is absent — even though the next line states that Git
operations are non-mutating to harness state. `skills/workflow/interview/SKILL.md:14`
carries the same defect. Consequence: in a repository without zharness installed,
neither skill can run at all, despite neither owning a harness entity
(`skills/workflow/README.md:66-67`). Also relevant: `SKILL.md` files are **not** part of
the embedded doc projection (`cli/docs/embedded/` holds only `AGENTS.md`, `WORKFLOW.md`,
`playbooks/`, `templates/`), so fixing them bumps no `docs_version` and forces no
`init --refresh-docs`. That is why this lands in P1 rather than P5.

**V7 — release mechanics.** `.github/workflows/cli-release.yml` publishes on a pushed
`cli/vX.Y.Z` tag; goreleaser publishes the release under the bare `vX.Y.Z` tag, and
`scripts/install-zharness.sh` resolves the latest published release. `version` is stamped
at build time and defaults to `dev` (`cli/cmd/zharness/main.go:9`), which is why local
builds always satisfy the gate.

## Phases and Verification

<!-- Phase and task definitions are immutable after to-plan. Do not add task status fields.
Append-only Progress is the sole task execution-status source. Only each phase lifecycle
status changes to mirror DB transitions. -->

- planning_status: planned
- phases:

### phase_slug: `p1-integrity-operability`
- status: planned
- goal: fix the two safety defects and remove the ceremony trap, without touching schema
  or contract
- depends_on: none
- waves:
  - wave 1 (parallel-safe):
    - task: `query traces [--run-id] [--tail]` — add `TraceView` + `QueryTraces` to
      `cli/internal/application/query.go`, a case and flags to
      `cli/internal/interfaces/query.go`, and update the `Short` view list
      | check: `cd cli && go test ./...`
    - task: `db status` — schema version, fence, per-table row counts, true pending
      changesets, plus a per-stage context-cost estimate (addresses G4)
      | check: `cd cli && go test ./...`
  - wave 2:
    - task: fix `ChangesetStatus` (`cli/internal/infrastructure/changeset.go:367-387`) to
      distinguish applied from below-fence-never-applied
      | check: test asserting a below-fence unapplied changeset is reported, not hidden
    - task: `db rebuild --yes` — delete `harness.db`, `-wal`, `-shm`; migrate; `Replay()`
      from empty | check: two-machine interleaved merge test reproducing the `b1` loss
  - wave 3:
    - task: `hasNonEmptyActivePlan` (`cli/internal/interfaces/preflight.go:78-90`) requires
      a phase actually `in-progress`, not merely a non-empty file
      | check: preflight returns `reduced` for an active plan with no in-progress phase
  - wave 4 (V6 — skill files only, no doc projection, no version bump):
    - task: `skills/workflow/git/SKILL.md:13` and `skills/workflow/interview/SKILL.md:14`
      degrade instead of hard-stopping — a missing binary proceeds without harness
      enrichment, an old one warns and continues
      | check: both skills usable in a repository that has never run `zharness init`
    - task: harness enrichment (`query check --latest`) is skipped, not failed, when
      `db: missing` | check: no error surfaced for an uninitialized repository
    - task: record the rule in `skills/workflow/README.md` — a skill that owns no harness
      entity must not hard-stop on the harness (`README.md:66-67` already names the two)
      | check: `bash scripts/verify-doc-links.sh`
- checks:
  - `cd cli && go test ./...`
  - `bash scripts/verify-doc-links.sh`

### phase_slug: `p2-complete-the-index`
- status: planned
- goal: every append-only markdown section has an index row at matching granularity
- depends_on: `p1-integrity-operability`
- waves:
  - wave 1:
    - task: migration 0007 re-creating `decisions`; register the entity in
      `cli/internal/infrastructure/changeset.go:29-55` (`entityTables`, `entityColumns`)
      | check: **a changeset written under schema 6 replays cleanly through schema 7** (R7)
    - task: `decision add` (repeatable `--decision`, plus `--rationale`, `--phase`/`--task`,
      `--run-id`) and `query decisions` | check: write-then-read round trip
  - wave 2:
    - task: `trace add` records at task granularity, matching `docs/playbooks/work.md:38`
      (addresses G1 — today the index is wave-level while the markdown is task-level, so a
      mid-wave interruption leaves the index blind) | check: task-level round trip
    - task: add the `intakes` ↔ plan link needed for lane resolution, then require
      `--judge independent` when the lane is `high-risk` (V2, addresses G2)
      | check: a `high-risk` check with `--judge same-session` is rejected
    - task: persist `exact_next_action` into `handoffs.anchors` (free-form JSON, no
      migration) | check: round trip through `handoff record` and `resume`
- checks:
  - `cd cli && go test ./...`
  - old-schema changeset replay test green (R7, risk R-B)

### phase_slug: `p3-cli-owns-the-pen`
- status: planned
- goal: one writer writes both the index row and its markdown line, in one transaction
- depends_on: `p2-complete-the-index`
- waves:
  - wave 1:
    - task: markdown section appender — locate `## X`, insert before the next `## `,
      following the parse precedent at `cli/internal/application/next.go:127-170` (V3).
      **Hand-written plans are the primary case, not an edge case**: every plan that
      exists today was written by an agent editing markdown directly, in whatever shape
      it chose. The appender must accept those, not just files it produced itself.
      | check: plans with reordered sections, missing sections, trailing content after the
      last heading, and CRLF line endings all append correctly or fail without data loss
  - wave 2:
    - task: `trace add`, `decision add`, `check record`, `handoff record` write index +
      markdown atomically | check: index and markdown cannot diverge; failure rolls back both
- checks:
  - `cd cli && go test ./...`
  - **Pause condition (risk R-A): if the appender cannot survive a hand-edited plan
    without data loss, stop and report rather than shipping a writer that corrupts the record.**

### phase_slug: `p4-stage-shaped-context`
- status: planned
- goal: one call returns the memory a stage needs (R4, R5)
- depends_on: `p3-cli-owns-the-pen`
- waves:
  - wave 1:
    - task: `preflight <stage>` returns a stage-shaped `context` object
      | check: each stage's packet contains exactly the fields its playbook references
    - task: add `version` to `PreflightView` (`cli/internal/application/preflight.go:24`),
      populated in `runPreflight` which already receives it
      (`cli/internal/interfaces/preflight.go:46`) | check: payload carries the version
  - wave 2:
    - task: window policy — current phase's traces in full, prior phases as story rows
      only, `--tail` capped at 30 | check: packet size bounded across phase counts
    - task: declare `omitted` with a fetch hint (R5) | check: truncation is never silent
  - wave 3:
    - task: update `cli/docs/CONTRACT.md:131-137` and `cli/docs/SCHEMA.md`
      | check: `bash scripts/verify-doc-links.sh`
- checks:
  - `cd cli && go test ./...`
  - `bash scripts/verify-doc-links.sh`

### phase_slug: `p5-harvest`
- status: planned
- goal: playbooks read the index instead of the whole plan — where the measured 230,656
  tokens actually disappear
- depends_on: `p4-stage-shaped-context`
- waves:
  - wave 1 (every edit applies to both `docs/playbooks/` and `cli/docs/embedded/playbooks/`):
    - task: `watzup.md:17-18` renders 1:1 from the packet, as `CONTRACT.md:137` already
      requires | check: projection-drift test green
    - task: `work.md:32` and `handoff.md:25` read the definition half plus the packet;
      **`check.md:33` is unchanged** (NG2) | check: projection-drift test green
  - wave 2:
    - task: vocabulary audit — every noun an execution step references must be a real
      field in the packet | check: manual review, recorded in Decisions
    - task: add a rehydrate step — after compaction, re-run `preflight` (addresses G5)
      | check: projection-drift test green
  - wave 3:
    - task: drop `zharness --version` from the 6 spine `SKILL.md`, the 6 playbooks, and
      the template at `skills/workflow/README.md:41-52` | check: both gates
    - task: defer rare branches out of the main playbook body — bounded mode, status
      routing (addresses G6) | check: projection-drift test green
  - wave 4 (the static outlier the audit measured but the earlier plan dropped):
    - task: thin-trigger `skills/workflow/git/SKILL.md` — 1,349 tokens, four times any
      spine trigger — down toward the ~300-token template at `skills/workflow/README.md:41-52`,
      moving operating logic into a git playbook alongside the six in `docs/playbooks`
      | check: both gates; projection-drift test green
    - task: prune `skills/workflow/git/references/**` against what the new playbook
      absorbs — 9,423 tokens of latent surface, 58% of the chain's total
      | check: `bash scripts/verify-doc-links.sh`
    - task: decide the four shell scripts under `skills/workflow/git/scripts/` (4,349
      tokens) — keep as executables the agent runs without reading, or delete as
      redundant with plain `git`; record the choice in Decisions
      | check: no skill references a deleted script
- checks:
  - `cd cli && go test ./...`
  - `bash scripts/verify-doc-links.sh`
  - `docs/playbooks/*.md` byte-identical to `cli/docs/embedded/playbooks/*.md`

### phase_slug: `p5b-release`
- status: planned
- goal: publish the binary the bumped version gate points at, in that order (risk R-D)
- depends_on: `p5-harvest`
- waves:
  - wave 1:
    - task: push a `cli/vX.Y.Z` tag and confirm `.github/workflows/cli-release.yml`
      publishes the release under the bare `vX.Y.Z` tag (V7)
      | check: `gh release list` shows the new release
    - task: only after the release is published, bump `MIN_ZHARNESS_VERSION` at
      `skills/workflow/README.md:35` and in every skill trigger that names it
      | check: `bash scripts/install-zharness.sh` installs a binary satisfying the new floor
    - task: run `zharness init --refresh-docs` against this repository and confirm
      preflight no longer reports `stale_docs`
      | check: `zharness preflight work --json` returns `docs: ready`
- checks:
  - `bash scripts/verify-doc-links.sh`
  - a fresh clone can install zharness and run `watzup` end to end

### phase_slug: `p6-measure-and-close`
- status: planned
- goal: prove the outcome with the same method that measured the baseline
- depends_on: `p5b-release`
- waves:
  - wave 1:
    - task: re-run the audit's lifecycle measurement; append before/after to
      `docs/audit/workflow-harness-ceremony-audit.md` | check: Success signals 1–3 verifiable
    - task: add F4–F7 and record the locked mental model in the audit doc
      | check: `bash scripts/verify-doc-links.sh`
    - task: realign the ROI table with what actually shipped | check: manual review
- checks:
  - both gates
  - `git status` shows only files named in this plan's scope
- standing obligation for every phase, not just this one: any durable `check` returning
  `REQUEST_CHANGES` appends one row per finding to `docs/evals/failures.md`, and any
  failure class appearing there a second time becomes a deterministic check under
  `scripts/` before this initiative closes. An initiative that repairs the harness must
  obey the harness's own graduation rule.

## Progress
<!-- Append-only durable entries record timestamp, phase, wave, task, task_status,
run_id, trace_id, exact verification/result, and changed surfaces or blocker. -->
- 2026-08-07, phase `p1-integrity-operability`, all 4 waves, task_status=DONE, run_id=none,
  trace_id=none (zharness is not installed in this working environment — see the plan's
  Pre-harness note; work was implemented and verified directly against `cd cli && go test
  ./...` and `bash scripts/verify-doc-links.sh` rather than through a live CLI-driven
  lifecycle). Changed surfaces: `cli/internal/application/query.go` (+`TraceView`,
  `QueryTraces`), `cli/internal/application/db_status.go` (new — `QueryDBStatus`,
  `ContextCostEstimate`), `cli/internal/interfaces/query.go` (`traces` view, `--run-id`,
  `--tail`), `cli/internal/interfaces/db.go` (`db rebuild --yes`, `db status`),
  `cli/internal/infrastructure/changeset.go` (`ChangesetStatus` now returns
  `unverifiedBelowFence`), `cli/internal/interfaces/repository_lock.go` (`db rebuild` added
  to the exclusive-lock command list), `cli/internal/interfaces/preflight.go`
  (`resolvePreflightRequestedMode`/`observePreflightState` now require a real in-progress
  story, not just a non-empty active-plan file, before auto-resolving `work` to full mode),
  `cli/docs/CONTRACT.md` (documents `query traces`, `db rebuild`, `db status`, the
  `unverified_below_fence` field, and the widened shared/exclusive lock lists),
  `skills/workflow/git/SKILL.md` + `skills/workflow/interview/SKILL.md` (version/preflight
  gates degrade instead of hard-stopping — neither skill owns a harness entity),
  `skills/workflow/README.md` (records the "no entity, no hard-stop" rule). Plus new/updated
  tests: `cli/internal/application/query_test.go`, `cli/internal/infrastructure/changeset_test.go`
  (`TestChangesetStatusFlagsInterleavedMachineChangesetNeverApplied` — reproduces the
  audit's F5 two-machine `b1` loss), `cli/internal/interfaces/db_test.go` (new — including
  `TestDBRebuildRecoversInterleavedMachineChangeset`, the CLI-level counterpart proving
  `db rebuild --yes` recovers the row), `cli/internal/interfaces/preflight_test.go`
  (`TestPreflightCommandWorkAutoUsesReducedWithActivePlanButNoInProgressPhase` and its
  positive control), `cli/internal/interfaces/read_only_commands_test.go`. Verification:
  `cd cli && go build ./...`, `go vet ./...`, `go test ./...` (all packages ok) and
  `bash scripts/verify-doc-links.sh` (0 findings) both green. No blocker.
- 2026-08-08, phase `p2-complete-the-index`, both waves, task_status=DONE, run_id=none,
  trace_id=none (same environment constraint as P1 — no live zharness binary in this
  session; verified via `cd cli && go test ./...` and `bash scripts/verify-doc-links.sh`).
  Three migrations landed, each its own focused change per the repo's one-change-per-
  migration convention (`0007_decisions` alone matches the plan's own phase name; two
  more followed within this same phase — see D10). `0007_decisions` re-creates
  `decisions`, dropped by `0003_drop_dead_surface`. `0008_trace_task_granularity` adds
  `traces.task`/`traces.task_status`, addressing G1. `0009_intake_plan_id` adds
  `intakes.plan_id`, the join `check record` needs for G2. Schema version is now 9, not
  7 — see D10. Wave 1: `decision add`/`query decisions`
  (`cli/internal/domain/decision.go`, `cli/internal/application/decision.go`,
  `cli/internal/interfaces/decision.go`) — a JSON-array batch call mirroring
  `check record --proof-links`'s existing precedent rather than the plan's literal
  "repeatable --decision flag" wording (see D11), so one call can record several
  decisions each with its own rationale/phase/task, not a shared one. Wave 2:
  `trace add --task/--task-status` (task-level granularity, G1); `check record`'s new
  `resolveLaneForRun` gates `--judge` to `independent` when a run resolves via
  `runs.plan_id` -> `intakes.plan_id` to a `high-risk` lane, additive and non-breaking
  (unresolvable or non-high-risk lanes are unaffected — see D12); `handoff record
  --next-action` persists into `anchors.exact_next_action` with no migration, readable
  back via the new `query handoff --latest` view added to close the round trip
  (`resume`'s own locked shape is untouched — see D13). R7's hard gate
  (`TestChangesetFromSchema6ReplaysCleanlyThroughCurrentSchema`, renamed from
  `...ThroughSchema7` since current is now 9) stayed green through all three migrations.
  Changed surfaces: `cli/internal/infrastructure/migrations.go`,
  `cli/internal/infrastructure/changeset.go` (entity registration plus `task_status`
  enum validation), `cli/internal/domain/{decision,trace,intake,handoff}.go`,
  `cli/internal/application/{decision,query,trace,intake,check_record,handoff}.go`,
  `cli/internal/interfaces/{decision,trace,intake,handoff,query,repository_lock,root}.go`,
  `cli/docs/CONTRACT.md`, `cli/docs/SCHEMA.md`. New/updated tests across all four
  layers, including `TestRecordDecisionsBatchIsAtomicOnValidationFailure`,
  `TestCreateTraceTaskGranularity`, `TestCheckRecordRequiresIndependentJudgeForHighRiskLane`
  (plus its unresolvable-lane and non-high-risk-lane controls), and
  `TestHandoffRecordNextActionRoundTripsThroughResumeAndQuery` (both application- and
  CLI-level). Verification: `cd cli && go build ./...`, `go vet ./...`, `go test ./...`
  (all packages ok) and `bash scripts/verify-doc-links.sh` (0 findings) both green.
  No blocker.

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- D1 (2026-08-07, planning): the DB is the compressed index of the markdown, not a copy.
  Rationale: the design already did this for `traces.summary` and `intakes.summary`
  without naming it; naming it makes the remaining gaps derivable rather than taste.
- D2 (2026-08-07, planning): re-add `decisions`, reversing migration
  `0003_drop_dead_surface`. Rationale: it was dropped as dead surface with no writer
  (V1), not rejected on merit. Under R1 it is now an obligation, and it ships with a
  writer this time.
- D3 (2026-08-07, planning): keep one plan file. Rationale: the measured problem is that
  the only way to read it is in full, not that it is one file. A slice/index read fixes
  the cost without losing the single diffable narrative.
- D4 (2026-08-07, planning): `check` keeps its full-plan read. Rationale: R2 — it is
  audit, and the same principle that cheapens the other stages protects this one.
- D5 (2026-08-07, planning): the trust model stays out of scope. `proof_links` is a
  declared string validated for shape only (`cli/internal/domain/check.go:42-45`);
  verifying it would trade ceremony against integrity, which is the owner's call (NG3).
- D6 (2026-08-07, verification): G2 moves from Phase 1 to Phase 2. Rationale: V2 — no
  join exists from a check to `intakes.lane`, so it is not the pure-validation change
  the earlier plan assumed.
- D7 (2026-08-07, planning): a skill that owns no harness entity must not hard-stop on
  the harness. Applies to exactly `git` and `interview` per the mapping table at
  `skills/workflow/README.md:66-67`; the 6 spine skills keep their hard stop because they
  write to the harness. Rationale: V6 — the CLI already degrades correctly, so the stop
  is a skill-file defect that makes `git` unusable in any repository without zharness,
  for no benefit.
- D8 (2026-08-07, planning): the version bump and the binary release are one ordered
  unit in their own phase, not two edits inside the harvest. Rationale: risk R-D — the
  bump alone points the gate at a version nobody can install.
- D9 (2026-08-07, planning): the `git` skill's static bloat returns to scope in P5 wave 4.
  Rationale: it was the audit's largest static outlier (1,349-token trigger, 9,423 tokens
  latent, 58% of the chain's latent surface) and fell out of the plan only because it is
  a bloat problem rather than a memory problem — which is not a reason to drop it.
- D10 (2026-08-08, implementation): P2 lands three migrations (0007-0009), one per
  focused change, rather than folding trace-granularity and intake-plan-id into 0007.
  Rationale: every existing migration in this codebase does exactly one thing
  (`0002_meta_docs_version`, `0003_drop_dead_surface`, etc.) — bundling unrelated schema
  changes into one migration would be the first violation of that convention. The plan's
  phase name ("schema 6 to 7") describes wave 1 accurately; it does not constrain wave 2
  to reuse the same version number. R7's replay test uses `CurrentSchemaVersion()`
  dynamically, so it required no changes to keep covering the invariant through 0008/0009.
- D11 (2026-08-08, implementation): `decision add` takes `--decisions` as a JSON array of
  complete `{decision, rationale, phase, task}` objects, not the plan's literal
  "repeatable --decision flag, shared --rationale" wording. Rationale: the codebase's
  established pattern for structured repeatable data is a JSON-array flag
  (`check record --proof-links`), and a shared rationale across a batch would misrepresent
  decisions that have genuinely different reasoning. The JSON-array form is a strict
  superset — a caller can still give every element the same rationale — and achieves the
  same ceremony goal (one call for N decisions).
- D12 (2026-08-08, implementation): the `check record` lane gate (G2) only fires when a
  lane is *resolved* — a run with no `plan_id` behaves exactly as before this feature
  existed, and a resolved non-high-risk lane is unaffected. Rationale: `intake --plan-id`
  is a new, optional flag that no playbook passes yet (that wiring is P5's job); making
  the gate strict would silently start blocking a scenario no agent currently produces
  correctly, for a repo that has done nothing wrong.
- D13 (2026-08-08, implementation): `resume`'s own JSON shape is left untouched; the
  `handoff record --next-action` round trip is completed by a new `query handoff --latest`
  view instead. Rationale: `resume`'s shape is CONTRACT.md-locked and explicitly reserved
  for P4's stage-shaped-context work (risk R-D); enriching it here would pre-empt that
  phase's own contract change. `resume.latest_handoff_id` already names which handoff to
  look up, so the round trip is complete without touching resume itself.

## Validation
<!-- Append-only durable entries record timestamp, phase, exact command/result/output,
run_id, check_id, verdict, and proof_gaps. -->
- none

## Current State and Next Action
- active_phase: `p2-complete-the-index` (both waves code-complete and verified; not
  DB-recorded as `in-progress`/`checked` — no live zharness in this session, so no `run
  create`/`check record` was possible; see Progress entries above for both P1 and P2)
- lifecycle_status: planned (honest: the plan-level phase `status:` field is intentionally
  left unchanged rather than claiming a DB transition this session could not perform)
- latest_run_id: none
- latest_trace_ids: []
- latest_check_id: none
- latest_handoff_id: none
- blockers: none
- open_items:
  - owner decision: does `.kit/changesets/` become committed? Decides whether
    `db rebuild` is a convenience or load-bearing (NG4)
  - owner decision: the trust model in D5/NG3
  - `p5b-release` is a deliberate release beat: publish the binary first, bump
    `MIN_ZHARNESS_VERSION` second, then `zharness init --refresh-docs` on consuming repos
  - suggested cut line: ending after `p3-cli-owns-the-pen` captures the whole ceremony
    reduction while touching no contract and forcing no repository to refresh its docs
  - once zharness is installed in a working environment: mint this plan's `id`/`intake_id`
    (frontmatter still has the pre-harness placeholders), run `zharness run create` for
    `p1-integrity-operability` and `p2-complete-the-index`, and route the diff through
    `check full` to record real DB-linked verdicts — durable bookkeeping this session
    could not produce. Schema is now at version 9 (migrations 0007-0009); `db rebuild`
    after minting IDs will replay everything cleanly (R7's own test proves this).
- exact_next_action: build `zharness` from source (`cd cli && go build -o zharness
  ./cmd/zharness`) in an environment that can run it, mint the plan/intake IDs, record
  runs/checks for `p1-integrity-operability` and `p2-complete-the-index` against this
  session's diff, then start `p3-cli-owns-the-pen` wave 1 — the appender wave, with its
  own Pause condition (risk R-A) if a hand-edited plan cannot survive the writer intact
