# Core Workflow Auto-Triggers

Invoke these skills proactively at the right moment — user does not need to type the slash command.

## Session Start
- Session has uncommitted changes, or an active plan under `docs/plans/active/` is mid-phase → invoke `watzup` before answering.

## Diagnosing Problems
- Error, crash, regression, or "it's not working" before any fix attempt → invoke `hunt`.

## Workflow

- For small tasks, act directly and verify the result.
- For non-trivial changes, explore relevant files, patterns, tests, and
  project-level instructions first.
- Check repo-level `AGENTS.md`, `CLAUDE.md`, README, docs, and test commands
  before making non-trivial edits in unfamiliar codebases.
- Use `rg` or `rg --files` for search.
- State a short plan only when it helps coordination or when the change is
  risky.
- Implement in focused increments.
- Run the smallest useful verification after meaningful changes, then broaden
  verification when the blast radius is larger.
- Close out with what changed, what verification ran, and any remaining risk.

## Design and Planning
- Intent is fuzzy or contradictory before planning → invoke `interview` to clarify first.
- New feature, system, or architectural question with scope > 3 files → invoke `brainstorm`.
- Architecture decision, "should we build X", "is this worth it" → invoke `think`.
- Plan locked in `docs/plans/active/{slug}.md` or user says "make a plan" / "what are the steps" → invoke `to-plan`.

## Implementation
- Approved roadmap or spec exists and user says "let's go" / "implement" / "cook" / "làm đi" → invoke `work`.

## Quality and Shipping
- Before any commit, push, or PR → invoke `check`.
- Commit, push, create PR, merge → invoke `git`.

## Session End
- User says done / wrapping up / "see you" / significant work left unmerged → invoke `handoff`.
