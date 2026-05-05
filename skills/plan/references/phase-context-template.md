# Phase Context Template

Use this structure for `.planning/phases/{phase-slug}/{phase-slug}-CONTEXT.md`:

```markdown
# Context: {phase name}

## Goal
Short restatement of the phase goal.

## Spec Hooks
- requirement(s) served
- boundary or constraint that matters here

## Locked Decisions
- decision 1
- decision 2

## Assumptions
- assumption 1
- assumption 2

## Canonical Refs
- `.planning/SPEC.md`
- `.planning/ROADMAP.md`
- other docs / files / APIs if relevant

## Rejected Options
- option + why rejected

## Deferred Ideas
- future idea intentionally not done now
```

Rules:
- keep only decisions that help implementation
- do not restate the entire spec
- rejected options should explain tradeoffs briefly
- deferred ideas should not leak back into scope
