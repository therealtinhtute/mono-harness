# Completed Execution Plans

Move a plan here only after the outcome exists, relevant validation passes, and handoff recorded an `absorb:` line (`absorb: none` is valid).

Retain a completed plan only when its recovery or transition details have a current independent audience — someone still needs the waves, commands, or dead ends. Ordinary work should rely on code, tests, decisions, pull requests, and Git history instead of accumulating permanent task narratives.

After absorb, a completed plan **may** be deleted when:

- an ADR and/or native guard already carries every lesson that must survive without this file; and
- no independent audience still needs the run log.

If unsure, **keep**. Do not delete a completed plan to “clean up” before absorb. Do not cite `docs/plans/completed/{slug}.md` as project knowledge; cite the ADR or guard instead.
