---
name: execution-discipline
description: "Global rule: keep agent execution lean — tool-call economy, check-in cadence, and stop-don't-guess escalation. Cuts wasted tokens and long silent sessions."
scope: global
applies_to: all_sessions
---

# Execution Discipline — Global Rule

Pure delta on SOUL (concise/verdict-first), Karpathy (minimal change), and Hard Rule #3 (prove completion). Do not restate those. Three imperatives:

## 1. Tool-call economy

- Call a tool only when the next step consumes its output. No speculative reads, greps, or "let me just check" calls.
- Never re-run a read-only check already run this session (`audit`, `status`, `diff`, `grep`, `--version`, test suites) — reuse the result you already have.
- Batch independent calls in one block. Don't serialize what has no dependency.
- Subagents have overhead; for small tasks a subagent costs more than it saves — see the Agent tool's own guidance, don't spawn unless asked or context-isolation clearly wins.

## 2. Check-in cadence

- Hook progress to the work's own boundaries (phase / wave / logical unit). At each: emit ≤3 lines of what's done + what's next, then continue or pause per the plan.
- Never run a multi-phase or multi-step initiative silently to completion. Long unbroken tool spirals with no surfaced progress are the failure mode.

## 3. Stop — don't guess

When the answer is not in context, the docs, or the code, and the next move is genuinely blocked: **do not fabricate.** Bias to asking the user early — it is the cheapest signal. If you consult instead, cap it at ONE advisor or research pass, then act or ask. Fire this on real blocks only, not minor uncertainty.

---

For "which stage calls what / verifies what," the stage → command → entity contract already lives in `skills/workflow/README.md`'s mapping table — reference it, don't re-derive it.
