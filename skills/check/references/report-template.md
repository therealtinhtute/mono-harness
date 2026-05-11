# Check Report Template

Use this for `.kit/reports/check/{YYYYMMDD-HHmm}-{slug}.md` when the repo uses harness artifacts or when the user wants a persisted gate result.

```markdown
# CHECK REPORT

Run ID: check-YYYYMMDD-HHmm-{slug}
Scope: gate | review | full
Artifact Alignment: aligned | drift | skipped
Review Verdict: APPROVED | APPROVE with requests | REQUEST CHANGES
Phase: {phase-slug} | none
Spec: .planning/SPEC.md | none
Plan: .planning/phases/{phase-slug}/{phase-slug}-PLAN.md | none
Workflow State: .kit/workflow-state.yml | none
Cook Run: .kit/runs/cook/{file}.md | none
Created At: YYYY-MM-DD HH:mm

## Gate Evidence
- tests: {command} → pass | fail | none
- types: {command} → pass | fail | none
- lint: {command} → pass | fail | none
- build: {command} → pass | fail | none

## Artifact Alignment
- status: aligned | drift | skipped
- notes:
  - spec coverage / gap
  - boundary compliance / drift
  - proof trail status

## Findings
### Critical
- none | finding

### Major
- none | finding

### Minor / Suggestions
- none | finding

## Next Action
- rerun `cook`
- refresh `plan phase {slug}`
- ready for PR
```

Rules:
- create one file per check run; do not overwrite older results from the same day unless the exact timestamp path is reused intentionally
- when harness artifacts are present, include the active phase and latest cook run path if known
- after persisting the report, update `.kit/workflow-state.yml` with `latest_check_report`, keep `current_phase` unchanged unless the gate closed the phase, and refresh `last_updated`
- keep the persisted report consistent with the chat sign-off block
