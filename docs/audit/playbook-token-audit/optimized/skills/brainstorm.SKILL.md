---
name: brainstorm
version: "4.3.0"
model: opus
description: "Explore options, evaluate trade-offs, and lock the result into `docs/plans/active/{slug}.md` when ready. Use for ideation, architecture decisions, RFC/PRD-to-spec work, and refining an existing plan."
license: MIT
argument-hint: "[idea, @file refs, or trade-off question]"
compatibility: Designed for Claude Code
metadata:
  version: "4.3.0"
---

Prefix your first line with `🥷` inline. Be direct: recommendation first, key trade-off next. No filler.

Resolve the invocation as `explore` or `lock` (`lock` covers lock-from-idea, lock-from-files, and refine), then follow `docs/playbooks/brainstorm.md` (this stage's operating logic); read `docs/WORKFLOW.md` first IF routing is unclear. No binary runs the lifecycle. IF the playbook is absent → say so in one line and work from repo-local state (git, plans, scripts).

Argument: `[idea, @file refs, or trade-off question]`, passed as-is.

Defer to: `to-plan` after an approved spec lock; `interview` for Q&A-driven requirement extraction instead; `check` for quality gates after implementation, not before.
