---
name: to-plan
version: "1.4.1"
model: opus
description: Generate the approach, phases, waves, and checks inside the locked `docs/plans/active/{slug}.md`. Use after `brainstorm` for a `high-risk` plan; a `normal` plan goes straight to `work`.
argument-hint: "[mode:full|phase] [phase-name?]"
compatibility: Designed for Claude Code
---

Prefix your first line with `🥷` inline. Be direct: executable steps, not planning prose. No filler.

Ask the user only through the agent's question tool (Claude Code `AskUserQuestion`, Codex `request_user_input`, Pi `ask_user_question`, Gemini CLI `ask_user`, omp `ask`): ≤4 per call, recommended option first and labelled `(Recommended)`; plain text only without one.

Resolve the mode, then follow `docs/playbooks/to-plan.md` (`docs/WORKFLOW.md` IF routing is unclear; IF the playbook is absent, say so and work from git and the plan).

Defer to: `work` for a normal plan (lock already wrote its phase); `brainstorm` when the locked input is missing or weak; `work` executes the phase next; `check` gates after implementation.
