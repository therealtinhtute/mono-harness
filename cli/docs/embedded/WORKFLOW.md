# Repository Workflow

`AGENTS.md` is the entrypoint; stage procedure lives under `docs/playbooks/`.

## Authority

Classify the request before mutation: read-only requests change nothing; change requests stay within the active stage and approved scope, and discovery grants no authority to fix adjacent findings. Repository docs, code, tests, and runtime behavior are the truth.

## Context

Read only the stage's playbook and the companion it routes to. `zharness` (install / update / uninstall) plays no part in running a stage.

| Stage | Playbook | Model tier |
|---|---|---|
| brainstorm | `docs/playbooks/brainstorm.md` | deep |
| to-plan (high-risk only) | `docs/playbooks/to-plan.md` | deep |
| work | `docs/playbooks/work.md` | standard |
| check | `docs/playbooks/check.md` | deep |
| handoff | `docs/playbooks/handoff.md` | fast |
| watzup | `docs/playbooks/watzup.md` | fast |

`git` keeps its own procedure. Tiers on Claude Code: deep = opus, standard = sonnet, fast = haiku. Do not switch model mid-phase.

## Execution boundary

Reduced modes write nothing durable; durable stages write only the plan sections their playbook names. Every proof claim names real output. Cite playbooks by section and step number, never line number. IF tooling and a playbook disagree → trust the repository and report the mismatch.
