# 0005 — Guard authored-document presence, not authored-document truth

## Status

Accepted. Implemented in phase `authored-docs-guard` of `docs/plans/active/consumer-doc-drift-gate.md` on 2026-08-21.

## Context

Commit `655c6ac` removed the repository's authored documentation while leaving the embedded managed set recoverable by `zharness init`. The managed half reappeared, but the authored half did not, so the repository looked healthy until its citations failed. The incident and its remaining gap are recorded in `docs/decisions/0004-docs-directory-deletion-655c6ac.md:9-13,45-50`.

Before this phase, `audit` composed database-backed lifecycle readers and exposed a fixed three-array JSON envelope (`cli/internal/application/audit.go:16-49`). The `managed_docs` table is a derived runtime record used by projection, not a durable source of repository state; a fresh clone followed by `db rebuild` cannot use it as the precondition for detecting lost authored documentation (`docs/plans/active/consumer-doc-drift-gate.md:42-49`).

## Decision

`zharness audit` guards authored documentation at the boundary of presence, not truth.

- The guard derives the managed path set from the binary's embedded filesystem and checks the repository filesystem. If a managed document remains on disk and `docs/` contains no Markdown outside that managed set, `audit` appends `authored_docs_missing` with `warning` severity to the existing `contract_violations` array. It does not read the runtime managed-document table (`cli/internal/application/audit.go:53-130`).
- The public `AuditReport` keeps its three top-level arrays. The finding's optional `identifier` and `severity` fields make it distinguishable without creating a new report channel (`cli/docs/CONTRACT.md:215-221`).
- `authored` remains the ownership class for consumer-written documentation. The consumer owns its content and its meaning. `zharness` neither creates, rewrites, repairs, or repins authored documents nor claims that their presence makes them correct. Semantic review, citation policy, and any deeper freshness analysis belong to the author or external tooling (`docs/README.md:21-40`).

The rejected alternative is to treat the database's managed-document rows as proof that authored documentation exists. That would make the guard disappear after `db rebuild`, precisely when committed filesystem state is the only durable evidence. Generating prose or adding a new top-level audit array was also rejected: the former violates the documentation boundary, and the latter breaks the public envelope.

## Consequences

- Deleting the authored half of `docs/` while managed documentation remains produces a visible, severity-graded warning instead of a green audit.
- A repository with no managed projection produces no authored-docs finding, and a repository with any Markdown outside the managed set clears it. The check reports presence only; it does not judge whether the document is useful, current, or accurate.
- Audit remains read-only: it writes no documentation, changes no pins, and leaves the database without WAL/SHM sidecars.
- The warning can surface an intentional consumer choice to keep no authored documentation. That is why it is a warning rather than a correctness failure; the consumer remains responsible for deciding whether to act.
- Authored-document quality remains deliberately unautomated by this guard. External link, review, or freshness tooling may add a separate signal without changing this ownership boundary.

## Authority

- `docs/decisions/0004-docs-directory-deletion-655c6ac.md:9-13,45-50` — incident and the unclosed detection gap.
- `docs/plans/active/consumer-doc-drift-gate.md:48-55` — accepted requirements R2 and R6-R9.
- `cli/internal/application/audit.go:16-21,30-130` — report envelope, embedded-set filesystem scan, finding identifier, and severity.
- `cli/docs/CONTRACT.md:215-221` — public `audit` contract.
- `docs/README.md:21-40` — documentation ownership classes and table.
