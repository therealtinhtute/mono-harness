---
id: {PLAN_ULID}
type: plan
intake_id: {INTAKE_ULID}
lane: {tiny|normal|high-risk}
status: {active|completed}
created: {YYYY-MM-DD}
updated: {YYYY-MM-DD}
---

# Plan: {initiative title}

## Outcome
- result: {observable initiative outcome}
- success_signals:
  - {checkable success signal}

## Authority and Requirements
- authority:
  - {repository source, approved decision, or owner instruction}
- requirements:
  - R1 [accepted]: {falsifiable requirement} | source: {authority}

## Non-goals
- NG1: {explicitly excluded scope}

## Approach and Risks
- approach: not-planned
- constraints:
  - none
- risks:
  - none

## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Do not add task status fields. Append-only Progress is the sole task execution-status source. Only each phase lifecycle status changes: to-plan=planned; work=in-progress; clean durable check=checked; closing handoff=done. Each planned phase records phase_slug, story_id, status, goal, depends_on, waves, tasks, and checks. -->
- planning_status: not-planned
- phases: none

## Progress
<!-- Append-only durable entries record timestamp, phase, wave, task, task_status, exact verification/result, and changed surfaces or blocker. -->
- none

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- none

## Validation
<!-- Append-only durable entries record timestamp, phase, exact command/result/output, verdict, judge, receipt, and proof_gaps. -->
- none

## Current State and Next Action
- active_phase: none
- lifecycle_status: not-planned
- blockers: none
- open_items: [to-plan must define stable phases, stories, waves, tasks, and checks]
- exact_next_action: to-plan
