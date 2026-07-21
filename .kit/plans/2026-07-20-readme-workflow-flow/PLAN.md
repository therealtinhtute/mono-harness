---
title: Refresh README for the zharness-backed usage flow
status: completed
created: 2026-07-20
scope: documentation-and-workflow-visual
---

# Refresh README for the zharness-backed usage flow

## Outcome

Make the root README teach the current user-facing workflow accurately: install the shared skills and `zharness`, initialize a project once, choose full or simple work, resume through `watzup`, and use the legacy import path only for old projects. Replace the obsolete workflow visual with a maintainable diagram using the current repository palette.

## Locked assumptions

- Diagram type: flowchart.
- Existing customized diagram palette is retained: white-smoke paper, jet-black ink, atomic-tangerine accent, blue-slate muted.
- Users invoke skills for normal work; internal `zharness intake`, `story`, trace, and hand-authored changeset commands are not taught as the primary workflow.
- The diagram remains within the complexity budget: at most 9 nodes, 12 arrows, and 2 accent elements.
- The HTML file is the canonical diagram source; the existing README SVG is refreshed from the same inline SVG so GitHub renders it directly.

## Files to change

1. `README.md`
   - Remove or replace stale claims about three layers and 13 skills.
   - Add `.kit/docs/` and clarify generated/ignored versus committed harness artifacts.
   - Split setup into repository bootstrap and required `zharness >= 0.4.1` installation.
   - Consolidate the duplicated workflow sections into one usage guide covering:
     - project initialization
     - full work: `brainstorm -> to-plan -> work -> check -> git/handoff`
     - simple work: `work simple`
     - resume: `watzup` as a readiness router
     - legacy adoption: `init -> import -> query -> validate -> audit`
   - Remove retired `plan`/`cook` terminology and direct `zharness story` onboarding.

2. `assets/workflow-usage-flow.html` — new
   - Self-contained diagram-design HTML with inline CSS and SVG.
   - Show current full, simple, and resume paths plus durable artifacts beneath spine stages.

3. `assets/spec-plan-workflow.svg`
   - Replace the obsolete drawing in place with the SVG used by the HTML source.
   - Preserve the current README asset path and GitHub inline rendering.

4. `docs/workflow-harness/migration.md`
   - Remove the obsolete requirement to create `.kit/` before `zharness init`.
   - Correct plain `init` versus `init --refresh-docs` behavior so the README's linked migration guide is consistent.

## Boundaries

### May change

- The four files listed above.
- README section ordering where needed to remove duplication.

### Must not change

- CLI behavior, embedded playbooks, skill implementations, or version constants.
- `CLAUDE.md` or unrelated skill documentation.
- Existing global configuration or installed skills.
- The unreferenced layer-stack PNG; the README may stop presenting it as authoritative, but the file will not be deleted.

## Implementation steps

1. Read the exact current contents of all four target files before editing.
2. Draft the flowchart in self-contained HTML using the diagram-design 4px grid, restrained accent usage, masked arrow labels, and current typography tokens.
3. Export the same inline SVG to `assets/spec-plan-workflow.svg` and confirm it remains readable at README width.
4. Rewrite the root README surgically around the canonical user flows and current artifact model.
5. Apply only the two factual migration-guide corrections needed by the README.
6. Verify documentation and visuals.

## Verification

- Render or parse the HTML and SVG; confirm valid markup, viewBox, fonts, labels, arrows, and mobile-width legibility.
- Verify every README-local link and asset path exists.
- Search for stale README terms and claims: `13 agent skills`, `3 abstraction layers`, retired `/plan` or `/cook`, direct `zharness story`, and the claim that `init` cannot create `.kit/`.
- Confirm the README states the `0.4.1` minimum and documents full, simple, resume, and legacy paths.
- Run `cd cli && go test ./...` because the migration statements describe live CLI behavior.
- Run `zharness validate --json` and `zharness audit --json`; record the expected pre-intake missing-SPEC baseline separately from documentation failures.
- Confirm `git diff --name-only` contains only the four planned files plus this plan/run evidence.

## Stop conditions

Pause before implementation if the current target files contradict the canonical flow, the SVG cannot be derived without adding a new tool dependency, or accurate README instructions require changing CLI or skill behavior.
