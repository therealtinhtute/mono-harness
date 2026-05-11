# ROADMAP Template

Use this structure for `.planning/ROADMAP.md`.

When writing or refreshing the roadmap, also initialize `.kit/workflow-state.yml` using `workflow-state-template.yml`.

```markdown
# ROADMAP: {title}

## Planning Basis
- source spec: `.planning/SPEC.md`
- planning mode: `full` | `phase`

## Phase 1: {phase name}
**Goal:** {what this phase achieves}

**Deliverables:**
- artifact or capability 1
- artifact or capability 2

**Dependencies:**
- upstream phase, file, system, or decision

**Risks / Watch-fors:**
- key risk 1
- key risk 2

## Phase 2: {phase name}
...
```

Rules:
- phase names should be short and stable
- order by dependency and risk reduction
- each deliverable should map back to the spec
- use as many phases as needed, but keep it lean
- set `entry_phase` and `current_phase` in `.kit/workflow-state.yml` to the recommended starting phase
