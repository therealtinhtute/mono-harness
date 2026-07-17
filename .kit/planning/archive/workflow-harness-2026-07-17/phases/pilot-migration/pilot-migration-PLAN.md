# Plan: pilot-migration

Phase: pilot-migration
Status: ready
Wave Count: 3
Execution Owner: work
Updated At: 2026-07-17 (refreshed — pilot target confirmed: dogfooding this repo, `Lab/skills`, itself; T1 rewritten with concrete import/rebuild commands against this repo's own workflow-state.yml)

## Goal
Pilot evidence + go/no-go published; migration docs shipped; legacy semantics purged.

## Inputs
- Phases 1–7 complete; `/tmp/zharness` dev build installable (no tagged release yet — dev build satisfies the version gate per existing skill convention)
- Pilot target confirmed by user 2026-07-17: **this repo (`Lab/skills`) itself**, dogfooding its own legacy `.kit/workflow-state.yml`-driven state

## Wave 1
### T1 — Execute pilot chain (import this repo's own legacy state)
- type: test
- inputs:
  - This repo's existing `.kit/workflow-state.yml` (validation-gate/continuity history) and `.kit/planning/**`; SPEC acceptance criteria as checklist
- touches:
  - This repo's own `.kit/` (creates `.kit/harness.db`, `.kit/changesets/**` — both already covered by the gitignore pattern established in continuity's e2e work)
- avoid:
  - hotfixing cli/ or skills mid-run — record gaps instead
- steps:
  1. `cd /Users/tinhtute/Lab/skills && /tmp/zharness init --json` then `/tmp/zharness import --json` against the existing `.kit/workflow-state.yml` + `.kit/planning/**` — this is the real "legacy project" acceptance criterion (SPEC: "legacy project: init && import && query state --json returns correct state derived from old workflow-state.yml")
  2. `/tmp/zharness query state --json` — verify `current_phase`/`entry_phase`/`latest_run_id`/`latest_check_id`/`latest_handoff_id` match this repo's actual `workflow-state.yml` values at import time
  3. Run the remainder of pilot-migration's own waves (T2–T4 below) through the zharness-backed chain: author/apply changesets for story/run/check/handoff as each wave completes, exactly per the mechanism `work`/`to-plan`/`check` skills now document
  4. Cross-machine resume: clone this repo to a scratch dir, install `zharness`, rebuild via `db changeset apply`, run `watzup` — assert recap matches the last handoff exactly (same procedure continuity's T4 already proved on a throwaway sample; this time on the real repo's real history)
  5. Capture evidence: changeset log, `query`/`resume` outputs, `validate` results, recap diff — commit under `docs/workflow-harness/pilot-evidence/`
- expected outputs:
  - This repo's own harness state live (`.kit/harness.db` + `.kit/changesets/**`), matching its pre-import `workflow-state.yml`; evidence bundle committed
- verification:
  - Every SPEC acceptance criterion checked off with a command output reference
- stop if:
  - `import` cannot correctly derive current state from this repo's `workflow-state.yml` (a real defect, not a hypothetical — record and escalate, don't hotfix silently)
- escalate to:
  - check (defect triage) then to-plan phase {owning-phase}

## Wave 2
### T2 — Gap issues + go/no-go
- type: docs
- inputs:
  - T1 evidence
- touches:
  - GitHub issues; `skills/workflow/README.md` (evidence summary section)
- avoid:
  - burying gaps in prose instead of filing them
- steps:
  1. File one GitHub issue per gap (`gh issue create`), linking evidence
  2. Write go/no-go section in skills/workflow/README.md with evidence links
- expected outputs:
  - Filed issues; published verdict
- verification:
  - `gh issue list --label workflow-harness` shows the gaps; README section exists
- stop if:
  - verdict is no-go
- escalate to:
  - brainstorm refine

## Wave 3
### T3 — Migration guide + quickstart
- type: docs
- inputs:
  - Go verdict; T1 evidence
- touches:
  - `docs/workflow-harness/migration.md` (new), root `README.md` (quickstart)
- avoid:
  - documenting aspirational behavior — only what the pilot proved
- steps:
  1. migration.md: install-first quickstart, legacy `.kit/` checklist (import → validate → trash workflow-state.yml), plain rollback notes, contributor playbook (adding commands/contracts without breaking changesets)
  2. Root README quickstart: install → init/import → run chain
- expected outputs:
  - Docs a new user can follow cold
- verification:
  - Dry-run the walkthrough on a clean scratch project following docs only; note zero out-of-doc steps needed
- stop if:
  - walkthrough requires undocumented knowledge
- escalate to:
  - check

### T4 — Purge legacy semantics
- type: migration
- inputs:
  - T3 merged (docs first, purge second)
- touches:
  - `CLAUDE.md`, `skills/workflow/to-plan/references/workflow-state-template.yml` (delete via `trash`), any remaining references
- avoid:
  - `rm` (hard rule: trash only); purging before docs land
- steps:
  1. `grep -rn 'workflow-state' --include='*.md' .` — update every live mention (historical notes in migration.md may remain, labeled legacy)
  2. `trash skills/workflow/to-plan/references/workflow-state-template.yml`
  3. Update CLAUDE.md pipeline description to zharness-backed flow
- expected outputs:
  - Zero live yml semantics in repo docs
- verification:
  - `grep -rn 'workflow-state.yml' --include='*.md' . | grep -v migration.md | grep -v '.kit/'` returns nothing
- stop if:
  - a consumer outside this repo depends on the template file
- escalate to:
  - user clarification

## Risks / Watch-fors
- Pilot on a toy task = false confidence — hold the "real repo, real task" line
- This repo's own `.kit/` (created by this very planning session) must itself be imported — dogfood before declaring migration done
