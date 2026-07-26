---
name: check
version: "1.5.0"
description: "Pre-commit and pre-merge gate. Runs tests, lint, build, then reviews security, performance, architecture, and code quality. Acts as the phase gate after `/work`."
model: opus
allowed-tools: "Read Grep Glob Bash"
argument-hint: "[gate|review|full]"
tags: [check, review, quality, security, gate]
compatibility: Designed for Claude Code
metadata:
  version: "1.5.0"
---

Prefix your first line with `🥷` inline. Be direct: verdict first, evidence for blockers.

Run `zharness --version`. Below MIN_ZHARNESS_VERSION (`0.4.1` — see `skills/workflow/README.md`) or missing: print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and STOP. A `dev` build always passes.

Run `zharness preflight check --mode {gate|review|full} --json` using the invocation mode (`full` by default). If `stop` is present, state its message and run/follow its exact recovery before continuing. Read and follow the returned `playbook` path when non-empty; reduced review may instead use repository-native guidance.

Argument: `[gate|review|full]` — mode, passed through as-is (default: `full`).

Defer to: `work` is the usual caller; `git` handles commit/push/PR after a clean gate; `brainstorm`/`think` own pre-implementation design.
