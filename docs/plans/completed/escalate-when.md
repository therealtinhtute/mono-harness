---
id: 01M0ESCWHENK4N2P8R
type: plan
intake_id: 01M0ESCINTAKEK4N2P8R
lane: tiny
status: completed
created: 2026-08-30
updated: 2026-08-30
---

# Plan: escalate-when — named stop predicates, ask the owner

## Outcome
- result: Agents grep `escalate_when` and get three stop predicates that require asking the owner instead of inventing: locked schema/requirements would change; the same verification command failed twice; a product rule conflicts. Retry cap stays one targeted fix. No new plan field. No hook.
- success_signals:
  - S1: `rg -n "escalate_when" cli/docs/embedded/playbooks/work.md docs/playbooks/work.md cli/docs/embedded/WORKFLOW.md docs/WORKFLOW.md` matches all four.
  - S2: The work.md block names the three predicates (schema/requirements change; same check failed twice; product-rule conflict) and the action is ask the owner / stop.
  - S3: `rg -n "One targeted fix is allowed after a failure" docs/playbooks/work.md` still matches; no “three retries” / “attempt <= 3”.
  - S4: `rg -n "escalate_when" cli/docs/embedded/templates/plan.md` prints nothing.
  - S5: `diff -q cli/docs/embedded/playbooks/work.md docs/playbooks/work.md` and `diff -q cli/docs/embedded/WORKFLOW.md docs/WORKFLOW.md` are silent.
  - S6: `bash scripts/verify-doc-links.sh` 0 findings; `bash scripts/test-guards.sh` still 0 failed.

## Authority and Requirements
- authority:
  - Owner, 2026-08-30: absorb only #1 from the article delta; prose `escalate_when`, not a YAML/plan field; not dual-encoded; #2–#6 Keep.
  - `docs/playbooks/work.md` step 6 — one targeted fix, then `BLOCKED_VERIFICATION`.
  - `docs/playbooks/brainstorm.md` step 4 — stop instead of inventing an unresolved product decision.
  - `docs/ARCHITECTURE.md` — playbook protocol, not a hook; tool gateway is host.
  - `docs/ARCHITECTURE.md` projection — edit `cli/docs/embedded/`, copy to `docs/`.
- requirements:
  - R1 [accepted]: Embedded `work.md` and its projection contain a labeled `escalate_when` block listing exactly three predicates: (a) the next edit would change locked schema or requirements; (b) the same verification command failed twice; (c) a product rule conflicts. Action for each: ask the owner and stop; do not invent. | source: owner lock #1
  - R2 [accepted]: Embedded `WORKFLOW.md` and its projection contain `escalate_when` naming the same three predicates (short form allowed). | source: owner lock #1
  - R3 [accepted]: work.md step 6 retry cap is unchanged: one targeted fix, then `BLOCKED_VERIFICATION`. Do not relax to three. | source: live work.md; owner keep cap 1
  - R4 [accepted]: No `escalate_when` field on the plan template, plan frontmatter, or Outcome YAML. | source: owner — not a per-plan field
  - R5 [accepted]: No hook, guard, or binary change. | source: ARCHITECTURE.md; owner no dual-encode
  - R6 [accepted]: Edit embedded sources, then copy to `docs/playbooks/work.md` and `docs/WORKFLOW.md`. | source: ARCHITECTURE.md projection

