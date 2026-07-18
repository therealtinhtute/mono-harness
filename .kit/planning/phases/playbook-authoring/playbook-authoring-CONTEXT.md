# Context: playbook-authoring

Phase: playbook-authoring
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: low
Expected Proof: inspection (command-accuracy + coverage review), unit deferred to cli-embed-scaffold

## Goal
Produce the canonical doc set that later phases embed: 6 stage playbooks, `AGENTS.md` shim, `CONTEXT_RULES.md`, `AUTHORITY.md` — each self-sufficient for a competent non-Claude agent.

## Scope Boundary
### Allowed Surfaces
- `cli/docs/embedded/**` (new directory — playbooks + shim + rules docs)
- Read-only: `skills/workflow/{brainstorm,to-plan,work,check,handoff,watzup}/**`, `cli/docs/STATE.md`, `cli/internal/interfaces/*.go` (to verify command surfaces)

### Forbidden Surfaces
- Any Go source change (`cli/internal/**`, `cli/cmd/**`)
- Any SKILL.md edit (that is thin-triggers' job)
- `skills/workflow/git/**`, `skills/workflow/interview/**`

## Spec Hooks
- R1 (doc set contents), R4 (self-sufficiency), R5 (CONTEXT_RULES mapping), R6 (authority model)
- Constraint: plain markdown, no agent-specific syntax

## Locked Decisions
- Source layout: `cli/docs/embedded/playbooks/{stage}.md` + `cli/docs/embedded/{AGENTS,CONTEXT_RULES,AUTHORITY}.md` — flat, predictable paths the embed package mirrors verbatim into `.kit/docs/`
- Playbook structure (uniform): Purpose → Preconditions → Steps (exact `zharness` commands) → Artifacts (paths + templates inline) → Exit / handoff conditions
- Existing `references/` templates (spec-template, run-artifact-template, report-template, handoff-template, output-contract) are absorbed *into* their stage playbook, not kept as separate files — one doc per stage
- `AGENTS.md` is a stable shim (short, rarely changes) linking to the playbooks; evolving content lives in the playbooks, mirroring upstream's shim pattern
- Every `zharness` command quoted in a playbook must be verified against `cli/internal/interfaces/*.go` in the same wave it is written

## Assumptions
- The 6 SKILL.md files + references are the complete de facto spec of each stage (SPEC assumption); anything ambiguous found during distillation escalates rather than gets invented
- Vietnamese-flavored output contracts (watzup's `Mức độ` table etc.) stay as-is — they are part of the stage contract, not Claude-specific syntax

## Canonical Refs
- `.kit/planning/SPEC.md`, `.kit/planning/ROADMAP.md`
- `skills/workflow/README.md` (4-layer model, entity mapping)
- `hoangnb24/repository-harness` docs (pattern source: AGENTS shim, CONTEXT_RULES, request-class authority) — pattern reference only, no text copying

## Rejected Options
- Keeping references/ as separate embedded files per stage — multiplies files and read hops; one self-sufficient playbook per stage is the R4 contract
- Authoring docs directly inside the Go package as string literals — unreviewable diffs; separate markdown files with `go:embed` keeps them readable

## Deferred Ideas
- Doc localization; playbooks for git/interview (out of scope per R8)

## Escalate If
- A SKILL.md contract cannot be expressed agent-agnostically (e.g. depends on AskUserQuestion tool semantics) → brainstorm refine: the authority doc may need an "interaction capability" note
- Playbook content for a stage exceeds what one coherent doc can carry → to-plan phase (split decision)
