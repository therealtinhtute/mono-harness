# Playbook: check

## Purpose

Quality gate for a direct change or a durable phase. Durable `gate` runs automated checks and records the verdict in the active plan's `## Validation`; `work` performs it after every phase (`work-full.md` step 8). `full` includes the gate and adds the complete Security, Performance, Architecture, and Code Quality review, required exactly once, on the initiative's final phase, before `handoff` closes it. `bounded` returns evidence in the response only.

## Preconditions and Modes

1. Preserve invocation intent:
   - `auto` (default) — a direct change or review outside a durable initiative → resolve to `bounded` without reading any plan. A request that names or continues a durable initiative → step 2, then `gate`; invalid initiative state never falls back to `bounded`. Ambiguous → ask. `auto` never resolves to `full`, and an active plan's mere existence never selects `gate`.
   - `gate` — durable automated phase gate; no complete manual review.
   - `full` — gate plus the complete manual review; `work` never performs it.
   - `bounded` (aliases: `simple`, `review`) — response-only check or review of a direct change, even when an active plan exists.
2. **Read-only preflight** (initiative-intent `auto`, explicit `gate`/`full`) — read `docs/playbooks/check-durable.md` now; it holds the preflight, Owned Plan State, and steps 4 and 8–11.
3. After the preflight succeeds, print the resolved mode before running checks or writing state (after any required prefix, same line): `mode: {resolved} ({one-line reason})`. A failed preflight reports its blocker instead.

**Zero-write rule:** `bounded` is always response-only and never appends to Validation or edits the plan. Invocation intent wins: an active plan never upgrades `bounded` into a durable gate.

## Steps

1. **Load scope without changing intent** — read the diff and repository verification instructions. Gate/full: also read the phase entry and the tails of `## Log` and `## Validation`. `bounded` may consult a plan for context but stays response-only.
2. **Classify depth and drift** — quick/standard/deep by blast radius, not only line count. Label scope on-target, drift, or incomplete. A phase-boundary violation blocks a clean durable verdict.
3. **Run the automated gate** — applicable tests, type checks, lint/static analysis, and build, in repository-defined order. IF `scripts/record-check.sh` exists → `bash scripts/record-check.sh -- "cmd1" "cmd2" …`. Else capture each command's output to a temp file and preserve its exit code (`rc=$?`; never pipe into `head`/`tail` in a way that clobbers `rc`). Validation bullets cite the raw commands, not the wrapper.
4. **Review plan alignment** (gate/full) — `check-durable.md` step 4.
5. **Manual review** — `full`: the complete Security, Performance, Architecture, and Code Quality review; for a class-of-bug fix, search sibling instances and state whether coverage is complete. `gate` does not perform that complete manual review. `bounded`: the requested or scope-appropriate review. `full`, or `bounded` on a diff: read `docs/playbooks/check-review.md` now.
6. **Evaluate required proof** — `tiny`: command output; `normal`: unit plus command output; `high-risk`: unit, integration, manual review, and command output. Name every missing class exactly.
7. **Choose the verdict** — any critical issue or material plan contradiction → `REQUEST_CHANGES`; major non-critical findings → at least `APPROVE_WITH_REQUESTS`; no blocking findings → `APPROVED`. Declare the judge (`same-session` IF the reviewer authored the diff, else `independent`) and the reviewing model identifier. A `same-session` `APPROVED`/`APPROVE_WITH_REQUESTS` must name at least one aspect not independently verified. A `full` entry, and any entry on a `high-risk` lane, must declare `judge: independent`.
8. **Evidence and plan state** (gate/full) — `check-durable.md` steps 8–11.

## Output Format

End the response with:

```text
mode: gate | full | bounded
depth: quick | standard | deep
scope: on target | drift | incomplete
gate: pass | fail
verdict: APPROVED | APPROVE_WITH_REQUESTS | REQUEST_CHANGES
judge: independent | same-session (judge_model: {model identifier})
enforcement: hook | ci | local-only
verification: exact command -> pass | fail | not-run
proof_gaps: none | exact missing classes
```

## Exit Conditions

- Gate: steps 1–4 and 6–11 ran; the Validation entry landed with judge and model and, IF `same-session`, what was not independently verified; plan statuses match what you wrote.
- Full: every gate condition, plus step 5's complete review and `judge: independent`.
- Bounded: honest proof and verdict in the response; zero plan, report, or markdown writes.
