# SPEC Template

Use this structure for `.kit/planning/SPEC.md`:

```markdown
---
id: {ULID}
type: spec
phase: none
lane: {tiny|normal|high-risk}
intake_id: {ULID returned by `zharness intake` at SPEC lock}
created: {YYYY-MM-DD}
updated: {YYYY-MM-DD}
---

# SPEC: {title}

Status: draft | locked
Input Type: new-spec | spec-slice | change-request | new-initiative | maintenance | harness-improvement
Lane: tiny | normal | high-risk
Risk Flags: auth, authorization, data-model, audit-security, external-systems, public-contract, cross-platform, existing-behavior, weak-proof, multi-domain
Affected Surfaces: api, browser, mobile, desktop, worker, db, provider, docs
Downstream: to-plan full | to-plan phase | work simple | none
Updated At: YYYY-MM-DD

## Source Mode
idea | files | refine

## Source Inputs
- user prompt summary
- @file references
- prior spec / decision if relevant

## Scenario
project bootstrap | feature bootstrap | module bootstrap | refine existing spec

## Goal
Short statement of what is being built and why.

## Users / Actors
Who this work serves or affects.

## Requirements
1. Requirement one
2. Requirement two
3. Requirement three

## Boundaries
### In Scope
- explicit included items

### Out of Scope
- explicit excluded items

## Constraints
Technical, product, timeline, policy, or dependency constraints.

## Acceptance Criteria
- concrete checks proving the spec is good enough

## Validation Expectations
- expected proof shape for downstream planning/execution
- unit / integration / e2e / platform expectations if already knowable

## Dependencies / Assumptions
- known dependencies
- assumptions made during clarification

## Key Decisions
- chosen path and rationale
- rejected alternative + why not chosen

## Open Questions
- unresolved but visible gaps

## Deferred Ideas
- intentionally excluded future ideas

## Ambiguity Report
- goal clarity
- scope clarity
- constraints clarity
- acceptance clarity
```

Rules:
- requirements should be numbered and falsifiable
- boundaries must be explicit
- acceptance criteria must be concrete enough for planning
- open questions must be visible, not hidden in prose
- risk flags should describe why the work is sensitive, not repeat the entire spec
- downstream should name the next intended workflow step
- frontmatter `id` is this SPEC's ULID; `phase` is always `none` — a SPEC precedes phase decomposition, `to-plan` assigns phases downstream
- frontmatter `lane` mirrors the body's `Lane:` field; keep both in sync when either changes
- frontmatter `intake_id` is the ULID returned by `zharness intake` when brainstorm fires it at the SPEC-lock step; absent only if the version gate blocked the skill (should not happen in normal operation)
