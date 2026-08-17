---
id: 01M08297HJ2YQZ0KGPQ7JXRTM9
type: plan
intake_id: 01M08297J2C9K64VB2E0VPJTKK
lane: high-risk
status: active
created: 2026-08-17
updated: 2026-08-17
---

# Plan: Markdown as source of truth, SQLite as derived index

## Outcome
- result: The harness enforces the "exactly one active plan" invariant in code instead of assuming it, and its durable state moves to markdown-authoritative with SQLite as a rebuildable derived index and the CLI narrowed to guardrails.
- success_signals:
  - `zharness scaffold plan` refuses to create a second file in `docs/plans/active/` while one already exists, with a non-zero exit and a diagnostic naming the existing plan.
  - Every caller of the active-plan resolver returns the same `Stop{Code:"ambiguous"}` contract; no caller errors out and no caller silently degrades.
  - Two active plans present: an agent resolves or reports the situation using ≤500 tokens of plan content, never by reading plan bodies.
  - `zharness plan complete` and `zharness plan abandon` exist as the only two exits from `docs/plans/active/`, and `docs/playbooks/handoff.md` calls one of them instead of instructing a manual file move.
  - `docs/playbooks/work.md` step 1 has an explicit recovery branch for an ambiguous/absent active plan, not only for `degraded: true`.
  - After P2, no lifecycle fact is written to both a markdown file and a DB row in the same operation; the DB row is derived from the markdown line.
  - `zharness db rebuild` reconstructs the index from committed repository content alone, and `.gitignore` no longer contradicts `docs/workflow-harness/migration.md`.
  - `cd cli && go test ./...` and `bash scripts/verify-doc-links.sh` pass at every phase boundary.

## Authority and Requirements
- authority:
  - `docs/audit/consumer-adoption-audit.md` — the Keep verdict and defects D1–D4, measured against consumer repo onedrive-cloud.
  - `cli/internal/application/plan_query.go` — the D1 defect site (`ambiguous_active_plan` hard error).
  - `cli/internal/application/next.go` — the correct `StopInfo` ambiguity contract, and `findActivePlans` which never reads `status:` frontmatter.
  - `cli/internal/application/plan_write.go` — the third caller, which silently degrades, and whose doc comment states the dual-write non-atomicity honestly.
  - `cli/internal/application/scaffold.go` — the only plan-creation guard; it blocks overwrite, not a second active plan.
  - `cli/internal/infrastructure/migrations.go` — the `managed_docs` SHA256 freshness pattern, reused for plan-index staleness.
  - `docs/workflow-harness/migration.md` line 37 — "Only `.kit/changesets/**` must be git-tracked inside `.kit/`", contradicted by `.gitignore` line 2.
  - `refactoring-draft/reference-models.md` — harness-experimental's retirement of its SQLite control plane, its delete-completed-plans rule, and the encoding-invariants diagnostic standard (violating item · broken rule · authority pointer · next action).
  - Owner decision, this session: markdown authoritative, SQLite derived, CLI as guardrail; P0 must ship alone and kill D1 in both directions; no `plan select` command.
