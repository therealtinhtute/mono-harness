# Plan: skill-adapters

Phase: skill-adapters
Status: ready
Wave Count: 2
Execution Owner: work
Updated At: 2026-07-17

## Goal
Three producing skills rewritten CLI-first and proven by a sample chain run that passes `zharness validate`.

## Inputs
- Phase 4 `zharness` binary installed
- `cli/docs/CONTRACT.md`; Phase 2 frontmatter templates; Phase 1 README mapping

## Wave 1
### T1 — Rewrite brainstorm SKILL.md
- type: implementation
- inputs:
  - CONTRACT.md intake schema; locked gate block text
- touches:
  - `skills/workflow/brainstorm/SKILL.md`, `skills/workflow/brainstorm/references/` (spec-template alignment notes only)
- avoid:
  - changing brainstorm's modes/gates UX; other skills
- steps:
  1. Add gate block at workflow start
  2. At SPEC-lock step: insert `zharness intake --type … --summary … --lane … --json` inline; write returned ULID into SPEC frontmatter `intake_id`
  3. Update Output Rules: lock modes no longer mention workflow-state.yml
- expected outputs:
  - CLI-first brainstorm SKILL.md
- verification:
  - `bash scripts/validate-skill.sh skills/workflow/brainstorm`; `grep -c 'zharness' skills/workflow/brainstorm/SKILL.md` ≥ 2; `grep 'workflow-state' skills/workflow/brainstorm/SKILL.md` empty
- stop if:
  - intake step has no natural home in the current flow
- escalate to:
  - brainstorm refine

### T2 — Rewrite to-plan SKILL.md
- type: implementation
- inputs:
  - CONTRACT.md init/story/query schemas
- touches:
  - `skills/workflow/to-plan/SKILL.md`, `skills/workflow/to-plan/references/planning-rules.md` (state instructions)
- avoid:
  - deleting workflow-state-template.yml (Phase 8); roadmap/phase template changes beyond state pointers
- steps:
  1. Add gate block
  2. Replace "initialize workflow-state.yml" instructions with: `zharness init` (if absent) + `zharness story --slug {phase} …` per phase + state recorded via CLI
  3. Roadmap/context/plan templates reference phase status from `zharness query state`
- expected outputs:
  - CLI-first to-plan SKILL.md
- verification:
  - `grep -n 'workflow-state.yml' skills/workflow/to-plan/SKILL.md` returns nothing; `grep -c 'zharness' …` ≥ 3
- stop if:
  - story flags can't express a roadmap phase
- escalate to:
  - to-plan phase harness-contracts

### T3 — Rewrite work SKILL.md
- type: implementation
- inputs:
  - CONTRACT.md trace schema; run-artifact template (Phase 2)
- touches:
  - `skills/workflow/work/SKILL.md`, `skills/workflow/work/references/execution-loop.md`, `run-artifact-template.md` alignment
- avoid:
  - check gate rewiring (Phase 6); routing changes
- steps:
  1. Add gate block
  2. Execution loop: `zharness trace add …` at each wave completion; capture trace ULIDs
  3. RUN artifact writing step includes trace IDs + plan_id in frontmatter
- expected outputs:
  - CLI-first work SKILL.md with trace-linked RUN artifacts
- verification:
  - `grep -n 'trace add' skills/workflow/work/SKILL.md`; `grep 'workflow-state' …` empty
- stop if:
  - wave model doesn't map to trace events cleanly
- escalate to:
  - brainstorm refine

## Wave 2
### T4 — Sample chain run + validate
- type: test
- inputs:
  - T1–T3; a scratch sample project
- touches:
  - scratch project `.kit/` only (not this repo's planning)
- avoid:
  - committing sample artifacts into this repo outside testdata
- steps:
  1. On a scratch project: run brainstorm→to-plan→work per the rewritten skills (agent-driven dry run)
  2. Run `zharness validate --json` and `zharness query state --json`
  3. Record evidence (commands + outputs) in the phase RUN artifact
- expected outputs:
  - Chain artifacts whose IDs cross-reference; validate exit 0
- verification:
  - `zharness validate` exit 0 on the sample; DB rows + changesets exist for intake/story/trace; zero writes to workflow-state.yml (`find` shows no such file)
- stop if:
  - validate failures trace to contract gaps rather than skill text
- escalate to:
  - to-plan phase harness-contracts

## Risks / Watch-fors
- The three SKILL.md rewrites are parallel (different files) but must share the identical gate block — copy exactly, don't paraphrase
- UX preservation is the acceptance most easily lost in a rewrite — diff the step structure, not just the added commands
