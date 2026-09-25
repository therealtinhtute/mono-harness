---
name: retro
description: "Session retrospective: evidence-cited, severity-ranked findings about the agent's environment, then one user-approved fix — a guard or a rerun-gated experiment. Use only when asked."
disable-model-invocation: true
---

Prefix your first line with `🥷` inline. Be direct: the top finding and its evidence first.

Follow `references/retro.md` — it holds this skill's operating logic. Start from `python3 scripts/session-digest.py [path|session-id]`, not the raw log. Change nothing until the user approves the `Next` finding; then fix only that one. A missing `zharness` binary is never a reason to stop.

Defer to: `work` when the fix grows into a locked initiative; `handoff` absorb when closing a plan.
