---
name: handoff
version: "1.4.0"
model: haiku
description: "Prospective: persist current session state into the active plan's Current State section so the next session can resume without context loss."
argument-hint: "[context]"
compatibility: Designed for Claude Code
---

Prefix your first line with `🥷` inline. Be direct: branch, blocker, next action first. No filler.

Ask every question to the user with the agent's question tool (Claude Code `AskUserQuestion`, Codex `request_user_input`, Pi `ask_user_question`, Gemini CLI `ask_user`, omp `ask`): at most 4 per call, recommended option first, labelled `(Recommended)`; plain text only when the agent has no such tool.

Follow `docs/playbooks/handoff.md` (this stage's operating logic); read `docs/WORKFLOW.md` first IF routing is unclear. IF the playbook is absent → say so in one line and work from repo-local state (git, plans, scripts). This stage writes to the plan's Current State section; there is no database row.

Argument: `[context]`, optional extra context, passed as-is.

Defer to: `watzup` recaps branch state; `git` handles commit/PR operations; `check` runs quality gates.
