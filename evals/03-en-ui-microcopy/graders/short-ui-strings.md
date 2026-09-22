---
type: llm
focus: last_message
weight: 1
---
Pass only if ALL hold:
1. Exactly three rewritten strings are returned (upload error, empty state, delete confirm), with little or no surrounding commentary.
2. Each string is 15 words or fewer.
3. None of them keeps filler such as "We're sorry, but unfortunately", "It looks like", "at the moment", "absolutely sure", or "Please be aware".
4. The upload error keeps the 25 MB limit, and the delete confirm still warns that deletion is permanent or can't be undone.
