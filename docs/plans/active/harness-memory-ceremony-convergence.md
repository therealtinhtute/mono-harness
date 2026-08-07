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
      following the parse precedent at `cli/internal/application/next.go:127-170` (V3)
      | check: malformed and hand-edited plans append correctly or fail without data loss
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
    - task: bump `MIN_ZHARNESS_VERSION` at `skills/workflow/README.md:35` | check: both gates
- checks:
  - `cd cli && go test ./...`
  - `bash scripts/verify-doc-links.sh`
  - `docs/playbooks/*.md` byte-identical to `cli/docs/embedded/playbooks/*.md`

### phase_slug: `p6-measure-and-close`
- status: planned
- goal: prove the outcome with the same method that measured the baseline
- depends_on: `p5-harvest`
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

## Progress
<!-- Append-only durable entries record timestamp, phase, wave, task, task_status,
run_id, trace_id, exact verification/result, and changed surfaces or blocker. -->
- none

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

## Validation
<!-- Append-only durable entries record timestamp, phase, exact command/result/output,
run_id, check_id, verdict, and proof_gaps. -->
- none

## Current State and Next Action
- active_phase: none
- lifecycle_status: planned
- latest_run_id: none
- latest_trace_ids: []
- latest_check_id: none
- latest_handoff_id: none
- blockers: none
- open_items:
  - owner decision: does `.kit/changesets/` become committed? Decides whether
    `db rebuild` is a convenience or load-bearing (NG4)
  - owner decision: the trust model in D5/NG3
  - Phases 4–6 end in a deliberate release beat: bump the version, release the binary,
    run `zharness init --refresh-docs` on consuming repos
- exact_next_action: mint the plan and intake IDs, then start `p1-integrity-operability`
  wave 1
