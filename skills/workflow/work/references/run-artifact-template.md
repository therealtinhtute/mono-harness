# Cook Run Artifact Template

Use this structure for `.kit/runs/work/{YYYYMMDD-HHmm}-{slug}.md`:

```markdown
# COOK RUN

Run ID: work-YYYYMMDD-HHmm-{slug}
Mode: full | simple
Status: running | blocked | passed | aborted
Spec: .kit/planning/SPEC.md | none
Roadmap: .kit/planning/ROADMAP.md | none
Workflow State: .kit/workflow-state.yml | none
Phase: {phase-slug} | none
Plan: .kit/planning/phases/{phase-slug}/{phase-slug}-PLAN.md | none
Started At: YYYY-MM-DD HH:mm

## Preflight
- scope drift: yes | no
- working tree note
- required artifacts present: yes | no
- selected phase / source prompt

## Wave / Task Log
### Wave 1
#### T1 — Task name
- status: DONE | DONE_WITH_CONCERNS | NEEDS_CONTEXT | BLOCKED
- changed files:
  - path
- verification:
  - command → pass | fail
- notes:
  - concern or blocker detail

## Summary
- passed tasks
- blocked tasks
- unresolved concerns

## Next Recommended Action
- `check full`
- `to-plan phase {slug}`
- `brainstorm refine`
- `handoff`
```

Rules:
- create one new run file per invocation
- never overwrite an older run artifact
- every task attempted should appear in the log
- blocker reasons should map to the stop taxonomy
- after creating the run, update `.kit/workflow-state.yml` with `current_phase`, `active_context`, `active_plan`, `latest_cook_run`, and `last_updated`
