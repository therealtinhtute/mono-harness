# Playbook: to-plan

## Purpose

Read a locked `.kit/planning/SPEC.md` and turn it into a phased implementation plan: a roadmap, per-phase context, and executable task waves. Own the HOW, but stay inside the spec's boundaries — never clarify product scope from scratch, execute code, or replace code review/testing (those belong to `brainstorm`, `work`, and `check` respectively).

## Preconditions

- **Version gate**: run `zharness --version` before anything else. A `dev` build always satisfies the gate. Otherwise, if the binary is missing or below `0.1.0` (`MIN_ZHARNESS_VERSION`), print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop.
- Require `.kit/planning/SPEC.md`. If missing, stop immediately and direct the user to run `brainstorm` first. Never invent plan artifacts from a vague prompt alone.

## Arguments

- `full` — generate roadmap + all phase context/plan artifacts (default)
- `phase [phase-name]` — refresh one named phase when the roadmap already exists; `[phase-name]` is required for this mode

## Steps

### Step 0 — Enforce precondition
Covered above: SPEC.md must exist.

### Step 1 — Read and normalize the spec
Extract at least: goal, actors, numbered requirements, in-scope/out-of-scope boundaries, constraints, acceptance criteria, validation expectations, dependencies/assumptions, sequencing questions, and intake metadata when present (input type, lane, risk flags, affected surfaces, downstream). If the spec is too weak for planning, stop and point back to `brainstorm` with the exact missing area.

### Step 2 — Build or refresh `.kit/planning/ROADMAP.md`
Use the Roadmap Template below.

Rules: split work into coherent phases; each phase needs a clear goal and deliverables; order respects dependencies and risk; do not create fake phases; the roadmap header names the current recommended entry phase and execution mode.

Run `zharness init` if no db exists yet (idempotent — `--json` reports `status: "exists"` when already initialized), then run `zharness story --slug {phase-slug} --goal "{phase goal}" [--depends-on {slug}]` once per roadmap phase — this is the durable record of roadmap/phase state. No command sets `current_phase`/`entry_phase` for a live project (only legacy `import` does); in `full` mode, after all stories are created, run `zharness id --json` and use that fresh ID as the filename for a one-line meta changeset (`.kit/changesets/{changeset-id}.changeset.jsonl`, `{"op":"update","entity":"meta","id":"meta","fields":{"current_phase":"{entry phase slug}","entry_phase":"{entry phase slug}"},"at":"{RFC3339 now}"}`), then apply it with `zharness db changeset apply {path} --json` — the same generic, already-shipped command `work` uses to register runs. In `phase` mode, only touch `current_phase` this way if the refreshed phase should now be active; mint a fresh changeset ID the same way.

### Step 3 — Create phase context files
For each roadmap phase, write `.kit/planning/phases/{phase-slug}/{phase-slug}-CONTEXT.md` using the Phase Context Template below. Each context file locks implementation decisions implied by the spec, phase-specific assumptions, canonical refs, rejected options, deferred ideas, allowed/forbidden surfaces, blast radius, expected proof class, and escalation conditions back to `brainstorm` or `to-plan`. If repo context is too unclear, note it explicitly as an open assumption.

### Step 4 — Create executable phase plans
For each roadmap phase, write `.kit/planning/phases/{phase-slug}/{phase-slug}-PLAN.md` using the Phase Plan Template below. Task rules: group tasks into waves; parallelize only when dependencies truly allow it; keep each task specific and actionable; include expected outputs, verification, touched/avoid surfaces, stop conditions, and escalation path; keep tasks inside spec boundaries; do not drift into post-hoc product design.

### Step 5 — State integrity + handoff guidance
Before finishing, run `zharness query state --json` and confirm `current_phase` matches the recommended entry phase (`full` mode) or the just-refreshed phase (`phase` mode, only if it should now be active); run `zharness query phases --json` and confirm one story exists per roadmap phase, slugs matching. Then suggest `work` to execute the phase next; `check` gates after implementation; `git` or `handoff` follow for wrap-up or transfer.

## Planning Rules

1. **Respect the spec** — every phase must trace back to the spec; do not add product scope just because it feels useful; out-of-scope items stay out.
2. **Phase by dependency, not by vibes** — earlier phases should reduce uncertainty or unlock later work; avoid phases that mix unrelated concerns; merge tiny phases when they don't need separate tracking.
3. **Context before task explosion** — lock the critical decisions for a phase before writing big task trees; if a key decision is unknown, surface it instead of bluffing.
4. **Waves should mean something** — Wave 1 is setup or independent foundation; Wave 2+ depends on earlier outputs; avoid fake parallelism.
5. **Verification is mandatory** — each meaningful task needs a verification path: tests, inspection steps, or explicit artifact checks.
6. **Suggest other skills only when useful** — `check` after implementation; `git`, `watzup`, `handoff` when wrapping or transferring context.

