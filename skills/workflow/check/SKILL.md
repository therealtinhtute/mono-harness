---
name: check
version: "1.6.0"
description: "Pre-commit and pre-merge gate. Runs tests, lint, build, then reviews security, performance, architecture, and code quality. Acts as the phase gate after `/work`."
model: opus
allowed-tools: "Read Grep Glob Bash"
argument-hint: "[gate|review|full]"
tags: [check, review, quality, security, gate]
compatibility: Designed for Claude Code
metadata:
  version: "1.6.0"
---

Prefix your first line with `🥷` inline. Be direct: verdict first, evidence for blockers.

Resolve the invocation mode (`full` by default), then follow `docs/playbooks/check.md` for that mode — it holds this stage's operating logic. Read `docs/WORKFLOW.md` first if the routing is unclear. The lifecycle needs no binary: `zharness` only installs and updates these managed docs and plays no part in running a stage. If the playbook is absent, say so in one line and work from repo-local state (git, plans, scripts).

Argument: `[gate|review|full]` — mode, passed through as-is (default: `full`).

Defer to: `work` is the usual caller; `git` handles commit/push/PR after a clean gate; `brainstorm`/`think` own pre-implementation design.
