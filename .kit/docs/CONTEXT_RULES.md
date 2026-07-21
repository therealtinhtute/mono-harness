# CONTEXT_RULES

Maps each lifecycle stage — and the two commands outside the spine — to exactly the docs it must read, so an agent never over-reads.

## Universal Reads (once per session, before entering any stage)

1. `AGENTS.md` — entrypoint: version gate, request classification pointer, lifecycle stage list
2. `AUTHORITY.md` — classify the current request as read-only or change before running anything beyond `--version`/`--help`
3. This document — resolve exactly which playbook the current stage needs

These three do not repeat per stage. Read them once at session start, then jump straight to the stage row below and stay there for the rest of that stage's work.

## Per-Stage Table

| Stage | Reads | Does Not Read |
|---|---|---|
| brainstorm | `playbooks/brainstorm.md` | the other 5 playbooks |
| to-plan | `playbooks/to-plan.md` | the other 5 playbooks |
| work | `playbooks/work.md` | the other 5 playbooks — `work`'s phase-gate step hands off to `check` by name; switch to the `check` row instead of opening `playbooks/check.md` while still inside `work` |
| check | `playbooks/check.md` | the other 5 playbooks |
| handoff | `playbooks/handoff.md` | the other 5 playbooks |
| watzup | `playbooks/watzup.md` | the other 5 playbooks |

Each row's playbook is self-sufficient: stage purpose, preconditions, exact `zharness` commands with arguments, artifact paths and templates, and exit/handoff conditions all live inside it. No playbook requires a second playbook open at the same time. Each playbook also names its own project-artifact reads (SPEC.md, ROADMAP.md, phase files, run logs, HANDOFF.md) as part of its own steps — those are stage-specific project state, not additional embedded docs to track here.

## Outside the Spine

`git` and `interview` carry no dedicated playbook (per `AGENTS.md`) — handle both with `AUTHORITY.md`'s general read-only/change rules, not a stage-specific doc.

`git` reads nothing beyond that: its one required step before commit/PR operations is `zharness query check --latest --json` — warn on a `REQUEST_CHANGES` verdict or an unavailable result (no binary, unreadable db, no check recorded yet), never block staging or committing on it. No playbook read is needed for this single step.

## Anti-Pattern

Reading a second playbook "just in case" while operating a stage — if a stage's own playbook doesn't cover a needed step, that is a gap in the playbook (escalate it), not a signal to go read another one.
