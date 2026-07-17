---
id: 01KXR93JGZTBHA5BVYPNAJ0N0X
type: run
phase: pilot-migration
lane: high-risk
plan_id: pilot-migration-PLAN.md
trace_ids: []
created: 2026-07-17
updated: 2026-07-17
---

# COOK RUN

Run ID: work-20260717-2200-pilot-migration
Mode: full
Status: passed
Spec: .kit/planning/SPEC.md
Roadmap: .kit/planning/ROADMAP.md
Phase: pilot-migration
Plan: .kit/planning/phases/pilot-migration/pilot-migration-PLAN.md
Started At: 2026-07-17 22:00

## Preflight
- scope drift: no — `.gitignore` narrowing was a pre-approved (AskUserQuestion) blocker fix, not scope creep; `.kit/` itself becoming git-tracked is the pilot's own deliverable
- working tree note: `.gitignore` modified (approved), `.kit/` untracked pending this phase's own commit
- required artifacts present: yes (SPEC, ROADMAP, pilot-migration CONTEXT/PLAN all refreshed 2026-07-17)
- selected phase: pilot-migration (full mode)

## Wave / Task Log

### Wave 1
#### T1 — Execute pilot chain (import this repo's own legacy state)
- status: DONE
- changed files:
  - `.gitignore` (narrowed `.kit/**` → `.kit/harness.db` + `.kit/cache/`, user-approved)
  - `.kit/harness.db` (new, gitignored)
  - `.kit/changesets/01KXR8RR9GAT35QFS05NCRJN8F.changeset.jsonl` (create story `continuity`)
  - `.kit/changesets/01KXR8RR9HFV81RGAQ42V898M3.changeset.jsonl` (create story `harness-concept`)
  - `.kit/changesets/01KXR8RR9HFV81RGAQ48PD21BE.changeset.jsonl` (create story `pilot-migration`)
  - `.kit/changesets/01KXR8RR9J1P5SP9KJD1R57MVZ.changeset.jsonl` (create run, continuity)
  - `.kit/changesets/01KXR8RR9J1P5SP9KJD32K4APV.changeset.jsonl` (update meta: current_phase/entry_phase/latest_run_id)
