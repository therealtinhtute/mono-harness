# Check Report Template

Use this for `.kit/reports/check/{YYYYMMDD-HHmm}-{slug}.md` when the repo uses harness artifacts or when the user wants a persisted gate result.

```markdown
---
id: {ULID}
type: check
phase: {phase-slug}
lane: {tiny|normal|high-risk}
run_id: {ULID of the RUN this check gates}
proof_links: [{command, output_ref, artifact_path}, ...]
created: {YYYY-MM-DD}
updated: {YYYY-MM-DD}
---

# CHECK REPORT

Run ID: check-YYYYMMDD-HHmm-{slug}
Scope: gate | review | full
Artifact Alignment: aligned | drift | skipped
Review Verdict: APPROVED | APPROVE with requests | REQUEST CHANGES
Phase: {phase-slug} | none
Spec: .kit/planning/SPEC.md | none
Plan: .kit/planning/phases/{phase-slug}/{phase-slug}-PLAN.md | none
Workflow State: .kit/workflow-state.yml | none
Cook Run: .kit/runs/work/{file}.md | none
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
- rerun `work`
- refresh `to-plan phase {slug}`
- ready for PR
```

Rules:
- create one file per check run; do not overwrite older results from the same day unless the exact timestamp path is reused intentionally
- when harness artifacts are present, include the active phase and latest work run path if known
- after persisting the report, update `.kit/workflow-state.yml` with `latest_check_report`, keep `current_phase` unchanged unless the gate closed the phase, and refresh `last_updated`
- keep the persisted report consistent with the chat sign-off block
- frontmatter `run_id` links to the RUN this check gates; each `proof_links` entry is `{command, output_ref, artifact_path}` — `command` is the exact verification command run, `output_ref` is where its output is recorded (inline in this report or a path), `artifact_path` is the file the command verified
