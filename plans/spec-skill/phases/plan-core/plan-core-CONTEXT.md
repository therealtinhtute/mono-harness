# Context: plan core

## Goal
Implement the `plan` skill as the second stage of the mini-GSD workflow: consume `SPEC.md`, then produce roadmap + context + executable phase plans.

## Gray Areas Reviewed
- whether `plan` should output one big plan or per-phase artifacts
- whether context locking is optional
- whether repo investigation belongs inside `plan`

## Locked Decisions
- `plan` requires an existing `.planning/SPEC.md`
- `plan` writes `ROADMAP.md` plus per-phase context and plan files
- context locking stays explicit in `{phase}-CONTEXT.md`
- wave-grouped tasks stay per-phase, not one giant global plan
- repo investigation may be suggested/invoked when needed, but is not the central artifact

## Canonical References
- `plans/spec-skill/SPEC.md`
- `plans/spec-skill/ROADMAP.md`
- `plans/spec-skill/research/get-shit-done-repo-research.md`

## Rejected Options
- planning without spec precondition
- skipping context lock entirely
- a single flat `PLAN.md` as the only artifact for every case

## Deferred Ideas
- whether tiny phases can skip `CONTEXT.md`
- whether `plan` should support single-phase mode in v1
