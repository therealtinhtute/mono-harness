---
name: handoff
version: "1.4.0"
model: sonnet
description: "Prospective: persist current session state into the active plan's Current State section so the next session can resume without context loss."
argument-hint: "[context]"
compatibility: Designed for Claude Code
metadata:
  version: "1.4.0"
---

Prefix your first line with `🥷` inline. Be direct: branch, blocker, next action first. No filler.

Follow `docs/playbooks/handoff.md` (this stage's operating logic); read `docs/WORKFLOW.md` first IF routing is unclear. No binary runs the lifecycle. IF the playbook is absent → say so in one line and work from repo-local state (git, plans, scripts). This stage writes to the plan's Current State section; there is no database row.

Argument: `[context]`, optional extra context, passed as-is.

Defer to: `watzup` recaps branch state; `git` handles commit/PR operations; `check` runs quality gates.
