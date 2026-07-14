# Skills Harness Architecture

Status: draft
Scope: core development-workflow skills only
Included skills: `brainstorm`, `to-plan`, `work`, `check`, `handoff`, `watzup`
Excluded from this document: utility/domain-capability skills such as `librarian`, `prompt-leverage`, `git`, `turbo-mono-platform`, and `bash-tui`
Last updated: 2026-05-11

## Why this document exists

This repository already has strong individual skills. What it needs next is a
clear harness contract so the core workflow behaves like one operating system,
not six unrelated prompts.

The goal of this document is to define:

1. the workflow roles of the six core skills
2. the canonical vs ephemeral artifacts they own
3. the boundaries between planning, execution, proof, and continuity
4. the refactor targets needed to make the skills harness-first

This doc is the shared contract for future skill changes.

## Sources and design influences

This direction is informed by two reference points:

- OpenAI's Harness Engineering article: repository as agent operating surface,
  humans steer while agents execute, plans become first-class artifacts,
  legibility matters more than chat-only context.
- `hoangnb24/harness-experimental`: intake-first, docs-first harness design,
  where prompts become product/plan/proof artifacts before implementation.

This repository is different in one important way: it is a **skills repo**, not
an application repo. So the job here is not to ship product docs, but to make
our workflow skills produce durable artifacts that app repos can rely on.

## Mental model

The core harness should behave like this:

```text
User intent
  -> brainstorm
  -> to-plan
  -> work
  -> check
  -> handoff / watzup
```

And each step should answer a different question:

- `brainstorm` — What exactly are we doing?
- `to-plan` — How should the work be executed safely?
- `work` — What happened during execution?
- `check` — What proof shows it is correct and on-contract?
- `handoff` — How does the next session resume safely?
- `watzup` — What friction, entropy, or risk should the harness learn from?

## Harness layers

### 1. Intake and contract formation
Owner: `brainstorm`

Purpose:
- classify the input
- clarify ambiguity
- surface options and tradeoffs
- lock the chosen contract into a reusable artifact

Output class:
- canonical planning artifacts

### 2. Execution design
Owner: `to-plan`

Purpose:
- convert a locked contract into phased work
- define execution boundaries and expected proof
- prepare tasks so `work` can execute without guessing

Output class:
- canonical planning artifacts

### 3. Execution runtime
Owner: `work`

Purpose:
- execute the plan inside defined boundaries
- verify per task
- stop on drift, proof failure, or missing context
- leave an inspectable execution trace

Output class:
- execution artifacts
- code / repo changes in the target project

### 4. Proof and alignment gate
Owner: `check`

Purpose:
- verify code quality and readiness
- verify alignment between code, plan, and contract
- distinguish working code from on-contract code

Output class:
- gate verdicts
- optional execution-quality artifacts

### 5. Continuity
Owner: `handoff`

Purpose:
- summarize the current state for the next session
- point directly to the active contract, run, and next safe action

Output class:
- continuity artifact

### 6. Reflection and entropy control
Owner: `watzup`

Purpose:
- summarize work done
- detect stale artifacts, repeated blockers, or workflow friction
- suggest harness improvements, not just session summaries

Output class:
- retrospective artifacts

## Artifact classes

We should use exactly five artifact classes for the core harness.

### A. Canonical planning artifacts
These are the source of truth for scoped work.

Location:
- `.kit/planning/`

Files:
- `.kit/planning/IDEA.md`
- `.kit/planning/SPEC.md`
- `.kit/planning/ROADMAP.md`
- `.kit/planning/phases/{phase-slug}/{phase-slug}-CONTEXT.md`
- `.kit/planning/phases/{phase-slug}/{phase-slug}-PLAN.md`

Rules:
- owned by `brainstorm` and `to-plan`
- downstream skills may read them but must not rewrite them casually
- should be stable enough that a later session can continue from them
- should be treated as **canonical**, not scratch notes

### B. Execution artifacts
These are runtime traces of what actually happened during implementation.

Location:
- `.kit/runs/`

Files:
- `.kit/runs/work/{YYYYMMDD-HHmm}-{slug}.md`
- optional later: `.kit/runs/check/{YYYYMMDD-HHmm}-{slug}.md`

Rules:
- owned by `work` (and optionally `check`)
- append new runs; do not overwrite prior runs by default
- not a replacement for planning artifacts
- primary use: inspectability, auditability, and resume support

### C. Continuity artifacts
These make the next session safe and fast.

Location:
- `.kit/`

Files:
- `.kit/HANDOFF.md`

Rules:
- owned by `handoff`
- latest state wins; overwrite the file on each handoff
- should link to canonical planning artifacts and latest execution run

### D. Workflow state manifest
This is the lightweight index every downstream skill can consult first.