- requirements:
  - R1 [accepted]: Creating a plan under `docs/plans/active/` fails when a non-empty active plan already exists there, and the failure names the existing path. | source: `cli/internal/application/scaffold.go`, owner decision
  - R2 [accepted]: All three `findActivePlans` callers share one resolver returning `Stop{Code:"ambiguous"}` on >1 and `Stop{Code:"none"}` on 0; no caller raises `ValidationError` and no caller returns `ok=false` without a Stop. | source: `next.go`, `plan_query.go`, `plan_write.go`
  - R3 [accepted]: Ambiguity resolution reads Tier 0 (index/traces, ~0 tokens), then Tier 1 (first 10 frontmatter lines per candidate), then Tier 2 (bounded candidate packet declaring an `omitted` entry). Plan bodies are never read to disambiguate. | source: owner decision
  - R4 [accepted]: When frontmatter is missing or unparseable, the fallback ordering signal is `git log -1 --format=%cI -- <path>`, never the file body, and `zharness validate` reports the missing frontmatter as a finding. | source: owner decision (premise collapse)
  - R5 [accepted]: Exactly two commands move a plan out of `docs/plans/active/`: `zharness plan complete` and `zharness plan abandon`. No selection command exists. | source: owner decision
  - R6 [accepted]: `docs/playbooks/work.md` step 1 and `docs/playbooks/handoff.md` reference those commands and the Stop contract; no playbook instructs a manual move of an active plan. | source: `docs/playbooks/handoff.md` lines 36–37
  - R7 [accepted]: Dead surfaces are removed, not deprecated: `intervention` (147 LOC, 0 callers) and `next` (2 historical references). | source: measured this session
  - R8 [accepted]: After P2, each of the six dual-write sites writes markdown first and derives the DB row from it; a failed markdown write produces no DB row. | source: `plan_write.go` doc comment, `trace.go`, `decision.go`, `check_record.go`, `handoff.go`
  - R9 [accepted]: `plan_index` stores path plus `sha256`, following the `managed_docs` column shape, and staleness is a 3-way comparison, not a timestamp guess. | source: `cli/internal/infrastructure/migrations.go`
  - R10 [accepted]: `zharness db rebuild` reconstructs the index from committed repository content alone; the changeset layer (1,715 LOC) is retired and `.gitignore` stops contradicting `migration.md`. | source: `migration.md` line 37, `.gitignore` line 2
  - R11 [accepted]: Every new diagnostic follows the encoding-invariants standard: violating item, broken rule, authority pointer, next action — never a bare `validation failed`. | source: `refactoring-draft/reference-models.md` section D
  - R12 [accepted]: Each phase is independently mergeable; P0 alone closes D1 in both directions even if P1–P4 never land. | source: owner decision

## Non-goals
- NG1: No edits to any consumer or reference repository. onedrive-cloud, harness-experimental, and codesight are cited as evidence only; the consumer-side D1 repair is the owner's to run.
- NG2: No `zharness plan select` command, and no pointer file naming the current plan. Both create a second source of truth for a fact the filesystem already states.
- NG3: D2 (unbounded `context.phases`), D3 (consumer state), and D4 (consumer `CLAUDE.md` size) are out of scope for this initiative and stay in `docs/audit/consumer-adoption-audit.md`.
- NG4: No migration path for existing consumer databases beyond `zharness db rebuild`; the index is derived and therefore disposable by construction.
- NG5: No adoption of harness-experimental or codesight as a dependency; only their mechanisms are borrowed.
- NG6: No change to the six-stage pipeline shape, the skill trigger contract, or MIN_ZHARNESS_VERSION gating beyond the version bump each phase requires.

## Approach and Risks
- approach: Invert the source of truth in five independently mergeable phases, ordered so the defect closes first and the architecture follows. P0 makes "exactly one active plan" an enforced invariant instead of an assumed one — it guards creation, unifies the three disagreeing resolvers behind one `Stop` contract, adds the two exits that let a plan leave `active/`, and teaches the ambiguity path a 3-tier ladder that resolves without reading plan bodies. P1 deletes the surfaces that have no callers, which shrinks what P2 has to convert. P2 stops the six dual-writes: markdown becomes the write target and the DB row is derived from it, with a `plan_index` table copying the `managed_docs` SHA256 shape so staleness is a comparison, not a guess. P3 retires the changeset layer once the index is derivable from committed content and repairs the `.gitignore`/`migration.md` contradiction. P4 adds the markdown-schema checks that only make sense once markdown is authoritative. Every phase leaves `cd cli && go test ./...` and `bash scripts/verify-doc-links.sh` green.
- constraints:
  - Consumer and reference repositories are read-only (NG1); every change lands in `/Users/tinhtute/Lab/mono-harness`.
  - `scripts/verify-doc-links.sh` validates backtick path tokens whose first segment is in `skills docs rules cli setup references` — write consumer paths as `onedrive-cloud/docs/...` so they fall outside the allowlist.
  - `application/next.go` holds `StopInfo`, `activePlan`, and `findActivePlans` alongside the dead `Next` command, so P0 must extract the shared resolver before P1 can delete the file.
  - Phase and task definitions are immutable after this stage; only phase lifecycle status changes.
