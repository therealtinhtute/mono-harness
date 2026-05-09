---
name: write
description: Write/edit EN/VI prose so it sounds natural, concise, audience-aware. Triggers - write, rewrite, shorten, polish, đổi giọng, sửa câu, bớt AI/báo cáo, docs/UI/report copy. Not for code or commits.
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

Pick exactly one mode. Decision order:

1. button/label/toast/error/empty-state → `ui`
2. help/instruction/how-to → `docs`
3. policy/HR/announcement/formal email → `formal`
4. landing/promo/social/hero copy → `marketing`
5. Notion report / research note / decision memo → `notion-report`
6. technical article / builder write-up → `builder`
7. user wants warmer, less dry prose → `playful` (VI) / `playful-lite` (EN)
8. EN+VI pair or bilingual consistency check → `bilingual`
9. otherwise → `clean`

Modes available: EN — `clean`, `builder`, `playful-lite`, `docs`. VI — `clean`, `builder`, `playful`, `docs`, `ui`, `marketing`, `formal`, `notion-report`. Shared — `bilingual`.

For `notion-report` with diagrams, also load `references/write-vi-notion-illustrations.md`. The writing skill owns whether a visual is needed, its job, and its caption — not visual-detail rules.

## Defer to other skills

- Commit messages, PR titles, branch names → `git`
- Improving raw prompts for AI agents → `prompt-leverage`
- Post-delivery code review or release gate → `check`

## Output

- Default: **return only the final prose**.
- If the user asks to compare tones: return **at most two versions** with short labels.
- If required context is missing: ask **one short blocking question** and stop.
