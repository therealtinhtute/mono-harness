---
name: watzup
version: "4.1.0"
model: haiku
description: "Recap: read branch state, committed + uncommitted changes, handoff context, and artifact chain — then recommend the next action."
argument-hint: "[branch]"
compatibility: Designed for Claude Code
metadata:
  version: "4.1.0"
---

Prefix your first line with `🥷` inline. Be direct: branch state and readiness first. No filler.

Run `zharness preflight watzup --json`. If `zharness` is missing from PATH or its `version` is below MIN_ZHARNESS_VERSION (`0.8.1` — see `skills/workflow/README.md`), degrade, never halt: print one line naming the fallback and follow `docs/playbooks/watzup.md` directly from repo-local state (git, plans, scripts). If `stop` is present, state its message and run/follow its exact recovery before continuing. Read and follow the returned `playbook` path when non-empty; reduced mode remains read-only.

Argument: `[branch]` — branch under review (default: current branch), passed through as-is.

Defer to: `handoff` writes resumable state; `check` runs the actual gate; `git` handles Git operations; `brainstorm` starts new work.
