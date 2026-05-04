# SPEC: `spec` + `plan` Skills

> Status: **DRAFT** — updated after review pivot away from single all-in-one skill

## 1. Concept & Vision

Không build **1 skill bự làm tất cả**.

Thay vào đó build một bộ skill planning gọn, bám spec-driven development:

1. **`spec`** — chốt WHAT
2. **`plan`** — chốt HOW + task breakdown, nhưng **phải dựa trên SPEC đã khóa**

Execution, review, verify, wrap-up vẫn tận dụng các skill hiện có (`investigator`, `strategist`, `reviewer`, `verifier`, `watzup`, `git`).

## 2. Goals

### `spec`
Nhận đầu vào từ một trong các nguồn sau:
- raw idea do user gõ
- file có sẵn (`@file:path`, PRD, note, README, issue, RFC...)
- bootstrap dự án mới
- bootstrap tính năng / module mới trong dự án hiện có

Output của `spec` là một **SPEC.md đủ rõ để planning không bị đoán mò**.

### `plan`
Đọc SPEC đã khóa, sau đó:
- chia phase
- khóa các decision xám cần thiết
- sinh task breakdown theo wave
- chỉ lập plan trong phạm vi SPEC

## 3. Skill Map

| Skill | Vai trò | Input | Output |
|------|---------|-------|--------|
| `spec` | chuẩn hóa idea thành spec | raw idea / file / bootstrap project / bootstrap feature | `.planning/SPEC.md` |
| `plan` | lập roadmap + plan từ spec | `.planning/SPEC.md` (+ optional codebase context) | `.planning/ROADMAP.md`, `phases/*-CONTEXT.md`, `phases/*-PLAN.md` |

## 4. Design Principles

1. **Spec-first** — không plan khi SPEC chưa đủ rõ
2. **Không nhồi hết vào 1 skill** — mỗi skill chỉ làm 1 job chính
3. **User control** — skill hỏi, gợi ý, khóa output; không auto-execute code
4. **Planning artifacts trước, implementation sau**
5. **Reuse existing skills** thay vì duplicate capability

## 5. File Structure

```text
.planning/
├── IDEA.md                     # optional raw idea / imported source
├── SPEC.md                     # locked requirements + boundaries + acceptance
├── ROADMAP.md                  # phases derived from spec
└── phases/
    └── {phase-slug}/
        ├── {phase}-CONTEXT.md  # locked implementation decisions
        └── {phase}-PLAN.md     # executable tasks for the phase
```

## 6. `spec` Skill Behavior

### 6.1 Inputs
`spec` should support these entry paths:

1. **Idea mode**
   - user describes an idea inline
2. **File mode**
   - user points to a file or set of files
3. **Project bootstrap mode**
   - user wants to start planning a new project
4. **Feature/module bootstrap mode**
   - user wants to plan a new capability inside an existing codebase

### 6.2 Modes

**fast**
- 1–2 interview rounds
- lighter ambiguity threshold
- compact SPEC

**deep**
- stronger Socratic interview
- optional research / codebase reading
- strict ambiguity gate before locking SPEC

### 6.3 Core flow

1. detect or create `.planning/IDEA.md` when useful
2. identify planning context:
   - new project?
   - existing project?
   - new module/feature?
   - file-derived spec?
3. run clarification loop
4. produce `.planning/SPEC.md`

### 6.4 Required sections in `SPEC.md`

```markdown
# SPEC: {title}

## Goal
## Users / Actors
## Requirements
## Boundaries
## Constraints
## Acceptance Criteria
## Open Questions
```

### 6.5 Ambiguity gate

`spec` should not finalize casually.

Suggested evaluation dimensions:
- goal clarity
- scope/boundary clarity
- constraints clarity
- acceptance criteria clarity

If still vague:
- in `fast`: allow proceed with explicit gaps
- in `deep`: keep asking until good enough or mark gaps loudly

## 7. `plan` Skill Behavior

### 7.1 Precondition
`plan` should require an existing `.planning/SPEC.md`.

If missing:
- tell user to run `spec` first
- or offer to generate spec from provided files/idea if the runtime supports chaining

### 7.2 Inputs
- required: `.planning/SPEC.md`
- optional: codebase/docs/repo context
- optional: phase target if user wants only one phase planned

### 7.3 Modes

**fast**
- derive compact roadmap
- minimal context locking
- shorter plan

**deep**
- derive roadmap carefully
- identify gray areas per phase
- lock implementation decisions into CONTEXT
- generate more rigorous wave-based task plan

### 7.4 Core flow

1. read SPEC
2. derive `ROADMAP.md`
3. for each selected phase:
   - identify gray areas
   - lock decisions into `{phase}-CONTEXT.md`
   - generate `{phase}-PLAN.md`

### 7.5 Required sections in `ROADMAP.md`

```markdown
# Roadmap: {title}

## Phase 1: {name}
Goal:
Deliverables:
Dependencies:

## Phase 2: ...
```

### 7.6 Required sections in `{phase}-PLAN.md`

```markdown
# Phase {N}: {name}

## Goal
## Inputs
## Wave 1
## Wave 2
## Verification
## Risks / Watch-fors
```

## 8. Relationship to Existing Skills

Do **not** rebuild these inside `spec` or `plan`:
- `investigator` → scout repo/files before locking details
- `strategist` → compare architectural options when needed
- `reviewer` → review implementation quality later
- `verifier` → run readiness / quality gates later
- `watzup` → summarize session changes later
- `git` → commit / push / PR later

This keeps the new system small and composable.

## 9. What This System Is NOT

- not an execution engine
- not an auto-coder
- not a giant GSD clone
- not a replacement for all existing workflow skills
- not forced multi-agent orchestration

## 10. Recommended Implementation Order

1. implement `spec` first
   - bootstrap modes
   - file-derived spec
   - feature/module spec flow
2. implement `plan` second
   - SPEC precondition
   - roadmap generation
   - phase context locking
   - phase plan generation
3. integrate lightly with existing skills via suggested next steps only

## 11. Open Questions

1. Should `spec` always create `IDEA.md`, or only when source is inline?
2. Should `plan` support single-file compact mode in `fast`?
3. How much repo investigation should `plan --deep` do before asking the user?
4. Do we want a shared `.planning/STATE.md` later, or keep artifacts separate?

## 12. Recommendation

Ship **2 skills only** for v1:
- `spec`
- `plan`

That is the smallest coherent system matching the direction:
- spec-driven
- planning-heavy
- project bootstrap capable
- reusable with the current skill set
