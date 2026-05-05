# Phase Plan Template

Use this structure for `.planning/phases/{phase-slug}/{phase-slug}-PLAN.md`:

```markdown
# Plan: {phase name}

## Goal
What this phase should accomplish.

## Inputs
- required files
- required prior decisions

## Wave 1
1. Task name
   - steps:
     1. first step
     2. second step
   - expected outputs:
     - file / endpoint / behavior
   - verify:
     - test, inspection, or command

## Wave 2
2. Task name
   - steps:
     1. first step
   - expected outputs:
     - output
   - verify:
     - proof of completion

## Risks / Watch-fors
- important coordination or sequencing risk
```

Rules:
- only place tasks in the same wave if they can proceed independently
- keep steps concrete enough for execution
- verification must be observable
- expected outputs should be explicit, not implied
