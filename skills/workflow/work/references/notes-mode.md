# Notes Mode (`--notes`)

## Output file
`.kit/implementation-notes.md` — append-only. Create on first entry if absent.

## File header (write once)
```
# Implementation Notes
/work --notes — {date}. Spec: {SPEC.md title or prompt slug}.
```

## Trigger: write one entry when a task involves
- A decision the spec didn't cover
- A change from what the plan specified
- A tradeoff between two valid options
- A wrong assumption discovered mid-execution
- Status `DONE_WITH_CONCERNS` or `NEEDS_CONTEXT`

No entry for tasks that ran exactly as planned.

## Entry format
```
## {timestamp} {phase/task}
**Decision**: what you chose and why (1-2 sentences).
**Spec gap**: what the spec or plan didn't cover.
**Tradeoff**: what you gave up.
**Risk**: one sentence.
```

## Rules
- Write the entry immediately after the task — not in a batch at phase end.
- Append only — never edit or delete prior entries.
- Flag absent → skip silently, zero overhead.
