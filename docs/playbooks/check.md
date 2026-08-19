# Playbook: check

## Purpose

Quality gate for a response-only review, a bounded diff, or a durable phase. Durable `gate` runs automated checks, audits lifecycle alignment, records the verdict in the DB, and synchronizes exact evidence and phase status in the active plan — this is the mode `work` performs itself, in-session, after every phase's waves complete (`work.md` step 11; `work` does not dispatch to the separate `/check` skill for this — see that step for why). `full` includes the gate and adds the complete Security, Performance, Architecture, and Code Quality review; `handoff` requires a clean `full` verdict exactly once, on an initiative's final phase, before closing it (`handoff.md` step 6) — a per-phase `full` would pay a cold prompt-cache model switch every phase for a review most phases don't need (F1 of the SDLC token-cache audit). `review` and bounded/simple modes return evidence in the response only.

## Preconditions and Modes

1. Preserve invocation intent:
   - `gate` — durable automated phase gate for `docs/plans/active/{slug}.md`; it does not perform the complete manual review. `work` performs this itself, in-session, per phase — not via the separate `/check` skill.
   - `full` — durable gate plus the complete Security, Performance, Architecture, and Code Quality review. `work` never performs this; `handoff` requires it exactly once, via the `/check` skill, on the initiative's final phase, before closing it.
   - `review` — response-only review, even when an active plan exists.
   - `bounded` (alias: `simple`) — response-only gate for a direct change with no durable initiative lifecycle.
2. Run `zharness preflight check --mode {gate|full|review|bounded} --json`. Missing binary: print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop. Otherwise check its `version` field — a `dev` build satisfies the gate; below MIN_ZHARNESS_VERSION (`0.8.1` — see `skills/workflow/README.md`), print the same message and stop. Then follow its stop/recovery result exactly.

**Zero-write rule:** review and bounded/simple modes create no lifecycle rows, plans, reports, changesets, or markdown artifacts. They do not call `zharness check record` and do not edit an active plan. Invocation intent wins: discovering an active plan never upgrades `review` or bounded/simple work into a durable gate.

## Owned Plan State

Only durable gate/full mode may:

- append to `## Validation`
- update the selected phase's lifecycle `status` field in `## Phases and Verification` to mirror the DB transition
- update lifecycle status, latest check ID, blockers/open items, and exact next action in `## Current State and Next Action`

Preserve every phase/task definition; the phase lifecycle status is the only mutable field inside a planned phase. Append-only `## Progress` is the sole task execution-status source, so check reads task state there and never adds or updates task-definition status fields.

Every Validation entry must include timestamp, stable phase slug, exact command/result and concise output, run ID, returned check ID, verdict, and proof gaps. Validation is append-only; never replace earlier failed evidence or verdicts.

## Review and Gate Steps

