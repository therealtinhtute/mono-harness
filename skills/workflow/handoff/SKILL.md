---
name: handoff
version: "1.3.0"
model: sonnet
description: "Prospective: persist current session state into the active plan's Current State section and a DB handoff row so the next session can resume without context loss."
argument-hint: "[context]"
compatibility: Designed for Claude Code
metadata:
  version: "1.3.0"
---

Prefix your first line with `🥷` inline. Be direct: branch, blocker, next action first. No filler.

Run `zharness --version`. A `dev` build always passes. Otherwise, if the binary is missing or reports a version below MIN_ZHARNESS_VERSION (`0.4.1` — see `skills/workflow/README.md`), print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and STOP.

Run `zharness preflight handoff --json`. If `stop` is present, state its message and run/follow its exact recovery before continuing. Read and follow the returned `playbook` path.

Argument: `[context]` — optional additional context to include in the handoff, passed through as-is.

Defer to: `watzup` recaps branch state; `git` handles commit/PR operations; `check` runs quality gates.
