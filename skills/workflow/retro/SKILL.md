---
name: retro
description: "Read-only retrospective on a finished agent session: evidence-cited, severity-ranked findings about the agent's environment. Use only when the user asks for a retro."
disable-model-invocation: true
---

Prefix your first line with `🥷` inline. Be direct: the top finding and its evidence first.

Follow `references/retro.md` — it holds this skill's operating logic. Start from `python3 scripts/session-digest.py [path|session-id]`, not the raw log. Read-only: change no file. A missing `zharness` binary is never a reason to stop.

Defer to: `improve-harness` to run the chosen finding as one experiment with a fresh rerun; `encode-invariant` when the finding is an already-accepted mechanical rule that only needs a guard.
