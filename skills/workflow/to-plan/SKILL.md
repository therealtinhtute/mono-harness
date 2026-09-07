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

Resolve the invocation mode (`full` by default), then follow `docs/playbooks/to-plan.md` for that mode — it holds this stage's operating logic. Read `docs/WORKFLOW.md` first if the routing is unclear. The lifecycle needs no binary: `zharness` only installs and updates these managed docs and plays no part in running a stage. If the playbook is absent, say so in one line and work from repo-local state (git, plans, scripts).

Arguments: `[mode:full|phase] [phase-name?]` — passed through as-is (default: `full`).

Defer to: `brainstorm` when the locked input is missing or weak; `work` executes the phase next; `check` gates after implementation.
