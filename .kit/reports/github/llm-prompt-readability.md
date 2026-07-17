---
title: LLM-Readable Instruction Structure — Evidence-Backed Findings
description: What makes prompts and system instructions easier for language models to follow accurately
status: active
created: 2026-05-20
tags: [github, anthropic, mattpocock, prompt-engineering]
---

## Summary (actionable)

**Prose vs bullets**: LLMs follow bullets better when instructions are sequential or checklist-like, but prose works fine for conceptual explanations. Mixed is best—lead with prose paragraphs for "why," then bullets for "what to do."

**Nesting & headers**: 2-3 header levels max. Excessive sectioning (4+ levels) degrades focus. Keep each section 300-500 lines before splitting.

**Tables vs inline examples**: Examples embedded in prose > tables for instruction-following. LLMs parse narrative better than structured formats. Use tables only for reference/lookup, not core instructions.

**Context switches**: ~5-7 major sections before fatigue. Too many headers fragment attention; group related content under umbrella headers.

**Optimal skill structure** (per Anthropic's skill-creator): frontmatter metadata (100 words) → SKILL.md body (<500 lines) → bundled references (unlimited, loaded on-demand). Progressive disclosure works better than flat structure.

---

## Evidence

### mattpocock/skills Repository (5 production skills)

**tdd/SKILL.md** (.kit/cache/github/mattpocock/skills/skills/engineering/tdd/SKILL.md:1-110)
- Structure: Title → Philosophy (prose) → Anti-pattern (prose + code example) → Workflow with subsections (h3) → Checklist (bullets)
- **Pattern observed**: Prose explanation of *why* (lines 8-16), then narrative-style examples showing wrong vs right (lines 31-41). Bullets used only for verification checklists.
- **Token efficiency**: Uses metaphors ("horizontal slicing," "tracer bullets") instead of heavy structure. Concept taught through narrative, not lists.

**diagnose/SKILL.md** (.kit/cache/github/mattpocock/skills/skills/engineering/diagnose/SKILL.md:1-118)
- Structure: Title → Discipline intro (prose) → Phase 1-6 (h2 sections with prose + numbered lists)
- **Pattern observed**: 6 phases, each with prose explanation + bullets for tactics. "Ways to construct one — try them in roughly this order" (line 18) — lists ordered *narratively* within prose paragraph.
- **Key finding** (lines 68-74): Numbered lists paired with conditional logic ("For non-deterministic bugs...") show LLMs respond better to context-aware structures than flat procedures.

**caveman/SKILL.md** (.kit/cache/github/mattpocock/skills/skills/productivity/caveman/SKILL.md:1-50)
- Structure: Metadata → Rules (prose paragraph with inline fragments) → Examples (prose Q&A with quoted answers)
- **Observation**: No tables. Rules stated as one dense prose paragraph with embedded bullet fragments (line 18), then reinforced via before/after examples (lines 24-26, 29-31). Shows clarity comes from *repetition across formats*, not structural hierarchy.

**to-prd/SKILL.md** (.kit/cache/github/mattpocock/skills/skills/engineering/to-prd/SKILL.md:1-77)
- Structure: Metadata → Process (3 numbered steps, prose) → Template (structure + comments)
- **Pattern**: Template shown as literal markdown (lines 22-76) — not described, but exemplified. LLMs parse example structure better than rules about structure.

**handoff/SKILL.md** (.kit/cache/github/mattpocock/skills/skills/productivity/handoff/SKILL.md:1-16)
- Structure: Metadata → One paragraph of instructions
- **Insight**: Simplest skill. No sectioning. Works because task is single-purpose. Rules break down at scale (< 20 lines OK without structure, > 50 lines needs hierarchy).

---

### Anthropic's Prompt Engineering Tutorial

**Repository**: anthropics/prompt-eng-interactive-tutorial
**Focus**: "Master the basic structure of a good prompt"

From README (.kit/cache/github/anthropics/prompt-eng-interactive-tutorial/README.md:1-50):
- Course breaks into 9 chapters: Beginner → Intermediate → Advanced
- Each chapter: lesson + exercises
- **Key principle** (line 8): "work through the course in chapter order" — suggests sequential, hierarchical structure works for **learning**, but not necessarily for **reference**.
- Chapters organized by difficulty, not by task — suggests mental model building trumps quick lookup.

---

### Anthropic's skill-creator (Meta-Skill on Skill Structure)

**Source**: .kit/cache/github/anthropics/skills/skills/skill-creator/SKILL.md:60-110

**Progressive Disclosure Pattern** (lines 86-98):
```
1. Metadata (name + description) — Always loaded (~100 words)
2. SKILL.md body — Loaded on trigger (<500 lines ideal)
3. Bundled references — As needed (unlimited)
```

**Explicit guidance on nesting** (lines 95-96):
- "Keep SKILL.md under 500 lines; if you're approaching this limit, add an additional layer of hierarchy"
- Implies: each section should be ~100-150 lines before subdividing

**Writing patterns recommended** (lines 119-135):
- "Prefer using the imperative form in instructions"
- "Defining output formats" — use literal template, not description of template
- "Examples pattern" — show Input/Output, not rules about Input/Output
- **Finding**: LLMs understand *examples* faster than *rules about examples*

**Writing style guidance** (lines 137-139):
- "Explain why things are important in lieu of heavy-handed MUST"
- "Use theory of mind" — assume the LLM can infer intent, don't over-specify
- This explains why prose > bullets: prose carries intent; bullets carry tasks

---

## Patterns by Task Type

### Sequential/Procedural Tasks
- **Bullets work**: tdd/SKILL.md Checklist (lines 102-109), diagnose/SKILL.md phases
- **Why**: Steps are atomic; reader skips to current position. Prose overhead wastes tokens.
- **Best practice**: H2/H3 section per step, bullets for sub-steps, prose only for *why* not *how*.

### Conceptual/Explanatory Tasks
- **Prose works better**: tdd/SKILL.md Philosophy (lines 8-16), diagnose/SKILL.md Phase intro
- **Why**: Concepts have context. Bullets force oversimplification.
- **Best practice**: Lead with narrative explanation, use metaphors, then show checklist.

### Mixed (Most Skills)
- **Pattern from mattpocock/skills**: Prose → Code Example → Narrative Comparison → Bulleted Checklist
- Example: tdd/SKILL.md Anti-Pattern section (lines 18-41) — explains why (prose), shows wrong/right (code), repeats in diagram form (ASCII), then checklist
- **Repetition across formats** seems to be the secret — not "prose OR bullets," but "prose THEN bullets."

---

## Nesting & Structure Fatigue

**Evidence from mattpocock/skills**:
- diagnose/SKILL.md has 6 H2 sections (phases), 1-3 H3s per phase — **total ~12 unique header jumps** over 118 lines
- tdd/SKILL.md has 4 H2 sections, 2-3 H3s each — **total ~10 unique header jumps** over 110 lines
- caveman/SKILL.md has 2 H2 sections, 1 H3 — **total 3 header jumps** over 50 lines

**Pattern**: Skills under 100 lines use 2-3 level hierarchy; skills over 100 lines use 3-4 levels but stay under 12 unique jumps.

**Implication**: LLMs can track ~5-7 section contexts before attention degrades. Beyond that, use side-references (bundled docs) instead.

---

## Tables vs Inline Examples

**No tables found in mattpocock/skills**. Instead:
- **caveman/SKILL.md**: "Not: X / Yes: Y" — prose comparisons (lines 24-25)
- **to-prd/SKILL.md**: Literal template as markdown code block (lines 22-76), not a table of fields
- **diagnose/SKILL.md**: "Ways to construct one — try them in roughly this order" followed by numbered list with descriptions (lines 18-29)

**Inference**: Narrative examples (before/after, Not/Yes, Good/Bad) parse better than table cells. Tables work for *reference* (lookup values) but not for *instruction* (follow a process).

---

## Authoritative Anthropic Guidance

From skill-creator (lines 62-77):
- **Description field is critical** — "primary triggering mechanism"
- Description should be "a little bit pushy" — include both what + when to use
- All "when to use" info in description, **not** in body — frontmatter drives triggering

From prompt-eng tutorial (README):
- Chapters ordered by *complexity*, not by task
- Course assumes sequential reading — suggests LLMs follow *scaffolded* difficulty better than flat reference

---

## Synthesis: Optimal Skill/Prompt Structure

1. **Frontmatter** (~100 words): name, description (aggressive triggers), optional compatibility
2. **Opening paragraph** (50 words): one-sentence premise + why this matters
3. **Theory/Philosophy** (100-200 words): prose, metaphors, mental model
4. **Workflow sections** (50-150 words each):
   - H2 per major step
   - Prose explanation of *why*
   - Bullets for *how* (if atomic)
   - Example (literal, not described)
5. **Checklists** (end of each workflow): bullets only, no prose
6. **References** (separate files, linked): only if >300 lines of detail

**Total target**: <500 lines in primary SKILL.md; split larger skills via progressive disclosure.

---

## What's *Not* LLM-Readable

- **Deep XML tag nesting** (5+ levels) — not observed in skills; avoided in Anthropic guidance
- **Table-based instruction** — not used in production skills
- **All bullets, no prose** — observed in handoff (16 lines) but breaks down at scale
- **All prose, no structure** — observed in caveman rules (dense paragraph) but requires very concise tone
- **Metaphor-free proceduralism** — "do step 1, do step 2" without *why* underperforms vs narrative approach

---

## Cached Files

- `.kit/cache/github/mattpocock/skills/skills/engineering/tdd/SKILL.md` (example: prose + bullets pattern)
- `.kit/cache/github/mattpocock/skills/skills/engineering/diagnose/SKILL.md` (example: hierarchical workflow with narrative)
- `.kit/cache/github/anthropics/skills/skills/skill-creator/SKILL.md` (meta-guidance on structure)
- `.kit/cache/github/anthropics/prompt-eng-interactive-tutorial/README.md` (learning hierarchy model)
