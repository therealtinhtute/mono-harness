# Context: harness-concept

Phase: harness-concept
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: low
Expected Proof: docs inspection

## Goal
Lock the workflow-harness mental model (lifecycle + 4-layer architecture) and the gap inventory in docs, including the story↔phase mapping decision that unblocks contract work.

## Scope Boundary
### Allowed Surfaces
- `skills/workflow/README.md` (new)
- `docs/workflow-harness/gap-matrix.md` (new, incl. directory creation)
- Root `README.md` (link/section updates only)

### Forbidden Surfaces
- Any `SKILL.md` or `references/` file
- `cli/` (does not exist yet — do not scaffold)
- `CLAUDE.md`

## Spec Hooks
- R1 (repo additions named), R22 (concept doc content), SPEC Open Question 1 (story↔phase mapping — resolved here)
- Constraint: preserve workflow chain UX; concept doc must state this explicitly

## Locked Decisions
- Lifecycle wording: Intent → Intake → Story/Plan → Work Trace → Validation Proof → Handoff/Resume
- 4 layers named exactly: harness / workflows / skills / cli
- Mapping table covers all 8 workflow skills, with `git` and `interview` marked minimal-integration
- Gap matrix groups fixed: intake, state, trace, validation, resume, CLI surface; row fields: gap, owner-skill, artifact, risk, phase

## Assumptions
- `~/Lab/harness-experimental/docs/HARNESS.md` remains the reference for lifecycle terminology
- Story↔phase recommendation: one harness story per to-plan phase, story slug = phase slug (confirm or overturn with rationale in the matrix doc)

## Canonical Refs
- `.kit/planning/SPEC.md`
- `.kit/planning/ROADMAP.md`
- `~/Lab/harness-experimental/docs/HARNESS.md`, `docs/FEATURE_INTAKE.md`
- Scratchpad 13-issue set (issues #1, #2) if still available

## Rejected Options
- Putting the concept doc in `docs/` instead of `skills/workflow/README.md` — skill users land in the skill tree first; SPEC R22 fixes the location
- Skipping the gap matrix ("we already know the gaps") — the matrix carries the story↔phase decision Phase 2 depends on

## Deferred Ideas
- Harness maturity/audit narrative (upstream HARNESS_MATURITY.md) — not needed for this initiative

## Escalate If
- Story↔phase mapping turns out to need schema changes beyond a slug convention → `to-plan phase harness-contracts` before writing SCHEMA.md
- Concept doc contradicts a SPEC boundary → `brainstorm refine`
