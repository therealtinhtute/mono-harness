# Refactoring Draft — Harness Durability & Knowledge Contract

Working directory for the mono-harness refactor brainstorm, 2026-08-15.
Nothing here is committed policy. This is the evidence and reasoning behind
`spec.md`, kept so the plan can be reviewed without re-deriving it.

## Read In This Order

| File | What it holds |
|---|---|
| [`spec.md`](spec.md) | **The deliverable.** Validated spec, four phases, hard gates. |
| [`audit-onedrive-cloud.md`](audit-onedrive-cloud.md) | Field evidence. Every defect the spec fixes, measured in a real consumer repo. |
| [`reference-models.md`](reference-models.md) | What harness-experimental and codesight do differently, and which mechanisms are worth taking. |
| [`decisions.md`](decisions.md) | Interview decision log — what was chosen, what was rejected, why. |

## Scope Boundary

Every file the spec touches lives under `/Users/tinhtute/Lab/mono-harness`.

These are **read-only reference sources**, cited as evidence, never edited:

- `/Users/tinhtute/Personal/onedrive-cloud` — real consumer repo running `zharness 0.9.1`
- `/Users/tinhtute/Lab/harness-experimental` — repository-as-system-of-record 
- `https://github.com/hoangnb24/repository-harness`  — repository-as-system-of-record version remote
- `github.com/Houseofmvps/codesight` — generated knowledge-base method

## Status

Spec drafted, not accepted. No implementation has started.
Next step is owner acceptance, then `/to-plan` for executable phase plans.
