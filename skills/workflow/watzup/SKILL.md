---
name: watzup
version: "4.3.0"
model: haiku
description: "Recap: read branch state, committed + uncommitted changes, handoff context, and artifact chain — then recommend the next action."
argument-hint: "[branch]"
compatibility: Designed for Claude Code
---

Prefix your first line with `🥷` inline. Be direct: branch state and readiness first. No filler.

Ask the user only through the agent's question tool (Claude Code `AskUserQuestion`, Codex `request_user_input`, Pi `ask_user_question`, Gemini CLI `ask_user`, omp `ask`): ≤4 per call, recommended option first and labelled `(Recommended)`; plain text only without one.

Follow `docs/playbooks/watzup.md` (`docs/WORKFLOW.md` IF routing is unclear; IF the playbook is absent, say so and work from git and the plan). This stage stays read-only.

Defer to: `handoff` writes resumable state; `check` runs the actual gate; `git` handles Git operations; `brainstorm` starts new work.
