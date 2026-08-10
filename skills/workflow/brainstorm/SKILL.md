---
name: brainstorm
version: "4.2.0"
model: opus
description: "Explore options, evaluate trade-offs, and lock the result into `docs/plans/active/{slug}.md` when ready. Use for ideation, architecture decisions, RFC/PRD-to-spec work, and refining an existing plan."
license: MIT
argument-hint: "[idea, @file refs, or trade-off question]"
compatibility: Designed for Claude Code
metadata:
  version: "4.2.0"
---

Prefix your first line with `🥷` inline. Be direct: recommendation first, key trade-off next. No filler.

Resolve the invocation as `explore` or `lock` (`lock` covers lock-from-idea, lock-from-files, and refine), then run `zharness preflight brainstorm --mode {explore|lock} --json`. Missing binary: print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and STOP. Otherwise check its `version` field — below MIN_ZHARNESS_VERSION (`0.8.1` — see `skills/workflow/README.md`): print the same message and STOP; a `dev` build always passes. If `stop` is present, state its message and run/follow its exact recovery before continuing. Read and follow the returned `playbook` path when non-empty.

Argument: `[idea, @file refs, or trade-off question]` — raw input, passed through as-is.

Defer to: `to-plan` after an approved spec lock; `interview` for Q&A-driven requirement extraction instead; `check` for quality gates after implementation, not before.