1. **Load scope without changing intent** — read the diff and repository verification instructions. For gate/full, read lifecycle position (`context.position`, `context.phases`, latest run/check/handoff IDs, drift) from the same `preflight check --mode {gate|full} --json` response (Preconditions step 2) instead of a separate `zharness resume --json` call (R6 of the SDLC token-cache audit), also read the active plan, and require a latest run for the phase. Review may consult an active plan for context but remains response-only.
2. **Classify depth and drift** — use quick/standard/deep based on blast radius, not only line count. Label scope on-target, drift, or incomplete before checks. A phase-boundary violation blocks a clean durable verdict.
3. **Run the automated gate** — execute applicable tests, type checks, lint/static analysis, and build in repository-defined order. Capture actual output; never self-certify.
4. **Review plan alignment when applicable** — compare the diff with accepted requirements, Non-goals, phase surfaces, task outputs, append-only Progress entries, and recorded Decisions. Read task execution status only from Progress. Missing planned proof is a finding even when local tests pass. If the optional failure ledger at docs/evals/failures.md exists, read it and, for every failure class recorded two or more times, state explicitly whether the current diff is clean of that class; a repository without that file skips this without it being an error.
5. **Apply mode-specific manual review** — `full` performs the complete Security, Performance, Architecture, and Code Quality review; for a class-of-bug fix, search for sibling instances and state whether coverage is complete. `gate` does not perform that complete manual review. `review` performs the requested response-only review, and bounded/simple performs only scope-appropriate review.
6. **Evaluate required proof** — for `tiny`, require command output; for `normal`, require unit plus command output; for `high-risk`, require unit, integration, manual review, and command output. Name every missing class exactly. A gate does not silently substitute automated checks for required manual-review evidence.
7. **Audit durable lifecycle links** — gate/full runs `zharness audit --json`. Treat pointer drift or contract violations touching the phase as findings; unlinked proof remains explicit context. Review and bounded/simple do not add lifecycle records merely to satisfy this step.
8. **Choose the verdict** — any critical issue or material plan contradiction is `REQUEST_CHANGES`; major non-critical findings are at least `APPROVE_WITH_REQUESTS`; no blocking findings is `APPROVED`. Declare the judge: `same-session` when the reviewing agent also authored the diff under review, `independent` otherwise, alongside the reviewing model's identifier. When the judge is `same-session`, an `APPROVED` or `APPROVE_WITH_REQUESTS` verdict must name at least one aspect that was not independently verified.
9. **Record only a durable gate/full check** — run `zharness check record --verdict {verdict} --run-id {run-id} --judge {independent|same-session} --judge-model {model identifier} --proof-links '[{"command":"{exact command}","output_ref":"Validation entry {timestamp}: {result}"}, ...]' --json`. `--judge` and `--judge-model` are required flags — the CLI rejects a missing or invalid value with `invalid_judge` or `missing_required_field` before recording anything. Do not pass or create an artifact path. Save the returned check ID. For `APPROVED`/`APPROVE_WITH_REQUESTS`, the CLI re-runs every proof command itself and requires exit 0 before recording anything (ceremony audit §15) — cite only commands safe to run a second time (tests, builds, lints), and a `proof_verification_failed` error means one of them doesn't actually pass; fix the proof or change the verdict, don't retry with a different command that wasn't actually run. `REQUEST_CHANGES` proof is never re-executed, so a failing command demonstrating the problem is fine to cite there.
10. **Synchronize durable plan state**:
    - For `APPROVED` or `APPROVE_WITH_REQUESTS`, immediately set the phase status and Current State lifecycle status to `checked`, record the returned check ID, append exact Validation evidence, and route to closing `handoff` or `git`.
    - For `REQUEST_CHANGES`, append the returned check ID and exact failed evidence to Validation, keep the phase and Current State lifecycle status `in-progress` to match the DB, record the findings as blockers/open items, and route back to `work`. When that ledger at docs/evals/failures.md exists, also append one ledger row per finding — durable gate/full only, so `review` and bounded/simple keep the zero-write rule above.
11. **Verify durable synchronization** — gate/full reruns `zharness query phases --json` and requires the plan phase status to match the DB. `check` never marks a phase `done`.

## Response-Only Review and Bounded Gate

Run the narrowest checks that prove the requested change, perform the requested or scope-appropriate review, and return the same evidence/verdict fields in the response. `review` is always response-only: it never calls `zharness check record` and never updates the plan, even if an active plan exists. Bounded/simple follows the same zero-write rule because it has no run row.

## Command Reference

- `zharness preflight check --mode {gate|full|review|bounded} --json` (step 1 — gate/full's `context` field replaces a separate `zharness resume --json` call)
- `zharness audit --json`
- `zharness query phases --json`
- `zharness check record --verdict {verdict} --run-id {run-id} --judge {independent|same-session} --judge-model {model} --proof-links '[...]' --json`

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
check_id: ULID | not-recorded
proof_gaps: none | exact missing classes
```

## Exit Conditions

- Gate: automated checks, plan alignment, required-proof evaluation, and lifecycle audit ran; `judge` and `judge_model` are both present in the output block, and a `same-session` judge named what it did not independently verify; the DB check row was recorded; exact evidence and verdict were appended to Validation; and plan/DB phase statuses match (`checked` for a clean verdict, `in-progress` for `REQUEST_CHANGES`). The complete manual review is not part of gate.
- Full: every gate condition holds and the complete Security, Performance, Architecture, and Code Quality review ran.
- Review or bounded/simple: the response contains honest proof and verdict with zero DB, changeset, plan, report, or markdown writes.
