# Context: pilot-migration

Phase: pilot-migration
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: high
Expected Proof: e2e (pilot evidence) + docs walkthrough

## Goal
Prove the whole chain on a real task with published evidence and a go/no-go verdict; then ship migration/adoption docs and purge legacy `workflow-state.yml` semantics.

## Scope Boundary
### Allowed Surfaces
- Pilot target: one real repo/task chosen by the user (its `.kit/` artifacts)
- `docs/workflow-harness/migration.md` (new)
- Root `README.md` (quickstart), `skills/workflow/README.md` (evidence summary section)
- `CLAUDE.md` (purge yml semantics), `skills/workflow/to-plan/references/workflow-state-template.yml` (delete via `trash`)
- GitHub issues (gap filing)

### Forbidden Surfaces
- `cli/**` and `skills/workflow/*/SKILL.md` — pilot files issues, it does not hotfix
- Purge tasks before the pilot go verdict (sequencing rule)

## Spec Hooks
- R14 (retire yml + template removal), R22 (docs set), R23 (CLAUDE.md purge), R24 (pilot early, evidence, go/no-go)
- Acceptance: new user completes install→import→watzup from docs alone

## Locked Decisions
- Pilot task must be a real change in a real repo (not a toy) — **confirmed 2026-07-17: user chose to dogfood this repo (`Lab/skills`) itself**, not a separate target. Real task = importing this repo's own legacy `.kit/workflow-state.yml`-driven state into zharness, then executing pilot-migration's own remaining waves (T2 gap-issues/go-no-go, T3 migration docs, T4 purge) through the zharness-backed chain instead of the markdown-only flow phases 1–7 used.
- Pilot runs the full chain: intake → to-plan → work → check → handoff → cross-machine watzup resume
- Every gap found becomes a filed GitHub issue; zero hotfixing mid-pilot
- Purge sequence: pilot go verdict → migration.md → README quickstart → CLAUDE.md purge → template deletion (trash, never rm)
- Rollback notes state plainly: markdown readable without CLI; DB rebuildable from changesets; abandoning the CLI loses machine state + deterministic gates

## Assumptions
- Phases 1–7 merged; `cli/v0.x` release installable (using `/tmp/zharness` local dev build this session — no tagged release yet, matches SPEC's Acceptance Criteria's own "dev build satisfies gate" allowance already used throughout phases 1–7)
- This repo's own `.kit/` migrates as part of the pilot itself (dogfooding) — confirmed as the actual pilot target, not a hypothetical

## Canonical Refs
- `.kit/planning/SPEC.md` acceptance criteria (the pilot's checklist)
- `docs/workflow-harness/gap-matrix.md` (compare found gaps vs predicted)

## Rejected Options
- Piloting on a synthetic sample — proves nothing about real intake ambiguity and real proof classes
- Purging legacy docs before the pilot verdict — rollback would then need doc restoration too

## Deferred Ideas
- Adoption telemetry / usage stats; multi-project harness index

## Escalate If
- Pilot no-go — stop the phase, route findings to `brainstorm refine` for scope correction
- Migration reveals a legacy field the import loses that users actually need → to-plan phase cli-core
