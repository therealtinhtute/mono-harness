# Brainstorm Examples

Worked examples for each mode. The 4 modes are documented in `mode-detection.md`.

---

## Explore mode — output layout

Save to `.kit/reports/brainstorm/{YYYYMMDD}-{slug}.md` with frontmatter:

```yaml
---
title: Brainstorm - {slug}
description: {one-line summary}
status: draft | active | completed
created: YYYY-MM-DD
tags: [brainstorm, {slug}]
---
```

Body order: recommendation first, problem statement, evaluated approaches with pros/cons, rationale, risks, next steps.

---

## Example 1 — Explore: Architecture decision (Monolith vs Microservices)

**Input**: "Should we build our new API as microservices or a monolith? Expected 10K users initially."

**Detected mode**: `explore` (vague trade-off question, no lock intent)

**Confirm**: ask via `AskUserQuestion` whether the user wants a recommendation report or wants to lock this into a project SPEC.

**Output** (excerpt of `.kit/reports/brainstorm/20260507-api-architecture.md`):

```markdown
# Brainstorm: API Architecture Decision

## Recommendation
Modular monolith. YAGNI applies — microservices complexity is unjustified at 10K users.

## Problem Statement
New API serving mobile and web clients, 10K initial users, single small team.

## Evaluated Approaches

### Option 1: Microservices
- Pros: independent scaling, technology flexibility, team autonomy
- Cons: massive operational overhead, network latency, distributed transaction
  complexity, requires DevOps expertise — overkill for 10K users

### Option 2: Modular Monolith ✅
- Pros: simple deployment, easy debugging, no network overhead, shared
  database transactions, can extract services later
- Cons: scales as one unit, requires discipline to keep modules loose

## Rationale
At 10K users with one team, operational simplicity dominates. Extract services
only when real evidence (100K+ users, specific bottlenecks) appears.

## Risks
- "Big ball of mud" — mitigate via lint-enforced module boundaries
- Premature optimization to microservices — wait for measured bottlenecks

## Next Steps
- Lock module boundaries (auth, users, billing) in a SPEC if proceeding
- Run `plan` after locking to derive phases
```

---

## Example 2 — Lock-from-idea: Project bootstrap

**Input**: "I want to build an AI inbox for small teams"

**Detected mode**: `lock-from-idea` (raw idea, no files)

**Workflow**:
1. Write `.planning/IDEA.md` capturing the idea verbatim
2. Clarify via `AskUserQuestion`: target team size, primary actor, in-scope channels (email/Slack/etc.), platform
3. Apply YAGNI to scope — recommend the smallest useful slice
4. Write `.planning/SPEC.md` using `spec-template.md`

**Output**: two files. `IDEA.md` preserves the original phrasing for future reference; `SPEC.md` contains numbered requirements, in/out-of-scope, constraints, acceptance criteria.

---

## Example 3 — Lock-from-files: Extract from PRD

**Input**: "lock @file:docs/auth-rfc.md @file:docs/notes.md"

**Detected mode**: `lock-from-files` (explicit file refs)

**Workflow**:
1. Read both files
2. Surface the proposed core change in one paragraph
3. Identify gaps via `clarification-rubric.md`: missing acceptance criteria, unclear actor boundaries, undefined constraints
4. Clarify gaps via `AskUserQuestion`
5. Write `.planning/SPEC.md` referencing the source files in `Dependencies / Assumptions`

**Output**: `.planning/SPEC.md` with traceability back to the source markdown. `IDEA.md` optional — only if the source files are loose notes rather than a structured RFC.

---

## Notes on mode upgrades

- If during `explore` the user says "lock the recommended option," upgrade to `lock-from-idea` using the recommendation as the seed idea. Confirm scope before writing.
- If during `lock-*` the user wants to revisit alternatives, drop into `explore` mode briefly, then resume locking with the chosen path.
- Refining an existing `.planning/SPEC.md` (mode `refine`) follows the lock workflow but only edits the affected sections; show the user a diff summary before writing.
