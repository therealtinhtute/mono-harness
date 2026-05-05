# TODO — Mini-GSD Planning System

## Goal
Build a **mini-GSD clone for Claude Code** focused on planning/spec work, using **2 core skills**:
- `spec`
- `plan`

## Product shape
- GSD-style `.planning/` folder
- artifact-first workflow
- `spec` locks WHAT
- `plan` derives HOW from locked spec
- existing skills stay available for repo scouting, review, verification, git, and wrap-up

## v1 must support

### `spec`
- [ ] input mode: `idea`
- [ ] input mode: `files`
- [ ] scenario: project bootstrap
- [ ] scenario: feature bootstrap
- [ ] scenario: module bootstrap
- [ ] write `.planning/IDEA.md`
- [ ] write `.planning/SPEC.md`
- [ ] ambiguity / clarity gate before lock

### `plan`
- [x] require `.planning/SPEC.md`
- [x] read spec and derive `.planning/ROADMAP.md`
- [x] create per-phase `{phase}-CONTEXT.md`
- [x] create per-phase `{phase}-PLAN.md`
- [x] keep tasks wave-grouped and executable

## v1 artifact chain
- [ ] IDEA.md
- [ ] SPEC.md
- [x] ROADMAP.md
- [x] phases/{slug}/{phase}-CONTEXT.md
- [x] phases/{slug}/{phase}-PLAN.md

## integration expectations
- [ ] suggest `plan` after `spec`
- [ ] suggest `investigator` when codebase understanding is weak
- [ ] suggest `reviewer` / `verifier` after implementation later
- [ ] suggest `git` / `watzup` / `handoff` when relevant

## explicitly not in v1
- [ ] full execution engine
- [ ] giant all-in-one skill
- [ ] forced multi-agent orchestration
- [ ] full GSD command ecosystem clone

## next implementation order
1. [ ] implement `spec`
2. [x] implement `plan`
3. [ ] wire suggestions into existing skill ecosystem