Location:
- `.kit/`

Files:
- `.kit/workflow-state.yml`

Rules:
- initialized by `to-plan` when roadmap + phase artifacts are created
- updated by downstream skills when canonical phase state changes
- should stay tiny: pointers and status only, never duplicate full artifact content
- latest state wins; overwrite in place

Recommended fields:
- `current_phase`
- `entry_phase`
- `spec`
- `roadmap`
- `active_context`
- `active_plan`
- `latest_cook_run`
- `latest_check_report`
- `handoff`
- `last_updated`

### E. Retrospective artifacts
These capture wrap-up, friction, and entropy observations.

Location:
- `.kit/reports/watzup/`

Files:
- `.kit/reports/watzup/{YYYYMMDD}-{slug}.md`
- `.kit/reports/watzup/{YYYYMMDD}-{slug}.html`

Rules:
- owned by `watzup`
- not canonical
- useful for review, history, and harness improvement proposals

## Canonical vs ephemeral

This distinction must stay sharp.

### Canonical
Canonical artifacts define the current work contract.

Examples:
- `.kit/planning/SPEC.md`
- `.kit/planning/ROADMAP.md`
- `.kit/planning/phases/...`

Properties:
- future skills depend on them
- changes should be intentional
- they answer "what is currently true?"

### Ephemeral
Ephemeral artifacts record support information, runtime traces, or reflections.

Examples:
- `.kit/workflow-state.yml`
- `.kit/runs/work/...`
- `.kit/HANDOFF.md`
- `.kit/reports/watzup/...`

Properties:
- useful for continuity and audit
- may accumulate over time
- they answer "what happened?" or "what should the next session know?"

## Ownership map

| Skill | Primary role | Reads | Writes | Artifact class |
| --- | --- | --- | --- | --- |
| `brainstorm` | intake + contract lock | user prompt, notes, markdown refs | `.kit/planning/IDEA.md`, `.kit/planning/SPEC.md`, explore reports | canonical planning |
| `to-plan` | execution design | `.kit/planning/SPEC.md` | `.kit/planning/ROADMAP.md`, phase context, phase plan, `.kit/workflow-state.yml` init | canonical planning + state |
| `work` | execution runtime | planning artifacts, workflow state | `.kit/runs/work/...`, `.kit/workflow-state.yml`, code changes | execution + state |
| `check` | proof + alignment gate | code diff, planning artifacts, latest work run, workflow state | console verdict, `.kit/reports/check/...`, `.kit/workflow-state.yml` | gate + state |
| `handoff` | continuity | branch state, planning artifacts, latest run/gate, workflow state | `.kit/HANDOFF.md`, `.kit/workflow-state.yml` | continuity + state |
| `watzup` | retrospective + entropy scan | repo state, planning artifacts, run artifacts, handoff, workflow state | `.kit/reports/watzup/...` | retrospective |

## Required metadata for canonical artifacts

To support machine legibility, canonical artifacts should include small,
consistent headers.

### `SPEC.md`
Recommended fields:

```text
Status: draft | locked
Input Type: new-spec | spec-slice | change-request | new-initiative | maintenance | harness-improvement
Lane: tiny | normal | high-risk
Risk Flags: auth, public-contract, weak-proof, ...
Affected Surfaces: api, ui, worker, db, docs
Downstream: to-plan full | to-plan phase | work simple | none
Updated At: YYYY-MM-DD
```

### `ROADMAP.md`
Recommended fields:

```text
Spec: .kit/planning/SPEC.md
Status: active | stale | superseded
Execution Mode: phased
Current Recommended Entry: {phase-slug}
Updated At: YYYY-MM-DD
```

### `PHASE-CONTEXT.md`
Recommended fields:

```text
Phase: {phase-slug}
Status: ready | stale | blocked
Scope Boundary: allowed / forbidden surfaces
Blast Radius: low | medium | high
Expected Proof: unit, integration, e2e, platform
Updated At: YYYY-MM-DD
```

### `PHASE-PLAN.md`
Recommended fields:

```text
Phase: {phase-slug}
Status: ready | stale | blocked
Wave Count: N
Execution Owner: work
Updated At: YYYY-MM-DD
```

## Required metadata for execution artifacts

### `work` run artifact
Recommended fields:

```text
Run ID: work-YYYYMMDD-HHmm-{slug}
Spec: .kit/planning/SPEC.md
Roadmap: .kit/planning/ROADMAP.md
Phase: {phase-slug}
Plan: .kit/planning/phases/{phase-slug}/{phase-slug}-PLAN.md
Mode: full | simple
Status: running | blocked | passed | aborted
Started At: YYYY-MM-DD HH:mm
```

Sections should include:
- preflight
- wave-by-wave execution
- per-task status
- verification result
- blockers
- next recommended action

