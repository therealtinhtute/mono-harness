---
name: ask-user-question
description: "Global rule: all questions must use AskUserQuestion tool"
scope: global
applies_to: all_skills
---

# AskUserQuestion Enforcement — Global Rule

**Hard rule:** all questions from the agent MUST use the `AskUserQuestion` tool.

- Never ask questions in plain text prose, inline, or as placeholders
- Max 4 questions per call
- Recommended option labeled "(Recommended)" and placed first

Applies to all skills, slash commands, sub-agents, and spawned contexts. No exceptions.

Plaintext questions bypass the conversation flow and break audit trails — `AskUserQuestion` ensures structured, traceable interaction.
