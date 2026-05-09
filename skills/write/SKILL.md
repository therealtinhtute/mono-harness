---
name: write
description: Edit or write prose in English or Vietnamese so it reads natural, concise, and context-aware. Use when the user explicitly asks to write, rewrite, shorten, polish, đổi giọng, sửa câu, bớt AI, bớt báo cáo, chỉnh copy/report/docs/UI text, or check bilingual consistency. Not for code comments, commit messages, or inline docs unless explicitly requested.
---

# write

Turn rough prose into the right prose for the right reader.

## Pre-flight

1. **Do we have the source text?** If the user wants an edit but did not provide the text, ask for the exact text and stop.
2. **Do we know the audience?** If reader/audience is unclear and cannot be inferred, ask before rewriting.
3. **Do we know the job?** Distinguish at least one of:
   - rewrite lightly
   - shorten
   - change tone
   - write from scratch
   - docs / UI / marketing / formal / report
4. **Detect language and mode.**
   - English prose → load `references/write-en-core.md` plus `references/write-en-style.md`
   - Vietnamese prose → load `references/write-vi-core.md` plus the mode file that fits best
   - Mixed Vietnamese/English → also load `references/write-bilingual.md`
   - If the task is a **Notion report with illustrations/diagrams**, also load `references/write-vi-notion-illustrations.md`
5. **If doing a final pass**, read `references/checklist-before-delivery.md`.

## Rewrite strength

Use one of three internal levels:
- `light` — minimal edits, keep structure nearly intact
- `medium` — trim filler, improve rhythm, rewrite locally
- `strong` — substantial rewrite while preserving meaning

Default:
- already decent text → `light`
- obvious AI/report tone → `medium`
- wrong mode or very clumsy draft → `strong`

## Hard rules

- **Meaning first, style second.**
- **Do not guess missing source text.**
- **Do not silently restructure large sections** unless the user asked for that.
- **Keep register and terminology consistent.**
- **Preserve source personality** if the original already has a clear voice.
- **Do not explain changes unless asked.** Default to returning the final prose only.
- **If two directions are both good, return at most two labeled versions** instead of one compromised hybrid.

## Routing

### English modes
- `clean`
- `builder`
- `playful-lite`
- `docs`

### Vietnamese modes
- `clean`
- `builder`
- `playful`
- `docs`
- `ui`
- `marketing`
- `formal`
- `notion-report`

### Shared mode
- `bilingual`

If the user does not specify a mode, prefer:
1. UI/app/help text → `ui` or `docs`
2. technical article / report / builder note → `builder`
3. wants lighter, warmer, less dry prose → `playful` / `playful-lite`
4. announcement / policy / HR / official tone → `formal`
5. landing / promo / social → `marketing`
6. otherwise → `clean`

For `notion-report` tasks that also involve diagrams/illustrations, keep the writing skill responsible for:
- deciding whether a visual is actually needed
- defining the visual's job and caption
- enforcing minimal, scan-friendly presentation

Use the illustration reference for those rules; do not overload `SKILL.md` with visual-detail guidance.

## Output

- Default: **return only the final prose**.
- If the user asks to compare tones: return **at most two versions** with short labels.
- If required context is missing: ask **one short blocking question** and stop.