## Artifacts

### ROADMAP.md — `.kit/planning/ROADMAP.md`

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

Rules: phase names short and stable; order by dependency and risk reduction; each deliverable maps back to the spec; use as many phases as needed, but keep it lean; the recommended starting phase is named in the roadmap header (`entry_phase` equivalent) — harness state (`zharness query state`) tracks `current_phase` durably instead of a written yml file.

### Phase Context — `.kit/planning/phases/{phase-slug}/{phase-slug}-CONTEXT.md`

```markdown
# Context: {phase name}

Phase: {phase-slug}
Status: ready | stale | blocked
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: low | medium | high
Expected Proof: unit, integration, e2e, platform

## Goal
Short restatement of the phase goal.

## Scope Boundary
### Allowed Surfaces
- files / modules / layers this phase may touch

### Forbidden Surfaces
- areas explicitly out of scope for this phase

## Spec Hooks
- requirement(s) served
- boundary or constraint that matters here

## Locked Decisions
- decision 1
- decision 2

## Assumptions
- assumption 1
- assumption 2

## Canonical Refs
- `.kit/planning/SPEC.md`
- `.kit/planning/ROADMAP.md`
- other docs / files / APIs if relevant

## Rejected Options
- option + why rejected

## Deferred Ideas
- future idea intentionally not done now

## Escalate If
- condition that should route back to `brainstorm` or `to-plan`
```

Rules: keep only decisions that help implementation; do not restate the entire spec; rejected options explain tradeoffs briefly; deferred ideas must not leak back into scope; allowed/forbidden surfaces must be concrete enough that `work` can detect drift.

### Phase Plan — `.kit/planning/phases/{phase-slug}/{phase-slug}-PLAN.md`

```markdown
# Plan: {phase name}

Phase: {phase-slug}
Status: ready | stale | blocked
Wave Count: {N}
Execution Owner: work
Updated At: YYYY-MM-DD

## Goal
What this phase should accomplish.

## Inputs
- required files
- required prior decisions

## Wave 1
### T1 — Task name
- type: implementation | test | migration | docs | refactor
- inputs:
  - required artifact or file
- touches:
  - files / modules / surfaces expected to change
- avoid:
  - forbidden or out-of-scope areas
- steps:
  1. first step
  2. second step
- expected outputs:
  - file / endpoint / behavior
- verification:
  - test, inspection, or command
- stop if:
  - ambiguity / drift / dependency condition
- escalate to:
  - brainstorm refine | to-plan phase | user clarification | check

## Wave 2
### T2 — Task name
(same shape as T1)

## Risks / Watch-fors
- important coordination or sequencing risk
```

Rules: only place tasks in the same wave if they can proceed independently; steps concrete enough for execution; verification must be observable; expected outputs explicit, not implied; each task specific enough that `work` can report task-level status without inventing new structure.

## Command Reference

- `zharness --version` — version gate
- `zharness id --json` — mint a fresh filename ID before every manually-authored meta changeset
- `zharness init` — idempotent; run before the first `story` if no db exists
- `zharness story --slug {phase-slug} --goal "{phase goal}" [--depends-on {slug}]` — once per roadmap phase
- `zharness db changeset apply {path} --json` — applies the meta changeset that sets `current_phase`/`entry_phase`
- `zharness query state --json` — confirm `current_phase` after writing
- `zharness query phases --json` — confirm one story per roadmap phase

## Exit / Handoff Conditions

Complete only when: `.kit/planning/SPEC.md` was enforced; `.kit/planning/ROADMAP.md` exists; every phase has both `-CONTEXT.md` and `-PLAN.md`; plans are wave-based and executable; phase contexts declare boundaries/proof expectations; task detail is sufficient for `work`; next-step suggestions are clear (`work` to execute, `check` to gate, `git`/`handoff` to wrap up).

If blocked, return a short fail-fast explanation naming the missing spec gap — do not silently produce a partial plan.

## Anti-Patterns

- Creating phases that are just task lists — phases need goals, boundaries, allowed/forbidden surfaces, and proof expectations
- Writing vague tasks ("implement the feature") — `work` can't execute what it can't verify; every task needs a verification command
- Adding scope beyond spec boundaries — scope creep via planning is still scope creep; route back to `brainstorm` instead
- Omitting verification commands per task — `work` treats missing verification as task-not-done
