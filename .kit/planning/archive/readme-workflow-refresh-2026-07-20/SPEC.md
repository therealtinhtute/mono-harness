---
id: 01KXYW686Z149RHC9E42V48BD4
type: spec
phase: none
lane: normal
intake_id: 01KXYW7JBX3ZPPFSB7VMAM6T46
created: 2026-07-20
updated: 2026-07-20
---

# SPEC: Refresh README for the zharness-backed usage flow

Status: locked
Input Type: change-request
Lane: normal
Risk Flags: public-contract, existing-behavior
Affected Surfaces: docs
Downstream: to-plan full
Updated At: 2026-07-20

## Source Mode

files

## Source Inputs

- User request to use `diagram-design` and update the root README for the new usage flow.
- `.kit/plans/2026-07-20-readme-workflow-flow/PLAN.md`
- `README.md`
- `skills/workflow/README.md`
- `docs/workflow-harness/migration.md`
- Diagram-design `style-guide.md` and `type-flowchart.md`

## Scenario

feature bootstrap

## Goal

Update the repository's primary public documentation and workflow visual so users can correctly install, initialize, start, resume, and complete work with the current zharness-backed skill chain.

## Users / Actors

- Developers installing this skills repository for Claude Code or another skills.sh-compatible agent.
- Contributors using the workflow harness in a new, existing, or legacy project.
- Maintainers who need the workflow diagram to remain editable and synchronized with the README.

## Requirements

1. The root README must describe the four-layer harness model and the current 14-skill repository without retaining the obsolete three-layer or 13-skill claims.
2. Setup instructions must separate global skill/config bootstrap from installation of `zharness >= 0.4.1`.
3. Project initialization must use `zharness init --json` directly and state that it creates `.kit/`, the local database, generated docs, and ignore entries as applicable.
4. The primary workflow must be skill-facing: `brainstorm -> to-plan -> work -> check -> git/handoff`.
5. The README must document bounded `work simple`, readiness-based resume through `watzup`, and the legacy `init -> import -> query -> validate -> audit` path.
6. The README must not teach retired `plan`/`cook` names or direct `zharness story` as the normal user path.
7. A self-contained HTML flowchart must show the full, simple, and resume paths plus durable artifact labels.
8. The flowchart must use the customized repository palette, a 4px grid, no more than 9 nodes, 12 arrows, or 2 accent elements, and diagram-design typography and masking rules.
9. The existing `assets/spec-plan-workflow.svg` must be replaced from the same inline SVG used by the HTML source so GitHub continues to render the diagram directly.
10. The linked migration guide must remove the obsolete pre-creation of `.kit/` and distinguish normal `init` from `init --refresh-docs`.

## Boundaries

### In Scope

- `README.md`
- `assets/workflow-usage-flow.html`
- `assets/spec-plan-workflow.svg`
- `docs/workflow-harness/migration.md`
- Harness planning, run, trace, and check evidence generated for this task

### Out of Scope

- CLI or skill behavior changes
- Embedded playbook or version-constant changes
- `CLAUDE.md` updates
- Deleting or regenerating `assets/diagram-layer-stack.png`
- New features beyond documenting the current flow

## Constraints

- Preserve the current public SVG path used by the README.
- Keep the HTML diagram self-contained with inline CSS and SVG and no JavaScript.
- Use only the existing Google Fonts dependency allowed by diagram-design.
- Keep README edits surgical and avoid duplicating the workflow in multiple sections.
- Do not add a rendering dependency solely to produce the SVG; derive it directly from the HTML's inline SVG.

## Acceptance Criteria

- `README.md` states the four layers, 14 skills, minimum CLI version, and full/simple/resume/legacy flows accurately.
- Every README-local link and asset path resolves.
- No stale README claim remains for 13 skills, three layers, retired `plan`/`cook`, direct `zharness story`, or inability of `init` to create `.kit/`.
- The HTML and SVG parse successfully and contain matching workflow nodes and labels.
- The SVG is legible at GitHub README width and the HTML remains usable at narrow viewport widths.
- The migration guide accurately documents current init behavior.
- `cd cli && go test ./...` passes.
- `git diff --name-only` remains inside the approved documentation and harness-evidence surfaces.

## Validation Expectations

- Unit: `cd cli && go test ./...` for the documented CLI behavior.
- Manual check: render or inspect HTML and SVG at desktop and narrow widths.
- Command output: parse markup, verify local links, search stale terms, compare diagram labels, and inspect diff boundaries.
- Harness: run `zharness validate --json` and `zharness audit --json`; distinguish task regressions from known contract limitations.

## Dependencies / Assumptions

- `zharness v0.4.1` remains the current minimum supported release.
- The repository's diagram style guide is already customized and does not require onboarding.
- GitHub renders the existing SVG asset path in README markdown.
- The approved custom plan is authoritative for scope and visual decisions.

## Key Decisions

- Chosen: keep a self-contained HTML source and refresh the existing SVG from the same diagram. This balances diagram-design maintainability with GitHub README rendering.
- Rejected: update only README prose and leave the old SVG. The visual would continue teaching retired names and topology.
- Rejected: add a second workflow diagram while retaining the old one. Duplicate visuals would drift and increase maintenance.
- Rejected: regenerate or delete the unrelated layer-stack PNG in this task. Removing its authoritative README role is sufficient and keeps scope focused.

## Open Questions

- None.

## Deferred Ideas

- Replace the legacy layer-stack PNG with a maintainable four-layer SVG in a separate documentation task.
- Fix unrelated root `AGENTS.md` shim links and embedded version-floor inconsistencies in a separate harness task.

## Ambiguity Report

- Goal clarity: clear; the public README and workflow visual must match current behavior.
- Scope clarity: clear; four documentation/asset files plus harness evidence.
- Constraints clarity: clear; preserve SVG path, use current diagram tokens, no behavior changes.
- Acceptance clarity: clear; stale-term, link, markup, visual, test, and diff checks are enumerated.
