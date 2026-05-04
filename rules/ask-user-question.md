---
name: ask-user-question
description: "Global rule: all questions must use AskUserQuestion tool"
scope: global
applies_to: all_skills
---

# AskUserQuestion Enforcement — Global Rule

This rule applies to **ALL skills** in all contexts.

## Hard Rule

**ALL questions from the agent MUST use `AskUserQuestion` tool.**

- Never ask questions in plain text prose
- Never output question text inline
- Never use placeholder text like "[Ask user about X]"
- Max 4 questions per AskUserQuestion call
- Recommended option MUST be labeled "(Recommended)" — placed first

## Why

Plaintext questions bypass the conversation flow, break audit trails, and violate the project communication standard. `AskUserQuestion` ensures structured, traceable, batched interaction.

## Scope

This rule is **global** — it applies to:
- All slash commands (/interview, /think, /hunt, etc.)
- All skills when invoked
- All sub-agents and spawned contexts
- Any tool call that would result in a question to the user

## Violation

If you catch yourself about to ask a question in plaintext — stop. Call AskUserQuestion instead. No exceptions.