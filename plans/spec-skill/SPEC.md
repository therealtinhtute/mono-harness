# SPEC: Spec Skill (SPDD + Spec-Driven Planning)

> Status: **DRAFT** — Pending review by external agent

## 1. Concept & Vision

Bộ skill cho **planning và discussion** theo Structured-Prompt-Driven Development (SPDD) + spec-driven development. User viết idea vào `.planning/IDEA.md`, chạy skill → output SPEC.md + ROADMAP.md → discuss decisions → executable PLAN.md. Execution tách riêng, user tự chạy skills khác khi sẵn sàng.

**Design principles:**
- Strict ambiguity gate (như GSD) — đảm bảo task đúng trước khi execute
- 2 execution modes: `fast` (minimal) và `deep` (full investigation)
- `.planning/` folder convention (GSD style)
- Output structure: `.planning/IDEA.md` → `SPEC.md` → `CONTEXT.md` → `PLAN.md`
- Skill gợi ý nhưng không force — user control freak

---

## 2. Skill Architecture

### 2.1 Skill Map

| Skill | Mode | Trigger | Output |
|-------|------|---------|--------|
| **`spec`** | fast/deep | User invoke `spec` skill | Full SPDD planning loop: IDEA → SPEC → ROADMAP → CONTEXT → PLAN |

**Single skill duy nhất** — `spec` — bao gồm tất cả planning phases. Không tách ra nhiều skills.

### 2.2 Execution Modes

**`fast` mode:**
- Minimal interview rounds (1-2)
- Research: skipped by default unless user asks
- Ambiguity gate: relaxed (proceed after minimal rounds)
- Output: compact SPEC + PLAN

**`deep` mode:**
- Full GSD-style interview với ambiguity scoring
- Research options surfaced via AskUserQuestion (web search, codebase grep)
- Strict ambiguity gate (≤0.20)
- Full CONTEXT.md + detailed PLAN.md

**Mode selection:** AskUserQuestion prompt khi skill invoke (header: "Mode")

---

## 3. File Structure

```
.planning/
├── IDEA.md                    # User's raw idea (user writes or @file: reference)
├── SPEC.md                    # Locked requirements + boundaries + acceptance criteria
├── ROADMAP.md                 # Phased execution roadmap
└── phases/
    └── {phase-slug}/
        ├── {phase}-CONTEXT.md  # Implementation decisions from discuss
        └── {phase}-PLAN.md      # Executable task plan for this phase
```

---

## 4. Phase-by-Phase Behavior

### 4.1 `spec` — Full Planning Loop

**Step 0: Mode Detection**
- If `--fast` or `--deep` in $ARGUMENTS → use that mode
- If no flag → AskUserQuestion: "Which mode?" (fast / deep)
- Both modes use AskUserQuestion for all choices

**Step 1: Initialize — Detect or Bootstrap `.planning/`**

```
if .planning/IDEA.md exists:
    Load IDEA.md content
    AskUserQuestion: "Found IDEA.md. Start planning?"
    → If yes: continue to Step 2
    → If no: exit
else:
    AskUserQuestion: "Where is your idea?"
    → "Write it now" → create .planning/IDEA.md, user types
    → "@file:path" → Read and use that file
    → "New project" → bootstrap fresh .planning/ structure
```

**Step 2: SPEC Phase (WHAT)**

Socratic interview loop với ambiguity scoring:

| Dimension | Weight | Minimum |
|-----------|--------|---------|
| Goal Clarity | 35% | 0.75 |
| Boundary Clarity | 25% | 0.70 |
| Constraint Clarity | 20% | 0.65 |
| Acceptance Criteria | 20% | 0.70 |

Ambiguity gate: ≤0.20 AND all dimensions ≥ minimum → proceed to write SPEC.md

**Interview perspectives** (rotate per round):
1. Researcher — "What exists today related to this?"
2. Simplifier — "What's the irreducible core?"
3. Boundary Keeper — "What will NOT be done?"
4. Failure Analyst — "What's the worst mis-requirement?"
5. Seed Closer — "What would make this completely clear?"

**After each round:** Update scores, display, check gate

**Max 6 rounds.** If gate not passed after 6 rounds:
- AskUserQuestion: "Write SPEC.md with gaps flagged?" / "Keep talking"
- `--fast`: Auto-proceed after 2 rounds regardless

