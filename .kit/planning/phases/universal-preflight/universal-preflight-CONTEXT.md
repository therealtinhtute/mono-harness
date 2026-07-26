# Context: universal preflight

Phase: universal-preflight
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: medium
Expected Proof: unit, integration, command-output, manual-check

## Goal
Give every active workflow skill one deterministic, read-only CLI rail guard without changing the current `.kit/docs` or `.kit/harness.db` layout yet.

## Scope Boundary
### Allowed Surfaces
- `cli/internal/domain/`
- `cli/internal/application/`
- `cli/internal/interfaces/{root,preflight}.go`
- matching Go tests
- `cli/docs/CONTRACT.md`
- eight active `skills/workflow/*/SKILL.md`
- `skills/workflow/README.md`

### Forbidden Surfaces
- DB path constants and `.gitignore`
- schema/migrations
- embedded workflow content beyond documenting preflight invocation
- lifecycle artifact paths/templates
- CI workflows

## Spec Hooks
- Requirements 3, 4, and 13.
- Preflight is mandatory; DB writes are not permitted for read-only/bounded routing.

## Locked Decisions
- Command: `zharness preflight <stage> [--mode <mode>] --json`.
- Eight stages: brainstorm, to-plan, work, check, handoff, watzup, git, interview.
- Missing DB yields reduced mode for read-only/bounded and blocked mode for durable mutation.
- Missing docs is reported; no read-only auto-init.
- Preflight returns the current playbook path so skills do not hardcode it.

## Assumptions
- Existing `next` remains work-specific and is not replaced in this phase.
- Existing lifecycle commands remain unchanged.

## Canonical Refs
- `.kit/planning/SPEC.md`
- `cli/internal/application/next.go`
- `cli/internal/interfaces/root.go`
- `cli/docs/CONTRACT.md`
- `skills/workflow/README.md`

## Rejected Options
- Put preflight logic separately in each skill: duplicates policy and drifts.
- Make preflight auto-init: mutates read-only invocations.
- Make every stage create an event: violates bounded-work simplicity.

## Deferred Ideas
- Root DB/docs path changes belong to root-layout.
- One-plan routing belongs to one-plan-lifecycle.

## Escalate If
- A stage cannot be classified as reduced or durable from invocation mode without new user-visible semantics.
- Existing command wiring forces a schema change.
