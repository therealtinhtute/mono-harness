# Phase Context Template

Use this structure for `.kit/planning/phases/{phase-slug}/{phase-slug}-CONTEXT.md`:

```markdown
# Context: {phase name}

Phase: {phase-slug}
Status: ready | stale | blocked
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: low | medium | high
Expected Proof: unit, integration, e2e, platform

## Goal
Short restatement of the phase goal.

## Scope Boundary
### Allowed Surfaces
- files / modules / layers this phase may touch

### Forbidden Surfaces
- areas explicitly out of scope for this phase

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
- `.kit/planning/SPEC.md`
- `.kit/planning/ROADMAP.md`
- other docs / files / APIs if relevant

## Rejected Options
- option + why rejected

## Deferred Ideas
- future idea intentionally not done now

## Escalate If
- condition that should route back to `brainstorm` or `plan`
```

Rules:
- keep only decisions that help implementation
- do not restate the entire spec
- rejected options should explain tradeoffs briefly
- deferred ideas should not leak back into scope
- allowed/forbidden surfaces should be concrete enough that `cook` can detect drift
