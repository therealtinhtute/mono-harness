# Roadmap: Mini-GSD Planning System

## Phase 1: `spec` core
**Goal:** Ship a usable `spec` skill that turns idea/files into a locked `.planning/SPEC.md`.

**Deliverables:**
- `skills/spec/SKILL.md`
- `.planning/IDEA.md` bootstrap behavior
- `.planning/SPEC.md` template
- source mode handling: `idea` / `files`
- scenario handling: project / feature / module bootstrap
- ambiguity-gate behavior

**Dependencies:** none

**Why this phase exists:** `plan` should not be implemented on shaky spec foundations.

## Phase 2: `plan` core
**Goal:** Ship a `plan` skill that reads locked spec and generates roadmap/context/phase plans.

**Deliverables:**
- `skills/plan/SKILL.md`
- `ROADMAP.md` generation
- per-phase `CONTEXT.md`
- per-phase `PLAN.md`
- SPEC precondition enforcement
- wave-based task breakdown

**Dependencies:** Phase 1

**Why this phase exists:** makes the system actually usable for implementation planning.

## Phase 3: integration + polish
**Goal:** Make the skill pair feel coherent inside the existing repo ecosystem.

**Deliverables:**
- handoff guidance between `spec` and `plan`
- suggestions to use `investigator`, `reviewer`, `verifier`, `git`, `watzup`
- wording cleanup
- examples and artifact consistency review

**Dependencies:** Phase 1, Phase 2

**Why this phase exists:** reduce ambiguity and make the system pleasant to adopt.
