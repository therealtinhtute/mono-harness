# Phase Plan Template

Use this structure for `.planning/phases/{phase-slug}/{phase-slug}-PLAN.md`:

```markdown
# Plan: {phase name}

Phase: {phase-slug}
Status: ready | stale | blocked
Wave Count: {N}
Execution Owner: cook
Updated At: YYYY-MM-DD

## Goal
What this phase should accomplish.

## Inputs
- required files
- required prior decisions

## Wave 1
### T1 — Task name
- type: implementation | test | migration | docs | refactor
- inputs:
  - required artifact or file
- touches:
  - files / modules / surfaces expected to change
- avoid:
  - forbidden or out-of-scope areas
- steps:
  1. first step
  2. second step
- expected outputs:
  - file / endpoint / behavior
- verification:
  - test, inspection, or command
- stop if:
  - ambiguity / drift / dependency condition
- escalate to:
  - brainstorm refine | plan phase | user clarification | check

## Wave 2
### T2 — Task name
- type: implementation | test | migration | docs | refactor
- inputs:
  - required artifact or file
- touches:
  - files / modules / surfaces expected to change
- avoid:
  - forbidden or out-of-scope areas
- steps:
  1. first step
- expected outputs:
  - output
- verification:
  - proof of completion
- stop if:
  - ambiguity / drift / dependency condition
- escalate to:
  - brainstorm refine | plan phase | user clarification | check

## Risks / Watch-fors
- important coordination or sequencing risk
```

Rules:
- only place tasks in the same wave if they can proceed independently
- keep steps concrete enough for execution
- verification must be observable
- expected outputs should be explicit, not implied
- each task should be specific enough that `cook` can report task-level status without inventing new structure
