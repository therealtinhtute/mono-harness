---
type: llm
focus: last_message
weight: 1
---
Pass only if ALL hold:
1. The response contains a commit message whose subject says the change fixes pagination's start index (off-by-one / 1-based page offset).
2. It does not claim changes absent from the diff (no new features, tests, or other files).
3. It is a commit message, not a prose rewrite or an essay about the code.
