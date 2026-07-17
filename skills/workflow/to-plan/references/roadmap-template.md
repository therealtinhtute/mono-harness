# ROADMAP Template

Use this structure for `.kit/planning/ROADMAP.md`.

When writing or refreshing the roadmap, also run `zharness init` (if no db yet) and one `zharness story --slug {phase-slug} --goal "..."` per phase below.

```markdown
# ROADMAP: {title}

## Planning Basis
- source spec: `.kit/planning/SPEC.md`
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
- the recommended starting phase should be named in the roadmap header (`entry_phase` equivalent); harness state (`zharness query state`) tracks `current_phase` durably instead of a written yml file
