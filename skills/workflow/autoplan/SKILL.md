---
name: autoplan
description: Start from a rough idea or markdown spec, clarify gaps, do light research, and produce an execution-ready plan. Trigger on autoplan, autoplan-lite, start from spec, or idea to plan.
---

# Autoplan

Use this skill when the user wants one wrapper flow from **idea/spec input** to **clean spec + phases + tasks + next action**.

## Use when

- the input is a rough idea, notes dump, chat transcript, or partial markdown spec
- the user wants `autoplan`, `autoplan-lite`, `workflow-start`, or `idea to plan`
- the work needs quick clarification before execution
- the user wants convergence, not open-ended brainstorming

## Trigger phrases

Strong triggers:
- `autoplan`, `autoplan-lite`, `workflow-start`
- `start from spec`, `idea to plan`, `turn this into a plan`
- `turn this spec into tasks`, `break this into phases`
- `shape this idea`, `make this execution-ready`
- `help me scope this`, `convert this into a roadmap`

Auto-trigger when the user:
- provides an idea or markdown spec and asks for a plan, roadmap, task breakdown, phases, or next steps
- asks to clarify scope, fill gaps, or make a request execution-ready before doing the work
- wants lightweight research plus a concrete execution structure

Do not auto-trigger when the user:
- is ready to execute a specific task immediately
- only wants free-form ideation with no convergence yet
- mainly needs supervised execution tracking; use `cook`
- is blocked by a conceptual knot and needs reframing first; use `brainstorm`

## Do not use when

- the task is already concrete enough to execute now
- the user only wants ideation without narrowing
- the task is mainly blocked by a hard conceptual knot; use `brainstorm` first
- the user already has a validated plan and mainly needs long supervised execution; use `cook`

## Auto-trigger priority

If both `autoplan` and another skill could fit, prefer:
- `autoplan` first when the missing piece is **structure**
- `cook` first when the missing piece is **supervised execution**
- `brainstorm` first when the missing piece is **clear thinking / abstraction**

## Fast dispatch

- **idea-first** → shape the problem, constraints, and success target before planning
- **spec-first** → audit the spec, fill gaps, then convert it into execution phases and tasks
- **unclear input** → ask the minimum questions needed to decide between the two

## Core workflow

1. Rewrite the user's goal in one tight sentence.
2. Classify the input as **idea-first** or **spec-first**.
3. Inspect first: identify missing scope, assumptions, dependencies, and decision points.
4. Ask only the smallest set of questions that materially changes the plan.
5. Do light live research when facts may have changed: APIs, competitors, libraries, constraints, costs, or operational details.
6. Normalize the result into `.kit/planning/SPEC.md` using the structure from `../brainstorm/references/spec-template.md`. Set `Status: locked` and `Downstream: plan full`. Map autoplan fields to SPEC fields: Goal→Goal, Context→Source Inputs, Assumptions→Dependencies/Assumptions, Open questions→Open Questions, Scope→In Scope, Non-goals→Out of Scope, Constraints→Constraints, Spec→Requirements, Risks→Risk Flags.
7. Convert the spec into phases, goals, and concrete tasks.
8. Recommend the next action: execute now, hand off to `cook`, or briefly use `brainstorm` then return.

## Required output

**Primary artifact:** `.kit/planning/SPEC.md` — locked spec in brainstorm's format (see `../brainstorm/references/spec-template.md`).

**Console summary** (chat response after writing SPEC.md):
```text
Goal (1 sentence)
Phases (preview — real breakdown comes from /plan)
Tasks (preview — grouped by phase)
Risks (if any)
Next: /plan → /cook full
```

The console phases/tasks are a preview for the user. The executable phase breakdown is `/plan`'s job.

## Planning rules

- Converge quickly; do not linger in brainstorming mode.
- Prefer one cleaned source of truth over scattered notes.
- Keep clarification questions grouped and minimal.
- Mark sourced facts separately from inference.
- If a decision is irreversible or product-defining, ask instead of inventing it.
- Tasks must be action-oriented, testable, and grouped by phase.
- If the user asked for "lite", keep research and decomposition lighter but keep the same output shape.

## Handoff rules

- After writing SPEC.md, suggest `/plan` to generate phase artifacts, then `/cook full` to execute. This is the default path.
- Suggest `/cook simple` only when the scope is tiny (≤5 files, ≤100 lines) and planning overhead is unnecessary.
- Use `brainstorm` when the plan is blocked by a bad abstraction or messy framing.
- Return to autoplan after any temporary detour so the final deliverable is still a clean plan.

## Done criteria

The skill is complete only when:
- `.kit/planning/SPEC.md` exists with `Status: locked`
- SPEC.md has numbered requirements, explicit boundaries, and acceptance criteria
- console summary shows phase/task preview and next action (`/plan` or `/cook simple`)
- no major unresolved gap hidden inside assumptions

## Anti-Patterns
- Skipping SPEC.md fields (Acceptance Criteria, Key Decisions) because "it's a quick plan" — partial spec breaks `/plan` downstream
- Lingering in clarification when user wants convergence — the skill's purpose is structure, not open-ended exploration
- Not verifying SPEC.md was written before reporting done — console output without the artifact is not done

## Read next when needed

- `references/autoplan-patterns.md` for input modes, question strategy, and task breakdown patterns
- `references/examples-core.md` for idea-first and spec-first example input/output patterns
- `references/examples-lite.md` for a lighter autoplan-lite example
- `../cook/SKILL.md` for supervised execution after planning
- `../brainstorm/SKILL.md` for conceptual reframing before returning to the plan
