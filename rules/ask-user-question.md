---
name: ask-user-question
description: "Global rule: ask every question with the agent's own question tool"
scope: global
applies_to: all_skills
---

# Question Tool — Global Rule

**Hard rule:** ask every question with the agent's own question tool: Claude Code `AskUserQuestion`, Codex `request_user_input`, Gemini CLI `ask_user`, omp `ask`.

- Max 4 questions per call; recommended option first, labelled "(Recommended)"
- Plain-text questions only when the agent has no such tool

Plaintext questions bypass the conversation flow and break audit trails — the question tool ensures structured, traceable interaction.
