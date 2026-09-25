---
name: brainstorm
version: "5.0.0"
model: opus
description: "Explore trade-offs, grill fuzzy intent or a plan, and lock the Goal into `docs/plans/active/{slug}.md`. Use for ideation, architecture decisions, grill/interview requests, and RFC/PRD-to-spec work."
license: MIT
argument-hint: "[explore|grill|raw|lock|refine] [idea, @file refs, plan path, or trade-off question]"
compatibility: Designed for Claude Code
metadata:
  version: "5.0.0"
---

Prefix your first line with `🥷` inline. Be direct: recommendation first, key trade-off next. No filler.

Resolve the mode from a leading subcommand (`explore`, `grill`, `raw`, `lock`, `refine`) or else from the request shape, then follow `docs/playbooks/brainstorm.md` (this stage's operating logic); read `docs/WORKFLOW.md` first IF routing is unclear. No binary runs the lifecycle. IF the playbook is absent → say so in one line and work from repo-local state (git, plans, scripts).

Argument: `[subcommand] [idea, @file refs, plan path, or trade-off question]`; the rest is passed as-is.

Defer to: `to-plan` after an approved lock; `check` for quality gates after implementation.