- risks:
  - R-A: P0 changes a public error contract — `query plan` stops returning `ValidationError{ambiguous_active_plan}` and returns a `Stop` instead. Mitigation: keep the non-zero exit and the JSON envelope shape; only the code and payload change; assert both in `query_plan_test.go`.
  - R-B: The 3-tier ladder assumes frontmatter is present and trustworthy. Mitigation is R4 — fall back to `git log -1 --format=%cI -- <path>`, never the file body, and make missing frontmatter a `validate` finding rather than a silent guess.
  - R-C: P2 inverts write ordering at six sites; a partial conversion leaves some facts markdown-first and others DB-first. Mitigation: convert all six in one phase, and assert in a test that a forced markdown-write failure leaves zero DB rows.
  - R-D: P3 retires 690 LOC of changeset code plus its recovery tests; existing local `.kit/changesets/` state becomes unreadable. Mitigation: `zharness db rebuild` must pass from committed content alone before any deletion, and the retired state is moved to trash, never `rm`.
  - R-E: `interventions` is referenced by `validate.go`, `layout_migration.go`, `layout_backfill.go`, `changeset.go`, and `repository_lock.go`, so P1 is a 6-site removal, not a 3-file delete. Mitigation: the phase's own task list names all six sites.

## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Do not add task status fields. Append-only Progress is the sole task execution-status source. Only each phase lifecycle status changes to mirror DB transitions: to-plan=planned; work after run create=in-progress; clean durable check=checked; closing handoff=done. Each planned phase records phase_slug, story_id, status, goal, depends_on, waves, tasks, and checks. -->
- planning_status: planned
- phases:
  - phase_slug: p0-single-active-plan
    story_id: 01M082F6NKJQXZTC9E5KJBVT47
    status: planned
    goal: Close both D1 entry paths — guard plan creation, unify the resolver behind one Stop contract, add the 3-tier disambiguation ladder, add plan complete/abandon, and give work.md a recovery branch.
    depends_on: none
    waves:
      - wave: 1 — extract the shared resolver
        tasks:
          - task: Create `cli/internal/application/plan_resolve.go` and move `StopInfo`, `activePlan`, `nextActivePlansGlob`, and `findActivePlans` out of `next.go` into it, unchanged. Leave `Next` and its helpers in place.
          - task: Add `ResolveActivePlan() (activePlan, *StopInfo, error)` to `plan_resolve.go` — the single entry point returning `Stop{Code:"ambiguous"}` for >1 plan and `Stop{Code:"none"}` for 0, each message following R11 (violating item · broken rule · authority pointer · next action).
        checks:
          - check: `cd cli && go build ./... && go test ./internal/application/...`
      - wave: 2 — unify the three callers
        tasks:
          - task: Rewrite `QueryPlanSection` in `plan_query.go` to call `ResolveActivePlan` and surface the returned Stop instead of raising `ValidationError{Code:"ambiguous_active_plan"}` or `{Code:"no_active_plan"}`.
          - task: Rewrite `activePlanForWrite` in `plan_write.go` to return the Stop instead of `ok=false`, so no caller degrades silently.
          - task: Rewrite `Next` in `next.go` to call `ResolveActivePlan`, deleting its own duplicated ambiguity branch.
          - task: Update `cli/internal/interfaces/query.go` so the Stop path still exits non-zero with the existing JSON envelope shape, and add a case to `query_plan_test.go` asserting exit code and payload for two active plans.
        checks:
          - check: `cd cli && go test ./...`
      - wave: 3 — guard creation
        tasks:
          - task: Extend the guard in `cli/internal/application/scaffold.go` so scaffolding into `docs/plans/active/` fails when any non-empty plan already exists there, with a diagnostic naming the existing path and pointing at `plan complete` / `plan abandon` as the exits.
          - task: Add a test in `scaffold_test.go` proving a second active plan with a different slug is refused, and that scaffolding into an empty `active/` still succeeds.
        checks:
          - check: `cd cli && go test ./internal/application/... -run Scaffold`
      - wave: 4 — the two exits
        tasks:
          - task: Add `zharness plan complete` — set `status: completed`, refresh `updated`, move the file from `docs/plans/active/` to `docs/plans/completed/`, and record the transition. Refuse when the plan has an open phase.
          - task: Add `zharness plan abandon` — same move with `status: abandoned` and a required `--reason`, for a plan that will never ship.
          - task: Register both in `cli/internal/interfaces/root.go` and add them to the read-only exclusion set in `repository_lock.go` so they take the write lock.
          - task: Add tests covering: complete on a clean plan, complete refused with an open phase, abandon without `--reason` refused, and `active/` empty afterward in both success cases.
        checks:
          - check: `cd cli && go test ./...`
      - wave: 5 — the disambiguation ladder and the playbooks
        tasks:
          - task: Implement the 3-tier ladder behind the ambiguous Stop — Tier 0 reads the index/traces already in the packet, Tier 1 reads the first 10 frontmatter lines of each candidate, Tier 2 emits a bounded candidate packet carrying an `OmittedField` that declares the plan bodies it refused to read. Never read a plan body.
          - task: Implement the R4 fallback — when frontmatter is absent or unparseable, order candidates by `git log -1 --format=%cI -- <path>`, and mark the candidate so `validate` can report it.
          - task: Add a test asserting the ambiguous packet stays under 500 tokens with two plans of 1,621 and 410 lines, and that it names both candidates.
          - task: Rewrite `docs/playbooks/work.md` step 1 to branch on the Stop contract — ambiguous and none each get an explicit recovery path naming `plan complete` / `plan abandon` — alongside the existing `degraded: true` branch.
          - task: Rewrite `docs/playbooks/handoff.md` lines 36–37 to call `zharness plan complete` instead of instructing the agent to set `status:` and move the file by hand.
          - task: Update `docs/playbooks/brainstorm.md` step 6 to state that scaffold now refuses a second active plan, and name the two exits.
        checks:
          - check: `cd cli && go test ./... && cd .. && bash scripts/verify-doc-links.sh`
  - phase_slug: p1-cut-dead-surfaces
    story_id: 01M082F6NW6CQCT9EPR3WRYF1A
    status: planned
    goal: Remove the intervention surface (136 LOC across 3 files plus 6 reference sites) and the dead next command, now that P0 owns the resolver.
    depends_on: p0-single-active-plan
    waves:
      - wave: 1 — intervention
        tasks:
          - task: Delete `cli/internal/application/intervention.go`, `cli/internal/domain/intervention.go`, `cli/internal/interfaces/intervention.go` and their tests, using `trash` rather than `rm`.
          - task: Remove the six reference sites — `interfaces/root.go:55` command registration, `interfaces/repository_lock.go:24` lock entry, `application/validate.go:75` UNION branch, `application/layout_backfill.go:66` backfill row, `application/layout_migration.go:278` migration query, and the two `infrastructure/changeset.go` maps at lines 35 and 53.
          - task: Add a migration dropping the `interventions` table, matching how `backlog` and `tools` were already dropped in `migrations.go`.
        checks:
          - check: `cd cli && grep -rn "ntervention" --include="*.go" . | grep -v migrations.go; go test ./...`
      - wave: 2 — next
        tasks:
          - task: Delete `cli/internal/application/next.go` (the resolver already lives in `plan_resolve.go`) and `cli/internal/interfaces/next.go`, plus their tests, using `trash`.
          - task: Unregister `next` from `interfaces/root.go` and drop it from `repository_lock.go` and `read_only_commands_test.go`.
          - task: Mark the two historical mentions in `docs/plans/completed/workflow-harness-history-2026-07.md` and `docs/audit/sdlc-gap-analysis.md` as retired rather than editing the historical record away.
        checks:
          - check: `cd cli && go test ./... && cd .. && bash scripts/verify-doc-links.sh`
  - phase_slug: p2-derive-index
    story_id: 01M082F6P4KY7QN6TPCVW432P8
    status: planned
    goal: Stop the six dual-writes — markdown becomes the write target and the DB row is derived from it — and add plan_index using the managed_docs SHA256 pattern.
    depends_on: p1-cut-dead-surfaces
    waves:
      - wave: 1 — plan_index
        tasks:
          - task: Add a `plan_index` migration in `cli/internal/infrastructure/migrations.go` with `path TEXT UNIQUE NOT NULL`, `sha256 TEXT NOT NULL`, `status TEXT NOT NULL`, `updated_at TEXT NOT NULL`, copying the `managed_docs` column shape at lines 131–137.
          - task: Implement the 3-way staleness comparison (on-disk hash / indexed hash / last-indexed timestamp) reusing the `managed_docs` helper rather than writing a second hashing path.
        checks:
          - check: `cd cli && go test ./internal/infrastructure/...`
      - wave: 2 — invert the six write sites
        tasks:
          - task: Invert `trace.go:60` and `trace.go:131` — write the `## Progress` markdown line first, then derive the DB row from the written line.
          - task: Invert `decision.go:62` — `## Decisions` markdown first, DB row derived.
          - task: Invert `check_record.go:118` and `check_record.go:136` — `## Validation` markdown first, DB row derived.
          - task: Invert `handoff.go:156` — `## Progress` markdown first, DB row derived.
          - task: Replace the honesty disclaimer in `plan_write.go` lines 28–38 with the guarantee the new ordering actually provides.
          - task: Add a test forcing the markdown write to fail (read-only plan file) and asserting zero DB rows were created, for each of the four record kinds.
        checks:
          - check: `cd cli && go test ./...`
      - wave: 3 — refresh on read
        tasks:
          - task: Make the read paths refresh `plan_index` when the on-disk hash differs, so the index can never serve a stale answer without saying so.
          - task: Add a `stale_index` drift finding to `resume.go` alongside the existing four, following the R11 diagnostic standard.
        checks:
          - check: `cd cli && go test ./... && cd .. && bash scripts/verify-doc-links.sh`
  - phase_slug: p3-retire-changesets
    story_id: 01M082F6PC49MBTSXMSVGMRCBD
    status: planned
    goal: Retire the changeset layer, make db rebuild work from committed content alone, and repair the .gitignore contradiction so D5 dissolves.
    depends_on: p2-derive-index
    waves:
      - wave: 1 — prove rebuild first
        tasks:
          - task: Rewrite `zharness db rebuild` to reconstruct every table from committed repository markdown plus `plan_index`, with no read of `.kit/changesets/`.
          - task: Add a test that wipes `harness.db`, rebuilds from a fixture repository containing only committed markdown, and asserts the resulting rows match the pre-wipe state.
        checks:
          - check: `cd cli && go test ./internal/application/... -run Rebuild`
      - wave: 2 — remove the layer
        tasks:
          - task: Delete `cli/internal/infrastructure/changeset.go` (598 LOC) and `cli/internal/application/changeset.go` (92 LOC) with their recovery tests, using `trash`.
          - task: Remove changeset writes from every lifecycle command and drop the changeset scaffolding from `zharness init`.
        checks:
          - check: `cd cli && go test ./...`
      - wave: 3 — repair the contradiction
        tasks:
          - task: Replace the bare `.kit/` entry on line 2 of `.gitignore` with entries that ignore only per-machine state, now that changesets no longer exist.
          - task: Update `docs/workflow-harness/migration.md` line 37 — the changeset-tracking requirement it states is retired; replace it with the committed-markdown guarantee.
          - task: Update `CLAUDE.md`'s state paragraph so the repo's three descriptions of `.kit/` finally agree.
        checks:
          - check: `git check-ignore -v .kit/ ; cd cli && go test ./... && cd .. && bash scripts/verify-doc-links.sh`
  - phase_slug: p4-markdown-schema
    story_id: 01M082F6PK73W3DP3N4QMYWFEA
    status: planned
    goal: Add markdown-schema checks to validate and bound query plan --section, now that markdown is authoritative.
    depends_on: p3-retire-changesets
    waves:
      - wave: 1 — validate the filesystem
        tasks:
          - task: Add checks 9–12 to `cli/internal/application/validate.go` — exactly one file in `docs/plans/active/`; required frontmatter keys present and parseable; every phase in `## Phases and Verification` has a matching story row; every append-only section is one of the four known headings.
          - task: Emit each finding in the R11 format — violating item, broken rule, authority pointer, next action — and add a negative fixture per check proving it fires.
        checks:
          - check: `cd cli && go test ./internal/application/... -run Validate`
      - wave: 2 — bound the section read
        tasks:
          - task: Cap `QueryPlanSection`'s not-found degrade path in `plan_query.go` so it returns the section list plus an `OmittedField` instead of the whole file body.
          - task: Add a test asserting a section miss on the 1,621-line fixture returns under 500 tokens and names the available sections.
        checks:
          - check: `cd cli && go test ./... && cd .. && bash scripts/verify-doc-links.sh`

