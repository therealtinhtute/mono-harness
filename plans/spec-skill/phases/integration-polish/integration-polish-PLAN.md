# Plan: integration + polish

## Goal
Polish the planning subsystem so the repo can move from design docs into actual skill authoring smoothly.

## Inputs
- `plans/spec-skill/SPEC.md`
- `plans/spec-skill/ROADMAP.md`
- phase outputs from spec-core and plan-core

## Wave 1
1. Add handoff rules between `spec` and `plan`.
   - verify: natural transition exists after spec completion

2. Add suggestion points to existing skills.
   - `investigator` when codebase context is missing
   - `reviewer` / `verifier` after implementation
   - `git` / `watzup` / `handoff` at wrap-up moments
   - verify: the subsystem feels composable

## Wave 2
3. Review wording/examples for consistency.
   - mini-GSD clone framing
   - source mode terminology
   - artifact names
   - verify: docs no longer mix old and new model

## Verification
- the initiative is ready to shift from planning docs into writing `skills/spec` and `skills/plan`

## Risks / Watch-fors
- polishing too long before implementation begins
- letting examples drift from actual artifact contracts
