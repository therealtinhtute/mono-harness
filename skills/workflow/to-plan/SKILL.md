---
name: to-plan
version: "1.2.0"
model: opus
description: Generate the approach, phases, waves, and checks inside the locked `docs/plans/active/{slug}.md`. Use after `brainstorm` for artifact-first implementation planning.
argument-hint: "[mode:full|phase] [phase-name?]"
compatibility: Designed for Claude Code
metadata:
  version: "1.2.0"
---

Prefix your first line with `🥷` inline. Be direct: executable steps, not planning prose. No filler.

Run `zharness --version`. Below MIN_ZHARNESS_VERSION (`0.4.1` — see `skills/workflow/README.md`) or missing: print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and STOP. A `dev` build always passes.

Run `zharness preflight to-plan --mode {full|phase} --json` using the invocation mode (`full` by default). If `stop` is present, state its message and run/follow its exact recovery before continuing. Read and follow the returned `playbook` path.

Arguments: `[mode:full|phase] [phase-name?]` — passed through as-is (default: `full`).

Defer to: `brainstorm` when the locked input is missing or weak; `work` executes the phase next; `check` gates after implementation.
