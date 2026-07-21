---
id: {ULID}
type: run
phase: {phase-slug} | none
lane: {tiny|normal|high-risk}
mode: {full|simple}
plan_id: {ULID of the phase PLAN this run executes} | none
trace_ids: [{ULID}, ...]
created: {YYYY-MM-DD}
updated: {YYYY-MM-DD}
---

# COOK RUN

Run ID: work-YYYYMMDD-HHmm-{slug}
Mode: full | simple
Status: running | blocked | passed | aborted
Spec: .kit/planning/SPEC.md | none
Roadmap: .kit/planning/ROADMAP.md | none
Phase: {phase-slug} | none
Plan: .kit/planning/phases/{phase-slug}/{phase-slug}-PLAN.md | none
Started At: YYYY-MM-DD HH:mm

## Preflight
- scope drift: yes | no
- working tree note
- required artifacts present: yes | no
- selected phase / source prompt

## Wave / Task Log
### Wave 1
#### T1 — Task name
- status: DONE | DONE_WITH_CONCERNS | NEEDS_CONTEXT | BLOCKED
- changed files:
  - path
- verification:
  - command → pass | fail
- notes:
  - concern or blocker detail

## Summary
- passed tasks
- blocked tasks
- unresolved concerns

## Next Recommended Action
- `check full`
- `to-plan phase {slug}`
- `brainstorm refine`
- `handoff`