## Progress
<!-- Append-only durable entries record timestamp, phase, wave, task, task_status, run_id, trace_id, exact verification/result, and changed surfaces or blocker. -->
- `2026-08-17T15:05:48.456Z` — handoff recorded. handoff: `01M0843RN8RAH4P549K6TBXBPX`. next action: work full p0-single-active-plan. open items: P5 (durable-memory) and P6 (retrieval-router) are decision-complete and approved but not yet appended to the plan — run: to-plan phase p5-durable-memory, then to-plan phase p6-retrieval-router; P0 has no run yet; no implementation code has been written for this initiative; onedrive-cloud has two plans with status: active since 2026-08-16 — owner must call which is live (ui-ux-audit-remediation.md, ten phases shipped in 0b32adb) vs dead (check-review-remediation.md, untouched 18 days); consumer-side repair only, does not block P0; DB story goals for p1-cut-dead-surfaces and p3-retire-changesets carry pre-measurement LOC figures (147 and 1,715); the plan file holds the measured values (136 LOC across 3 files plus 6 reference sites, and 690 LOC core) — trust the plan file; P5 must follow P3: deleting a completed plan is only reversible through git once db rebuild works from committed content alone.

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- none

## Validation
<!-- Append-only durable entries record timestamp, phase, exact command/result/output, run_id, check_id, verdict, and proof_gaps. -->
- none

