# SPEC: Mini-GSD Planning System for Claude Code

> Status: **DRAFT** — rewritten after review

## 1. Vision

Mục tiêu là build một **mini-GSD clone** cho Claude Code, nhưng **không phải 1 skill monolith làm hết mọi thứ**.

Thay vào đó, hệ thống planning/spec sẽ có **2 skill core**:

1. **`spec`** — biến raw idea hoặc file inputs thành spec rõ ràng, khóa WHAT
2. **`plan`** — đọc SPEC đã khóa, rồi sinh roadmap + context + phase plans, khóa HOW trong phạm vi spec

Tinh thần là **GSD-style**, nhưng nhỏ hơn, dễ maintain hơn, và fit với repo skills hiện tại.

---

## 2. Product Definition

### 2.1 What this system is

Một **spec-driven development subsystem** cho Claude Code với:
- `.planning/` state folder
- Socratic clarification
- ambiguity gating trước khi plan
- roadmap derivation từ spec
- context locking trước khi task breakdown
- phase plans dạng executable waves

### 2.2 What this system is not

- không phải execution engine
- không auto-code
- không thay toàn bộ GSD ecosystem
- không làm mọi workflow trong 1 skill duy nhất
- không bắt buộc subagent orchestration ở v1

---

## 3. Core Skill Map

| Skill | Role | Input | Output |
|------|------|-------|--------|
| `spec` | clarify + lock WHAT | idea prompt / specified files | `.planning/IDEA.md`, `.planning/SPEC.md` |
| `plan` | derive HOW from locked spec | `.planning/SPEC.md` | `.planning/ROADMAP.md`, `phases/*-CONTEXT.md`, `phases/*-PLAN.md` |

### Key principle
**`plan` never exists without `SPEC.md`.**

---

## 4. GSD-Style Conventions to Keep

Các thứ nên giữ theo hướng GSD:

1. **`.planning/` as first-class state**
2. **spec-first** — không plan nếu spec chưa khóa
3. **structured artifacts** thay vì chat-only planning
4. **Socratic discussion before commitment**
5. **phase-based breakdown**
6. **context engineering before task generation**
7. **explicit user choices at transitions**

---

## 5. `.planning/` File Structure

```text
.planning/
├── IDEA.md
├── SPEC.md
├── ROADMAP.md
└── phases/
    └── {phase-slug}/
        ├── {phase}-CONTEXT.md
        └── {phase}-PLAN.md
```

### Purpose of each file
- `IDEA.md` — raw source material captured from prompt or files
- `SPEC.md` — locked requirements, boundaries, constraints, acceptance
- `ROADMAP.md` — ordered phases derived from SPEC
- `{phase}-CONTEXT.md` — implementation decisions / locked gray-area decisions
- `{phase}-PLAN.md` — executable tasks in waves, grounded in SPEC + CONTEXT

---

## 6. `spec` Skill

## 6.1 Mission
`spec` converts a vague or semi-structured starting point into a **locked planning artifact**.

It is responsible for the **WHAT**, not the implementation task breakdown.

## 6.2 Input Modes

`spec` should support **2 primary modes**:

### Mode A — `idea`
User gives an idea directly in prompt.

Examples:
- “anh muốn build hệ thống inbox AI cho team nhỏ”
- “let’s add provenance graph view”
- “bootstrap project for X”

### Mode B — `files`
User points `spec` at one or more files.

Examples:
- `@file:docs/idea.md`
- `@file:README.md @file:PRD.md`
- issue spec / RFC / note dump / meeting notes

**This is the default non-chat-native mode.**

## 6.3 Supported planning scenarios
Within `idea` or `files`, `spec` should classify the job into one of these scenarios:

1. **new project bootstrap**
2. **new feature bootstrap**
3. **new module / subsystem bootstrap**
4. **refinement of an existing rough spec**

The mode is still `idea` or `files`; the scenario affects the interview questions and SPEC framing.

## 6.4 `spec` workflow

### Step 0 — detect source mode
- if explicit files are provided → `files`
- otherwise → `idea`

### Step 1 — create / refresh `IDEA.md`
- in `idea` mode: write the raw user idea into `.planning/IDEA.md`
- in `files` mode: summarize or mirror the provided source material into `.planning/IDEA.md` with references

### Step 2 — classify scenario
Pick one:
- project bootstrap
- feature bootstrap
- module bootstrap
- refine existing spec

### Step 3 — Socratic clarification loop
Ask questions until the planning surface is clear enough.

Focus areas:
- goal
- target user / actor
- boundaries
- constraints
- success / acceptance
- dependencies / assumptions

### Step 4 — ambiguity scoring
Suggested dimensions:
- goal clarity
- scope clarity
- constraints clarity
- acceptance clarity

If clarity is too low:
- keep interviewing
- or write SPEC with explicit unresolved gaps if user accepts

### Step 5 — write `SPEC.md`
Lock the WHAT.

## 6.5 Required `SPEC.md` structure

