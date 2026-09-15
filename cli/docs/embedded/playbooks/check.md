# Playbook: check

## Purpose

Quality gate for a response-only review, a bounded diff, or a durable phase. Durable `gate` runs automated checks and records the verdict in the active plan's `## Validation`; `work` performs it in-session after every phase (`work-full.md` step 11). `full` includes the gate and adds the complete Security, Performance, Architecture, and Code Quality review, required exactly once, on the initiative's final phase, before `handoff` closes it (`handoff.md` step 6). `review` and bounded/simple return evidence in the response only.

## Preconditions and Modes

1. Preserve invocation intent:
   - `auto` (default) — classify initiative intent from the request and conversation first. Direct change outside a durable initiative → resolve to `bounded` without reading any plan. Request names or continues a durable initiative → validate its plan and phase in step 2, then resolve to `gate`; invalid initiative state never falls back to `bounded`. Ambiguous intent or phase → stop and ask for the selection. `auto` never resolves to `full`; request `full` by name. The mere existence of an active plan never selects `gate`.
   - `gate` — durable automated phase gate; no complete manual review.
   - `full` — gate plus the complete manual review; `work` never performs it.
   - `review` — response-only review, even when an active plan exists.
   - `bounded` (alias: `simple`) — response-only gate for a direct change with no durable lifecycle.
2. **Read-only preflight** (initiative-intent `auto`, explicit `gate`/`full`) — count every non-empty markdown plan under `docs/plans/active/` before matching the requested initiative. Require exactly one active plan in total; IF multiple → list every candidate and stop even when only one matches the request. The sole plan must match the request. Read only its selected phase and `## Current State and Next Action`; require the phase to read `in-progress` and Current State to agree. IF a summary or compaction happened since the last read → re-read those sections before resolving the mode. IF the plan is missing/mismatched, candidates are multiple, the phase is absent, ambiguous, or in any other status (including unstarted, `checked`, or `done`), or Current State conflicts → stop before checks or writes; report the exact mismatch and the required planning, phase-start, selection, reconciliation, or closing step. Never change phase status to pass preflight. Explicit `review` and `bounded`/`simple` bypass this durable preflight.
3. After the read-only preflight succeeds, print the resolved mode before running checks or writing state (after any required prefix, same line): `mode: {resolved} ({one-line reason})`. Direct bounded requests need no plan read first. A failed preflight reports its blocker instead of claiming a resolved mode.

**Zero-write rule:** `review` is always response-only, and bounded/simple matches it: no plans, reports, or markdown artifacts; never appends to Validation or edits the plan. Invocation intent wins: an active plan never upgrades `review` or bounded/simple into a durable gate. These modes run the narrowest checks that prove the change and return the Output Format fields in the response.

## Owned Plan State

Only durable gate/full may: append to `## Validation`; update the selected phase's lifecycle `status` in `## Phases and Verification`; update lifecycle status, latest anchors, blockers/open items, and exact next action in `## Current State and Next Action`. Phase lifecycle status is the only mutable field in a planned phase. Append-only `## Progress` is the sole task execution-status source: read task state there; never add or update task-definition status fields.

## Review and Gate Steps

