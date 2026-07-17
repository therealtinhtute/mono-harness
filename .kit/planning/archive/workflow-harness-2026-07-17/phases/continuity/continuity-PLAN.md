# Plan: continuity

Phase: continuity
Status: ready
Wave Count: 2
Execution Owner: work
Updated At: 2026-07-17

## Goal
watzup + handoff unified on the continuity contract; git/interview minimally integrated; cross-machine resume proven.

## Inputs
- Phase 4 `resume`/`check record` (+ `handoff record`), Phase 6 gate outputs
- `cli/docs/STATE.md` recovery table; Phase 2 handoff template

## Wave 1
### T1 — Rewrite handoff SKILL.md
- type: implementation
- inputs:
  - CONTRACT.md handoff schema; handoff-template frontmatter
- touches:
  - `skills/workflow/handoff/SKILL.md`, `references/handoff-template.md`, `references/continuity-sources.md`
- avoid:
  - watzup files; prose anchors resume can't parse
- steps:
  1. Add gate block
  2. Close-out flow: `zharness handoff record --json` with anchors (state, latest run/check ULIDs, open items) → then write HANDOFF.md with entity ULID in frontmatter
  3. continuity-sources.md: entity is canonical, markdown is narrative
- expected outputs:
  - CLI-first handoff skill
- verification:
  - `grep -n 'handoff record' skills/workflow/handoff/SKILL.md`; `grep 'workflow-state' …` empty
- stop if:
  - CLI lacks `handoff record`
- escalate to:
  - to-plan phase cli-domain

### T2 — Rewrite watzup SKILL.md
- type: implementation
- inputs:
  - resume JSON shape; STATE.md recovery table
- touches:
  - `skills/workflow/watzup/SKILL.md`, `references/output-contract.md`, `references/artifact-recap.md`
- avoid:
  - handoff files; independent prose re-derivation of state
- steps:
  1. Add gate block with `no-harness` routing (install vs import)
  2. Recap flow renders `zharness resume --json` 1:1: position, latest IDs, readiness state, drift findings + recovery steps
  3. output-contract.md documents the 4 readiness states and recovery action table
- expected outputs:
  - CLI-first watzup skill
- verification:
  - `grep -n 'resume' skills/workflow/watzup/SKILL.md`; output-contract lists all 4 states (inspection)
- stop if:
  - a recap section has no snapshot field
- escalate to:
  - to-plan phase cli-domain

### T3 — Minimal integration: git + interview
- type: implementation
- inputs:
  - CONTRACT.md query schema
- touches:
  - `skills/workflow/git/SKILL.md` (one step: query latest check verdict, warn on FAIL/missing), `skills/workflow/interview/SKILL.md` (gate block only)
- avoid:
  - restructuring either skill; blocking behavior in git
- steps:
  1. git: before commit/PR flow, `zharness query check --latest --json`; warn-not-block on FAIL
  2. interview: add gate block
- expected outputs:
  - All 8 workflow skills CLI-aware
- verification:
  - `grep -l 'zharness' skills/workflow/*/SKILL.md | wc -l` = 8
- stop if:
  - query lacks a latest-check view
- escalate to:
  - to-plan phase cli-domain

## Wave 2
### T4 — Cross-machine resume e2e
- type: test
- inputs:
  - T1–T3; Phase 5 sample project in a git repo
- touches:
  - scratch clones only
- avoid:
  - copying harness.db between machines (must rebuild from changesets)
- steps:
  1. On sample project: run work partially, run handoff (entity + HANDOFF.md + commit changesets)
  2. Clone to a fresh directory (simulating another machine), install zharness, `zharness db changeset apply` / rebuild, run watzup flow
  3. Assert recap matches the handoff anchors exactly; then stale a pointer deliberately and assert the specific recovery step prints
- expected outputs:
  - Recorded e2e evidence (commands + outputs)
- verification:
  - Recap fields == handoff entity fields (diff the JSON); staled fixture prints non-generic recovery text
- stop if:
  - rebuild-from-changesets diverges from original DB
- escalate to:
  - to-plan phase cli-core

## Risks / Watch-fors
- The e2e is the SPEC's flagship acceptance — don't fake the "fresh machine" (fresh dir + empty PATH shim minimum)
- Warn-not-block in git must not silently become block — chain UX constraint
