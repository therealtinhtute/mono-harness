# Global Claude Code Rules

User: technical engineer, cross-stack, non-native English speaker.

## Hard Rules (override everything)

1. **Questions** → `AskUserQuestion` tool only. Max 4 per call. Recommended option first.
2. **Deletion** → `trash`, never `rm`.
3. **Prove before done** → run tests, check output. No "done" without evidence.
4. **Read before edit** → never modify unread code.
5. **Minimal change** → no unrequested features, refactors, or abstractions.

## Workflow (non-trivial: 3+ steps or multi-file)

1. Explore → read relevant files
2. Plan → short numbered plan in `.kit/plans/`, wait for confirmation
3. Implement → test after each meaningful step
4. Verify → prove it works
5. Close → mark done, update `.kit/HANDOFF.md` if needed

## Questioning

For small tasks with a clear default: state assumption inline, proceed.
Use `AskUserQuestion` for: ambiguous scope, risky actions, missing required info.
Never fabricate API behavior or external system behavior.

## Output Conventions

All generated artifacts → `.kit/`. Plans → `.kit/plans/{date}-{slug}/`. Reports → `.kit/reports/{skill}/`.

## Session Hygiene

- Compact at ~50% context; `/clear` and restart at ~70%.
- After any correction → save to memory system.
- At session end → update `.kit/HANDOFF.md`.

## Subagents

Cannot use `AskUserQuestion`. Resolve all ambiguity before delegating.
Define expected output format in every delegation prompt.

## Rules

- `rules/karpathy-guidelines.md` — Think Before Coding, Simplicity First, Surgical Changes, Goal-Driven
- `rules/english.md` — passive correction at end of reply
- `rules/ask-user-question.md` — AskUserQuestion enforcement details
