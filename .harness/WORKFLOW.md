# Repository Workflow

`AGENTS.md` is the entrypoint; stage procedures belong to the six installed workflow skills.

## Authority

Classify the request before mutation: read-only requests change nothing; change requests stay within the active stage and approved scope, and discovery grants no authority to fix adjacent findings. Repository docs, code, tests, and runtime behavior are the truth.

## Context

Resolve the selected stage through the host's skill metadata. Read only its primary reference and the companions it routes to, relative to that installed skill. Resolve repository data paths from the consumer root. Install brainstorm, to-plan, work, check, handoff and watzup independently before upgrade. If a required skill or reference is unavailable, report the missing dependency and stop; retained legacy procedures are not a fallback. `zharness` neither detects nor installs skills and plays no part in running a stage.

| Stage | Reference in the installed skill’s references/ directory |
|---|---|
| brainstorm | `brainstorm.md` |
| to-plan (high-risk only) | `to-plan.md` |
| work | `work.md` |
| check | `check.md` |
| handoff | `handoff.md` |
| watzup | `watzup.md` |

`git` keeps its own procedure. None of the six stages above requires a named model; the host's current model runs it. Model use is a capability role: the lead decides and integrates; a researcher answers a narrow read-only question with explicit ownership; a reviewer or judge is independent of the author where the plan requires it. Any capable model may fill a role (for example, as of 2026-10-21, Claude Haiku 5.5 for narrow research). When work moves to another model or session, hand over a transfer brief: goal and scope, plan path and phase, source identity (HEAD SHA plus a hash of the tracked dirty diff and relevant untracked inputs), evidence pointers (files changed, last passing proof), open decisions, and the exact next action.

## Execution boundary

Reduced modes write nothing durable; durable stages write only the plan sections their procedure names. Every proof claim names real output. Lane → required proof, final check mode, and judge independence come from `check.md` § Lane Matrix alone. Cite procedures by section and step number, never line number. IF tooling and a procedure disagree → trust the repository and report the mismatch.