**Output:** `.planning/SPEC.md`
- `## Requirements` — numbered, falsifiable (current state → target state → acceptance criterion)
- `## Boundaries` — In scope / Out of scope (explicit lists)
- `## Ambiguity Report` — dimensions below minimum flagged for planner

**Step 3: ROADMAP Phase**

Derive phases from SPEC.md requirements.

```
AskUserQuestion (multiSelect):
  "How should we split this into phases?"
  → Option: [phase name] — [brief goal]
  → Add phase, remove phase, reorder
```

**Output:** `.planning/ROADMAP.md`
```
# Roadmap: {project title}

## Phase 1: {name}
**Goal:** {from SPEC.md}
**Deliverables:** {list}

## Phase 2: {name}
...
```

**Step 4: CONTEXT Phase (HOW — Implementation Decisions)**

Per phase (iterate from ROADMAP.md):

```
gray_areas = identify_gray_areas(phase_goal, SPEC.md, prior_context)

AskUserQuestion (multiSelect: true):
  "Which areas to discuss?"
  → Show 3-5 phase-specific gray areas
```

**Then for each selected area:**
- AskUserQuestion: concrete question about implementation choice
- Accumulate decisions into CONTEXT.md
- Canonical refs: add docs/specs user references during discussion

**Output:** `.planning/phases/{slug}/{phase}-CONTEXT.md`
```
<domain>
<decisions>
  ### [Category]
  - [Locked decision]
<canonical_refs>
<deferred_ideas>
```

**Step 5: PLAN Phase (Tasks)**

Read: SPEC.md + CONTEXT.md + ROADMAP.md phase entry

**Research step (deep mode only):**
```
AskUserQuestion:
  "Research before planning?"
  → "Skip" / "Quick grep" / "Full search"
```
Research findings → appended to phase PLAN.md

**Planning step:**
Generate executable tasks:

```
## Phase {N}: {name}
**Goal:** {from ROADMAP}

### Wave 1 (parallel)
1. [Task] — {steps} — verify: {criterion}
2. [Task] — {steps} — verify: {criterion}

### Wave 2 (sequential)
3. [Task] — {steps} — verify: {criterion}
```

**Task structure:**
- Description (specific, actionable)
- Steps (numbered, atomic)
- Verification approach (how to confirm done)
- Estimated complexity (wave assignment)

**Output:** `.planning/phases/{slug}/{phase}-PLAN.md`

---

## 5. Output Summary (end of `spec`)

```
.spec complete.

.planning/
├── IDEA.md
├── SPEC.md          — {N} requirements, ambiguity: {score}
├── ROADMAP.md       — {M} phases
└── phases/
    └── {phase-slug}/
        ├── {phase}-CONTEXT.md
        └── {phase}-PLAN.md    — {P} tasks across {W} waves

Next: [suggest execution skills, e.g. /check /review /watzup]
```

---

## 6. Key Design Decisions

1. **Single skill `spec`** — not split into multiple skills. `fast`/`deep` flags handle variation.
2. **Strict ambiguity gate** — GSD-style scoring ≥0.20 threshold. Non-negotiable for quality.
3. **`.planning/` prefix** — GSD convention, keeps repo clean.
4. **AskUserQuestion for ALL choices** — mode, research, phase selection, continue/abort. Never assume.
5. **Skill suggests, user controls** — exit message lists next steps but doesn't auto-execute.
6. **Research is opt-in per phase** — not a separate phase. Deep mode asks per phase.
7. **CONTEXT.md per phase** — locked implementation decisions feed into PLAN.md.
8. **PLAN.md per phase** — wave-grouped executable tasks. Not one big PLAN.md.

---

## 7. What's NOT Built

- No execution phase (execute, verify, complete — use existing skills)
- No slash command aliases — direct skill invocation only
- No auto-commit — user commits when ready
- No subagent orchestration — single agent, no gsd-* subagents
- No multi-runtime support (GSD-style Claude Code/Copilot/etc.) — Claude Code only

---

## 8. Unknowns (deferred)

| Item | Reason deferred | Owner |
|------|-----------------|-------|
| `--auto` mode (skip all questions) | Need `--fast` and `--deep` stable first | User decides later |
| Integration with existing skills (`/check`, `/review`) | Execution side not in scope yet | User |
| VS Code extension / TUI | Not needed for initial release | — |
| Persistent checkpoint resume | Nice-to-have for long projects | User |
