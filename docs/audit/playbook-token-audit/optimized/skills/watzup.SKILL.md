---
name: watzup
version: "4.2.0"
model: haiku
description: "Recap: read branch state, committed + uncommitted changes, handoff context, and artifact chain — then recommend the next action."
argument-hint: "[branch]"
compatibility: Designed for Claude Code
metadata:
  version: "4.2.0"
---

Prefix your first line with `🥷` inline. Be direct: branch state and readiness first. No filler.

Follow `docs/playbooks/watzup.md` (this stage's operating logic); read `docs/WORKFLOW.md` first IF routing is unclear. No binary runs the lifecycle. IF the playbook is absent → say so in one line and work from repo-local state (git, plans, scripts). This stage stays read-only.

Argument: `[branch]` (default: current branch), passed as-is.

Defer to: `handoff` writes resumable state; `check` runs the actual gate; `git` handles Git operations; `brainstorm` starts new work.
