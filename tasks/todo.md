# TODO — `spec` + `plan` Skills

## Goal
Pivot PR #2 from a single all-in-one planning skill to a smaller spec-driven skill pair:
- `spec`
- `plan`

## Must keep
- `spec` can start from idea, file(s), project bootstrap, or new feature/module bootstrap
- `plan` must depend on `.planning/SPEC.md`
- existing skills remain in the system for execution/review/verify/git

## Implementation checklist

### Phase 1 — Finalize design
- [ ] review updated `plans/spec-skill/SPEC.md`
- [ ] confirm `.planning/` artifact structure
- [ ] confirm `fast` vs `deep` behavior for both skills

### Phase 2 — Implement `spec`
- [ ] create `skills/spec/SKILL.md`
- [ ] define invocation / argument shape
- [ ] define bootstrap modes
- [ ] define `SPEC.md` output template
- [ ] define ambiguity gate behavior

### Phase 3 — Implement `plan`
- [ ] create `skills/plan/SKILL.md`
- [ ] enforce SPEC precondition
- [ ] define `ROADMAP.md` generation
- [ ] define `{phase}-CONTEXT.md` generation
- [ ] define `{phase}-PLAN.md` generation

### Phase 4 — Integration polish
- [ ] define how `spec` suggests `plan`
- [ ] define how `plan` suggests `investigator` / `strategist` / `verifier` / `reviewer`
- [ ] keep the system composable, not monolithic

## Explicitly out of scope for now
- [ ] auto-execution
- [ ] giant all-in-one skill
- [ ] multi-agent orchestration
- [ ] CI / packaging work
