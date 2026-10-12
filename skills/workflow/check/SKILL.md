---
name: check
version: "1.8.0"
description: "Pre-commit and pre-merge gate. Runs tests, lint, build, then reviews security, performance, architecture, and code quality. Acts as the phase gate after `/work`."
model: opus
allowed-tools: "Read Grep Glob Bash"
argument-hint: "[auto|gate|full|bounded]"
tags: [check, review, quality, security, gate]
compatibility: Designed for Claude Code
---

Prefix your first line with `🥷`. verdict first, evidence for blockers.

Ask the user only through the agent's question tool (Claude Code `AskUserQuestion`, Codex `request_user_input`, Pi `ask_user_question`, Gemini CLI `ask_user`, omp `ask`): ≤4 per call, recommended option first and labelled `(Recommended)`; plain text only without one.

Resolve the mode, then follow `docs/playbooks/check.md` (`docs/WORKFLOW.md` IF routing is unclear; IF the playbook is absent, say so and work from git and the plan).

Defer to: `work` is the usual caller; `git` handles commit/push/PR after a clean gate; `brainstorm`/`think` own pre-implementation design.
