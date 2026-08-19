# Decision Records

Architecture decisions that are already made, already implemented, and expensive to reverse. Each one exists because the reasoning was otherwise reachable only through a commit message or a session transcript.

An ADR here is a record, not a proposal. It is written after the decision has landed in code, and it states what was decided, why, what was rejected, and what it costs. Nothing in this directory is a plan — plans live in `docs/plans/`.

## Index

| # | Decision | Status |
|---|---|---|
| [0001](0001-markdown-as-source-of-truth.md) | Markdown is the source of truth; `harness.db` is a derived, rebuildable index | Accepted |
| [0002](0002-single-active-plan-resolver.md) | One resolver owns the "at most one active plan" invariant, returning a Stop contract | Accepted |
| [0003](0003-durable-memory-not-wired-into-playbooks.md) | Durable memory ships as an opt-in CLI surface, unwired from the spine playbooks | Accepted |
| [0004](0004-docs-directory-deletion-655c6ac.md) | Recovery position after commit `655c6ac` deleted `docs/` | Accepted |

## Writing a new one

Number it sequentially, name it `NNNN-kebab-case-title.md`, and add a row above. Keep the same five headings: Status, Context, Decision, Consequences, Authority. Every structural claim carries a `path:line` citation that resolves at merge time.
