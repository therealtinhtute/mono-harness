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

Run `zharness preflight handoff --json`. If `zharness` is missing from PATH or its `version` is below MIN_ZHARNESS_VERSION (`0.8.1` — see `skills/workflow/README.md`), degrade, never halt: print one line naming the fallback and follow `docs/playbooks/handoff.md` directly from repo-local state (git, plans, scripts). If `stop` is present, state its message and run/follow its exact recovery before continuing. Read and follow the returned `playbook` path.

Argument: `[context]` — optional additional context to include in the handoff, passed through as-is.

Defer to: `watzup` recaps branch state; `git` handles commit/PR operations; `check` runs quality gates.
