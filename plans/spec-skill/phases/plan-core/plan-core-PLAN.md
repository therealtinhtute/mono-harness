# Plan: plan core

## Goal
Implement `skills/plan/SKILL.md` so it can turn a locked spec into roadmap/context/phase plans that are immediately useful for implementation.

## Inputs
- `plans/spec-skill/SPEC.md`
- `plans/spec-skill/phases/plan-core/plan-core-CONTEXT.md`
- `.planning/SPEC.md` from real usage

## Wave 1
1. Define `skills/plan/SKILL.md` frontmatter and precondition behavior.
   - enforce `.planning/SPEC.md` existence
   - define fail-fast message when spec is missing
   - verify: planner never invents context without spec

2. Define `ROADMAP.md` generation rules.
   - split by coherent phases
   - include goal / deliverables / dependencies / risks
   - verify: phases are understandable and orderable

3. Define `{phase}-CONTEXT.md` contract.
   - locked decisions
   - canonical refs
   - rejected options
   - deferred ideas
   - verify: context reduces implementation ambiguity

## Wave 2
4. Define `{phase}-PLAN.md` task contract.
   - wave-based tasks
   - steps
   - expected outputs
   - verification method
   - verify: a coding agent can execute from the plan without further macro-planning

5. Add examples.
   - plan a new project from spec
   - plan a feature from spec
   - verify: examples cover at least the main workflows

## Verification
- `plan` refuses to proceed without spec
- generated artifacts stay inside `.planning/`
- output is phase-based, not one giant vague checklist

## Risks / Watch-fors
- letting `plan` drift into execution instructions too early
- generating plans that ignore spec boundaries
