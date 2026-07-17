# Context: validation-gate

Phase: validation-gate
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: medium
Expected Proof: unit + integration (determinism fixtures)

## Goal
Research commands (score-trace, score-context, audit, propose) land in `zharness`; `check` gates deterministically on an evidence matrix and records verdicts via CLI.

## Scope Boundary
### Allowed Surfaces
- `cli/internal/**`, `cli/testdata/**` (research commands)
- `skills/workflow/check/SKILL.md`
- `skills/workflow/check/references/gate-checklist.md`, `artifact-alignment.md`, `report-template.md` (matrix alignment only)

### Forbidden Surfaces
- brainstorm/to-plan/work skills (done in Phase 5)
- watzup/handoff (Phase 7)
- Adopting `propose`/`score-context` into any skill (reserved — documented only)

## Spec Hooks
- R6 (research commands), R19 (deterministic verdicts, matrix), R24 sequencing note (pilot doesn't wait for this phase's polish)
- Open Question 2: trace tiers default to upstream definitions

## Locked Decisions
- Validation matrix axes: intake lane (tiny/normal/high-risk) × proof class (unit, integration, e2e, manual-check, command-output); every cell `required|optional|n-a`
- check gate flow: run `zharness audit --json` + `zharness score-trace --json` → evaluate matrix against CHECK report proof links → `zharness check record` with structured verdict
- Missing required proof ⇒ FAIL naming the missing evidence — no judgment override inside the skill; overrides are a human decision recorded via `intervention`
- `audit` report sections: pointer drift, contract violations, unlinked proofs, entropy score (upstream-style)

## Assumptions
- Upstream trace quality tiers (TRACE_SPEC.md) port unchanged; deviations noted in CONTRACT.md
- Phase 5 RUN artifacts already carry trace IDs (input to score-trace)

## Canonical Refs
- `cli/docs/CONTRACT.md` (research command schemas from Phase 2)
- `~/Lab/harness-experimental/docs/TRACE_SPEC.md`, `docs/TEST_MATRIX.md`
- `skills/workflow/check/references/review-dimensions.md` (existing dimensions to keep)

## Rejected Options
- Verdict from LLM judgment with matrix as "guidance" — SPEC R19 requires deterministic outcomes
- Building a separate maintenance skill to own audit — deferred in SPEC

## Deferred Ideas
- `propose`-driven improvement loop; score-context adoption
- Custom trace tiers tuned to this repo (post-pilot)

## Escalate If
- Matrix cannot express an existing check dimension without judgment → brainstorm refine
- audit needs DB fields SCHEMA.md lacks → to-plan phase harness-contracts
