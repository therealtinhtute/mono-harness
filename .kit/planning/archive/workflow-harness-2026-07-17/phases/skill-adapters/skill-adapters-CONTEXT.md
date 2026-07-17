# Context: skill-adapters

Phase: skill-adapters
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: high
Expected Proof: integration (sample chain run + zharness validate) + skill lint

## Goal
brainstorm, to-plan, work rewritten CLI-first: explicit `zharness` calls inline in each flow, mandatory version gate, zero `workflow-state.yml` writes.

## Scope Boundary
### Allowed Surfaces
- `skills/workflow/brainstorm/SKILL.md` + its `references/`
- `skills/workflow/to-plan/SKILL.md` + its `references/` (except deleting workflow-state-template.yml — Phase 8)
- `skills/workflow/work/SKILL.md` + its `references/`

### Forbidden Surfaces
- check, watzup, handoff, git, interview skills (Phases 6–7)
- `cli/**` (flag changes go back through Phase 4)
- Skill UX: mode structure, order, and intent of each skill must remain recognizable

## Spec Hooks
- R17 (CLI-first inline, mandatory gate, no fallback), R18 (adapter behaviors), R14 (no yml writes)
- Constraint: skills.sh self-containment — each SKILL.md carries its own gate block

## Locked Decisions
- Standard gate block (identical text in all three): check `zharness --version` ≥ MIN_ZHARNESS_VERSION (constant documented in skills/workflow/README.md); on failure print `bash scripts/install-zharness.sh` guidance and STOP
- brainstorm: `zharness intake` fires at SPEC lock (step where classification happens today); intake ULID written into SPEC frontmatter `intake_id`
- to-plan: `zharness init` if no db; one `zharness story` per phase (slug = phase slug); no workflow-state.yml initialization — replace that instruction with state-via-CLI
- work: `zharness trace add` at wave completion; RUN artifact frontmatter carries trace IDs
- Command snippets appear inline at the exact step they run, with `--json` and expected output noted
- `scripts/validate-skill.sh` verified functional against a real SKILL.md (`skills/workflow/brainstorm/SKILL.md` → PASS, full structure/frontmatter/token checks ran) — safe to use as T1's lint step without a pre-check

## Assumptions
- Phase 4 binary installed locally for the sample run

## Canonical Refs
- `cli/docs/CONTRACT.md` (only source for flags)
- `skills/workflow/README.md` (lifecycle + mapping, Phase 1)
- Phase 2 templates (frontmatter keys)

## Rejected Options
- references/cli-adapter.md pattern — SPEC decision: explicit inline CLI-first
- Soft-fail when binary missing — SPEC decision: mandatory, stop with install guidance

## Deferred Ideas
- Auto-install prompt inside skills — install stays an explicit user action

## Escalate If
- A skill step has no matching CLI action in CONTRACT.md → to-plan phase harness-contracts
- Inline rewrites materially change a skill's UX flow → brainstorm refine