## Non-goals
- NG1: No YAML/frontmatter field on plans.
- NG2: No hook or guard for escalate_when.
- NG3: No tool-observation schema, receipt cost/tool_calls, or metrics (#2 #5 #6).
- NG4: No playbook slice/compiler; work.md and check.md stay whole-file maps (#3).
- NG5: No mid-wave absorb; no `failure_class` → retry mapping (#4).
- NG6: No SQLite, lifecycle CLI, or installer rewrite.
- NG7: No edits to ADR bodies or dated `docs/audit/*`.

## Approach and Risks
- approach: Insert a labeled `escalate_when` block into embedded `work.md` (next to Status Routing, where stops already live) listing the three predicates and “ask the owner / stop”. Add a short form of the same three predicates to embedded `WORKFLOW.md` under Execution boundary. Copy both to `docs/`. Do not touch the plan template, hooks, or other playbooks.
- constraints:
  - Edit `cli/docs/embedded/`, then copy to `docs/`.
  - work.md step 6 retry sentence stays byte-identical.
  - Lane tiny; proof is `rg` + doc-links + guards unchanged.
- rejected alternatives:
  - YAML/`escalate_when` field on every plan — rejected: same three bullets every lock (NG1, R4).
  - Fail-closed hook — rejected: “ask the owner” is not grepable from staged bytes (R5, NG2).
  - Also edit `brainstorm.md` — rejected this phase: lock-time stop already exists at step 4; R1/R2 name work.md + WORKFLOW.md only.
- risks:
  - Agents skip the prose block. Mitigation: labeled `escalate_when` heading; still not a hook (R5).
  - Retry cap accidentally relaxed while inserting. Mitigation: R3 greps the existing sentence; stop if it moves or changes.
- recovery: revert the two embedded files and their projections. Hooks untouched.
## Phases and Verification
- planning_status: planned
- phases:
  - phase_slug: p1-write-escalate-when
    story_id: 01M0ESCWHENPH1K4N2P8R
    status: done
    goal: `escalate_when` is grep-able from work.md and WORKFLOW.md with the three predicates; retry cap and plan template unchanged.
    depends_on: none
    surfaces_allowed: cli/docs/embedded/playbooks/work.md, docs/playbooks/work.md, cli/docs/embedded/WORKFLOW.md, docs/WORKFLOW.md
    surfaces_avoided: cli/cmd/**, cli/internal/**, scripts/install-git-hooks.sh, cli/docs/embedded/templates/plan.md, docs/playbooks/brainstorm.md, docs/playbooks/check.md, docs/playbooks/handoff.md, docs/decisions/**, docs/audit/**
    requirements: R1, R2, R3, R4, R5, R6
    waves:
      - wave: 1
        goal: Write the block and the short form; project both.
        tasks:
          - task: In embedded work.md, add a `## escalate_when` section next to Status Routing listing exactly three predicates (locked schema/requirements would change; same verification command failed twice; product rule conflicts) with action ask the owner and stop. Keep step 6 “One targeted fix is allowed after a failure” unchanged. Copy to docs/playbooks/work.md.
            verify: `rg -n "escalate_when" cli/docs/embedded/playbooks/work.md docs/playbooks/work.md` matches both; `rg -n "One targeted fix is allowed after a failure" docs/playbooks/work.md` matches; `diff -q cli/docs/embedded/playbooks/work.md docs/playbooks/work.md` is silent.
          - task: In embedded WORKFLOW.md, add a short `escalate_when` form under Execution boundary naming the same three predicates. Copy to docs/WORKFLOW.md.
            verify: `rg -n "escalate_when" cli/docs/embedded/WORKFLOW.md docs/WORKFLOW.md` matches both; `diff -q cli/docs/embedded/WORKFLOW.md docs/WORKFLOW.md` is silent.
        stop_condition: the insert changes the retry cap or adds a plan-template field — revert the hunk.
        escalation: product-rule wording needs a new predicate — ask the owner; do not invent a fourth.
    checks:
      - `rg -n "escalate_when" cli/docs/embedded/playbooks/work.md docs/playbooks/work.md cli/docs/embedded/WORKFLOW.md docs/WORKFLOW.md`
      - `rg -n "escalate_when" cli/docs/embedded/templates/plan.md`
      - `rg -n "attempt <= 3|three retries" docs/playbooks/work.md cli/docs/embedded/playbooks/work.md`
      - `bash scripts/verify-doc-links.sh`
      - `bash scripts/test-guards.sh`
## Progress
<!-- Append-only durable entries record timestamp, phase, wave, task, task_status, exact verification/result, and changed surfaces or blocker. -->
- [2026-08-30] (p1-write-escalate-when, wave 1): task_status=in-progress. Phase start.
- [2026-08-30] (p1-write-escalate-when, wave 1): task In embedded work.md, add a `## escalate_when` section next to Status Routing. task_status=DONE. Three predicates + ask the owner / stop. Copied to docs/playbooks/work.md. `diff -q` silent. Retry sentence unchanged.
- [2026-08-30] (p1-write-escalate-when, wave 1): task In embedded WORKFLOW.md, add a short `escalate_when` form under Execution boundary. task_status=DONE. Same three predicates. Copied to docs/WORKFLOW.md. `diff -q` silent.
- [2026-08-30] wave-summary: p1-write-escalate-when wave 1 done. S1 four-file rg exit 0; S4 plan template rg exit 1 (no field); guards 22 passed 0 failed; doc-links 0 findings.

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- [2026-08-30] absorb: none
- [2026-08-30] FullJudge3 is the independent `full`. The earlier FullJudge Validation was recorded after the authoring session had already formed a verdict.

## Validation
<!-- Append-only durable entries record timestamp, phase, exact command/result/output, verdict, judge, receipt, and proof_gaps. -->
- [2026-08-30 (gate)] mode `gate` verdict `APPROVED` — judge: `same-session` (lane: tiny). p1-write-escalate-when. Scope on-target. Surfaces: embedded+projected work.md and WORKFLOW.md only. Retry cap sentence unchanged. Plan template has no escalate_when field (rg exit 1). Same-session: whether agents will obey the prose block is not independently verified.
  - `rg -n "escalate_when" cli/docs/embedded/playbooks/work.md docs/playbooks/work.md cli/docs/embedded/WORKFLOW.md docs/WORKFLOW.md`
  - `rg -n "One targeted fix is allowed after a failure" docs/playbooks/work.md`
  - `diff -q cli/docs/embedded/playbooks/work.md docs/playbooks/work.md`
  - `diff -q cli/docs/embedded/WORKFLOW.md docs/WORKFLOW.md`
  - `bash scripts/verify-doc-links.sh`
  - `bash scripts/test-guards.sh`
  receipt:
    context_sources: [plan p1, work.md, WORKFLOW.md]
    policy: docs/playbooks/check.md
    judge: same-session
    judge_model: grok-4.6
    retries: 0
    rollback_point: none
    failure_ledger: absent
    not_independently_verified: prose-only escalate_when obedience
  proof_gaps: none
- [2026-08-30 (initiative full)] mode `full` verdict `APPROVED` — judge: `independent` (lane: tiny). FullJudge. Security: docs-only stop predicates; no secrets, hooks, auth. Performance: no runtime path. Architecture: playbook protocol not hook; projections byte-identical; no plan-template field. Code Quality: three predicates + ask the owner/stop; step 6 retry sentence unchanged. S3/S4 nested as test -z (exit 0 when empty).
  - `rg -n "escalate_when" cli/docs/embedded/playbooks/work.md docs/playbooks/work.md cli/docs/embedded/WORKFLOW.md docs/WORKFLOW.md`
  - `rg -n "One targeted fix is allowed after a failure" docs/playbooks/work.md`
  - `diff -q cli/docs/embedded/playbooks/work.md docs/playbooks/work.md`
  - `diff -q cli/docs/embedded/WORKFLOW.md docs/WORKFLOW.md`
  - `test -z "$(rg -n 'escalate_when' cli/docs/embedded/templates/plan.md || true)"`
  - `test -z "$(rg -n 'attempt <= 3|three retries' docs/playbooks/work.md cli/docs/embedded/playbooks/work.md || true)"`
  - `bash scripts/verify-doc-links.sh`
  - `bash scripts/test-guards.sh`
  receipt:
    context_sources: [plan p1, work.md, WORKFLOW.md, FullJudge]
    policy: docs/playbooks/check.md
    judge: independent
    judge_model: grok-4.6
    retries: 0
    rollback_point: none
    failure_ledger: absent
    not_independently_verified: none
  proof_gaps: none
- [2026-08-30 (re-review full)] mode `full` verdict `APPROVED` — judge: `independent` (lane: tiny). FullJudge3. Security: docs-only stop predicates; no secrets, auth, or privilege paths. Performance: docs only; no runtime path. Architecture: markdown protocol + embedded projection; no plan-template or hook drift. Code Quality: three predicates; step 6 retry cap preserved; projections byte-identical. proof_gaps none. surface_drift none.
  - `rg -n "escalate_when" cli/docs/embedded/playbooks/work.md docs/playbooks/work.md cli/docs/embedded/WORKFLOW.md docs/WORKFLOW.md`
  - `rg -n "One targeted fix is allowed after a failure" docs/playbooks/work.md`
  - `diff -q cli/docs/embedded/playbooks/work.md docs/playbooks/work.md`
  - `diff -q cli/docs/embedded/WORKFLOW.md docs/WORKFLOW.md`
  - `test -z "$(rg -n 'escalate_when' cli/docs/embedded/templates/plan.md || true)"`
  - `test -z "$(rg -n 'attempt <= 3|three retries' docs/playbooks/work.md cli/docs/embedded/playbooks/work.md || true)"`
  - `bash scripts/verify-doc-links.sh`
  - `bash scripts/test-guards.sh`
  receipt:
    context_sources: [plan p1, work.md, WORKFLOW.md, FullJudge3]
    policy: docs/playbooks/check.md
    judge: independent
    judge_model: grok-4.6
    retries: 0
    rollback_point: none
    failure_ledger: absent
    not_independently_verified: none
  proof_gaps: none

## Current State and Next Action
- active_phase: none
- lifecycle_status: done
- blockers: none
- open_items: none
- exact_next_action: none (initiative completed; absorb: none)
