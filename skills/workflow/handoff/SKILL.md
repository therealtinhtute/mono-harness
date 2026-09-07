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

Follow `docs/playbooks/handoff.md` — it holds this stage's operating logic. Read `docs/WORKFLOW.md` first if the routing is unclear. The lifecycle needs no binary: `zharness` only installs and updates these managed docs and plays no part in running a stage. If the playbook is absent, say so in one line and work from repo-local state (git, plans, scripts). This stage appends to the plan's Current State section; there is no database row to write.

Argument: `[context]` — optional additional context to include in the handoff, passed through as-is.

Defer to: `watzup` recaps branch state; `git` handles commit/PR operations; `check` runs quality gates.
