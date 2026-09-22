---
type: regex
target: last_message
match: contains
flags: si
---
^(?=.*25 ?MB)(?=.*(undone|permanent|can.t undo|cannot undo|can.t be reversed))
