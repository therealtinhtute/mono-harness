# Repository Workflow

`AGENTS.md` is the entrypoint; stage procedures belong to the six installed workflow skills.

## Authority

Classify the request before mutation: read-only requests change nothing; change requests stay within the active stage and approved scope, and discovery grants no authority to fix adjacent findings. Repository docs, code, tests, and runtime behavior are the truth.

## Context

Resolve the selected stage through the host's skill metadata. Read only its primary reference and the companions it routes to, relative to that installed skill. Resolve repository data paths from the consumer root. Install brainstorm, to-plan, work, check, handoff and watzup independently before upgrade. If a required skill or reference is unavailable, report the missing dependency and stop; retained legacy procedures are not a fallback. `zharness` neither detects nor installs skills and plays no part in running a stage.

| Stage | Reference in the installed skill’s references/ directory | Model tier |
|---|---|---|
| brainstorm | `brainstorm.md` | deep |
| to-plan (high-risk only) | `to-plan.md` | deep |
| work | `work.md` | standard |
| check | `check.md` | deep |
| handoff | `handoff.md` | fast |
| watzup | `watzup.md` | fast |

`git` keeps its own procedure. Tiers on Claude Code: deep = opus, standard = sonnet, fast = haiku. Do not switch model mid-phase.

## Execution boundary

Reduced modes write nothing durable; durable stages write only the plan sections their procedure names. Every proof claim names real output. Cite procedures by section and step number, never line number. IF tooling and a procedure disagree → trust the repository and report the mismatch.
