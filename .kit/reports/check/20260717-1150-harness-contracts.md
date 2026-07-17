---
id: 01J0000000000000000HARNCT
type: check
phase: harness-contracts
lane: normal
run_id: work-20260717-1100-harness-contracts
proof_links: [{command: "grep -c '^### `' cli/docs/CONTRACT.md", output_ref: "inline below", artifact_path: "cli/docs/CONTRACT.md"}, {command: "grep -c '^- Errors:' cli/docs/CONTRACT.md", output_ref: "inline below", artifact_path: "cli/docs/CONTRACT.md"}, {command: "grep -iE 'TBD|TODO' cli/docs/STATE.md", output_ref: "inline below", artifact_path: "cli/docs/STATE.md"}, {command: "grep -l '^id:' across 4 templates", output_ref: "inline below", artifact_path: "skills/workflow/*/references/*-template.md"}]
created: 2026-07-17
updated: 2026-07-17
---

# CHECK REPORT

Run ID: check-20260717-1150-harness-contracts
Scope: full
Artifact Alignment: aligned (after remediation; drift on first pass)
Review Verdict: APPROVED
Phase: harness-contracts
Spec: .kit/planning/SPEC.md
Plan: .kit/planning/phases/harness-contracts/harness-contracts-PLAN.md
Workflow State: .kit/workflow-state.yml
Cook Run: .kit/runs/work/20260717-1100-harness-contracts.md
Created At: 2026-07-17 11:50

## Gate Evidence
- tests: none → not applicable (docs-only phase, no Go code exists yet)
- types: none → not applicable
- lint: none → not applicable
- build: none → not applicable
- task verification commands (docs-phase equivalent of gate evidence):
  - `grep -iE 'TBD|TODO' cli/docs/STATE.md` → no matches → pass
  - `grep -l '^id:' skills/workflow/{brainstorm/references/spec-template.md,work/references/run-artifact-template.md,check/references/report-template.md,handoff/references/handoff-template.md}` → all 4 → pass
  - `grep -c '^### `' cli/docs/CONTRACT.md` → 20 (19 commands, `db` split into 2 subcommand headings) → pass
  - `grep -c '^- Errors:' cli/docs/CONTRACT.md` → 20, matching 20 headings 1:1 → pass
  - SCHEMA.md table↔entity cross-check: 11 tables, 11 entity rows; CONTRACT.md's 19 commands split 8 entity-producing + 3 meta-only + 1 replay-special + 7 read-only = 19 → pass

## Artifact Alignment
- status: aligned (post-remediation)
- notes:
  - spec coverage: R4, R6-R16 all addressed (STATE.md, CONTRACT.md, SCHEMA.md, 4 template frontmatter contracts)
  - boundary compliance: first pass had 1 violation — CONTRACT.md documented 20 commands vs the phase's own locked 19 (`CONTEXT.md` Goal + `PLAN.md` T3's explicit `avoid`/verification clauses); remediated by reverting to 19 and replacing the "implemented deviation" framing with a deferred escalation note, consistent with `continuity-CONTEXT.md`'s own "escalate, don't shim" guidance
  - proof trail: all 4 tasks' verification commands re-run independently this check pass (not self-certified from the run log alone)
  - known, correctly-handled gap (not blocking): PLAN→SPEC cross-link has no carrier file this phase (`phase-plan-template.md` outside Allowed Surfaces) — flagged in `CONTRACT.md`'s `validate` section and `.kit/implementation-notes.md`, with T2's own named escalation path recorded
  - known, correctly-handled asymmetry (not blocking): `SCHEMA.md`'s `handoffs` table (SPEC R13) has no producing command among CONTRACT.md's 19 (SPEC R6 vs R18 gap) — table defined now to avoid a future breaking migration, command deferred to whichever phase resolves the gap

## Findings
### Critical
- none

### Major
- none (1 found on first pass — CONTRACT.md's 20-vs-19 command count — remediated before this report; see Artifact Alignment)

### Minor / Suggestions
- none blocking; 1 advisory noted above (PLAN→SPEC cross-link gap) — already correctly handled, no action needed

## Next Action
- ready for PR / advance to `cli-core` phase
