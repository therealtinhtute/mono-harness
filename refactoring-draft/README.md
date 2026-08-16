# Refactoring Draft — Harness Durability & Knowledge Contract

Working directory for the mono-harness refactor brainstorm, 2026-08-15.
Nothing here is committed policy. This is the evidence and reasoning behind
`spec.md`, kept so the plan can be reviewed without re-deriving it.

## Read In This Order

| File | What it holds |
|---|---|
| [`spec.md`](spec.md) | **The deliverable.** Accepted spec — four phases (P0–P3), three hard gates. |
| [`decisions.md`](decisions.md) | **Read second.** D1–D31 — what was chosen, what was rejected, why. Later entries supersede earlier ones in place, with pointers. |
| [`work-shape.md`](work-shape.md) | M8 — why a small task costs 142 lines of markdown, and how docs / tooling / context / CLI combine into one authority chain. |
| [`open-questions.md`](open-questions.md) | Resolution record for the four post-interview corrections. Kept for the reasoning, not for open work. |
| [`audit-onedrive-cloud.md`](audit-onedrive-cloud.md) | Field evidence. Every defect the spec fixes, measured in a real consumer repo. |
| [`reference-models.md`](reference-models.md) | What harness-experimental, repository-harness, and codesight do differently, and which mechanisms are worth taking. Note: its codesight section describes a generator that was later cut (D31). |

## Scope Boundary

Every file the spec touches lives under `/Users/tinhtute/Lab/mono-harness`.

These are **read-only reference sources**, cited as evidence, never edited:

- `/Users/tinhtute/Personal/onedrive-cloud` — real consumer repo running `zharness 0.9.1`
- `/Users/tinhtute/Lab/harness-experimental` — repository-as-system-of-record 
- `https://github.com/hoangnb24/repository-harness`  — repository-as-system-of-record version remote
- `github.com/Houseofmvps/codesight` — generated knowledge-base method

## Status

**Accepted. No implementation has started.** Tracked in
[PR #59](https://github.com/therealtinhtute/mono-harness/pull/59) — documentation only,
no code touched.

### The program

| Phase | What it does |
|---|---|
| **P0** | Work-shape routing — four targeted fixes, starting with `next.go`'s auto-routing |
| **P1** | Durability core — state consolidates into `.harness/`, plus `audit` hardening |
| **P2** | Decision durability — ADRs promoted at handoff, filtered by trigger |
| **P3** | Docs contract — the source-of-truth hierarchy gets declared and enforced |

Three hard gates: **work shape** (a bounded fix produces 0 markdown files), **durability**
(a fresh clone holds every durable item as a tracked file), **resident cost** (≤900
entrypoint, ≤2,400 full path).

### How it got here

Drafted as seven durability improvements (M1–M7). Three rounds of correction since:

1. **Post-interview research** (Q1–Q4 → D13–D16) — the ≤1,000-token gate was 2.2x tighter
   than the reference model it cited; blanket ADR promotion would have flooded
   `docs/decisions/`; and the owner reframed the root cause from *durability* to *ceremony
   weight*, which put M8 in front as P0.
2. **Acceptance interview** (D17–D23) — six ambiguities resolved, including who declares
   work shape and where `harness.db` lives. D20 resolved a contradiction that made D1 and
   the Standing Constraint mutually impossible.
3. **Process brainstorm** (D24–D31) — every finding made the program *smaller*:
   - **D24** — P0's mechanism already shipped (`work.md:15`). The defect is routing:
     `next.go` picks the light path only when no plan is open, so work size never enters
     the decision. P0 shrank to four fixes.
   - **D27** — authority is three tiers. Resolved a contradiction where the architecture
     declared the database authoritative while Gate #1 assumed tracked markdown was.
   - **D29** — `audit` gains a docs-adherence check. `work.md:17` states a five-file rule
     that exists nowhere in the Go source; 2 of 56 error messages cite a doc.
   - **D31** — **P4 is cut.** The generator ran on one stack and could not run on
     mono-harness itself. The rule replacing it matters more: no hand-maintained document
     may hold derived facts.

**Net effect:** five phases with a large unknown build → four phases with no genuine
unknowns left. Everything remaining fixes something that has already caused data loss,
drift, or daily friction.

### Known debts

- **G2 (subagent fan-out)** — the audit deferred it *"until R1–R4 land and remeasure."*
  They landed on 2026-08-11 and nobody remeasured. An overdue deferral, not a live decision.
- **D26** — D19 removes proof capture from a bounded mode that currently performs it. The
  decision was taken before that was known; it stands, and it is recorded as a weakening
  rather than a formalization.

Next step is `/to-plan` against `spec.md`, starting with P0.
