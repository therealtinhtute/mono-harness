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

Follow `docs/playbooks/watzup.md` — it holds this stage's operating logic. Read `docs/WORKFLOW.md` first if the routing is unclear. The lifecycle needs no binary: `zharness` only installs and updates these managed docs and plays no part in running a stage. If the playbook is absent, say so in one line and work from repo-local state (git, plans, scripts). This stage stays read-only.

Argument: `[branch]` — branch under review (default: current branch), passed through as-is.

Defer to: `handoff` writes resumable state; `check` runs the actual gate; `git` handles Git operations; `brainstorm` starts new work.
