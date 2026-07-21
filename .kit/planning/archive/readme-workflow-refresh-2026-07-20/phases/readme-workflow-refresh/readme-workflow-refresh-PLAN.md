# Plan: README workflow refresh

Phase: readme-workflow-refresh
Status: ready
Wave Count: 3
Execution Owner: work
Updated At: 2026-07-20

## Goal

Ship an accurate, maintainable README usage guide and workflow diagram for the current zharness-backed lifecycle.

## Inputs

- `.kit/planning/SPEC.md`
- `.kit/planning/phases/readme-workflow-refresh/readme-workflow-refresh-CONTEXT.md`
- `.kit/plans/2026-07-20-readme-workflow-flow/PLAN.md`
- Current `README.md`, workflow SVG, migration guide, workflow docs, and diagram-design references

## Wave 1

### T1 — Build the canonical workflow flowchart

- type: docs
- inputs:
  - current `assets/spec-plan-workflow.svg`
  - diagram-design style guide and flowchart reference
  - canonical flow from the locked SPEC
- touches:
  - `assets/workflow-usage-flow.html`
  - `assets/spec-plan-workflow.svg`
- avoid:
  - `assets/diagram-layer-stack.png`
  - project source, skills, CLI, and setup files
- steps:
  1. Read the current SVG before replacing it.
  2. Design a top-to-bottom flowchart for full work, simple work, and resume routing.
  3. Keep the diagram within 9 nodes, 12 arrows, and 2 accent elements.
  4. Create a self-contained HTML file with inline CSS and SVG.
  5. Copy the same SVG markup into the existing README SVG asset with required namespace and font import support.
- expected outputs:
  - canonical self-contained workflow HTML
  - matching GitHub-renderable SVG at the existing path
- verification:
  - `python3` XML/HTML parser check confirming both files parse, the expected node labels exist in both, node/arrow/accent counts remain within budget, and the SVG viewBox supports README-width scaling
- stop if:
  - the visual needs more than 9 nodes or requires a new rendering dependency
  - HTML and SVG cannot share the same diagram markup
- escalate to:
  - `to-plan phase readme-workflow-refresh`

## Wave 2

### T2 — Rewrite the README around the current user flow

- type: docs
- inputs:
  - completed workflow diagram
  - `README.md`
  - canonical workflow and CLI references listed in phase context
- touches:
  - `README.md`
- avoid:
  - `CLAUDE.md`
  - skill and CLI implementation files
- steps:
  1. Read the full current README before editing.
  2. Remove the stale three-layer/13-skill visual claim from its authoritative position.
  3. Update the `.kit/` artifact model with generated and tracked paths.
  4. Separate bootstrap from required CLI installation.
  5. Consolidate normal, simple, resume, and legacy guidance into one user-facing workflow section.
  6. Preserve the existing workflow SVG path and add a link to the canonical HTML source.
- expected outputs:
  - one concise, non-duplicated README usage guide matching the locked requirements
- verification:
  - script checking README-local links, required flow/version terms, skill count, and absence of stale terms and retired commands
- stop if:
  - current behavior differs from the locked canonical flow
  - README accuracy requires implementation changes
- escalate to:
  - `brainstorm refine`

### T3 — Correct linked migration guidance

- type: docs
- inputs:
  - `docs/workflow-harness/migration.md`
  - current `zharness init` contract and implementation behavior
- touches:
  - `docs/workflow-harness/migration.md`
- avoid:
  - unrelated migration sections
  - CLI and embedded docs
- steps:
  1. Read the migration guide before editing.
  2. Remove the obsolete `mkdir -p .kit` requirement from the new-adopter path and legacy checklist.
  3. State that plain `init` scaffolds missing docs while `--refresh-docs` explicitly refreshes existing generated docs.
- expected outputs:
  - migration guide consistent with the README and current CLI behavior
- verification:
  - targeted stale-term search plus `cd cli && go test ./...`
- stop if:
  - tests contradict the documented init behavior
- escalate to:
  - `brainstorm refine`

## Wave 3

### T4 — Verify documentation, visual parity, and boundaries

- type: test
- inputs:
  - all Wave 1 and Wave 2 outputs
- touches:
  - `.kit/runs/work/`
  - `.kit/reports/check/` through the phase gate
- avoid:
  - implementation files outside approved scope
- steps:
  1. Parse HTML and SVG and compare required labels and structural budgets.
  2. Verify README-local links and search for all stale claims named by the SPEC.
  3. Run `cd cli && go test ./...`.
  4. Run `zharness validate --json` and `zharness audit --json`; classify any findings against the current full-harness artifacts.
  5. Inspect the final diff and confirm only approved documentation/assets plus harness evidence changed.
  6. Invoke `check full` for the phase gate.
- expected outputs:
  - command-backed verification evidence
  - clean or explicitly classified harness findings
  - phase gate report
- verification:
  - combined documentation validation script, Go tests, harness validate/audit, and `git diff --name-only`
- stop if:
  - any required link is broken, stale term remains, markup fails to parse, Go tests fail, visual budgets are exceeded, or diff scope drifts
- escalate to:
  - `work` for one targeted fix, then `check`

## Risks / Watch-fors

- Git may detect the replaced SVG as a large rewrite; verify content rather than relying only on diff size.
- Google Fonts are acceptable in HTML but the SVG must remain understandable when fonts fall back on GitHub.
- `zharness validate` findings must be evaluated against the new full lifecycle rather than treated as the previous empty-harness baseline.
