# 0006 — v0.15 deleted the derived index; live authority is ARCHITECTURE + CONTRACT

## Status

Accepted. 2026-08-30. Overlay on ADRs 0001–0003 and 0005, which remain historical records and are not rewritten.

## Context

v0.15 deleted the lifecycle CLI and SQLite from source (`docs/plans/completed/zharness-v015-slim.md`). The live map did not move with it. Playbooks still named a lifecycle ledger, DB-mirroring frontmatter, and ULID check ids. ADRs 0001, 0002, 0003, and 0005 still describe `harness.db`, `ResolveActivePlan`, `zharness memory`, and `zharness audit` as if they were the running system. Agents following those files reconstructed a control plane that no longer exists.

The invariant “at most one active plan” had been encoded in `ResolveActivePlan` (`docs/decisions/0002-single-active-plan-resolver.md`). That function is gone. Leaving the rule in playbook prose only reopened consumer-adoption D1.

## Decision

Current authority for how the harness works is `docs/ARCHITECTURE.md` plus `cli/docs/CONTRACT.md`.

ADRs 0001–0003 and 0005 stay as accepted historical records of the 0.14 dual-write / resolver / memory-CLI / audit-CLI decisions. They are not current runbooks. Do not rewrite their bodies; a rewritten record is a falsified record.

The “at most one non-empty file under `docs/plans/active/`” invariant is dual-encoded again: playbook guidance, plus `zharness_guard_at_most_one_active_plan` in ZGUARD-CORE. It is not a lifecycle CLI.

Rejected: resurrecting `harness.db`, `preflight`, `zharness validate`, or `ResolveActivePlan` to “make the old ADRs true again.”

## Consequences

- A new session that reads ARCHITECTURE + CONTRACT + playbooks cannot invent `zharness query` as a required step.
- Dated audits under `docs/audit/` that describe 0.14 remain evidence of residue, not instructions.
- The cost is one more ADR to open before trusting 0001–0003/0005 as “how it works now.”

## Authority

- `docs/ARCHITECTURE.md` — v0.15 slim: three verbs, markdown is the record, hook guards.
- `cli/docs/CONTRACT.md` — command surface and where the guarantees live.
- `docs/plans/completed/zharness-v015-slim.md` — the deletion.
- `docs/audit/harness-engineering-gap-audit.md` — H1/H2/H4, 2026-08-30.
- Owner lock `docs/plans/completed/playbook-truth-and-guards.md` R3.