## Current State and Next Action
- active_phase: p0-single-active-plan
- lifecycle_status: planned
- latest_run_id: none
- latest_trace_ids: []
- latest_check_id: none
- latest_handoff_id: 01M0843RN8RAH4P549K6TBXBPX
- blockers: none
- open_items:
  - P5 (durable-memory) and P6 (retrieval-router) are decision-complete and approved but not yet appended — run `to-plan phase p5-durable-memory`, then `to-plan phase p6-retrieval-router`. P0–P4 stay immutable.
  - P0 has no run yet; no implementation code has been written for this initiative.
  - onedrive-cloud has two plans with `status: active` since 2026-08-16 — the owner must call which is live (`onedrive-cloud/docs/plans/active/ui-ux-audit-remediation.md`, ten phases shipped in commit 0b32adb) versus dead (`onedrive-cloud/docs/plans/active/check-review-remediation.md`, untouched 18 days). Consumer-side repair only; does not block P0.
  - The DB story goals for `p1-cut-dead-surfaces` and `p3-retire-changesets` carry pre-measurement LOC figures (147 and 1,715). The measured values are in this file: 136 LOC across 3 files plus 6 reference sites, and 690 LOC core. Trust this file.
  - P5 must follow P3 — deleting a completed plan is only reversible through git once `db rebuild` works from committed content alone.
- exact_next_action: work full p0-single-active-plan
