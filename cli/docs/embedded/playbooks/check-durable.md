# Playbook: check (durable gate and full)

Loaded by `docs/playbooks/check.md` step 2 for durable `gate`/`full`. `bounded` never loads it. Numbering follows `check.md`: Preconditions step 2, then Steps 4 and 8–11.

## Preflight

2. **Read-only preflight** (initiative-intent `auto`, explicit `gate`/`full`) — count every non-empty markdown plan under `docs/plans/active/` before matching the requested initiative. Require exactly one active plan in total, matching the request; IF multiple → list every candidate and stop even when only one matches the request. Read only its selected phase and `## Current State and Next Action`, re-reading after any compaction; require the phase to read `in-progress` and Current State to agree. A missing or ambiguous phase, any other status (including unstarted, `checked`, or `done`), or a conflict → stop before checks or writes and report the exact mismatch. Never change phase status to pass preflight.

## Owned Plan State

Only durable gate/full may: append to `## Validation`; update the selected phase's lifecycle `status:` line in `## Phases and Verification`; update lifecycle status, blockers/open items, and exact next action in `## Current State and Next Action`. `## Log` is the sole task execution-status source: read task state there; never add task-definition status fields.

## Durable Steps

4. **Review plan alignment** (gate/full) — compare the diff with the Goal's requirements and non-goals, phase surfaces, task checks, and Log `decision` entries. Missing planned proof is a finding even when local tests pass. For each requirement the phase `goal:` names, judge its `acceptance:` (none → judge the requirement text) and record it in the entry's `requirements:` line as `met`, `not met`, or `partial (→ <phase>)` when a later phase's `goal:` also names it. `not met` is a material plan contradiction; `partial` is not a finding. `full` on the final phase also judges the Goal's `success_signal:`. Record `rollback_point:` as the commit the phase can be reverted to. IF `docs/evals/failures.md` exists → for every failure class recorded two or more times, state whether the diff is clean of it.
8. **Append durable evidence** — read `docs/playbooks/check-validation.md`, then write the Validation entry by hand in that format. Never self-certify: cite only commands whose real output you captured. REQUEST_CHANGES entries may cite deliberately failing commands.
9. **Declare enforcement honestly** — no verifier runs here. The repository's pre-commit hook is the sole proof guarantee where installed (`bash scripts/install-git-hooks.sh --force`): it re-executes every proof of a newly added APPROVED/APPROVE_WITH_REQUESTS entry before it can commit and rejects a disallowed `same-session` judge; CI re-runs the same guard core where a checked-in workflow extracts it. Without the hook, declare `enforcement: local-only`; never assert enforcement the repository does not have.
10. **Synchronize plan state**:
    - `APPROVED` or `APPROVE_WITH_REQUESTS` on a non-final phase → set the phase status and Current State lifecycle status to `checked`, and route to `handoff` or `git`. On the final phase after `full`, set the same statuses to `checked` and route to closing `handoff`.
    - `REQUEST_CHANGES` → keep the phase and Current State lifecycle status `in-progress`, record findings as blockers/open items, and route back to `work`. IF `docs/evals/failures.md` exists → append one ledger row per finding.
11. **Verify synchronization** — re-read the phase entry and require the statuses you wrote; confirm Current State agrees with the Validation tail. `check` never marks a phase `done`.
