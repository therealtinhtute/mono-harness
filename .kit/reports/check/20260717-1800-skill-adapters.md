# CHECK REPORT

Run ID: check-20260717-1800-skill-adapters
Scope: full
Artifact Alignment: aligned
Review Verdict: APPROVE with requests
Phase: skill-adapters
Spec: .kit/planning/SPEC.md
Plan: .kit/planning/phases/skill-adapters/skill-adapters-PLAN.md
Workflow State: none (retired this phase — harness `zharness query state`/`query phases` is now the durable index; see Findings)
Cook Run: .kit/runs/work/20260717-1630-skill-adapters.md
Created At: 2026-07-17 18:00

## Scope

- depth: standard (`git diff --stat HEAD -- skills/workflow/brainstorm skills/workflow/to-plan skills/workflow/work skills/workflow/README.md` → 7 files changed, 120 insertions(+), 18 deletions(-))
- drift: none — every changed file traces to skill-adapters-PLAN.md's T1–T4 (`brainstorm/SKILL.md` + `spec-template.md` [T1], `to-plan/SKILL.md` + `roadmap-template.md` [T2], `work/SKILL.md` + `run-artifact-template.md` [T3+T4 amendment], `skills/workflow/README.md` [new Version Gate section — not literally in Allowed Surfaces, but directed by CONTEXT.md's own Locked Decisions: "constant documented in skills/workflow/README.md"])
- the raw working-tree diff (`git status -sb`) also shows the entirety of harness-contracts/cli-core/cli-domain's uncommitted work, since nothing has been committed across any phase yet — same known artifact of deferred commits already flagged in cli-core's and cli-domain's own check reports, not new scope drift

## Gate Evidence
- tests: n/a — no test suite for `skills/workflow/**` (markdown-only diff, no `cli/**` touched)
- types: n/a
- lint: `bash scripts/validate-skill.sh skills/workflow/brainstorm/SKILL.md` → PASS; `bash scripts/validate-skill.sh skills/workflow/to-plan/SKILL.md` → PASS; `bash scripts/validate-skill.sh skills/workflow/work/SKILL.md` → PASS WITH WARNINGS (156/150 lines, soft token-efficiency warning only, not a hard fail)
- build: n/a
- gate block identity: `awk` extraction of the `<version-gate>...</version-gate>` block from all 3 rewritten SKILL.md files → byte-identical text in all three (PLAN.md's own Risk/Watch-for: "must share the identical gate block — copy exactly, don't paraphrase")
- zero-yml-writes claim: `grep -n 'workflow-state' skills/workflow/brainstorm/SKILL.md skills/workflow/to-plan/SKILL.md skills/workflow/work/SKILL.md` → empty (exit 1)
- live integration proof (T4): built `zharness` (`dev`) from `cli/`, ran a full scratch dry-run (brainstorm-lock → to-plan full → work full) against a throwaway project outside this repo. Final: `zharness validate --json` → `valid:true`, exit 0 (one finding, the pre-existing documented SPEC->PLAN gap); `zharness query state --json` → `current_phase`/`entry_phase`/`latest_run_id` all correctly populated; `zharness resume --json` → `readiness:"in-progress"`, `drift:[]`; `find . -name workflow-state.yml` → empty

## Artifact Alignment
- status: aligned
- notes:
  - all 4 tasks (T1–T4) logged DONE in the run artifact with changed files, verification commands, and outcomes matching each task's declared scope in skill-adapters-PLAN.md
  - Scope Boundary respected: no edits to `check`/`watzup`/`handoff`/`git`/`interview` skills or `cli/**` (both Forbidden Surfaces); `workflow-state-template.yml` deletion correctly deferred to Phase 8 per CONTEXT.md's own Scope Boundary note
  - `.kit/implementation-notes.md` carries dated entries for every off-spec decision this phase: T1's brainstorm-needs-init gap, `MIN_ZHARNESS_VERSION` invention, T2's roadmap-template scope extension, T3's run-registration gap (Gap #1), T4's current_phase/entry_phase gap (Gap #2) and latest_run_id gap (Gap #3)
  - Gap #3's fix directly closes a Minor finding carried forward from the prior phase's own check report (`.kit/reports/check/20260717-1600-cli-domain.md`: "no ported command advances `meta.current_phase`/`latest_run_id`/`latest_check_id` going forward... carried forward from T3's own note as a gap `skill-adapters` will likely need to own") — confirms this phase picked up exactly the gap the previous gate flagged for it, rather than a new surprise

## Findings

### Critical
- none found

### Major
- none found

### Minor / Suggestions
- 💡 three real, same-class gaps were discovered and fixed during T3/T4 rather than anticipated in skill-adapters-PLAN.md: (1) no live command creates a `runs` row — `work` now authors one via `db changeset apply`; (2) no live command sets `meta.current_phase`/`entry_phase` — `to-plan` now authors one; (3) no live command sets `meta.latest_run_id` — folded into `work`'s run-registration changeset as a second JSONL line. All three reuse the already-shipped, generic `db changeset apply` command (zero `cli/**` changes) and are confirmed by `SCHEMA.md`'s own producer-mapping table as the intended (if previously unwired) design, not inventions. Fully logged in `.kit/implementation-notes.md` with Decision/Spec gap/Tradeoff/Risk/Verification for each.
- 💡 `skills/workflow/README.md` was edited (Version Gate section) even though it isn't literally listed in skill-adapters-CONTEXT.md's Allowed Surfaces list — justified because the same CONTEXT.md's Locked Decisions explicitly names that file as where `MIN_ZHARNESS_VERSION` should be documented; flagging for visibility rather than treating as silent scope creep
- 💡 pre-existing `plan_id` cross-link gap (PLAN artifacts carry no frontmatter at all) remains untouched — already flagged in cli-domain's own notes and CONTRACT.md's Known Gap text; correctly out of scope for skill-adapters
- 💡 doc-debt items carried forward unchanged from prior phases (Security note on changeset trust boundaries, `query phases`/`query artifacts --json` shape lock, Error Codes table amendments for `no_check_found`/`invalid_input`/`invalid_proof_links`) — still not actioned, still not newly blocking

## Knowledge Sync
- no new safety-relevant invariant introduced this phase
- doc debt: unchanged from cli-domain's list (see above) — no new items besides the already-tracked ones; recommend bundling all into one CONTRACT.md/SCHEMA.md amendment pass during `validation-gate` or `continuity` rather than trickling in phase by phase (same recommendation cli-domain's report already made)

## Next Action
- advance `current_phase` from `skill-adapters` to `validation-gate` per the standing pipeline
- `cli/v0.1.0` tag/release remains deliberately deferred from cli-core, unchanged this phase
- carry forward: bundle the accumulated doc-debt list into one amendment pass before pilot-migration
