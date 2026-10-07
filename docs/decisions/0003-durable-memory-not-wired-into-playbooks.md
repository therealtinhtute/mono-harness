# 0003 — Durable memory ships unwired from the spine playbooks

## Status

Superseded by [0006](0006-v015-authority.md) — historical record, not a current runbook. The `zharness memory` CLI surface and the `memories` index it describes were deleted in v0.15; `docs/memory/` remains an authored markdown surface with no index. Accepted and recorded as NG3 of the `durable-memory` initiative, left standing after `retrieval-router` shipped ranking on top of it (`git show 3bab3c0:docs/plans/completed/durable-memory.md`, `git show 3bab3c0:docs/plans/completed/retrieval-router.md`).

## Context

The `durable-memory` initiative added a repo-scoped, cross-session memory store: markdown entries under `docs/memory/`, a `memories` index derived from them (`cli/internal/infrastructure/migrations.go:235`), and three CLI subcommands — `memory add`, `memory get`, `memory query`. `retrieval-router` then added keyword ranking to `memory query`.

The obvious next move was to make the six spine playbooks call it: have `watzup` read memory at session start, have `handoff` write memory at session end. That move was deliberately not made.

## Decision

No spine playbook gains a mandatory memory step. Memory writes are opt-in, and whether the playbooks should call them is a decision for a later initiative.

The reasoning is that every mandatory playbook step is paid on every invocation, forever, by every consumer repo. A step that reads memory at session start costs tokens in every session including the many that need no memory at all, and a step that writes memory at handoff produces an entry whether or not anything durable was learned. Neither cost is recoverable once the step is embedded in a released binary's playbooks, because the playbooks are projected into consumer repos and a removal needs its own release.

Wiring it in is cheap to do later and expensive to undo. Leaving it out is the reversible direction.

## Consequences

- `zharness memory` works fully and is invoked only when an agent or owner decides a fact is worth keeping. It is a tool, not a ceremony.
- The memory store may stay near-empty on repos where nobody chooses to write to it. That is an accepted outcome, not a defect: an empty store costs nothing, while a store filled by an automatic handoff step costs a read on every future session.
- Any future initiative that does wire memory into the playbooks must justify the recurring cost, not just the feature.
- The separate `~/.claude` personal auto-memory is untouched and unrelated (NG4). The two systems are not merged and are not intended to be.

## Authority

- `git show 3bab3c0:docs/plans/completed/durable-memory.md` — NG3 (no mandatory playbook step), NG2 (CLI-only surface), NG4 (no merge with personal memory), R4 (rebuildable from committed markdown).
- `git show 3bab3c0:docs/plans/completed/retrieval-router.md` — added ranking without adding a playbook step.
- `cli/internal/infrastructure/migrations.go:235` — the `memories` table, mirroring `plan_index`.
