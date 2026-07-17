---
id: 01KXQKT39GV8YK5QBBD819Y5GD
type: spec
phase: none
lane: normal
created: 2026-01-01
updated: 2026-01-01
---

# SPEC: sample-phase fixture (broken chain)

Status: locked
Input Type: new-spec
Lane: normal
Risk Flags: none
Affected Surfaces: docs
Downstream: to-plan full
Updated At: 2026-01-01

## Source Mode
idea

## Source Inputs
- fixture data for cli-domain T4 validate tests

## Scenario
project bootstrap

## Goal
Fixture SPEC for exercising `zharness validate` against a chain with exactly one broken cross-link (CHECK->RUN).

## Users / Actors
cli-domain test suite

## Requirements
1. Provide a chain with exactly one broken cross-link for validate's failure path

## Boundaries
### In Scope
- fixture-only content

### Out of Scope
- real project work

## Constraints
None.

## Acceptance Criteria
- `zharness validate --json` reports valid=false naming the broken RUN->CHECK link

## Validation Expectations
- none beyond the fixture's own frontmatter

## Dependencies / Assumptions
- none

## Key Decisions
- none

## Open Questions
- none

## Deferred Ideas
- none

## Ambiguity Report
- goal clarity: n/a (fixture)
- scope clarity: n/a (fixture)
- constraints clarity: n/a (fixture)
- acceptance clarity: n/a (fixture)
