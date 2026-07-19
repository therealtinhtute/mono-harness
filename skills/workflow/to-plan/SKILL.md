---
name: to-plan
version: "1.1.0"
model: opus
description: Generate roadmap, phase context, and executable phase plans from a locked `.kit/planning/SPEC.md`. Use after `brainstorm` for artifact-first implementation planning.
argument-hint: "[mode:full|phase] [phase-name?]"
compatibility: Designed for Claude Code
metadata:
  version: "1.1.0"
---

Prefix your first line with `🥷` inline. Be direct: executable steps, not planning prose. No filler.

Run `zharness --version`. Below MIN_ZHARNESS_VERSION (`0.4.0` — see `skills/workflow/README.md`) or missing: print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and STOP. A `dev` build always passes.

Ensure docs are present: run `zharness init` if `.kit/docs/` is missing (idempotent — always safe to run).

Read `.kit/docs/playbooks/to-plan.md` and follow it exactly — that file is the operating logic; this file only triggers it.

Arguments: `[mode:full|phase] [phase-name?]` — passed through as-is (default: `full`).

Defer to: `brainstorm` when `.kit/planning/SPEC.md` is missing or too weak to plan from; `work` executes the phase next; `check` gates after implementation.
