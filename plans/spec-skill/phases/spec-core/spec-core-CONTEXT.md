# Context: spec core

## Goal
Implement the `spec` skill as the first locked artifact generator in the mini-GSD workflow.

## Gray Areas Reviewed
- whether to keep `fast/deep` or switch primary mode to source mode
- whether bootstrap is a separate skill or a scenario inside `spec`
- whether `IDEA.md` should always exist

## Locked Decisions
- primary intake mode is **source mode**: `idea` or `files`
- bootstrap stays inside `spec` as a **scenario**, not a separate skill
- `spec` always owns creation/update of `.planning/IDEA.md`
- `spec` writes `.planning/SPEC.md` and stops there
- ambiguity handling remains part of `spec`; planning breakdown is not

## Canonical References
- `plans/spec-skill/SPEC.md`
- `plans/spec-skill/research/get-shit-done-repo-research.md`
- `plans/spec-skill/research/gsd-2-repo-research.md`

## Rejected Options
- single giant planning skill
- separate bootstrap skill for v1
- allowing `plan` to run without `SPEC.md`

## Deferred Ideas
- visible ambiguity score formatting details
- whether to preserve source excerpts verbatim inside `IDEA.md`
