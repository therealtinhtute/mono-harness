# Core Workflow Auto-Triggers

Invoke these skills proactively at the right moment — user does not need to type the slash command.

## Session Start
- Session has uncommitted changes, or `zharness resume` reports `drifted`/`in-progress` → invoke `watzup` before answering.

## Diagnosing Problems
- Error, crash, regression, or "it's not working" before any fix attempt → invoke `hunt`.

## Design and Planning
- Intent is fuzzy or contradictory before planning → invoke `interview` to clarify first.
- New feature, system, or architectural question with scope > 3 files → invoke `brainstorm`.
- Architecture decision, "should we build X", "is this worth it" → invoke `think`.
- Plan locked in `docs/plans/active/{slug}.md` or user says "make a plan" / "what are the steps" → invoke `to-plan`.

## Implementation
- Approved roadmap or spec exists and user says "let's go" / "implement" / "直接改" → invoke `work`.

## Quality and Shipping
- Before any commit, push, or PR → invoke `check`.
- Commit, push, create PR, merge → invoke `git`.

## Session End
- User says done / wrapping up / "see you" / significant work left unmerged → invoke `handoff`.
