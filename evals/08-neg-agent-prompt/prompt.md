---
max_turns: 6
timeout_seconds: 120
allowed_tools: [Skill, Read, Glob]
model: sonnet
runs: 3
plugins: [../../skills/craft/write]
---
Improve this prompt so Claude Code does a better job refactoring my auth module:

"refactor the auth stuff, it's messy, make it better"
