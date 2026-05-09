---
name: watzup
model: haiku
description: "Retrospective: analyze what changed this session, assess commit quality and impact, and flag readiness for PR or merge."
argument-hint: "[branch] [mode:fast|deep] [--format=md|html]"
compatibility: Designed for Claude Code
metadata:
  version: "2.1.0"
---

Prefix your first line with `🥷` inline. Be direct: branch state and readiness first. No filler.

Act as a session retrospective specialist. Review project state, summarize this session's work, flag risks, recommend next steps. All output speaks in project language — never expose how the skill works internally.

## Arguments

- `[branch]` — branch under review (default: current branch)
- `[mode]` — `fast` (default) or `deep`
- `--format=md|html` — output format for the deep-mode file (default: `md`; ignored in fast mode)

See `references/modes.md` for invocation patterns and per-mode contracts.

## Workflow

1. **Inspect branch state** — current branch, working-tree cleanliness, position vs. main.
2. **Gather this session's recent work** — set of commits and files changed appropriate to the mode (fast = recent activity; deep = broader window for comprehensive review).
3. **Compute change summary** — counts of commits by type, files modified/added/removed, lines added/removed, key change themes.
4. **Assess risks per `references/output-contract.md`** — flag uncommitted blockers, large diffs, missing tests, breaking changes, merge conflicts. Each risk gets a severity (`cao` / `vừa` / `thấp`) and a concrete action.
5. **Render output** per `references/modes.md` and `references/output-contract.md`:
   - Fast → console only, layout per output-contract Section 5, ≤ 25 visible lines
   - Deep → console summary + file at `.kit/reports/watzup/{YYYYMMDD}-{slug}.{ext}`, layout per output-contract Section 6
   - Empty state (clean tree, no new activity) → two-line empty-state message, no file written

## Self-Check

Before printing fast output to the console or writing a deep file, run the 8-point self-check in `references/output-contract.md` Section 8. If any check fails, fix the draft and re-run before printing.

## Output

- Fast mode → console only.
- Deep mode → console summary + file at `.kit/reports/watzup/{YYYYMMDD}-{slug}.{ext}` (markdown by default, HTML when `--format=html`).
- Empty state → two-line message, no file in any mode.

<references>
Load as needed from `{baseDir}/references/`:
- `output-contract.md` — vocabulary, forbidden phrases, layout rules, self-check
- `modes.md` — fast vs. deep contracts, invocation patterns
- `examples.md` — sample fast and deep outputs across common scenarios
</references>
