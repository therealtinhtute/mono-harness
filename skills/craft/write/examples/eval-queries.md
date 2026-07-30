# Trigger eval queries

Use this set to validate that the `description` field activates `write` for the right requests and stays quiet for the wrong ones. Mix EN/VI, formal/casual, short/contextual, and explicit/implicit naming. A reasonable trigger threshold is 0.5.

## Should trigger

1. `Sửa lại đoạn này cho tự nhiên hơn, bớt AI giúp anh.`
2. `Rewrite this update for engineers. Keep it concise.`
3. `Polish these button labels and error messages.`
4. `Viết lại báo cáo này cho dễ scan trên Notion.`
5. `Check the English and Vietnamese versions for consistency.`
6. `Soạn lại thông báo này cho lịch sự hơn.`
7. `My intro paragraph sounds robotic — can you make it warmer?`
8. `Đoạn copy landing này nghe sáo quá, viết lại cho gần tai hơn.`
9. `Shorten this 4-paragraph release note to a tight summary.`
10. `Help me draft a Notion research note from these bullet points.`

Implicit (no naming, but should still trigger):
- `Cái này đọc nó hơi cứng, em xem giúp anh được không.`
- `This reads like a brochure, can you fix it.`

## Should NOT trigger

1. `Write a Python function to deduplicate a list.` — code, not prose
2. `Draft a commit message for these staged changes.` — defer to `git`
3. `Improve this prompt I'm sending to Claude.` — defer to `prompt-leverage`
4. `Add JSDoc comments to this file.` — inline doc comments excluded
5. `Review my PR before I merge.` — defer to `check`

## How to use

1. Run each query through the agent with `write` installed.
2. Observe whether the agent loads `skills/craft/write/SKILL.md`.
3. A should-trigger query passes if loaded ≥50% of runs; should-not-trigger passes if loaded <50%.
4. If trigger rate is off, edit the `description` field in `SKILL.md` and re-run a fresh batch (do not reuse the same queries used for tuning).
