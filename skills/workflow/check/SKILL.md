---
name: check
version: "1.7.0"
description: "Pre-commit and pre-merge gate. Runs tests, lint, build, then reviews security, performance, architecture, and code quality. Acts as the phase gate after `/work`."
model: opus
allowed-tools: "Read Grep Glob Bash"
argument-hint: "[auto|gate|full|bounded]"
tags: [check, review, quality, security, gate]
compatibility: Designed for Claude Code
metadata:
  version: "1.7.0"
---

Prefix your first line with `🥷` inline. Be direct: verdict first, evidence for blockers.

Resolve the mode, then follow `docs/playbooks/check.md` (this stage's operating logic); read `docs/WORKFLOW.md` first IF routing is unclear. No binary runs the lifecycle. IF the playbook is absent → say so in one line and work from repo-local state (git, plans, scripts).

Argument: `[auto|gate|full|bounded]` (default: `auto`; `review` and `simple` are aliases of `bounded`), passed as-is.

Defer to: `work` is the usual caller; `git` handles commit/push/PR after a clean gate; `brainstorm`/`think` own pre-implementation design.
