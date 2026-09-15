---
name: to-plan
version: "1.3.0"
model: opus
description: Generate the approach, phases, waves, and checks inside the locked `docs/plans/active/{slug}.md`. Use after `brainstorm` for artifact-first implementation planning.
argument-hint: "[mode:full|phase] [phase-name?]"
compatibility: Designed for Claude Code
metadata:
  version: "1.3.0"
---

Prefix your first line with `🥷` inline. Be direct: executable steps, not planning prose. No filler.

Resolve the mode, then follow `docs/playbooks/to-plan.md` (this stage's operating logic); read `docs/WORKFLOW.md` first IF routing is unclear. No binary runs the lifecycle. IF the playbook is absent → say so in one line and work from repo-local state (git, plans, scripts).

Arguments: `[mode:full|phase] [phase-name?]` (default: `full`), passed as-is.

Defer to: `brainstorm` when the locked input is missing or weak; `work` executes the phase next; `check` gates after implementation.
