---
name: work
model: sonnet
version: "1.4.0"
description: "Execution orchestrator after `brainstorm` and `to-plan`. Runs phases wave-by-wave from locked artifacts, verifies each task, and hands off to `check`, `git`, or `handoff`."
license: MIT
argument-hint: "[mode:auto|full|simple|phase] [phase-name?] [--notes?]"
compatibility: Designed for Claude Code
metadata:
  version: "1.4.0"
---

Prefix your first line with `🥷` inline. Be direct: state, next move, evidence. No filler.

Resolve the invocation mode (`auto` by default), then follow `docs/playbooks/work.md` for that mode — it holds this stage's operating logic. Read `docs/WORKFLOW.md` first if the routing is unclear. The lifecycle needs no binary: `zharness` only installs and updates these managed docs and plays no part in running a stage. If the playbook is absent, say so in one line and work from repo-local state (git, plans, scripts).

Arguments: `[mode:auto|full|simple|phase] [phase-name?] [--notes?]` — passed through as-is (default: `auto`).

Defer to: `brainstorm`/`to-plan` when durable artifacts are missing; `check` for the phase gate; `git`/`handoff` are suggested after a clean gate, never automatic.
