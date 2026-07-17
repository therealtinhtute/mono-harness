# Plan: harness-concept

Phase: harness-concept
Status: ready
Wave Count: 2
Execution Owner: work
Updated At: 2026-07-16

## Goal
Concept doc + gap matrix locked; root README linked; story↔phase mapping decided.

## Inputs
- `.kit/planning/SPEC.md` (R1, R22, Open Question 1)
- `~/Lab/harness-experimental/docs/HARNESS.md`, `docs/FEATURE_INTAKE.md`
- Current `skills/workflow/*/SKILL.md` (read-only, for the mapping table)

## Wave 1
### T1 — Write skills/workflow/README.md concept doc
- type: docs
- inputs:
  - SPEC Goal + R22; upstream HARNESS.md lifecycle wording
- touches:
  - `skills/workflow/README.md` (new)
- avoid:
  - implementation detail (schemas, flags) — that is Phase 2
- steps:
  1. Write lifecycle section: Intent→Intake→Story/Plan→Trace→Proof→Handoff/Resume, one paragraph per stage naming the owning skill
  2. Write 4-layer model section (harness/workflows/skills/cli) with the preserved-UX statement
  3. Write mapping table: 8 skills × (harness artifact, zharness command group, entity); mark git/interview minimal
  4. Add in-scope/out-of-scope section mirroring SPEC boundaries
- expected outputs:
  - `skills/workflow/README.md` with lifecycle, 4-layer model, mapping table, scope section
- verification:
  - `grep -c 'watzup\|brainstorm\|to-plan\|work\|interview\|check\|git\|handoff' skills/workflow/README.md` — all 8 present in mapping table
  - `grep -iE 'TBD|TODO' skills/workflow/README.md` returns nothing
- stop if:
  - mapping requires a command not in SPEC R6's surface
- escalate to:
  - brainstorm refine

## Wave 2
### T2 — Write docs/workflow-harness/gap-matrix.md
- type: docs
- inputs:
  - T1 output; current workflow skill references (read-only)
- touches:
  - `docs/workflow-harness/gap-matrix.md` (new)
- avoid:
  - fixing any gap; editing skills
- steps:
  1. Build matrix rows for the 6 groups (intake, state, trace, validation, resume, CLI surface): gap, owner-skill, artifact, risk, phase
  2. Decide and document story↔phase mapping (default: 1 story per phase, story slug = phase slug) with rationale and rejected alternative
- expected outputs:
  - Gap matrix with zero empty cells; story↔phase decision section
- verification:
  - Inspect: every row has all 5 fields; `grep -i 'story' docs/workflow-harness/gap-matrix.md` shows the mapping decision
- stop if:
  - a gap requires new product scope beyond SPEC
- escalate to:
  - brainstorm refine

### T3 — Link from root README
- type: docs
- inputs:
  - T1 file path
- touches:
  - `README.md` (root — add workflow-harness section/link only)
- avoid:
  - rewriting install/quickstart (that is Phase 8)
- steps:
  1. Add a short "Workflow Harness" subsection linking `skills/workflow/README.md` and `docs/workflow-harness/`
- expected outputs:
  - Root README references both docs
- verification:
  - `grep -n 'workflow-harness\|workflow/README' README.md`
- stop if:
  - README restructure feels needed — defer to Phase 8
- escalate to:
  - to-plan phase pilot-migration

## Risks / Watch-fors
- Concept doc scope creep into contract detail — keep flags/schemas out
- Story↔phase left undecided blocks Phase 2's SCHEMA.md
