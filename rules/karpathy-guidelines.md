---
name: karpathy-guidelines
description: "Global rule: Karpathy's coding guidelines plus output and close-out discipline"
scope: global
applies_to: all_sessions
---

# Karpathy Guidelines — Global Rule

From [Andrej Karpathy's observations](https://x.com/karpathy/status/2015883857489522876) on LLM coding pitfalls.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

- State assumptions before coding: inline for minor ambiguity, then proceed; when truly uncertain, ask via the agent's question tool.
- Multiple interpretations? Present them — don't pick silently.
- Simpler approach, or a flawed one? Say so in one sentence first.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No unrequested features, abstractions, configurability, or error handling for impossible cases.
- 200 lines that could be 50? Rewrite it.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

- Don't "improve" adjacent code, comments, or formatting.
- Match existing style. Don't refactor what isn't broken.
- Remove imports/vars YOUR changes made unused; leave pre-existing dead code alone.
- Test: every changed line traces to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Turn tasks into verifiable goals first:
- "Add validation" → write tests for invalid inputs, make them pass
- "Fix the bug" → write a test reproducing it, make it pass
- "Refactor X" → ensure tests pass before and after

For risky or multi-step tasks, state the plan inline: `1. [step] → verify: [check]`. Small tasks: act directly, then verify.

## 5. Output and Close-out

- Code, commands, diffs, and concrete decisions over prose; prose only for decisions, risks, blockers, or non-obvious rationale; no basics, generic closers, or filler.
- Done is production-ready: error handling, types, and edge cases the task can hit. No placeholders or TODOs unless requested; comment only non-obvious logic.
- Flag security, data-loss, or correctness risk in one line.
- Before non-trivial edits, read the relevant files, patterns, and tests (in an unfamiliar codebase also `AGENTS.md`/`CLAUDE.md`, README, test commands). Search with `rg`.
- Implement in focused increments; verify smallest-first, widening with blast radius.
- Close out with what changed, what verification ran, and remaining risk.
