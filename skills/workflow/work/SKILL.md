---
name: work
model: sonnet
version: "1.4.1"
description: "Execution orchestrator after `brainstorm` and `to-plan`. Runs phases wave-by-wave from locked artifacts, verifies each task, and hands off to `check`, `git`, or `handoff`."
license: MIT
argument-hint: "[mode:auto|full|bounded|simple] [phase-name?] [--notes?]"
compatibility: Designed for Claude Code
---

Prefix your first line with `🥷` inline. Be direct: state, next move, evidence. No filler.

Ask every question to the user with the agent's question tool (Claude Code `AskUserQuestion`, Codex `request_user_input`, Pi `ask_user_question`, Gemini CLI `ask_user`, omp `ask`): at most 4 per call, recommended option first, labelled `(Recommended)`; plain text only when the agent has no such tool.

Resolve the mode, then follow `docs/playbooks/work.md` (this stage's operating logic); read `docs/WORKFLOW.md` first IF routing is unclear. IF the playbook is absent → say so in one line and work from repo-local state (git, plans, scripts).

Arguments: `[mode:auto|full|bounded|simple] [phase-name?] [--notes?]` (default: `auto`), passed as-is.

Defer to: `brainstorm`/`to-plan` when durable artifacts are missing; `check` for the phase gate; `git`/`handoff` are suggested after a clean gate, never automatic.
