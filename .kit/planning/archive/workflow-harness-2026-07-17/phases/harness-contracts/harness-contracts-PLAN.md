# Plan: harness-contracts

Phase: harness-contracts
Status: ready
Wave Count: 2
Execution Owner: work
Updated At: 2026-07-17

## Goal
STATE.md, CONTRACT.md, SCHEMA.md written; 4 artifact templates carry frontmatter contracts; every legacy yml field mapped.

## Inputs
- `.kit/planning/SPEC.md` (R4, R6–R16)
- `docs/workflow-harness/gap-matrix.md` (story↔phase decision)
- `~/Lab/harness-experimental/crates/harness-cli/src/interface.rs` + `docs/TRACE_SPEC.md`

## Wave 1
### T1 — cli/docs/STATE.md (state contract v1)
- type: docs
- inputs:
  - SPEC R12, R14; current `to-plan/references/workflow-state-template.yml` (read-only)
- touches:
  - `cli/docs/STATE.md` (new)
- avoid:
  - editing the legacy yml template itself
- steps:
  1. Define state model: current/entry phase, status enum `planned|in-progress|checked|done`, artifact pointers by ULID, schema_version
  2. Write writer-ownership table (to-plan/work/check/handoff write; watzup/git read)
  3. Write stale-pointer rules: missing file, unknown phase slug, out-of-order run/check IDs — each with named recovery action
  4. Write legacy mapping table: every `workflow-state.yml` field → DB representation or "dropped, because…"
- expected outputs:
  - STATE.md with model, ownership, drift rules, complete legacy mapping
- verification:
  - Every field in `workflow-state-template.yml` appears in the mapping table: cross-check by inspection; `grep -iE 'TBD|TODO' cli/docs/STATE.md` empty
- stop if:
  - a legacy field has no sensible destination and dropping it loses user data
- escalate to:
  - user clarification

### T2 — Artifact frontmatter contracts in 4 templates
- type: docs
- inputs:
  - SPEC R15, R16; locked frontmatter keys (CONTEXT)
- touches:
  - `skills/workflow/brainstorm/references/spec-template.md`
  - `skills/workflow/work/references/run-artifact-template.md`
  - `skills/workflow/check/references/report-template.md`
  - `skills/workflow/handoff/references/handoff-template.md`
- avoid:
  - SKILL.md files; changing the human-facing body structure
- steps:
  1. Add required frontmatter block (`id`, `type`, `phase`, `lane`, `created`, `updated`) to each template
  2. Add per-template link keys: spec_id / plan_id+trace_ids / run_id+proof links / run_id+check_id
  3. Document the proof-link shape in the check report template (command, output ref, artifact path)
- expected outputs:
  - 4 templates with frontmatter contracts, bodies unchanged in intent
- verification:
  - `grep -l '^id:' skills/workflow/{brainstorm/references/spec-template.md,work/references/run-artifact-template.md,check/references/report-template.md,handoff/references/handoff-template.md}` lists all 4
- stop if:
  - a template's existing structure contradicts the contract keys
- escalate to:
  - to-plan phase harness-contracts

## Wave 2
### T3 — cli/docs/CONTRACT.md (19 command schemas)
- type: docs
- inputs:
  - T1 (state semantics), upstream interface.rs args
- touches:
  - `cli/docs/CONTRACT.md` (new)
- avoid:
  - inventing commands beyond SPEC R6's surface
- steps:
  1. Document all 19 commands: args, `--json` output shape, exit/error codes
  2. For each command, name its consumer skill or mark "reserved — adopted later" (propose, score-context)
  3. Add the workflow-step→CLI-action 1:1 mapping table
  4. Note deviations from upstream semantics explicitly
- expected outputs:
  - CONTRACT.md covering 19 commands, zero consumer-less commands
- verification:
  - Count command headings = 19; every heading section contains `--json` and an error table (inspection)
- stop if:
  - upstream semantics unclear for a command after reading interface.rs + domain.rs
- escalate to:
  - user clarification

### T4 — cli/docs/SCHEMA.md (DB + changesets)
- type: docs
- inputs:
  - T1, T3; SPEC R7–R9, R13
- touches:
  - `cli/docs/SCHEMA.md` (new)
- avoid:
  - premature index/optimization design (deferred in SPEC)
- steps:
  1. Define tables: ported (stories, decisions, backlog, traces, intakes, interventions, tools) + workflow (phases, runs, checks, handoffs)
  2. Map every table to a changeset entity type; define line shape `{op, entity, id, fields, at}` and file naming `{ulid}.changeset.jsonl`
  3. Specify replay ordering (ULID filename order), idempotency key rules, and the epoch-fence adaptation
- expected outputs:
  - SCHEMA.md with complete table/entity/changeset spec
- verification:
  - Every table named in SCHEMA.md has an entity type row; every CONTRACT.md mutating command names its entity (cross-check inspection)
- stop if:
  - a workflow entity cannot be expressed append-only
- escalate to:
  - brainstorm refine

## Risks / Watch-fors
- T3/T4 must cross-agree with T1/T2 — one pass of cross-checking IDs and enum values before closing the phase
- Do not let contract docs silently redefine what Phase 1 locked
