---
name: work
model: opus
version: "1.3.0"
description: "Execution orchestrator after `brainstorm` and `to-plan`. Runs phases wave-by-wave from locked artifacts, verifies each task, and hands off to `check`, `git`, or `handoff`."
license: MIT
argument-hint: "[mode:auto|full|simple|phase] [phase-name?] [--notes?]"
compatibility: Designed for Claude Code
metadata:
  version: "1.3.0"
---

Prefix your first line with `🥷` inline. Be direct: state, next move, evidence. No filler.

Run `zharness --version`. Below MIN_ZHARNESS_VERSION (`0.4.1` — see `skills/workflow/README.md`) or missing: print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and STOP. A `dev` build always passes.

Run `zharness preflight work --mode {auto|full|simple|phase} --json` using the invocation mode (`auto` by default). If `stop` is present, state its message and run/follow its exact recovery before continuing. Read and follow the returned `playbook` path when non-empty; reduced mode may instead use repository-native guidance.

Arguments: `[mode:auto|full|simple|phase] [phase-name?] [--notes?]` — passed through as-is (default: `auto`).

Defer to: `brainstorm`/`to-plan` when durable artifacts are missing; `check` for the phase gate; `git`/`handoff` are suggested after a clean gate, never automatic.
