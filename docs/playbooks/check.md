# Playbook: check

## Purpose

Quality gate for a response-only review, a bounded diff, or a durable phase. Durable `gate` runs automated checks and records the verdict in the active plan's Validation section — this is the mode `work` performs itself, in-session, after every phase's waves complete (`work.md` step 11). `full` includes the gate and adds the complete Security, Performance, Architecture, and Code Quality review; `handoff` requires a clean `full` verdict exactly once, on an initiative's final phase, before closing it (`handoff.md` step 6) — a per-phase `full` would pay a cold prompt-cache model switch every phase for a review most phases don't need (F1 of the SDLC token-cache audit). `review` and bounded/simple modes return evidence in the response only.

## Preconditions and Modes

1. Preserve invocation intent:
   - `gate` — durable automated phase gate for `docs/plans/active/{slug}.md`; it does not perform the complete manual review. `work` performs this itself, in-session, per phase.
   - `full` — durable gate plus the complete Security, Performance, Architecture, and Code Quality review. `work` never performs this; it belongs to the initiative's final phase via handoff closure. A `full` Validation entry must declare `judge: independent`.
   - `review` — response-only review, even when an active plan exists.
   - `bounded` (alias: `simple`) — response-only gate for a direct change with no durable initiative lifecycle.
2. For durable gate/full, require exactly one non-empty active plan whose selected phase reads `in-progress`; a summary or compaction since the last read invalidates earlier-read anchors until the plan is re-read.

**Zero-write rule:** review and bounded/simple modes create no plans, reports, or markdown artifacts. They do not append to Validation and do not edit an active plan. Invocation intent wins: discovering an active plan never upgrades `review` or bounded/simple work into a durable gate.

## Owned Plan State

Only durable gate/full mode may:

- append to `## Validation`
- update the selected phase's lifecycle `status` field in `## Phases and Verification`
- update lifecycle status, latest anchors, blockers/open items, and exact next action in `## Current State and Next Action`

Preserve every phase/task definition; the phase lifecycle status is the only mutable field inside a planned phase. Append-only `## Progress` is the sole task execution-status source, so check reads task state there and never adds or updates task-definition status fields.

Every Validation entry must include timestamp, stable phase slug, exact command/result and concise output, verdict, judge declaration, reviewing model identifier, proof gaps, and a grep-able `receipt:` block (`context_sources`, `policy`, `judge`, `judge_model`, `retries`, `rollback_point`, `failure_ledger: absent|{path}`, `not_independently_verified`) — with every proof command written as a nested sub-bullet under the entry so the commit-time guard can find them. Validation is append-only; never replace earlier failed evidence or verdicts.

## Review and Gate Steps

