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
    status: checked
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
          - task: Wire story status transitions (story create, and every `update story` changeset write in `check_record.go`/`handoff.go`) to also write the new status into the matching phase block's `- status:` line in `## Phases and Verification`, reusing `plan_write.go`'s atomic-write pattern — stories become markdown-first the same way trace/decision/check/handoff did in P2.
          - task: Rewrite `zharness db rebuild` to reconstruct `stories` from phase blocks, `traces`/`decisions`/`checks`/`handoffs` from `## Progress`/`## Decisions`/`## Validation`, and `intakes` from every plan's `intake_id`/`lane` frontmatter — scanning `docs/plans/{active,completed}/*.md` plus `plan_index`, with no read of `.kit/changesets/`.
          - task: Reconstruct `runs` only from backreferences (`` run: `<id>` `` in Progress/Validation entries); a run with no backreference is not reconstructed — nothing durable in committed content proves it existed, consistent with markdown-as-truth. Leave `meta` pointers (`current_phase`, `latest_run_id`, `latest_check_id`, `docs_version`) unset after rebuild — already a supported state via the existing `unknown_phase`/`out_of_order`/`stale_docs` drift tolerance, so no new pointer-recovery logic is needed.
          - task: Add a test that wipes `harness.db`, rebuilds from a fixture repository containing only committed markdown, and asserts stories/traces/decisions/checks/handoffs/intakes (and any run reachable via a backreference) match the pre-wipe state.
        checks:
          - check: `cd cli && go test ./internal/application/... -run Rebuild`
      - wave: 2 — remove the layer
        tasks:
          - task: Delete `cli/internal/infrastructure/changeset.go` (598 LOC) and `cli/internal/application/changeset.go` (92 LOC) with their recovery tests, using `trash`.
          - task: Remove changeset writes from every lifecycle command — trace/decision/check/handoff already write markdown from P2; story create/run create/intake/import now write their markdown equivalent from wave 1 instead — and drop the changeset scaffolding from `zharness init`.
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
- `2026-08-17T15:09:36Z` — wave 1. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Phase p0-single-active-plan started (in-progress).
- `2026-08-17T15:11:29Z` — wave 1, task Create plan_resolve.go, move StopInfo/activePlan/nextActivePlansGlob/findActivePlans out of next.go unchanged. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Moved 4 symbols to new cli/internal/application/plan_resolve.go; next.go still has Next/resolveFullMode/parseNextArgument/etc. Removed now-unused os/path/filepath/sort imports from next.go..
- `2026-08-17T15:11:29Z` — wave 1, task Add ResolveActivePlan() single entry point returning Stop{ambiguous}/Stop{none}. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Implemented in plan_resolve.go with R11-format messages (violating item, broken rule + D1 authority pointer, next action). go build ./... && go test ./internal/application/... pass..
- `2026-08-17T15:11:35Z` — wave 1. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Wave 1 done: resolver extracted to plan_resolve.go, ResolveActivePlan added; go build+test green.
- `2026-08-17T15:15:42Z` — wave 2, task Rewrite QueryPlanSection to call ResolveActivePlan and surface Stop instead of raising ambiguous_active_plan/no_active_plan ValidationError. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: plan_query.go signature now (PlanSectionView, *StopInfo, error); unknown_section/missing_required_field stay real ValidationErrors. Updated 8 call sites + 2 rewritten tests in plan_query_test.go..
- `2026-08-17T15:15:42Z` — wave 2, task Rewrite activePlanForWrite to return Stop instead of ok=false. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: plan_write.go: activePlanForWrite now (path string, stop *StopInfo, err error) via ResolveActivePlan; preparePlanAppend adapted, same no-op-write behavior on zero/ambiguous preserved (dual-write ordering fix is P2/R8, out of scope)..
- `2026-08-17T15:15:42Z` — wave 2, task Rewrite Next to call ResolveActivePlan, deleting its own duplicated ambiguity branch. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: next.go: removed local plans/len(plans)>1 block; ambiguous Stop now reused directly from ResolveActivePlan (message/code identical, tests unchanged). auto-mode-no-plan-falls-back-to-simple and full-mode no-plan/no-plan code preserved as next-specific logic on top of stop.Code=="none"..
- `2026-08-17T15:15:42Z` — wave 2, task Update query.go Stop path to exit non-zero with existing JSON envelope shape; add ambiguous-plan test. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Added mapStop() in errors.go (StopInfo->cliError, exit 1) and wired it into runQuery plan branch; added TestQueryPlanAmbiguousActivePlanExitsNonZero in interfaces/query_plan_test.go asserting code=ambiguous, exit=1, message names both plans. go test ./... passes all packages..
- `2026-08-17T15:15:44Z` — wave 2. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Wave 2 done: three callers unified behind ResolveActivePlan; query.go Stop path preserves prior exit-1 envelope; go test ./... green.
- `2026-08-17T15:16:54Z` — wave 3, task Extend scaffold.go guard so scaffolding into docs/plans/active/ fails when a non-empty plan already exists. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Added isActivePlanDir + existingActivePlanPath (reuses findActivePlans), guard fires only for kind=="plan" whose dir is docs/plans/active; ValidationError{Code:"active_plan_exists"} names the existing path and both exits (plan complete/abandon)..
- `2026-08-17T15:16:54Z` — wave 3, task Add scaffold_test.go coverage: second active plan refused, empty active/ still succeeds. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Added TestScaffoldArtifact_RefusesSecondActivePlan and TestScaffoldArtifact_EmptyActiveDirStillSucceeds. go test ./internal/application/... -run Scaffold: all 14 pass..
- `2026-08-17T15:16:54Z` — wave 3. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Wave 3 done: scaffold guards a second active plan; -run Scaffold green.
- `2026-08-17T15:24:22Z` — wave 4, task Add zharness plan complete — set status: completed, refresh updated, move file to docs/plans/completed/, record the transition. Refuse when the plan has an open phase.. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Added PlanComplete (plan_lifecycle.go): resolves active plan via ResolveActivePlan, refuses via ValidationError{open_phase} when selectActivePhase(db, slugs) returns a non-done phase, else rewrites frontmatter status/updated (setFrontmatterFields, scoped to lines between the two --- delimiters) and appends a ## Decisions entry via AppendToPlanSection, then moves the file with writeFileAtomically+os.Remove..
- `2026-08-17T15:24:22Z` — wave 4, task Add zharness plan abandon — same move with status: abandoned and a required --reason, for a plan that will never ship.. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Added PlanAbandon(reason): ValidationError{missing_required_field} when --reason is blank, otherwise same transitionActivePlan move with status: abandoned and the reason recorded in the Decisions entry; no open-phase check (abandon means never finishing)..
- `2026-08-17T15:24:22Z` — wave 4, task Register both in cli/internal/interfaces/root.go and add them to the read-only exclusion set in repository_lock.go so they take the write lock.. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Added interfaces/plan.go (newPlanCmd, complete/abandon subcommands, mapValidationError/mapStop error mapping mirroring run.go); registered root.AddCommand(newPlanCmd()) in root.go; added plan complete and plan abandon to exclusiveMutationCommandPaths in repository_lock.go — wrapExclusiveMutationCommands inventory check passes (go test ./internal/interfaces/... green, no panic)..
- `2026-08-17T15:24:22Z` — wave 4, task Add tests covering: complete on a clean plan, complete refused with an open phase, abandon without --reason refused, and active/ empty afterward in both success cases.. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Added plan_lifecycle_test.go: TestPlanCompleteMovesCleanPlan, TestPlanCompleteRefusedWithOpenPhase, TestPlanAbandonRequiresReason, TestPlanAbandonMovesPlanRegardlessOfPhaseStatus — all assert docs/plans/active/ is empty after a successful transition. go test ./... passes across all 6 packages..
- `2026-08-17T15:24:26Z` — wave 4. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Wave 4 done: plan complete/abandon exits implemented, registered, write-locked; go test ./... green across all packages..
- `2026-08-17T15:28:37Z` — wave 5, task Implement the 3-tier ladder behind the ambiguous Stop — Tier 0, Tier 1 (first 10 frontmatter lines), Tier 2 (bounded candidate packet carrying an OmittedField). Never read a plan body.. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: plan_resolve.go: buildAmbiguousStop replaces the plain ambiguous Stop; Tier 0 is documented as already-free agent context (no code needed); Tier 1 is frontmatterPreview (first 10 lines between the --- delimiters) + frontmatterPreviewField(updated); Tier 2 packs StopInfo.Candidates ([]PlanCandidate: path/updated/ordered_by/frontmatter_ok) plus OmittedField="plan bodies" into the Message, sorted newest-first..
- `2026-08-17T15:28:37Z` — wave 5, task Implement the R4 fallback — when frontmatter is absent or unparseable, order candidates by git log -1 --format=%cI -- <path>, and mark the candidate so validate can report it.. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: resolveCandidate falls back to gitLogCommitTime (exec git log -1 --format=%cI, metadata only) when frontmatterPreview fails or has no updated: line, setting OrderedBy=git_log_fallback; FrontmatterOK=false is left on the candidate for a future P4 validate check to consume..
- `2026-08-17T15:28:37Z` — wave 5, task Add a test asserting the ambiguous packet stays under 500 tokens with two plans of 1,621 and 410 lines, and that it names both candidates.. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: plan_resolve_test.go: TestResolveActivePlanAmbiguousPacketStaysBounded builds 1621/410-line fixtures, asserts stop.Message under 500 tokens (~2000 chars), names both paths, OmittedField set, newest-first ordering by frontmatter updated:. TestResolveActivePlanAmbiguousFallsBackToGitLogOrdering covers R4 for a candidate with no frontmatter block..
- `2026-08-17T15:28:37Z` — wave 5, task Rewrite docs/playbooks/work.md step 1 to branch on the Stop contract — ambiguous and none each get an explicit recovery path naming plan complete / plan abandon — alongside the existing degraded: true branch.. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Added an ambiguous/none sub-branch to work.md step 1, naming the Candidates packet and the plan complete/plan abandon and brainstorm lock recoveries..
- `2026-08-17T15:28:37Z` — wave 5, task Rewrite docs/playbooks/handoff.md lines 36-37 to call zharness plan complete instead of instructing the agent to set status: and move the file by hand.. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Replaced the two manual status:/move bullets in step 6 with a single zharness plan complete --json call; added it to the Command Reference list..
- `2026-08-17T15:28:37Z` — wave 5, task Update docs/playbooks/brainstorm.md step 6 to state that scaffold now refuses a second active plan, and name the two exits.. task_status: `DONE`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Extended the scaffold plan bullet in step 6 to document the active_plan_exists refusal and name plan complete/plan abandon as the exits..
- `2026-08-17T15:28:45Z` — wave 5. run: `01M084AAB3Y8MBBAHJ6HA6897J`. summary: Wave 5 done: 3-tier disambiguation ladder (Tier 1 frontmatter preview, Tier 2 bounded packet, R4 git-log fallback) implemented in plan_resolve.go; work.md/handoff.md/brainstorm.md playbooks updated; go test ./... and verify-doc-links.sh both green. Phase p0-single-active-plan waves 1-5 complete..

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- `2026-08-17T15:33:00Z` — Copied wave-5 playbook edits (work.md/handoff.md/brainstorm.md) from docs/playbooks/ into cli/docs/embedded/playbooks/, the actual source of truth for the projection-drift test (phase: `p0-single-active-plan`), task: gate. rationale: Wave 5 edited the root doc projection directly instead of the embedded source; TestProjectionDrift_RootDocsMatchEmbed caught the mismatch during the phase gate (go test ./... -count=1). Embedded copy is authoritative — root docs are a generated projection..
- `2026-08-18T00:00:00Z` — Executed p1-cut-dead-surfaces without full-mode lifecycle ceremony (phase: `p1-cut-dead-surfaces`), task: gate. rationale: this worktree's harness.db and .kit/changesets were both absent (fresh worktree checkout — per-machine state, gitignored, never carried P0's run/check history), so the P0-close/P1-run-create/decision-add/trace-add chain had no DB rows to attach to and `db rebuild` had no changesets to rebuild from. Owner approved skipping the ceremony and executing the phase's task list directly against the plan definition, verified by the phase's own check commands. Deviations from the written task list, discovered only by executing it: `checkExists` (intervention.go) is also called by handoff.go — moved to check_record.go, next to `runExists`'s equivalent placement. `selectActivePhase`/`parseActivePlanPhaseOrder`/`activePlanPhaseSlug` (next.go) are also called by plan_lifecycle.go's `PlanComplete` (a P0 wave-4 addition the P1 task list predates) — moved to plan_lifecycle.go. `chdirFixture`/`writeFile`/`writeActivePlan`/`seedStory` (next_test.go) are shared test fixtures used by decision_test.go, handoff_test.go, check_record_test.go, plan_query_test.go, scaffold_test.go, plan_lifecycle_test.go, plan_resolve_test.go, and trace_test.go — moved to helpers_test.go. The wave-2 task "mark the two historical mentions in docs/plans/completed/workflow-harness-history-2026-07.md and docs/audit/sdlc-gap-analysis.md as retired" was skipped as inapplicable: neither file exists in this worktree (docs/plans/completed/ doesn't exist at all; docs/audit/ has only consumer-adoption-audit.md) — pre-existing drift from master's docs/ deletion (655c6ac), unrelated to P1. `cd cli && go test ./...` and `bash scripts/verify-doc-links.sh` both green after the phase's own two check commands passed per-wave; the pre-existing 16 broken doc cross-references (verified present before this phase's changes too) are untouched by this phase..
- `2026-08-18T02:43:30Z` — Executed p2-derive-index wave 3 (refresh on read) by refreshing plan_index on the write path, not inside resume.go, task: gate. rationale: This worktree's harness.db has no story/run for p2-derive-index (same fresh-worktree gap P1 hit), so decisions/traces are recorded without phase/run linkage, matching the owner-approved P1 precedent. Wave-3 task 1 (refresh plan_index on read) was implemented inside preparePlanAppend's write closure (trace/decision/check/handoff call sites) instead of resume.go, because resume opens its db read-only in production (interfaces/resume.go OpenReadOnly) — an initial attempt to call the writing refreshPlanIndex from resume.go broke 5 internal/interfaces tests with "attempt to write a readonly database (8)". resume.go instead calls the existing read-only planIndexStaleness comparison to detect drift and emits a new stale_index DriftFinding per the R11 diagnostic standard (violating item: plan_index missing/out of date; broken rule: R9 3-way hash comparison; authority pointer: docs/plans/active/harness-markdown-truth.md; next action: record any trace/decision/check/handoff to refresh, or confirm the file changed outside the CLI). cd cli && go test ./... and bash scripts/verify-doc-links.sh both pass; the pre-existing 16 broken doc cross-references (documented in P1's decision entry, drift from master's docs/ deletion 655c6ac) are untouched by this phase..
- `2026-08-18T03:11:32Z` — Amended P3 wave 1-2 task text before implementing (task list is normally immutable after to-plan; owner approved this deviation), task: wave 1. rationale: Wave 1 as originally written ("reconstruct every table from committed repository markdown plus plan_index, with no read of .kit/changesets/") assumed stories/runs/intakes/meta already had a markdown representation. Grepping story_create.go, run_create.go, intake.go, import.go, and init.go for preparePlanAppend/writeFileAtomically found none — only trace/decision/check/handoff became markdown-first in P2. Since wave 2 as originally written deletes the entire changeset layer and removes changeset writes from every lifecycle command, following the original text as-is would have left stories/runs/intakes/meta with no durable source of truth at all post-P3. Resolved by reframing rather than inventing new sections: stories piggyback on the existing per-phase story_id/status fields already in `## Phases and Verification` (only the write-back was missing); intakes are already fully derivable from every plan file's existing intake_id/lane frontmatter; runs are reconstructed only from backreferences (`run: `id`` in Progress/Validation entries) and a run with zero footprint is correctly dropped under this initiative's own markdown-is-truth philosophy, not a bug; meta pointers (current_phase/latest_run_id/latest_check_id/docs_version) are left unset post-rebuild, which is already a supported and tested state via the existing unknown_phase/out_of_order/stale_docs drift tolerance (TestResumeStaleDocsMissing and friends). Wave 1 and wave 2 task text updated to reflect this; R8-R12 and every other phase are unchanged..
- `2026-08-18T06:33:19Z` — db rebuild reconstructs stories/intakes/checks/runs/traces/handoffs/decisions from committed plan markdown alone (docs/plans/{active,completed}/*.md), no read of .kit/changesets/, task: Rewrite zharness db rebuild (p3-retire-changesets) to reconstruct all tables from phase blocks and Progress/Decisions/Validation/frontmatter, no changeset read. rationale: markdown is the source of truth per P3; traces/decisions get freshly minted ids on rebuild since formatTraceProgressEntry/formatDecisionEntry never embed one; a run is only reconstructed when a Validation (check) entry backreferences it, since that is the only entry shape carrying story_slug (runs.story_slug NOT NULL); intakes.type/summary have no markdown home and are synthesized placeholders, while intakes.id/plan_id/lane are recovered exactly via a plan frontmatter's intake_id/id/lane fields; meta pointers are left unset since no committed markdown proves which run/check is latest..
- `2026-08-18T07:50:16Z` — Completed P3 waves 2 and 3: retired the changeset write path (story/run/check/handoff/decision/trace/intake/import/layout-migration/init all write direct SQL now; infrastructure/changeset.go, application/changeset.go, `db changeset apply`/`db changeset status`, and ~30 dependent test files deleted or converted) and repaired the `.gitignore`/`CLAUDE.md` `.kit/` contradiction (D5), task: wave 2 and wave 3. rationale: This worktree's harness.db was again absent (same fresh-worktree gap P1/P2/wave-1 hit), so this entry is recorded directly in the plan rather than via `zharness decision add`, per the same owner-approved precedent. CreateRun/RecordCheck/RecordHandoff switched from changeset-derived ULID ordering to independent `ulid.Make()`+`time.Now()` minting, wrapped in explicit `db.Begin`/`tx.Exec`/`tx.Commit` to preserve the atomicity `ApplyChangeset`'s single transaction used to provide for each function's multi-row write (run+story+meta; check+story+meta; handoff+story). `db_status.go`'s `Pending`/`UnverifiedBelowFence`/`Fence` fields were dropped (changeset-only concepts with no successor); `import.go` dropped `changesetDir`/`ChangesetsWritten`; `CONTRACT.md` updated to match the real current `import`/`db status`/`migrate layout` JSON shapes. Converting layout_backfill.go/layout_migration.go off changeset-replay surfaced a real pre-existing data-loss bug — `migrate layout` was silently dropping `checks.judge`/`judge_model`, `traces.task`/`task_status`, and the entire `decisions` table during a v1→v2 layout migration — fixed as part of this wave, caught only because two tests' before/after DB comparisons started failing once the surrounding replay logic was converted. `.gitignore` line 2's bare `.kit/` was replaced with per-machine-state-only entries (`.kit/harness.db`, `.kit/cache/`, `.kit/conflicts/`, `.kit/log/`); the stale local `.kit/changesets/` directory this left untracked was moved to trash (never `rm`, per R-D's explicit mitigation). `docs/workflow-harness/migration.md` line 37 — the doc that stated the original `.gitignore` contradiction — could not be updated: the file was deleted by an earlier unrelated commit (655c6ac), part of the same pre-existing 16-broken-doc-link drift already documented in P1's decision entry; CLAUDE.md's `.kit/` paragraph was rewritten instead, and the dangling `docs/workflow-harness/migration.md` cross-reference in that same paragraph is left as-is, already covered by that drift baseline. `cd cli && go test ./... -count=1`, `go vet ./...`, and `bash scripts/verify-doc-links.sh` all pass; the pre-existing 16 broken doc cross-references are unchanged. P3 is now complete: all three waves shipped and verified..

## Validation
<!-- Append-only durable entries record timestamp, phase, exact command/result/output, run_id, check_id, verdict, and proof_gaps. -->
- `2026-08-17T15:33:45.891Z` — check. verdict: `APPROVED`. check: `01M085PYS31ZEN0RYJ6E8N7GX5`. run: `01M084AAB3Y8MBBAHJ6HA6897J`. phase: `p0-single-active-plan`. judge: `same-session` (claude-sonnet-5).
  - `cd cli && go test ./... -count=1` → all 7 packages ok, including TestProjectionDrift_RootDocsMatchEmbed
  - `bash scripts/verify-doc-links.sh` → doc links OK (0 findings)

## Current State and Next Action
- active_phase: p0-single-active-plan
- lifecycle_status: checked
- latest_run_id: 01M084AAB3Y8MBBAHJ6HA6897J
- latest_trace_ids: [01M085DSPJA6WQ9041AH0231H4]
- latest_check_id: 01M085PYS31ZEN0RYJ6E8N7GX5
- latest_handoff_id: 01M0843RN8RAH4P549K6TBXBPX
- blockers: none
- open_items:
  - P5 (durable-memory) and P6 (retrieval-router) are decision-complete and approved but not yet appended — run `to-plan phase p5-durable-memory`, then `to-plan phase p6-retrieval-router`. P0–P4 stay immutable.
  - onedrive-cloud has two plans with `status: active` since 2026-08-16 — the owner must call which is live (`onedrive-cloud/docs/plans/active/ui-ux-audit-remediation.md`, ten phases shipped in commit 0b32adb) versus dead (`onedrive-cloud/docs/plans/active/check-review-remediation.md`, untouched 18 days). Consumer-side repair only; does not block P0.
  - The DB story goals for `p1-cut-dead-surfaces` and `p3-retire-changesets` carry pre-measurement LOC figures (147 and 1,715). The measured values are in this file: 136 LOC across 3 files plus 6 reference sites, and 690 LOC core. Trust this file.
  - P5 must follow P3 — deleting a completed plan is only reversible through git once `db rebuild` works from committed content alone.
  - check verdict `APPROVED` was same-session (I authored and gated the P0 diff); the complete manual Security/Performance/Architecture/Code Quality review (`check full`) was explicitly not performed here per `work.md` step 11 — it is deferred to the initiative's final phase as a `handoff` closure precondition.
- exact_next_action: handoff — close phase p0-single-active-plan (checked in both DB and plan, clean APPROVED gate, no blockers) via `zharness handoff record --run-id 01M084AAB3Y8MBBAHJ6HA6897J --check-id 01M085PYS31ZEN0RYJ6E8N7GX5 --open-items '[]' --close-phase`, unlocking p1-cut-dead-surfaces
