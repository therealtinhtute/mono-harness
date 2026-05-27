# Check Examples

## Example 1 — Phase gate after work

**Input**: `check full`

**Observed context**:
- phase diff touches `src/inbox/*`
- `.kit/planning/SPEC.md` exists
- `.kit/planning/phases/inbox-ui/inbox-ui-PLAN.md` exists
- latest `.kit/runs/work/20240901-1010-inbox-ui.md` shows `pnpm test inbox.spec.ts`

**Response shape**:
```text
🥷 APPROVE with requests — code quality is fine, artifact proof is incomplete.

scope: on target
artifact_alignment: ❌ drift
- spec coverage is consistent with inbox-ui requirements
- phase boundary respected
- work run log is missing verification output for task T5

next:
- rerun `/work phase inbox-ui` or append proof to the run artifact
- then re-run `/check full`
```

## Example 2 — Boundary drift blocks clean gate

**Input**: `check full`

**Observed context**:
- phase plan allows `apps/web/inbox/**`
- diff also edits `packages/billing/src/ledger.ts`

**Response shape**:
```text
🥷 REQUEST CHANGES — phase boundary drift.

artifact_alignment: ❌ drift
- changed file outside allowed surfaces: `packages/billing/src/ledger.ts`
- no matching task or approved scope expansion found

next:
- split the unrelated change out, or
- refresh `/plan phase inbox-ui` if billing is genuinely in scope
```

## Example 3 — Non-harness repo

**Input**: `check review`

**Observed context**:
- no `.kit/planning/` or `.kit/runs/work/` artifacts exist

**Response shape**:
```text
🥷 APPROVE — standard review only.

artifact_alignment: skipped
- harness artifacts not present in this repo flow
- ran normal gate + code review only
```
