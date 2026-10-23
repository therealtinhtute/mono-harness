# Playbook: work

## Preconditions and Modes

1. Resolve mode:
   - `auto` (default) — `full` IF the request names or continues a durable initiative; else `bounded` unless rejected below; else route to `brainstorm`. An active plan's existence never selects `full`.
   - `full [phase {slug}]` — execute the next phase of the one non-empty plan under `.harness/plans/active/`; IF several → list them and stop.
   - `bounded` (alias: `simple`) — a direct change with a known verification path. Decide from uncertainty, impact, dependency and recoverability, not size: a large mechanical change with a known check stays bounded. Reject IF the intent or a product choice is unresolved, the change is hard to reverse or touches a public contract, security or data, or it needs ordered dependent steps; a small change can still be rejected.
2. Before reading any plan, print the resolved mode as your first output line: `mode: {resolved} ({reason})`.
3. IF context was compacted since you read the plan → re-read it.

IF the mode is `bounded` → read [work-bounded.md](work-bounded.md) now and follow it.
IF the mode is `full` → read [work-full.md](work-full.md) now and follow it.

## Memory

Write `docs/memory/{id}.md` only for a fact correction (set the old file's `superseded_by`), a durable lesson, or a lasting owner preference. Never store secret values; record only that a secret exists, its scope, and where to fetch it.

escalate_when: ask the owner and stop — locked schema or requirements would change; a verification command fails again with no new evidence since its last run; a product rule conflicts.
