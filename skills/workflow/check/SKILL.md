---
name: check
version: "1.4.0"
description: "Pre-commit and pre-merge gate. Runs tests, lint, build, then reviews security, performance, architecture, and code quality. Acts as the phase gate after `/work`."
model: opus
allowed-tools: "Read Grep Glob Bash"
argument-hint: "[gate|review|full]"
tags: [check, review, quality, security, gate]
compatibility: Designed for Claude Code
metadata:
  version: "1.4.0"
---

Prefix your first line with `🥷` inline. Be direct: verdict first, evidence for blockers.

Run `zharness --version`. Below MIN_ZHARNESS_VERSION (`0.4.0` — see `skills/workflow/README.md`) or missing: print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and STOP. A `dev` build always passes.

Ensure docs are present: run `zharness init` if `.kit/docs/` is missing (idempotent — always safe to run).

Read `.kit/docs/playbooks/check.md` and follow it exactly — that file is the operating logic; this file only triggers it.

Argument: `[gate|review|full]` — mode, passed through as-is (default: `full`).

Defer to: `work` is the usual caller (per-phase gate); `git` handles commit/push/PR after a clean gate; `brainstorm`/`think` for pre-implementation design instead of post-hoc review.
