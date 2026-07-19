---
name: handoff
version: "1.2.0"
model: sonnet
description: "Prospective: capture current session state into .kit/HANDOFF.md so the next session can resume without context loss."
argument-hint: "[context]"
compatibility: Designed for Claude Code
metadata:
  version: "1.2.0"
---

Prefix your first line with `🥷` inline. Be direct: branch, blocker, next action first. No filler.

Run `zharness --version`. A `dev` build always passes. Otherwise, if the binary is missing or reports a version below MIN_ZHARNESS_VERSION (`0.4.0` — see `skills/workflow/README.md`), print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and STOP.

Ensure docs are present: run `zharness init` if `.kit/docs/` is missing (idempotent — always safe to run).

Read `.kit/docs/playbooks/handoff.md` and follow it exactly — that file is the operating logic; this file only triggers it.

Argument: `[context]` — optional additional context to include in the handoff, passed through as-is.

Defer to: `watzup` recaps branch state at the start of a new session (reads what handoff writes); `git` handles commit operations and PR creation; `check` runs the code quality audit and gate checks.
