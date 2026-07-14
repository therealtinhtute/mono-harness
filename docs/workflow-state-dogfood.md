# Workflow State Dogfood Example

This reference shows one minimal end-to-end harness slice where `.kit/workflow-state.yml` stays coherent across `to-plan` → `work` → `check` → `handoff` → `watzup`.

## Scenario

Branch slug: `feat/workflow-state-operational-hooks`

Phase slug: `workflow-state-deeper-support`

Goal: tighten deeper operational support for the workflow-state manifest without adding new commands.

## 0) Before `/to-plan`

`/brainstorm` may lock `.planning/SPEC.md`, but it does not initialize `.kit/workflow-state.yml`. The manifest starts at `to-plan`, not earlier.

### Workflow-state status

```text
.kit/workflow-state.yml does not exist yet
```

What matters here:
- scope can be locked before any execution-state index exists
- `brainstorm` stays focused on planning artifacts only
- `to-plan` is the first skill that creates the manifest

## 1) After `/to-plan`

`/to-plan` writes or refreshes the canonical phase artifacts, then initializes the manifest as the first lookup index.

### `.kit/workflow-state.yml`

```yaml
current_phase: workflow-state-deeper-support
entry_phase: workflow-state-deeper-support
spec: .kit/planning/SPEC.md
roadmap: .kit/planning/ROADMAP.md
active_context: .kit/planning/phases/workflow-state-deeper-support/workflow-state-deeper-support-CONTEXT.md
active_plan: .kit/planning/phases/workflow-state-deeper-support/workflow-state-deeper-support-PLAN.md
latest_cook_run: none
latest_check_report: none
handoff: none
last_updated: 2026-05-11 21:52
```

What changed:
- `entry_phase` and `current_phase` both point at the recommended starting phase
- phase pointers are exact
- downstream artifacts stay `none` until they exist

## 2) After `/work full phase workflow-state-deeper-support`

`/work` reads the manifest first, verifies the pointed phase files, creates a new work run, then refreshes only the runtime pointers.

### New run artifact

Path:

```text
.kit/runs/work/20260511-2202-workflow-state-deeper-support.md
```

### `.kit/workflow-state.yml`

```yaml
current_phase: workflow-state-deeper-support
entry_phase: workflow-state-deeper-support
spec: .kit/planning/SPEC.md
roadmap: .kit/planning/ROADMAP.md
active_context: .kit/planning/phases/workflow-state-deeper-support/workflow-state-deeper-support-CONTEXT.md
active_plan: .kit/planning/phases/workflow-state-deeper-support/workflow-state-deeper-support-PLAN.md
latest_cook_run: .kit/runs/work/20260511-2202-workflow-state-deeper-support.md
latest_check_report: none
handoff: none
last_updated: 2026-05-11 22:09
```

What changed:
- `latest_cook_run` now points at the exact run artifact just created
- phase pointers stay unchanged because `work` does not guess the next phase

## 3) After `/check full`

`/check` reads the manifest first, verifies the phase + run chain, persists a report, then refreshes only the gate pointer fields.

### New check report

Path:

```text
.kit/reports/check/20260511-2211-workflow-state-deeper-support.md
```

### `.kit/workflow-state.yml`

```yaml
current_phase: workflow-state-deeper-support
entry_phase: workflow-state-deeper-support
spec: .kit/planning/SPEC.md
roadmap: .kit/planning/ROADMAP.md
active_context: .kit/planning/phases/workflow-state-deeper-support/workflow-state-deeper-support-CONTEXT.md
active_plan: .kit/planning/phases/workflow-state-deeper-support/workflow-state-deeper-support-PLAN.md
latest_cook_run: .kit/runs/work/20260511-2202-workflow-state-deeper-support.md
latest_check_report: .kit/reports/check/20260511-2211-workflow-state-deeper-support.md
handoff: none
last_updated: 2026-05-11 22:11
```

What changed:
- `latest_check_report` now points at the persisted gate verdict
- `current_phase` stays put unless the gate explicitly closes the phase

## 4) After `/handoff`

`/handoff` reads the manifest first, summarizes the active phase/run/gate chain, writes `.kit/HANDOFF.md`, then refreshes the handoff pointer.

### `.kit/workflow-state.yml`

```yaml
current_phase: workflow-state-deeper-support
entry_phase: workflow-state-deeper-support
spec: .kit/planning/SPEC.md
roadmap: .kit/planning/ROADMAP.md
active_context: .kit/planning/phases/workflow-state-deeper-support/workflow-state-deeper-support-CONTEXT.md
active_plan: .kit/planning/phases/workflow-state-deeper-support/workflow-state-deeper-support-PLAN.md
latest_cook_run: .kit/runs/work/20260511-2202-workflow-state-deeper-support.md
latest_check_report: .kit/reports/check/20260511-2211-workflow-state-deeper-support.md
handoff: .kit/HANDOFF.md
last_updated: 2026-05-11 22:14
```

What changed:
- `handoff` now points at the fresh continuity artifact
- all earlier pointers remain intact for the next session

## 5) During `/watzup deep`

`/watzup` reads the manifest first, then the linked phase/run/check/handoff artifacts to summarize readiness and drift.

### Read pattern

```text
.kit/workflow-state.yml
→ .kit/planning/ROADMAP.md
→ .kit/planning/phases/workflow-state-deeper-support/*
→ .kit/runs/work/20260511-2202-workflow-state-deeper-support.md
→ .kit/reports/check/20260511-2211-workflow-state-deeper-support.md
→ .kit/HANDOFF.md
```

### Example deep report conclusions

- `artifact_chain: complete`
- `workflow_state: current`
- `readiness: ready-for-pr`
- `active_phase: workflow-state-deeper-support`

## Operational rules this example demonstrates

- `to-plan` initializes the manifest after writing canonical planning artifacts
- `work` updates only `current_phase`, phase pointers, `latest_cook_run`, and `last_updated`
- `check` updates only `latest_check_report` and `last_updated` unless it explicitly closes the phase
- `handoff` updates `handoff`, preserves the active phase anchor, and refreshes `last_updated`
- `watzup` reads the manifest first but does not mutate it
- the manifest stays pointer-only; detailed evidence lives in the linked artifacts
