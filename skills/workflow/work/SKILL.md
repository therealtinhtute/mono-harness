---
name: work
model: opus
version: "1.2.0"
description: "Execution orchestrator after `brainstorm` and `to-plan`. Runs phases wave-by-wave from locked artifacts, verifies each task, and hands off to `check`, `git`, or `handoff`."
license: MIT
argument-hint: "[mode:auto|full|simple|phase] [phase-name?] [--notes?]"
compatibility: Designed for Claude Code
metadata:
  version: "1.2.0"
---

Prefix your first line with `🥷` inline. Be direct: state, next move, evidence. No filler.

Run `zharness --version`. Below MIN_ZHARNESS_VERSION (`0.4.0` — see `skills/workflow/README.md`) or missing: print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and STOP. A `dev` build always passes.

Ensure docs are present: run `zharness init` if `.kit/docs/` is missing (idempotent — always safe to run).

Read `.kit/docs/playbooks/work.md` and follow it exactly — that file is the operating logic; this file only triggers it.

Arguments: `[mode:auto|full|simple|phase] [phase-name?] [--notes?]` — passed through as-is (default: `auto`).

Defer to: `brainstorm`/`to-plan` when required artifacts are missing (full mode); `check` for the per-phase gate; `git`/`handoff` — suggested on clean gate, never run automatically.
