# Artifact-Aware Retrospective

Use this when the repo follows the harness flow.

## Sources to Read

1. `.kit/workflow-state.yml` when present
2. `.kit/planning/ROADMAP.md`
3. active phase `-CONTEXT.md` and `-PLAN.md`
4. latest `.kit/runs/cook/*.md`
5. latest `.kit/HANDOFF.md` when present
6. latest `.kit/reports/check/*.md`

## What to Summarize

### Session Chain
- which phase moved forward
- whether execution reached `cook`, `check`, `handoff`
- whether the branch is ready for PR, needs proof cleanup, or needs plan refresh
- whether the workflow-state manifest stayed in sync with the artifact chain

### Recurring Patterns
Track repeated issues across the chain:
- `BLOCKED_CONTEXT`
- `BLOCKED_SCOPE`
- `BLOCKED_VERIFICATION`
- `BLOCKED_CONTRACT_DRIFT`
- proof gaps from `check`
- repeated phase-boundary drift

### Risk Mapping

| Pattern | Default severity | Example action |
|--------|------------------|----------------|
| Spec contradiction | cao | Refresh spec or phase plan before more code |
| Boundary drift | cao | Split unrelated changes or refresh phase scope |
| Missing verification proof | vừa | Re-run `cook` or append proof before merge |
| Repeated handoff churn | thấp | Tighten `handoff` next-step anchor |

## Deep Report Additions

When writing a deep report, include:
- `Artifact Chain` — spec / phase / cook / check / handoff status
- `Recurring Drift` — repeated blockers or proof gaps
- `Readiness` — `ready-for-pr` / `needs-proof` / `needs-plan-refresh` / `blocked`

## Skip Rule

If harness artifacts do not exist, say `artifact_chain: skipped` and fall back to normal session summary.
