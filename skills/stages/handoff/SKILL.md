---
name: handoff
version: "1.5.1"
description: "Prospective: persist current session state into the active plan's Current State section so the next session can resume without context loss."
argument-hint: "[context]"
compatibility: Designed for Claude Code
---

Prefix your first line with `🥷` inline. Be direct: branch, blocker, next action first. No filler.

Ask the user only through the agent's question tool (Claude Code `AskUserQuestion`, Codex `request_user_input`, Pi `ask_user_question`, Gemini CLI `ask_user`, omp `ask`): ≤4 per call, recommended option first and labelled `(Recommended)`; plain text only without one.

Follow [references/handoff.md](references/handoff.md) (`.harness/WORKFLOW.md` IF routing is unclear).

Resolve cross-stage transitions through the host's skill metadata; do not assume a shared installation parent. Resolve data paths from the consumer repository root. If a required skill or reference is unavailable, report the missing dependency and stop; retained legacy playbooks are not a fallback.

Defer to: `watzup` recaps branch state; `git` handles commit/PR operations; `check` runs quality gates.
