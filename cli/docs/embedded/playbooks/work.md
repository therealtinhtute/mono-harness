# Playbook: work

## Purpose

Execute approved work: the next phase of a durable initiative from `docs/plans/active/{slug}.md` (`full`), or a direct change (`bounded`). Full mode appends execution state to the same plan; bounded mode changes only the requested product files, with no lifecycle bookkeeping.

## Preconditions and Modes

1. Resolve mode:
   - `auto` (default) — `full` IF the request names or continues a durable initiative or one of its phases; else `bounded` IF no bounded rejection below applies; else stop and route through `brainstorm` and `to-plan`. The mere existence of an active plan never selects `full`.
   - `full [phase {stable-phase-slug}]` — execute from the active plan. Requires exactly one non-empty plan under `docs/plans/active/`; IF several → report every candidate and stop.
   - `bounded` (alias: `simple`) — known subsystem, bounded files, direct success criterion; no plan. Reject IF scope is unclear, crosses an unfamiliar subsystem, exceeds five files or ~100 changed lines, needs multi-phase coordination, or lacks a known verification path → route through `brainstorm` and `to-plan`.
2. Before reading any plan, print the resolved mode as your first output line (after any required prefix, same line): `mode: {resolved} ({one-line reason})`.
3. IF context was compacted or summarized since you last read the plan → re-read before trusting any earlier anchor.

**Zero-write rule:** bounded/simple mode creates no plans, reports, changesets, or markdown artifacts and never edits an active plan. Its evidence is the Git diff plus captured proof.

## Full Mode

IF the resolved mode is `full` → read `docs/playbooks/work-full.md` now and follow it; it owns the plan sections, execution steps, and status routing. Bounded/simple never loads it.

## Memory

Write `docs/memory/{id}.md` (plain committed file) only on one of three triggers:

- **Fact correction** — write the corrected entry, then set the old file's frontmatter `superseded_by` to the new id, with the date.
- **Durable lesson** — a cross-session learning, architecture decision, or gotcha that would otherwise be rediscovered.
- **Owner preference** — an explicit owner instruction on style, process, or scope meant to persist beyond the session.

NEVER store credentials, secrets, API keys, or token values in a body; record only that a secret exists, its scope, and where to fetch it. Retrieve by grep over `docs/memory/*.md`; discount superseded files but keep them for lineage.

## Exit Conditions

- Full: the phase is `checked` in both the status field and Current State after a clean non-final gate, else `in-progress`; every attempted task has a `## Progress` entry; material decisions are recorded with rationale; each completed wave has a summary line; Current State is resumable; completed work is gated in-session per `work-full.md` step 11.
- Bounded/simple: code and proof shown in the response; zero lifecycle or markdown writes.

escalate_when: ask the owner and stop; never invent — locked schema or requirements would change; the same verification command failed twice; a product rule conflicts.
