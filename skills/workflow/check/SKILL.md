---
name: check
version: "1.8.1"
description: "Pre-commit and pre-merge gate. Runs tests, lint, build, then reviews security, performance, architecture, and code quality. Acts as the phase gate after `/work`."
model: opus
allowed-tools: "Read Grep Glob Bash"
argument-hint: "[auto|gate|full|bounded]"
tags: [check, review, quality, security, gate]
compatibility: Designed for Claude Code
---

Prefix your first line with `🥷`. Verdict first, evidence for blockers.

Ask the user only through the agent's question tool (Claude Code `AskUserQuestion`, Codex `request_user_input`, Pi `ask_user_question`, Gemini CLI `ask_user`, omp `ask`): ≤4 per call, recommended option first and labelled `(Recommended)`; plain text only without one.

Resolve the mode, then follow [references/check.md](references/check.md) (`docs/WORKFLOW.md` IF routing is unclear).

Resolve cross-stage transitions through the host's skill metadata; do not assume a shared installation parent. Resolve data paths from the consumer repository root. If a required skill or reference is unavailable, report the missing dependency and stop; retained legacy playbooks are not a fallback.

Defer to: `work` is the usual caller; `git` handles commit/push/PR after a clean gate; `brainstorm`/`think` own pre-implementation design.
