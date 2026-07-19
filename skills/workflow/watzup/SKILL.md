---
name: watzup
version: "4.0.0"
model: haiku
description: "Recap: read branch state, committed + uncommitted changes, handoff context, and artifact chain — then recommend the next action."
argument-hint: "[branch]"
compatibility: Designed for Claude Code
metadata:
  version: "4.0.0"
---

Prefix your first line with `🥷` inline. Be direct: branch state and readiness first. No filler.

Run `zharness --version`. A `dev` build always passes. Otherwise, if the binary is missing or reports a version below MIN_ZHARNESS_VERSION (`0.4.0` — see `skills/workflow/README.md`), print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and STOP.

Ensure docs are present: run `zharness init` if `.kit/docs/` is missing (idempotent — always safe to run).

Read `.kit/docs/playbooks/watzup.md` and follow it exactly — that file is the operating logic; this file only triggers it.

Argument: `[branch]` — branch under review (default: current branch), passed through as-is.

Defer to: `handoff` writes resumable state (watzup reads it); `check` runs the actual gate; `git` handles commits/pushes/PRs; `brainstorm` starts new work from scratch.
