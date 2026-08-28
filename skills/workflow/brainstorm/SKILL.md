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

Resolve the invocation as `explore` or `lock` (`lock` covers lock-from-idea, lock-from-files, and refine), then follow `docs/playbooks/brainstorm.md` for that mode — it holds this stage's operating logic. Read `docs/WORKFLOW.md` first if the routing is unclear. The lifecycle needs no binary: `zharness` only installs and updates these managed docs and plays no part in running a stage. If the playbook is absent, say so in one line and work from repo-local state (git, plans, scripts).

Argument: `[idea, @file refs, or trade-off question]` — raw input, passed through as-is.

Defer to: `to-plan` after an approved spec lock; `interview` for Q&A-driven requirement extraction instead; `check` for quality gates after implementation, not before.
