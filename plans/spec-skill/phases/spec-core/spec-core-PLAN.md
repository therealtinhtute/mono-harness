# Plan: spec core

## Goal
Implement `skills/spec/SKILL.md` so it can reliably generate `.planning/IDEA.md` and `.planning/SPEC.md` from idea/files inputs.

## Inputs
- `plans/spec-skill/SPEC.md`
- `plans/spec-skill/phases/spec-core/spec-core-CONTEXT.md`
- existing repo skill conventions

## Wave 1
1. Define `skills/spec/SKILL.md` frontmatter and role.
   - include clear description
   - define accepted source modes: `idea`, `files`
   - define artifact targets
   - verify: skill intent is obvious from frontmatter alone

2. Define `spec` workflow.
   - detect source mode
   - classify scenario: project / feature / module / refine-existing
   - define question flow
   - verify: workflow can be followed step-by-step without hidden assumptions

3. Define `SPEC.md` output contract.
   - title
   - source mode
   - scenario
   - requirements / boundaries / constraints / acceptance / open questions
   - verify: output is specific enough for `plan`

## Wave 2
4. Define ambiguity gate behavior.
   - what “good enough to lock” means
   - what happens when clarity is still low
   - verify: `spec` does not finalize vague planning artifacts silently

5. Add examples for each scenario.
   - new project
   - new feature
   - new module
   - file-derived refinement
   - verify: each scenario has at least one concrete invocation example

## Verification
- `spec` can be invoked from either idea prompt or files input
- output always lands in `.planning/`
- output shape is stable enough that `plan` can consume it

## Risks / Watch-fors
- over-designing interview flow before first usable artifact exists
- making `spec` secretly do planning work that belongs to `plan`
