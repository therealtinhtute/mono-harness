---
name: brainstorm
version: "5.1.0"
model: opus
description: "Explore trade-offs, grill fuzzy intent or a plan, and lock the Goal into `docs/plans/active/{slug}.md`. Use for ideation, architecture decisions, grill/interview requests, and RFC/PRD-to-spec work."
license: MIT
argument-hint: "[explore|grill|raw|lock|refine] [idea, @file refs, plan path, or trade-off question]"
compatibility: Designed for Claude Code
---

Prefix your first line with `🥷`. Recommendation first, key trade-off next.

Ask the user only through the agent's question tool (Claude Code `AskUserQuestion`, Codex `request_user_input`, Pi `ask_user_question`, Gemini CLI `ask_user`, omp `ask`): ≤4 per call, recommended option first and labelled `(Recommended)`; plain text only without one.

Resolve the mode, then follow `docs/playbooks/brainstorm.md` (`docs/WORKFLOW.md` IF routing is unclear; IF the playbook is absent, say so and work from git and the plan).

Defer to: `to-plan` after an approved lock; `check` for quality gates after implementation.
