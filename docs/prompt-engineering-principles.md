# Prompt Engineering Principles

Reference for writing effective prompts, skills (SKILL.md), and rules. Read this before creating or editing any skill or rule file.

Complements: `skills/craft/skill-creator/references/writing-effective-instructions.md` (HOW), `token-efficiency-criteria.md` (SIZE). This file covers WHY and SYNTAX.

## Core Principle: Context Engineering

Context = all tokens the LLM sees at inference (system prompt + tools + examples + history + data). Not just "the prompt."

- LLMs have a finite **attention budget** — n² pairwise relationships degrade at scale (context rot)
- Guiding rule: find the **smallest set of high-signal tokens** that maximize desired outcome
- Instruction budget: ~150-200 for frontier thinking models; ~50 already consumed by Claude Code harness
- Degradation is **uniform** — adding instructions weakens ALL of them, not just the new ones
- Smaller models degrade exponentially; frontier models degrade linearly

## Formatting Syntax

### XML Tags vs Markdown

| Use | For |
|-----|-----|
| XML tags | Boundaries between content blocks — documents, examples, output schemas, data |
| Markdown headers | Hierarchy within a single instruction block |
| Never both | Don't wrap XML then add redundant Markdown headers inside |

### Inside XML Tags

| Content type | Format | Example |
|---|---|---|
| Rules/constraints | Dash list, no indent, no blank lines | `- Max 200 words\n- Vietnamese only` |
| Examples | Flat key:value, no prose | `input: ...\noutput: ...` |
| Documents/data | Attributes for metadata, flat content | `<document source="x" date="y">content</document>` |
| Output schema | Template directly, no explanation | `{verdict}: {reason}` |
| Context/background | Compact paragraph, no headers inside | One flowing paragraph |

### Bullets vs Paragraphs

- ≥4 independent items (order doesn't matter) → **bullets** (compliance drops in paragraph form)
- ≤3 related items with causal dependency → **paragraph** (preserves logic chain)
- Reasoning context + discrete rules → **paragraph first, then bullets**

Token note: dash (`-`) = 1 token; numbered (`1.`) = 2+ tokens. Prefer dashes.

## Language Rule

| Content | Language | Rationale |
|---------|----------|-----------|
| Instructions, rules, constraints | English | Highest compliance, lowest token cost |
| Few-shot example outputs | Target language | Model captures tone/style |
| Data, documents | Original language | Don't translate — meaning loss |

Vietnamese instructions cost ~50% more tokens for the same semantics due to tokenizer optimization for Latin script.

## Few-Shot Patterns

- 3-4 diverse examples > 8 homogeneous (diversity > quantity)
- Each example must cover a different facet or edge case
- Use XML with `id` attributes: `<example id="1">...</example>`
- Audit examples whenever instructions change — stale examples silently contradict new rules (contamination)
- Token trade-off: each example costs context. Remove if budget is tight.

## System Prompt Altitude

The Goldilocks zone between two failure modes:

| Extreme | Problem |
|---------|---------|
| Brittle if-else logic | Fragile, high maintenance, breaks on new scenarios |
| Vague guidance ("be helpful") | Non-deterministic, model must guess |

**Optimal**: specific enough to guide behavior, flexible enough for model heuristics.

Practice:
1. Start with minimal prompt (essential intent only)
2. Test with best available model
3. Add instructions ONLY when testing shows a failure mode
4. Test: "would removing this line change model behavior?" If no → delete it

## Positive Framing

Negative instructions ("don't hallucinate", "never do X") trigger the **Pink Elephant effect** — increases model focus on the banned behavior.

| ❌ Negative | ✅ Positive |
|---|---|
| Don't hallucinate | Only cite sources from provided documents |
| Never use informal language | Use formal academic tone |
| Don't include code | Respond in prose only |

## Constraint-First Pattern

Put constraints BEFORE the task. Model reads constraints first → applies during generation.

```
Format: 3 bullet points, each under 15 words. Vietnamese. No introduction.

Task: Summarize the attached article.
```

Not:
```
Summarize the attached article. Keep it to 3 bullet points...
```

## Anti-Patterns

| Anti-Pattern | Why it fails | Fix |
|---|---|---|
| Instruction stacking (>10 rules) | Attention degrades monotonically per rule | 5-8 critical rules; progressive disclosure for rest |
| Conflicting instructions | Model oscillates or picks randomly | Audit for contradictions; use conditional framing |
| Overly prescriptive (line-by-line HOW) | Limits model problem-solving | Specify WHAT + WHY, let model decide HOW |
| Cargo-culting CoT | Adds variability on simple tasks | Measure effectiveness before applying |
| Missing output schema | Downstream parsers break silently | Always define schema explicitly |
| Auto-generated instructions | Bloated, not universally applicable | Craft every line deliberately |
| Using prompt as linter | Expensive, unreliable, wastes budget | Use deterministic tools (Biome, eslint) + hooks |

## Cross-Model Awareness

| Model type | Prompting style |
|---|---|
| Reasoning (o-series, extended thinking) | High-level goals. Trust the model. Don't say "think step-by-step" — it manages its own reasoning. |
| Standard GPT / non-thinking | Explicit step-by-step. Define exact output. |
| Latest Claude (Opus 4.7) | Calibrates length to complexity. Follows "be conservative" literally. Shorter by default. |

## The Iteration Rule

- First prompt is never optimal — treat as code: version, test, iterate
- Define success criteria before writing the prompt
- Add instructions only when testing shows they're needed
- Pin model versions in production (e.g., `gpt-4.1-2025-04-14`)
- Re-evaluate prompts on model upgrades — behavior shifts between versions

## Application to Skills & Rules

When writing SKILL.md or rules/*.md in this repo:

- SKILL.md body < 150 lines — move detail to `references/`
- Write instructions in English, imperative form
- One rule per dash-item — atomic, testable
- Use progressive disclosure: SKILL.md → references/ → scripts/
- Constraint-first in every instruction block
- Include 2-3 examples in target output language
- Delete any line that doesn't change model behavior
