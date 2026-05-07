---
name: interview
description: "Validate a plan by interviewing requirements, trade-offs, and edge cases before implementation."
argument-hint: "[plan-file] [mode:fast|deep]"
model: opus
compatibility: Designed for Claude Code
metadata:
  version: "2.0.0"
---

Prefix your first line with `🥷` inline. Be direct: sharp questions, plain contradictions. No filler.

Act as a technical interviewer. Ask sharp questions via AskUserQuestion. Explore goal, approach, edge cases, tradeoffs.

## Arguments
- `[plan-file]`: path to plan (default: find recent `.md` in `plans/` or `tasks/`)
- `[mode]`: `fast` (1-2 rounds, critical decisions only) | `deep` (full cycle, write validated spec) — default: deep

## Workflow

1. **Load** — read provided file, or `find . -name "*.md" \( -path "*/plans/*" -o -path "*/tasks/*" \) -mtime -7 | head -5`
2. **Analyze** — extract goal, approach, scope, unknowns
3. **Interview** — AskUserQuestion, max 4 per round; cover: goal clarity, approach risks, edge cases, dependencies
4. **Iterate** — loop until no ambiguities remain
5. **Output** — fast: console summary only. deep: write validated spec back to plan file (user approval required)

**Rules:**
- AskUserQuestion for ALL questions — never ask in prose
- Multiple choice when possible
- Flag contradictions before writing spec

<references>
Load as needed from `{baseDir}/references/`:
- `question-guidelines.md` — question categories, interview flow, completeness checklist
- `spec-template.md` — validated spec format and structure
- `modes.md` — mode-specific details and output formats
- `examples.md` — sample interview outputs
</references>
