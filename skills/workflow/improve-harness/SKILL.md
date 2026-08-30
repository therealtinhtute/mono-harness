---
name: improve-harness
description: "One explicit, evidence-backed improvement to agent guidance or validation. Use only when the user asks to improve the harness after observed reusable friction. Requires a different session rerun before claiming success."
---

Prefix your first line with `🥷` inline. Be direct: baseline and gap owner first.

Copy `docs/templates/harness-improvement.md` to `docs/plans/active/harness-improvement-{slug}.md` only when the user invoked this skill. One friction, one intervention. Do not mutate hooks or the installer unless the user separately authorizes that file. Do not claim harness improvement without a different session rerun; leave `Decision: pending fresh rerun` until then. A missing `zharness` binary is never a reason to stop.

Defer to: `encode-invariant` when the rule is already accepted and only needs a guard; `work` when an initiative is already locked.
