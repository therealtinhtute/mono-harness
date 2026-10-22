---
name: mono-mode
version: "1.0.0"
description: "Entry router: match a task to a playbook (feature, bug-fix, investigation, session-pickup) and run its skill steps as a todo list. Use when the route is unclear."
argument-hint: "[task]"
compatibility: Designed for Claude Code
---

Prefix your first line with `🥷` inline. Name the matched playbook first.

Ask the user only through the agent's question tool (Claude Code `AskUserQuestion`, Codex `request_user_input`, Pi `ask_user_question`, Gemini CLI `ask_user`, omp `ask`): ≤4 per call, recommended option first and labelled `(Recommended)`; plain text only without one.

| Task | Playbook |
|---|---|
| Build or change behavior | [playbooks/feature.md](playbooks/feature.md) |
| Error, crash, failing test, regression | [playbooks/bug-fix.md](playbooks/bug-fix.md) |
| Question, review, diagnosis; nothing to change | [playbooks/investigation.md](playbooks/investigation.md) |
| Resume, "where was I", an active plan exists | [playbooks/session-pickup.md](playbooks/session-pickup.md) |
| Learn from recent sessions, retrospective | [playbooks/retro.md](playbooks/retro.md) |

Read the matched playbook. Open a todo list whose first items are its steps copied verbatim, in order. A step you choose not to do stays in the list as `skip: <reason>`; never drop it silently. No playbook fits → say so and ask which one applies.

Each step names a skill; that skill owns the procedure. Resolve it through the host's skill metadata. If a required skill is unavailable, report the missing dependency and stop.

Gated skills — `retro`, `create-cli`, `turbo-mono-platform` — run only from a playbook step here or when the user names them. In `feature`, call `create-cli` for CLI design and `turbo-mono-platform` for that monorepo stack.

Defer to: the skill each playbook step names; `git` only on the owner's explicit request.
