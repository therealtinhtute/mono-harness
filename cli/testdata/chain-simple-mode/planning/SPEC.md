---
id: 01KXX1VGCY2896YRK9FP76HZCD
type: spec
phase: none
lane: tiny
created: 2026-01-01
updated: 2026-01-01
---

# SPEC: chain-simple-mode fixture

Status: locked
Input Type: new-spec
Lane: tiny
Risk Flags: none
Affected Surfaces: docs
Downstream: none
Updated At: 2026-01-01

## Source Mode
idea

## Source Inputs
- fixture data for harness-mode-parity T1 validate tests

## Scenario
project bootstrap

## Goal
Fixture SPEC for exercising `zharness validate` against a simple-mode-produced chain: a phase-less, plan-less RUN and an unregistered CHECK, both carrying `mode: simple`.

## Users / Actors
cli-domain test suite

## Requirements
1. Provide a simple-mode RUN/CHECK pair that validate accepts as valid despite no DB rows and no real phase/plan

## Boundaries
### In Scope
- fixture-only content

### Out of Scope
- real project work

## Constraints
None.

## Acceptance Criteria
- `zharness validate --json` reports valid=true against this fixture

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