```markdown
# SPEC: {title}

## Source Mode
idea | files

## Scenario
project bootstrap | feature bootstrap | module bootstrap | refine existing spec

## Goal
## Users / Actors
## Requirements
## Boundaries
### In Scope
### Out of Scope
## Constraints
## Acceptance Criteria
## Dependencies / Assumptions
## Open Questions
## Ambiguity Report
```

## 6.6 `spec` quality bar

A good `SPEC.md` must make these things explicit:
- what is being built
- who it is for
- what success means
- what is intentionally excluded
- what planner is allowed to assume vs not assume

---

## 7. `plan` Skill

## 7.1 Mission
`plan` transforms a locked SPEC into an executable planning structure.

It is responsible for the **HOW**, but only inside the boundaries already defined by `SPEC.md`.

## 7.2 Hard precondition
Before doing anything, `plan` must verify:
- `.planning/SPEC.md` exists
- SPEC is readable enough to derive phases

If not:
- stop and tell user to run `spec`
- do not invent planning context from thin air

## 7.3 Inputs
Required:
- `.planning/SPEC.md`

Optional:
- codebase files
- architecture docs
- selected phase target only
- extra user constraints

## 7.4 `plan` workflow

### Step 0 — read SPEC
Extract:
- requirements
- boundaries
- acceptance criteria
- constraints
- scenario

### Step 1 — derive `ROADMAP.md`
Split the work into phases.

Each phase should have:
- goal
- deliverables
- dependencies
- why this phase exists

### Step 2 — identify gray areas per phase
For the chosen phase(s), identify implementation decisions that are still ambiguous.

Examples:
- architecture choices
- data/storage decisions
- API shape
- UI flow
- migration strategy
- test strategy

### Step 3 — lock decisions into `{phase}-CONTEXT.md`
This is the GSD-style context-engineering step.

A `CONTEXT.md` should preserve:
- chosen approach
- rejected alternatives if important
- canonical refs
- deferred ideas

### Step 4 — generate `{phase}-PLAN.md`
Turn the phase into executable tasks grouped into waves.

## 7.5 Required `ROADMAP.md` structure

```markdown
# Roadmap: {title}

## Phase 1: {name}
Goal:
Deliverables:
Dependencies:
Risks:

## Phase 2: ...
```

## 7.6 Required `{phase}-CONTEXT.md` structure

```markdown
# Context: {phase}

## Goal
## Gray Areas Reviewed
## Locked Decisions
## Canonical References
## Rejected Options
## Deferred Ideas
```

## 7.7 Required `{phase}-PLAN.md` structure

```markdown
# Plan: {phase}

## Goal
## Inputs
## Wave 1
## Wave 2
## Verification
## Risks / Watch-fors
```

### Task style inside PLAN
Each task should include:
- what to do
- concrete steps
- expected output
- verification method

---

## 8. Relationship to Existing Skills

This mini-system should **compose with** current skills, not replace them.

Suggested handoffs:
- after `spec` → suggest `plan`
- during `plan` when repo understanding is weak → suggest `investigator`
- after implementation later → suggest `reviewer` / `verifier`
- end-of-session → suggest `watzup` / `handoff`
- shipping / git ops → suggest `git`

### Explicit note
Không cần giữ `strategist` trong core flow nếu thấy dư.
Nếu `spec` + `plan` đã đủ rõ thì cứ để `strategist` là optional, thậm chí bỏ khỏi main narrative cũng được.

---

## 9. GSD-Style Recommendations for v1

Nếu bám theo style của GSD, bản mini này nên giữ các phẩm chất sau:

1. **artifact-driven** chứ không chỉ conversational
2. **spec gate before execution**
3. **context locking before task generation**
4. **planning files easy to diff**
5. **same folder convention every time**
6. **phase-by-phase progression**

Không cần clone nguyên hệ slash-command/phân lớp runtime của GSD v1/v2 ở bản đầu.
Nhưng phải clone được **tư duy workflow** của nó.

---

## 10. Recommended v1 Scope

### Build now
- `spec`
- `plan`
- `.planning/` conventions
- idea/files input modes
- project/feature/module bootstrap scenarios
- SPEC → ROADMAP → CONTEXT → PLAN artifact chain

### Do later
- execution-side skills tightly integrated
- auto-resume / persistent state helpers
- richer discussion variants
- explicit review / verify hooks in planning artifacts
- deeper GSD-style command ecosystem

---

## 11. Open Decisions

1. In `files` mode, should `IDEA.md` be a summary or a stitched source digest?
2. Should `plan` support planning just one selected phase first?
3. Should `CONTEXT.md` always be required, or optional for tiny phases?
4. Should `spec` keep a visible ambiguity score in the final artifact?

---

## 12. Final Recommendation

Rewrite the initiative as:

> **A mini-GSD planning subsystem for Claude Code built from two core skills: `spec` and `plan`.**

That gives anh exactly the intended shape:
- spec-driven
- GSD-like
- bootstrap-capable
- planning-heavy
- artifact-first
- but still small enough to ship and iterate safely
