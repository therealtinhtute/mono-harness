# Playbook: work

## Preconditions and Modes

1. Resolve mode:
   - `auto` (default) — `full` IF the request names or continues a durable initiative; else `bounded` unless rejected below; else route to `brainstorm`. An active plan's existence never selects `full`.
   - `full [phase {slug}]` — execute the next phase of the one non-empty plan under `docs/plans/active/`; IF several → list them and stop.
   - `bounded` (alias: `simple`) — a direct change with a known verification path. Reject IF scope is unclear, crosses an unfamiliar subsystem, exceeds five files or ~100 changed lines, or needs phases.
2. Before reading any plan, print the resolved mode as your first output line: `mode: {resolved} ({reason})`.
3. IF context was compacted since you read the plan → re-read it.

**Zero-write rule:** bounded/simple mode creates no plans, reports, changesets, or markdown artifacts and never edits an active plan. Evidence: the diff plus captured proof in the response.

IF the mode is `full` → read [work-full.md](work-full.md) now and follow it.

## Memory

Write `docs/memory/{id}.md` only for a fact correction (set the old file's `superseded_by`), a durable lesson, or a lasting owner preference. Never store secret values; record only that a secret exists, its scope, and where to fetch it.

escalate_when: ask the owner and stop — locked schema or requirements would change; the same verification command failed twice; a product rule conflicts.
