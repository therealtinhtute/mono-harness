# Refactoring Draft — Harness Durability & Knowledge Contract

Working directory for the mono-harness refactor brainstorm, 2026-08-15.
Nothing here is committed policy. This is the evidence and reasoning behind
`spec.md`, kept so the plan can be reviewed without re-deriving it.

## Read In This Order

| File | What it holds |
|---|---|
| [`spec.md`](spec.md) | **The deliverable.** Accepted spec — five phases (P0–P4), three hard gates. |
| [`open-questions.md`](open-questions.md) | Resolution record for the four post-interview corrections. Kept for the reasoning, not for open work. |
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

**Accepted 2026-08-15.** No implementation has started.

All four post-interview corrections are resolved and folded into `spec.md`, logged as
D13–D16 in [`decisions.md`](decisions.md):

- **Q1 → D13** — resident cost is two benchmarked gates (≤900 entrypoint, ≤2,400 full
  path), not one aspirational ≤1,000
- **Q2 → D14** — ADR promotion filters against five durability triggers; the handoff gate
  asserts promotion was *considered*, not that everything was promoted
- **Q3 → D15** — **M8 work-shape gating leads as P0.** The owner reframed the root cause
  from *durability* to *ceremony weight* ("a small task generates too many files"), so the
  symptom ships before the diagnosis
- **Q4 → D16** — pruning discipline folds into P3; enforcement ladder and diagnostic
  standard fold into P1 as `audit` hardening

Next step is `/to-plan` against `spec.md`, starting with P0.