1. **Load scope without changing intent** — read the diff and repository verification instructions. For gate/full, also read the active plan's Phases entry for the phase and its recent Progress/Validation tails; require a started phase (`in-progress`). Review may consult an active plan for context but remains response-only.
2. **Classify depth and drift** — use quick/standard/deep based on blast radius, not only line count. Label scope on-target, drift, or incomplete before checks. A phase-boundary violation blocks a clean durable verdict.
3. **Run the automated gate** — execute applicable tests, type checks, lint/static analysis, and build in repository-defined order. If `scripts/record-check.sh` exists, run them via `bash scripts/record-check.sh -- "cmd1" "cmd2" …` (timeout → gtimeout → unbounded; keeps the proof's exit code; 3-line pass tail / 10-line fail tail). If it is absent, run each command with output captured to a temp file, print the same tails, and preserve the command's exit code (`rc=$?`; do not pipe into `head`/`tail` in a way that clobbers `rc`). Never self-certify. Nested Validation bullets cite the raw commands, not the wrapper, so the hook re-executes them.
4. **Review plan alignment when applicable** — compare the diff with accepted requirements, Non-goals, phase surfaces, task outputs, append-only Progress entries, and recorded Decisions. Read task execution status only from Progress. Missing planned proof is a finding even when local tests pass. If the optional failure ledger at docs/evals/failures.md exists, read it and, for every failure class recorded two or more times, state explicitly whether the current diff is clean of that class; a repository without that file skips this without it being an error.
5. **Apply mode-specific manual review** — `full` performs the complete Security, Performance, Architecture, and Code Quality review; for a class-of-bug fix, search for sibling instances and state whether coverage is complete. `gate` does not perform that complete manual review. `review` performs the requested response-only review, and bounded/simple performs only scope-appropriate review.
6. **Evaluate required proof** — for `tiny`, require command output; for `normal`, require unit plus command output; for `high-risk`, require unit, integration, manual review, and command output. Name every missing class exactly. A gate does not silently substitute automated checks for required manual-review evidence.
7. **Choose the verdict** — any critical issue or material plan contradiction is `REQUEST_CHANGES`; major non-critical findings are at least `APPROVE_WITH_REQUESTS`; no blocking findings is `APPROVED`. Declare the judge: `same-session` when the reviewing agent also authored the diff under review, `independent` otherwise, alongside the reviewing model's identifier. When the judge is `same-session`, an `APPROVED` or `APPROVE_WITH_REQUESTS` verdict must name at least one aspect that was not independently verified. A `full` entry must declare `judge: independent`; the hook rejects a newly added `mode: full` entry that also declares `same-session`.
8. **Append durable evidence first (mandatory)** — write the Validation entry into the active plan by hand: timestamp, phase slug, exact commands and results, verdict, judge, judge model, proof gaps, and nested sub-bullets carrying each proof command exactly as run. Never self-certify; cite only commands whose real output you captured. `REQUEST_CHANGES` entries may cite failing commands on purpose.
9. **Commit-time proof guarantee** — no verifier runs here. The repository's pre-commit hook is the sole proof guarantee: it parses this Validation entry from staged bytes and re-executes every nested proof command itself before an APPROVED/APPROVE_WITH_REQUESTS verdict can be committed (`scripts/install-git-hooks.sh`; CI re-runs it). Cite only commands whose real output you captured, written as nested sub-bullets exactly as run — REQUEST_CHANGES entries may cite deliberately failing commands.
10. **Synchronize durable plan state**:
    - For `APPROVED` or `APPROVE_WITH_REQUESTS`, immediately set the phase status and Current State lifecycle status to `checked`, complete the Validation entry's exact evidence and `receipt:` block, and route to closing `handoff` or `git`.
    - For `REQUEST_CHANGES`, keep the phase and Current State lifecycle status `in-progress`, record the findings as blockers/open items, and route back to `work`. When that ledger at docs/evals/failures.md exists, also append one ledger row per finding — durable gate/full only.
11. **Verify durable synchronization** — re-read the plan's Phases entry and require the statuses you wrote to appear there; catch partial writes by confirming Current State agrees with the Validation tail. `check` never marks a phase `done`.

## Response-Only Review and Bounded Gate

Run the narrowest checks that prove the requested change, perform the requested or scope-appropriate review, and return the same evidence/verdict fields in the response. `review` is always response-only: it never appends to Validation and never updates the plan, even if an active plan exists. Bounded/simple follows the same zero-write rule.

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
receipt: context_sources / policy / judge / judge_model / retries / rollback_point / failure_ledger: absent|{path} / not_independently_verified
proof_gaps: none | exact missing classes
```

## Exit Conditions

- Gate: automated checks, plan alignment, required-proof evaluation ran; the Validation entry landed with judge/model declared, a `receipt:` block, and a `same-session` judge naming what it did not independently verify; and plan statuses match what you wrote (`checked` for a clean verdict, `in-progress` for `REQUEST_CHANGES`). The complete manual review is not part of gate.
- Full: every gate condition holds, the complete Security, Performance, Architecture, and Code Quality review ran, and the judge is `independent`.
- Review or bounded/simple: the response contains honest proof and verdict with zero plan, report, or markdown writes.
