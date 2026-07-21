---
id: {ULID}
type: check
phase: {phase-slug} | none
lane: {tiny|normal|high-risk}
mode: {full|simple}
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
