---
name: brainstorm
version: "4.1.0"
model: opus
description: "Explore options, evaluate trade-offs, and lock the result into `.kit/planning/SPEC.md` when ready. Use for ideation, architecture decisions, RFC/PRD-to-spec work, and refining an existing spec."
license: MIT
argument-hint: "[idea, @file refs, or trade-off question]"
compatibility: Designed for Claude Code
metadata:
  version: "4.1.0"
---

Prefix your first line with `🥷` inline. Be direct: recommendation first, key trade-off next. No filler.

Run `zharness --version`. Below MIN_ZHARNESS_VERSION (`0.2.0` — see `skills/workflow/README.md`) or missing: print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and STOP. A `dev` build always passes.

Ensure docs are present: run `zharness init` if `.kit/docs/` is missing (idempotent — always safe to run).

Read `.kit/docs/playbooks/brainstorm.md` and follow it exactly — that file is the operating logic; this file only triggers it.

Argument: `[idea, @file refs, or trade-off question]` — raw input, passed through as-is.

Defer to: `to-plan` after an approved spec lock; `interview` for Q&A-driven requirement extraction instead; `check` for quality gates after implementation, not before.