## Overwrite vs append rules

### Overwrite in place
Use overwrite semantics for canonical planning artifacts:
- `.kit/planning/IDEA.md`
- `.kit/planning/SPEC.md`
- `.kit/planning/ROADMAP.md`
- phase context/plan files
- `.kit/HANDOFF.md`

Reason: they represent the current best known contract or current resume state.

### Append / create new files
Use append-by-new-file semantics for historical execution and retrospective
artifacts:
- `.kit/runs/work/...`
- `.kit/reports/watzup/...`

Reason: these should preserve history and support inspection.

## Skill-by-skill refactor targets

### `brainstorm`
Current strength:
- good at clarifying and locking the WHAT

Needed upgrade:
- become the formal intake gate

Refactor targets:
- add input-type classification before lock
- add lane/risk classification before lock
- add affected-surface declaration
- add downstream route declaration
- ensure `SPEC.md` is both human-readable and machine-legible

Definition of success:
- a later session can read the spec and know exactly what kind of work it is,
  how risky it is, and which skill should run next

### `to-plan`
Current strength:
- good at phased breakdown and execution intent

Needed upgrade:
- become the machine-readable execution design contract

Refactor targets:
- add explicit scope boundaries per phase
- add blast radius and proof expectations per phase
- give each task a stable mini-schema: inputs, touched surfaces, expected
  output, verification, stop conditions, escalation path
- optimize artifacts so `work` can execute without inventing missing structure

Definition of success:
- `work` can execute a phase by reading artifacts, not by reconstructing the
  plan from chat context

### `work`
Current strength:
- good orchestration shape
- already routes back upstream when artifacts are missing

Needed upgrade:
- become an execution runtime, not just an execution prompt

Refactor targets:
- always create a run artifact
- record preflight status before execution
- add drift detection before touching code
- standardize stop reasons
- support resume from the latest relevant run
- map per-task execution to verification output

Recommended stop reasons:
- `BLOCKED_CONTEXT`
- `BLOCKED_SCOPE`
- `BLOCKED_VERIFICATION`
- `BLOCKED_CONTRACT_DRIFT`
- optional later: `BLOCKED_EXTERNAL`, `BLOCKED_DEPENDENCY`

Definition of success:
- execution is inspectable, resumable, and scoped without relying on chat
  history alone

### `check`
Current strength:
- strong code gate and review posture

Needed upgrade:
- become a proof + alignment gate

Refactor targets:
- read the relevant planning artifacts before final verdict
- read the latest work run when present
- report not only code quality, but contract alignment
- add artifact-state verdicts: current, stale, missing
- add proof-gap detection when code changed but expected evidence is weak

Definition of success:
- a passing gate means both "the code works" and "the work stayed on-contract"

### `handoff`
Current strength:
- already writes a continuation artifact

Needed upgrade:
- become a continuation contract, not a loose session note

Refactor targets:
- link active phase explicitly
- link latest work run and latest gate result
- categorize blockers
- provide next safe action instead of vague next steps

Definition of success:
- the next session can resume by opening 2-3 artifacts, not by reading the full
  prior transcript

### `watzup`
Current strength:
- good session summary surface

Needed upgrade:
- become the entropy and friction scanner for the harness

Refactor targets:
- inspect stale planning artifacts
- inspect unresolved run blockers
- detect repeated friction patterns
- suggest harness improvements when patterns recur
- distinguish product risk from harness risk

Definition of success:
- the repository gets better over time because wrap-up surfaces structural
  friction, not just completed work

## Non-goals for this phase

To avoid overbuilding, this phase should not introduce:

- JSON schemas for every artifact
- a database for run state
- CI-only automation before the artifact contract is stable
- new utility skills
- a full application harness template inside this skills repo

We only need a clear markdown contract first.

## Recommended implementation order

### PR 1 — this document
Create and agree on the artifact architecture and role boundaries.

### PR 2 — `brainstorm` + `to-plan`
Add intake metadata, contract headers, and machine-readable planning structure.

### PR 3 — `work`
Add execution run artifacts, preflight drift checks, and stop taxonomy.

### PR 4 — `check`
Add artifact alignment gate and proof-gap reporting.

### PR 5 — `handoff` + `watzup`
Add continuity linking, entropy scanning, and harness friction reporting.

## Decision summary

This repository should treat the six core workflow skills as a single harness
system with distinct layers:

- `brainstorm` locks the contract
- `to-plan` makes the contract executable
- `work` records what execution did
- `check` proves alignment and readiness
- `handoff` preserves continuity
- `watzup` improves the harness from friction

The most important near-term leverage points are:

1. codify artifact ownership and classes
2. make `work` produce durable execution traces
3. make `brainstorm` and `to-plan` more machine-legible for downstream skills
