# ROADMAP: README workflow refresh

Entry Phase: readme-workflow-refresh
Execution Mode: full

## Planning Basis

- source spec: `.kit/planning/SPEC.md`
- planning mode: `full`

## Phase 1: README workflow refresh

**Goal:** Replace stale workflow documentation and visuals with the current zharness-backed install, full, simple, resume, and legacy flows.

**Deliverables:**

- Self-contained HTML flowchart using the current diagram-design tokens.
- GitHub-renderable SVG derived from the same flowchart.
- Consolidated and corrected root README usage guide.
- Migration guide corrections for current `zharness init` behavior.
- Verification evidence for markup, links, stale terms, CLI tests, harness state, and diff boundaries.

**Dependencies:**

- Locked `.kit/planning/SPEC.md`.
- Existing diagram-design style guide and flowchart conventions.
- Current CLI behavior in `cli/` and workflow behavior in `skills/workflow/`.

**Risks / Watch-fors:**

- HTML and SVG diverging after edits.
- README teaching internal CLI commands instead of skill-facing workflow.
- Diagram exceeding its visual complexity budget or becoming unreadable at README width.
- Documentation edits drifting into unrelated CLI, skill, or project-instruction changes.
