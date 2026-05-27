# Artifact-Aware Recap

Use this when the repo follows the harness flow (`.kit/planning/` artifacts present).

## Sources to Read

1. `.kit/workflow-state.yml` — first lookup index
2. `.kit/planning/ROADMAP.md` — active phase order
3. Active phase `-CONTEXT.md` and `-PLAN.md` — locked decisions, remaining tasks
4. Latest `.kit/runs/work/*.md` — task statuses, blockers, proof trail
5. `.kit/HANDOFF.md` — previous session context
6. Latest `.kit/reports/check/*.md` — gate verdict

## What to Summarize

### Session Context
- Which phase is active
- Whether execution reached `work`, `check`, `handoff`
- Whether the branch is `ready-for-pr`, `needs-work`, `needs-plan-refresh`, or `blocked`
- Whether workflow-state pointers are still valid

### Stale Detection
Check each pointer in `.kit/workflow-state.yml`:
- File exists? → valid
- File missing? → stale, report the missing path
- Timestamps inconsistent? → flag as potential drift

### Risk Mapping from Artifacts

| Pattern | Default severity | Example action |
|---------|-----------------|----------------|
| Spec contradiction | cao | Refresh spec or phase plan before more code |
| Boundary drift | cao | Split unrelated changes or refresh phase scope |
| Missing verification proof | vừa | Re-run work or append proof before merge |
| Stale workflow-state pointers | vừa | Refresh pointers or re-run the stale phase |

## Output Integration

In the recap output:
- **Context** section: include phase name, work run status, check verdict
- **Risks** section: include artifact-derived risks alongside code-derived risks
- **Readiness** state: factor artifact chain health into the overall assessment

## Skip Rule

If harness artifacts do not exist, omit artifact chain from the output entirely. Do not say "artifact chain: skipped" — just leave it out.