1. **Load scope without changing intent** — read the diff and repository verification instructions. Gate/full: also read the phase's Phases entry and recent Progress/Validation tails. `review` may consult a plan for context but stays response-only.
2. **Classify depth and drift** — quick/standard/deep by blast radius, not only line count. Label scope on-target, drift, or incomplete before checks. A phase-boundary violation blocks a clean durable verdict.
3. **Run the automated gate** — applicable tests, type checks, lint/static analysis, and build, in repository-defined order. IF `scripts/record-check.sh` exists → `bash scripts/record-check.sh -- "cmd1" "cmd2" …` (timeout → gtimeout → unbounded; keeps the exit code; 3-line pass / 10-line fail tail). Else capture each command's output to a temp file, print the same tails, and preserve its exit code (`rc=$?`; never pipe into `head`/`tail` in a way that clobbers `rc`). Validation bullets cite the raw commands, not the wrapper.
4. **Review plan alignment** (when applicable) — compare the diff with accepted requirements, Non-goals, phase surfaces, task outputs, Progress entries, and Decisions. Missing planned proof is a finding even when local tests pass. IF `docs/evals/failures.md` exists → for every failure class recorded two or more times, state whether the diff is clean of it; an absent file is not an error.
5. **Apply mode-specific manual review** — `full`: the complete Security, Performance, Architecture, and Code Quality review; for a class-of-bug fix, search sibling instances and state whether coverage is complete. `gate` does not perform that complete manual review. `review`: the requested review. Bounded/simple: scope-appropriate review only.
6. **Evaluate required proof** — `tiny`: command output; `normal`: unit plus command output; `high-risk`: unit, integration, manual review, and command output. Name every missing class exactly. Never substitute automated checks for required manual-review evidence.
7. **Choose the verdict** — any critical issue or material plan contradiction → `REQUEST_CHANGES`; major non-critical findings → at least `APPROVE_WITH_REQUESTS`; no blocking findings → `APPROVED`. Declare the judge (`same-session` IF the reviewer authored the diff, else `independent`) and the reviewing model identifier. A `same-session` `APPROVED`/`APPROVE_WITH_REQUESTS` must name at least one aspect not independently verified. A `full` entry must declare `judge: independent`.
8. **Append durable evidence first (mandatory)** — read `docs/playbooks/check-validation.md`, then write the Validation entry into the plan by hand in that format. Never self-certify: cite only commands whose real output you captured, as nested sub-bullets exactly as run. REQUEST_CHANGES entries may cite deliberately failing commands.
9. **Declare enforcement honestly** — no verifier runs here. The repository's pre-commit hook is the sole proof guarantee where installed (`bash scripts/install-git-hooks.sh --force`): it parses the staged Validation entry, re-executes every nested proof before an APPROVED/APPROVE_WITH_REQUESTS verdict can commit, and rejects a newly added `mode: full` entry declaring `same-session`; CI re-runs the same guard core where a checked-in workflow extracts it. Without the hook the level is `Optional hook` on the ladder in `docs/patterns/encoding-invariants.md`: honor these rules as authoring discipline and declare that level; never assert enforcement the repository does not have.
10. **Synchronize durable plan state**:
    - `APPROVED` or `APPROVE_WITH_REQUESTS` → immediately set the phase status and Current State lifecycle status to `checked`, complete the entry's evidence and `receipt:` block, and route to closing `handoff` or `git`.
    - `REQUEST_CHANGES` → keep the phase and Current State lifecycle status `in-progress`, record findings as blockers/open items, and route back to `work`. IF `docs/evals/failures.md` exists → append one ledger row per finding (durable gate/full only).
11. **Verify durable synchronization** — re-read the Phases entry and require the statuses you wrote; confirm Current State agrees with the Validation tail. `check` never marks a phase `done`.

## Output Format

End the response with:

```text
mode: gate | full | review | bounded
scope: on target | drift | incomplete
depth: quick | standard | deep
gate: pass | fail
review: APPROVED | APPROVE_WITH_REQUESTS | REQUEST_CHANGES
judge: independent | same-session
judge_model: {model identifier}
blockers: N critical, N major
verification: exact command -> pass | fail | not-run
receipt: context_sources / policy / judge / judge_model / retries / rollback_point / failure_ledger: absent|{path} / enforcement: hook | ci | local-only / not_independently_verified
proof_gaps: none | exact missing classes
```

## What the Guards Cannot Check

A passing guard is not evidence of these; name any that applies in `proof_gaps:`.

- **`judge: independent` is testimony, not proof.** The guard rejects `same-session` on a high-risk lane or in `full` mode, but nothing establishes that an `independent` claim is true.
- **Unparsed entries are ignored, not rejected.** A misspelled, buried, or novel verdict token means no verdict and no proof re-execution; guard silence can mean "clean" or "unparsed".
- **A range guard compares endpoints.** An entry added and removed inside one push or pull request, or history rewritten before the base, is out of scope.

## Exit Conditions

- Gate: steps 1-4 and 6-11 ran; the Validation entry landed with judge/model, a `receipt:` block, and, IF `same-session`, what was not independently verified; plan statuses match what you wrote (`checked` clean, `in-progress` for `REQUEST_CHANGES`).
- Full: every gate condition, plus step 5's complete review and `judge: independent`.
- Review or bounded/simple: honest proof and verdict in the response; zero plan, report, or markdown writes.
