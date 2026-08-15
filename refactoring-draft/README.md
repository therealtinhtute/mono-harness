# Refactoring Draft — Harness Durability & Knowledge Contract

Working directory for the mono-harness refactor brainstorm, 2026-08-15.
Nothing here is committed policy. This is the evidence and reasoning behind
`spec.md`, kept so the plan can be reviewed without re-deriving it.

## Read In This Order

| File | What it holds |
|---|---|
| [`spec.md`](spec.md) | **The deliverable.** Validated spec, four phases, hard gates. |
| [`open-questions.md`](open-questions.md) | **Read second.** Four items research has put back in play, each contradicting something in `spec.md`. |
| [`work-shape.md`](work-shape.md) | M8 — why a small task costs 142 lines of markdown, and how docs / tooling / context / CLI combine into one authority chain. |
| [`audit-onedrive-cloud.md`](audit-onedrive-cloud.md) | Field evidence. Every defect the spec fixes, measured in a real consumer repo. |
| [`reference-models.md`](reference-models.md) | What harness-experimental, repository-harness, and codesight do differently, and which mechanisms are worth taking. |
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

`spec.md` has **four open corrections** on it — see [`open-questions.md`](open-questions.md).
The largest: the owner reframed the root cause from *durability* to *ceremony weight*
("a small task generates too many files"), which adds M8 and likely reorders the phases.

Next step is resolving Q1–Q3, then owner acceptance, then `/to-plan`.
