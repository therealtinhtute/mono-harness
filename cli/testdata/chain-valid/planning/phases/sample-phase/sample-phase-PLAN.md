# Plan: sample-phase fixture

Phase: sample-phase
Status: ready
Wave Count: 1
Execution Owner: work
Updated At: 2026-01-01

## Goal
Fixture PLAN so validate's PLAN->RUN file-existence check has a real target.

## Inputs
- planning/SPEC.md

## Wave 1
### T1 — Fixture task
- type: docs
- inputs:
  - planning/SPEC.md
- touches:
  - testdata fixture files only
- avoid:
  - real project surfaces
- steps:
  1. exist as a fixture
- expected outputs:
  - this file
- verification:
  - `zharness validate --json` finds this file at the derived PLAN path
- stop if:
  - n/a
- escalate to:
  - n/a

## Risks / Watch-fors
- none — fixture only
