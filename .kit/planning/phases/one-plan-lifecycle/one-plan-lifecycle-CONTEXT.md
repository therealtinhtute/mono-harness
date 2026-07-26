# Context: one plan lifecycle

Phase: one-plan-lifecycle
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: high
Expected Proof: unit, integration, lifecycle e2e, deletion coverage, manual-check

## Goal
Make one evolving plan the only durable initiative markdown while typed DB rows retain lifecycle position, proof links, and recovery.

## Scope Boundary
### Allowed Surfaces
- lifecycle schema fields and command/application logic
- scaffold/validate/resume/query behavior
- embedded plan template and six stage playbooks
- root plans/decisions directories
- legacy history summary and approved legacy artifact removal

### Forbidden Surfaces
- typed table collapse
- bounded-work DB writes
- changeset compaction
- extra legacy archive directory
- CI workflows

## Spec Hooks
- Requirements 5, 6, 7, 10, and 13.

## Locked Decisions
- One plan file carries outcome through handoff.
- Full lifecycle retains intake/story/run/trace/check/handoff/intervention rows.
- Bounded work creates no lifecycle row or markdown artifact.
- Old detailed files are covered by one completed history summary, then removed; Git history is the raw archive.

## Assumptions
- Existing rows remain queryable after artifact paths become nullable/deprecated.
- Plans use their own ULID and carry the intake ID.

## Canonical Refs
- `cli/docs/{SCHEMA,STATE,CONTRACT}.md`
- `cli/internal/application/{run_create,check_record,handoff,resume,validate}.go`
- `cli/docs/embedded/playbooks/*.md`
- current `.kit/planning`, `.kit/plans`, `.kit/runs`, `.kit/reports`, `.kit/HANDOFF.md`

## Rejected Options
- Keep SPEC + plan: still duplicates intent.
- DB-only durable work: weakens repository legibility.
- Archive all old files under docs: preserves the clutter being removed.

## Deferred Ideas
- Multi-plan orchestration; independent workstreams use independent initiatives.

## Escalate If
- A legacy file cannot be represented in the completed history summary without losing a still-current requirement.
- Removing artifact paths makes an existing query impossible without a replacement DB field.
