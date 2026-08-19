# 0001 — Markdown is the source of truth; `harness.db` is a derived index

## Status

Accepted. Implemented by the `harness-markdown-truth` initiative (`docs/plans/completed/harness-markdown-truth.md`), phases P2 and P3.

## Context

The harness kept lifecycle state in two places at once. Six write sites — plan, trace, decision, check, handoff, intake — wrote a markdown section *and* a row in `harness.db`, and neither was declared authoritative. When they disagreed, nothing said which one to believe.

A separate mechanism made this worse rather than better: an append-only changeset log under `.kit/changesets/`, replayed from empty to rebuild the database. It was 1,715 lines of code whose only job was to reconstruct state that the committed markdown already described. It also had to be git-tracked, which meant `.kit/` could not simply be per-machine scratch — and this repo's `.gitignore` contradicted its own migration doc on exactly that point.

## Decision

Markdown is authoritative. `harness.db` is a derived index over it.

- Every dual-write site writes the markdown first and derives the DB row from what it wrote. A failed markdown write produces no DB row (`harness-markdown-truth` R8).
- `zharness db rebuild --yes` reconstructs the entire database from committed repository content alone — plan markdown under `docs/plans/`, memory entries under `docs/memory/` — with no read of any non-committed state (R10, and `durable-memory` R4).
- The changeset layer is retired, not deprecated. `db changeset apply` and `db changeset status` no longer exist; `zharness db` offers only `rebuild` and `status`.
- `plan_index` tracks each active plan by path and SHA256, copying the column shape already used by `managed_docs` (`cli/internal/infrastructure/migrations.go:216`). Staleness is a three-way hash comparison, never a timestamp guess.

## Consequences

- Losing `harness.db` loses nothing. It is gitignored on purpose (`.gitignore:3`) and rebuilt on demand.
- A field that lands only in the DB and never in markdown silently vanishes on the next rebuild. This is the failure mode any new mutating command must be checked against.
- `.kit/` became purely per-machine — cache, conflicts, logs — and is fully gitignored (`.gitignore:6-15`). Nothing inside it needs to survive a clone.
- The CLI owns the pen. `trace add` and `decision add` append their own `## Progress` and `## Decisions` lines; agents do not hand-write them, because a hand-written line has no derived row behind it.
- An agent with no CLI at all can still read the whole lifecycle, because the whole lifecycle is committed markdown.

## Authority

- `docs/plans/completed/harness-markdown-truth.md` — R8 (markdown-first dual writes), R9 (`plan_index` shape), R10 (rebuild from committed content, changeset retirement).
- `cli/internal/infrastructure/migrations.go:216` — `plan_index` table, following `managed_docs` at line 129.
- `docs/plans/completed/durable-memory.md` — R4 extends the same rule to the `memories` index.