- verification:
  - `cd /Users/tinhtute/Lab/skills && /tmp/zharness init --json` → `{"db_path":".kit/harness.db","schema_version":1,"status":"created"}`
  - `/tmp/zharness import --json` → `{"imported":5,"skipped":0,"changesets_written":[5 ULID paths]}`
  - `/tmp/zharness query state --json` → `{"current_phase":"pilot-migration","entry_phase":"harness-concept","schema_version":1,"latest_run_id":"01KXR8RR9J1P5SP9KJCY27BHMK","latest_check_id":null}` — matches this repo's pre-import `workflow-state.yml` (`current_phase: pilot-migration`, `entry_phase: harness-concept`) exactly
  - `/tmp/zharness resume --json` → `{"position":{"current_phase":"pilot-migration","status":"planned"},"latest_run_id":"01KXR8RR9J1P5SP9KJCY27BHMK","latest_check_id":null,"latest_handoff_id":null,"drift":[],"readiness":"in-progress"}`
  - `sqlite3 .kit/harness.db "select id, slug, status from stories;"` → 3 rows: continuity (in-progress), harness-concept (planned), pilot-migration (planned) — matches FK-driven minimal import (not a bug; `import.go` only creates stories for slugs referenced by current_phase/entry_phase/latest run's phase, confirmed by reading source doc comment)
  - `latest_check_id: null` confirmed correct-by-design — `import` never synthesizes a `checks` row since `checks.verdict` is NOT NULL and yml has no check-report body to map
- notes:
  - SPEC acceptance criterion "legacy project: init && import && query state --json returns correct state derived from old workflow-state.yml" — satisfied on this repo's real historical state, not a scratch fixture

#### T1 (cont.) — steps 3-5
- status: DONE
- changed files:
  - `.kit/changesets/01KXR94MPXB2JF7Y9MH6ACBEEB.changeset.jsonl` (create run, pilot-migration)
  - `.kit/changesets/01KXR95D04X6ZE6WSN0YSDHF2V.changeset.jsonl` (update meta.latest_run_id)
  - `docs/workflow-harness/pilot-evidence/2026-07-17-lab-skills-import.md` (new)
- verification:
  - step 3: hand-authored + applied run+meta changesets for this session's own pilot-migration work (`zharness db changeset apply` ×2, both `{"applied":1,"skipped_already_applied":0}`); `resume --json` confirms `latest_run_id` now points at this run
  - step 4: copied `.kit/` (excl. `harness.db`/`cache/`) to a scratch dir, `zharness init` + `db changeset apply` on all 7 changesets in ULID order → `resume --json` byte-identical to the original — zero divergence
  - step 5: evidence bundle committed at `docs/workflow-harness/pilot-evidence/2026-07-17-lab-skills-import.md` (changeset log, init/import/query outputs, rebuild-from-changesets proof, validate/audit findings)
- notes:
  - `zharness validate --json` / `zharness audit --json` surfaced a real, non-hypothetical gap: every phase 1-6 RUN/CHECK artifact (pre-dating continuity's CLI-first rewrite) fails cross-link validation (missing/placeholder `id`/`phase`/`plan_id`/`run_id`/`check_id`); `entropy_score: 100`, zero `pointer_drift` (DB pointers themselves are consistent — gap is in markdown frontmatter only)
  - per this task's own `avoid` rule ("hotfixing cli/ or skills mid-run — record gaps instead"), NOT fixed here — routed to Wave 2 T2 as a filed gap/issue, full detail in the evidence bundle
  - T1's own stop condition ("`import` cannot correctly derive current state") did NOT trigger — state derivation is correct; this is an artifact-completeness gap, not an import defect

### Wave 2
#### T2 — Gap issues + go/no-go
- status: DONE
- changed files:
  - `skills/workflow/README.md` (new "Pilot Evidence & Go/No-Go" section)
- verification:
  - `gh issue create` ×2 → https://github.com/therealtinhtute/skills/issues/24 (resume.go/STATE.md mismatch), https://github.com/therealtinhtute/skills/issues/25 (phase 1-6 artifact backfill gap)
  - `gh issue list --label workflow-harness --limit 20` → both #24 and #25 listed
  - `grep -n "Go/No-Go" skills/workflow/README.md` → line 67, section present with verdict **GO** and both issue links
- notes:
  - verdict: GO — both filed gaps are non-blocking (state derivation correct, rebuild-from-changesets byte-exact); neither breaks the chain's core promise

### Wave 3
#### T3 — Migration guide + quickstart
- status: DONE
- changed files:
  - `docs/workflow-harness/migration.md` (new)
  - `README.md` (new "Quickstart: zharness" section)
- verification:
  - Dry-ran the documented quickstart on a clean scratch project (`/private/tmp/.../scratchpad/docs-dryrun`), docs-only, no prior knowledge assumed:
    - `cd cli && go build -o /tmp/zharness-dryrun ./cmd/zharness && zharness --version` → `zharness version dev` — matches doc exactly
    - `zharness init --json` on an empty scratch dir → **failed** (`db_not_writable`) — `.kit/` directory didn't exist yet; this was a genuine undocumented step, not doc-following error (confirmed by inspecting the empty scratch dir)
    - Fixed both docs (`README.md` quickstart, `migration.md` checklist step 0) to add `mkdir -p .kit` before `init` on new projects
    - Re-ran the full walkthrough on a fresh scratch dir following the corrected docs verbatim: `mkdir -p .kit && zharness init && zharness story && zharness query state && zharness resume` — all 4 commands succeeded, zero out-of-doc steps needed
  - filed a 3rd gap ([#26](https://github.com/therealtinhtute/skills/issues/26)) for the missing `cli/v*` tagged release that `install-zharness.sh` depends on — `gh release list --repo therealtinhtute/skills --limit 20` confirmed zero releases exist; documented both the intended and current-actual install paths in `migration.md` rather than presenting an unusable command as canonical
- notes:
  - the `mkdir -p .kit` gap was found and fixed in-line because it was a **docs** fix within this phase's own Allowed Surfaces (not a `cli/` hotfix) — directly actionable per T3's own stop condition ("walkthrough requires undocumented knowledge")

#### T4 — Purge legacy semantics
- status: DONE
- changed files:
  - `CLAUDE.md` (added harness-backed state paragraph to Skill Pipeline section)
  - `skills/workflow/to-plan/references/workflow-state-template.yml` (deleted via `trash`)
  - `README.md` (harness artifact layout section: `workflow-state.yml` demoted from "lightweight pointer index" to explicitly retired/replaced by `harness.db`+`changesets/`)
- verification:
  - `grep -rn 'workflow-state-template' .` → only hits in `.kit/planning/**` and `.kit/runs/**` (historical phase records correctly describing when/why it was deferred to this phase) and this run's own log — zero live-doc dangling references; `to-plan/SKILL.md`'s own References list already stopped pointing at it back in skill-adapters (phase 5), confirmed via the same grep
  - `ls skills/workflow/to-plan/references/` → template file gone, 4 remaining files intact
  - `grep -rn 'workflow-state.yml' --include='*.md' . | grep -v migration.md | grep -v '.kit/'` (plan's literal T4 verification command) → **not empty** — 10 remaining hits reviewed individually: all describe `workflow-state.yml` in explicitly historical/retired/legacy-mapping context (this repo's own `README.md` now says "retired"; `docs/workflow-harness/gap-matrix.md` describes the original problem in past tense; `pilot-evidence/*.md` and `skills/workflow/README.md`'s go/no-go section narrate this session's real import event; `cli/docs/STATE.md`'s "Legacy Field Mapping" section and `cli/testdata/legacy-kit/**` are intentionally-preserved legacy-mapping references, `cli/**` being Forbidden Surface anyway) — none claim it is still the live/current mechanism, satisfying the plan's stated intent ("zero live yml semantics") even though the literal grep isn't empty
- notes:
  - **scope deviation from PLAN.md, staying within CONTEXT.md**: PLAN.md T4 step 1 said "update every live mention" with no file-list qualifier, but CONTEXT.md's own Forbidden Surfaces explicitly list `skills/workflow/*/SKILL.md` for this phase, and Allowed Surfaces name only `CLAUDE.md` for the purge (plus the template deletion). The live `workflow-state.yml` mentions remaining in `handoff/SKILL.md` and its `references/continuity-sources.md`, `check/SKILL.md`, and `watzup/SKILL.md`'s anti-pattern note were **not touched** — continuity's own check gate (2026-07-17) already confirmed these specific mentions are an intentional, locked dual-write/fast-index design decision, not leftover cruft, and `skills/workflow/*/SKILL.md` is Forbidden Surface for pilot-migration regardless. Followed CONTEXT.md's boundary over PLAN.md's broader wording per this initiative's own rule ("if it surfaces a contradiction, stop and escalate per that phase's own CONTEXT.md, don't silently patch") — resolved by scoping down to the narrower, authoritative boundary rather than expanding into Forbidden Surface

## Summary
- passed tasks: T1 (5 steps), T2, T3, T4 — all DONE
- blocked tasks: none
- unresolved concerns: 3 filed gaps (#24, #25, #26), tracked not blocking; all non-critical per go/no-go verdict (GO)

## Next Recommended Action
- `check full` — gate the pilot-migration phase diff before commit
