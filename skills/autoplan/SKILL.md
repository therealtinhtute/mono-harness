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
- mainly needs supervised execution tracking; use `goal`
- is blocked by a conceptual knot and needs reframing first; use `problem-solving`

## Do not use when

- the task is already concrete enough to execute now
- the user only wants ideation without narrowing
- the task is mainly blocked by a hard conceptual knot; use `problem-solving` first
- the user already has a validated plan and mainly needs long supervised execution; use `goal`

## Auto-trigger priority

If both `autoplan` and another skill could fit, prefer:
- `autoplan` first when the missing piece is **structure**
- `goal` first when the missing piece is **supervised execution**
- `problem-solving` first when the missing piece is **clear thinking / abstraction**

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
6. Normalize the result into one cleaned spec.
7. Convert the spec into phases, goals, and concrete tasks.
8. Recommend the next action: execute now, hand off to `goal`, or briefly use `problem-solving` then return.

## Required output

```text
Goal
Context
Assumptions
Open questions
Scope
Non-goals
Constraints
Spec
Phases
Tasks
Risks
Next recommended action
```

## Planning rules

- Converge quickly; do not linger in brainstorming mode.
- Prefer one cleaned source of truth over scattered notes.
- Keep clarification questions grouped and minimal.
- Mark sourced facts separately from inference.
- If a decision is irreversible or product-defining, ask instead of inventing it.
- Tasks must be action-oriented, testable, and grouped by phase.
- If the user asked for "lite", keep research and decomposition lighter but keep the same output shape.

## Handoff rules

- Use `goal` when execution is large, long-running, or needs pass/fail supervision.
- Use `problem-solving` when the plan is blocked by a bad abstraction or messy framing.
- Return to autoplan after any temporary detour so the final deliverable is still a clean plan.

## Done criteria

The skill is complete only when there is:
- one clear normalized spec
- explicit phases
- concrete tasks under those phases
- a recommended next move
- no major unresolved gap hidden inside assumptions

## Read next when needed

- `references/autoplan-patterns.md` for input modes, question strategy, and task breakdown patterns
- `references/examples-core.md` for idea-first and spec-first example input/output patterns
- `references/examples-lite.md` for a lighter autoplan-lite example
- `../goal/SKILL.md` for supervised execution after planning
- `../problem-solving/SKILL.md` for conceptual reframing before returning to the plan
