---
name: karpathy-guidelines
description: "Global rule: apply Karpathy's 4 behavioral guidelines every session when writing, reviewing, or refactoring code"
scope: global
applies_to: all_sessions
---

# Karpathy Guidelines — Global Rule

Derived from [Andrej Karpathy's observations](https://x.com/karpathy/status/2015883857489522876) on LLM coding pitfalls. Apply every session. Bias toward caution; use judgment for trivial tasks.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

- State assumptions before writing code. Ask via `AskUserQuestion` if uncertain.
- Multiple interpretations? Present them — don't pick silently.
- Simpler approach exists? Say so. Push back when warranted.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No unrequested features, abstractions, configurability, or error handling for impossible cases.
- If you write 200 lines and it could be 50, rewrite it.
- Test: would a senior engineer call this overcomplicated?

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

- Don't "improve" adjacent code, comments, or formatting.
- Match existing style. Don't refactor what isn't broken.
- Remove imports/vars YOUR changes made unused; leave pre-existing dead code alone.
- Test: every changed line traces to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals before starting:
- "Add validation" → write tests for invalid inputs, make them pass
- "Fix the bug" → write a test reproducing it, make it pass
- "Refactor X" → ensure tests pass before and after

For multi-step tasks, state plan inline: `1. [step] → verify: [check]`.

If you catch yourself writing code before stating assumptions, adding unrequested features, or starting without success criteria — stop and correct course.
