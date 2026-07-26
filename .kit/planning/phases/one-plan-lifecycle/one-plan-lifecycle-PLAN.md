# Plan: one plan lifecycle

Phase: one-plan-lifecycle
Status: ready
Wave Count: 4
Execution Owner: work
Updated At: 2026-07-26

## Goal
Replace parallel lifecycle markdown with one active plan and make DB state complete without report files.

## Inputs
- completed root-layout phase
- one-plan-lifecycle-CONTEXT.md
- legacy artifact inventory and DB rows

## Wave 1
### T1 — Add one-plan schema and scaffold contract
- type: migration + implementation + test
- touches:
  - migrations/changeset allowlists
  - intake/run schema/application code
  - scaffold command/templates/tests
  - `cli/docs/{SCHEMA,CONTRACT}.md`
- avoid:
  - deleting legacy files before migration proof
- steps:
  1. Add intake plan path and deprecate nullable artifact paths additively.
  2. Add `scaffold plan` with the approved nine-section template.
  3. Ensure plan IDs/intake IDs are stable and validateable.
  4. Prove existing changesets still replay.
- expected outputs:
  - schema and template support one evolving plan
- verification:
  - `cd cli && go test ./... -run 'Plan|Scaffold|Replay|Migration'`
- stop if:
  - backward compatibility needs a destructive migration
- escalate to:
  - check

## Wave 2
### T2 — Remove markdown dependencies from lifecycle commands
- type: implementation + test
- touches:
  - run/check/handoff/story application and interfaces
  - resume/query/validate/audit
  - corresponding tests
- avoid:
  - bounded-work entity creation
- steps:
  1. Allow durable run creation without run artifact path.
  2. Record checks/handoffs entirely in DB with plan/story anchors.
  3. Implement automatic story status transitions to `done`.
  4. Make resume/validate/query operate without report files.
  5. Add a bounded route test proving no DB/checksum/changeset change.
- expected outputs:
  - complete DB lifecycle independent of report files
- verification:
  - `cd cli && go test ./... -run 'Run|Check|Handoff|Status|Resume|Validate|Bounded'`
- stop if:
  - a read-only command mutates state
- escalate to:
  - check

## Wave 3
### T3 — Rewrite stage playbooks around one plan
- type: docs + integration test
- touches:
  - `cli/docs/embedded/playbooks/{brainstorm,to-plan,work,check,handoff,watzup}.md`
  - plan template and projection tests
- avoid:
  - final global mental-model rewrite reserved for Phase 4
- steps:
  1. Make brainstorm create the active plan outcome/requirements.
  2. Make to-plan update the same file with approach/phases/verification.
  3. Make work/check/handoff update progress/validation/current-state sections.
  4. Remove RUN/CHECK/HANDOFF file creation and simple-mode report output.
  5. Add a fixture proving one markdown file across a full lifecycle.
- expected outputs:
  - new workflow emits one durable plan only
- verification:
  - `cd cli && go test ./... -run 'Lifecycle|Projection|OnePlan'`
- stop if:
  - any stage requires a second durable markdown source
- escalate to:
  - to-plan phase one-plan-lifecycle

## Wave 4
### T4 — Consolidate and remove legacy lifecycle files
- type: migration + docs
- touches:
  - `docs/plans/completed/workflow-harness-history-2026-07.md`
  - approved `.kit` legacy artifact trees
- avoid:
  - `.kit/changesets`
  - `docs/workflow-harness/`
- steps:
  1. Inventory every tracked legacy lifecycle file.
  2. Write one completed history summary covering initiatives, decisions, gates, and provenance links.
  3. Verify DB replay/resume/validate without report-path dependencies.
  4. Use `trash` to remove `.kit/planning`, `.kit/plans`, `.kit/runs`, `.kit/reports`, `.kit/HANDOFF.md`, and old `.kit/docs` projection.
  5. Confirm `.kit` contains tracked changesets only, plus ignored temporary conflict state when present.
- expected outputs:
  - clean one-plan repository history model
- verification:
  - `git ls-files '.kit/**' && test "$(git ls-files '.kit/**' | grep -vc '^.kit/changesets/')" -eq 0 && cd cli && go test ./...`
- stop if:
  - inventory coverage cannot be demonstrated
- escalate to:
  - check

## Risks / Watch-fors
- The phase deletes many tracked files; verify summary coverage and replay before each trash operation.
- Global installed skills remain on old playbooks until release/install; use repository sources and the phase binary for proof.
