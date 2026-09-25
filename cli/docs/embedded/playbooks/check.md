# Playbook: check

## Purpose

Quality gate for a direct change or a durable phase. Durable `gate` runs automated checks and records the verdict in the active plan's `## Validation`; `work` performs it after every phase (`work-full.md` step 8). `full` includes the gate and adds the complete Security, Performance, Architecture, and Code Quality review, required exactly once, on the initiative's final phase, before `handoff` closes it. `bounded` returns evidence in the response only.

## Preconditions and Modes

1. Preserve invocation intent:
   - `auto` (default) — a direct change or review outside a durable initiative → resolve to `bounded` without reading any plan. A request that names or continues a durable initiative → step 2, then `gate`; invalid initiative state never falls back to `bounded`. Ambiguous → ask. `auto` never resolves to `full`, and an active plan's mere existence never selects `gate`.
   - `gate` — durable automated phase gate; no complete manual review.
   - `full` — gate plus the complete manual review; `work` never performs it.
   - `bounded` (aliases: `simple`, `review`) — response-only check or review of a direct change, even when an active plan exists.
2. **Read-only preflight** (initiative-intent `auto`, explicit `gate`/`full`) — count every non-empty markdown plan under `docs/plans/active/` before matching the requested initiative. Require exactly one active plan in total, matching the request; IF multiple → list every candidate and stop even when only one matches the request. Read only its selected phase and `## Current State and Next Action`, re-reading after any compaction; require the phase to read `in-progress` and Current State to agree. A missing or ambiguous phase, any other status (including unstarted, `checked`, or `done`), or a conflict → stop before checks or writes and report the exact mismatch. Never change phase status to pass preflight.
3. After the preflight succeeds, print the resolved mode before running checks or writing state (after any required prefix, same line): `mode: {resolved} ({one-line reason})`. A failed preflight reports its blocker instead.

**Zero-write rule:** `bounded` is always response-only and never appends to Validation or edits the plan. Invocation intent wins: an active plan never upgrades `bounded` into a durable gate.

## Owned Plan State

Only durable gate/full may: append to `## Validation`; update the selected phase's lifecycle `status:` line in `## Phases and Verification`; update lifecycle status, blockers/open items, and exact next action in `## Current State and Next Action`. `## Log` is the sole task execution-status source: read task state there; never add task-definition status fields.

## Steps

1. **Load scope without changing intent** — read the diff and repository verification instructions. Gate/full: also read the phase entry and the tails of `## Log` and `## Validation`. `bounded` may consult a plan for context but stays response-only.
2. **Classify depth and drift** — quick/standard/deep by blast radius, not only line count. Label scope on-target, drift, or incomplete. A phase-boundary violation blocks a clean durable verdict.
3. **Run the automated gate** — applicable tests, type checks, lint/static analysis, and build, in repository-defined order. IF `scripts/record-check.sh` exists → `bash scripts/record-check.sh -- "cmd1" "cmd2" …`. Else capture each command's output to a temp file and preserve its exit code (`rc=$?`; never pipe into `head`/`tail` in a way that clobbers `rc`). Validation bullets cite the raw commands, not the wrapper.
4. **Review plan alignment** (gate/full) — compare the diff with the Goal's requirements and non-goals, phase surfaces, task checks, and Log `decision` entries. Missing planned proof is a finding even when local tests pass. For each requirement the phase `goal:` names, judge its `acceptance:` (none → judge the requirement text) and record it in the entry's `requirements:` line as `met`, `not met`, or `partial (→ <phase>)` when a later phase's `goal:` also names it. `not met` is a material plan contradiction; `partial` is not a finding. `full` on the final phase also judges the Goal's `success_signal:`. Record `rollback_point:` as the commit the phase can be reverted to. IF `docs/evals/failures.md` exists → for every failure class recorded two or more times, state whether the diff is clean of it.
5. **Manual review** — `full`: the complete Security, Performance, Architecture, and Code Quality review, run as the two axes below; for a class-of-bug fix, search sibling instances and state whether coverage is complete. `gate` does not perform that complete manual review. `bounded`: the requested or scope-appropriate review, on the same two axes when it reviews a diff.
   - **Two axes, kept apart** — default: this agent runs them one after the other on the diff (`git diff <base>...HEAD`), finishing one report before starting the next. Sub-agents are opt-in, for `full` only: IF the request asks for sub-agents or parallel review → use them; IF it asks for a single agent → stay single; IF it says neither → before step 5, ask once with the structured-question tool (single agent (Recommended) / two parallel sub-agents) and follow the answer. Each sub-agent gets the diff command and only its own axis's sources. `bounded` never spawns sub-agents unless the request names them. Sub-agents keep the axes apart; they do not change the `judge:` declaration.
   - **Spec axis** — sources: the Goal's requirements and non-goals, and the phase `goal:`. Report requirements missing or partial, behavior nobody asked for (scope creep), and requirements that look implemented but wrong; cite the requirement ID for each.
   - **Standards axis** — sources: the repository's documented standards (AGENTS.md, docs it points to), plus Security, Performance, Architecture, Code Quality, and this baseline. Documented repo standards win over the baseline; skip whatever tooling already enforces. Each baseline item is a judgement label ("possible Feature Envy"), never a hard violation:
     - Smells: Mysterious Name, Duplicated Code, Feature Envy, Data Clumps, Primitive Obsession, Repeated Switches, Shotgun Surgery, Divergent Change, Speculative Generality, Message Chains, Middle Man, Refused Bequest.
     - Tests: implementation-coupled (mocks internals, asserts through a side channel), tautological (recomputes the expected value the way the code does), horizontal (tests written in bulk ahead of the code they test).
   - Report under `Spec` and `Standards` headings without merging or reranking across them; step 7 weighs both.
6. **Evaluate required proof** — `tiny`: command output; `normal`: unit plus command output; `high-risk`: unit, integration, manual review, and command output. Name every missing class exactly.
7. **Choose the verdict** — any critical issue or material plan contradiction → `REQUEST_CHANGES`; major non-critical findings → at least `APPROVE_WITH_REQUESTS`; no blocking findings → `APPROVED`. Declare the judge (`same-session` IF the reviewer authored the diff, else `independent`) and the reviewing model identifier. A `same-session` `APPROVED`/`APPROVE_WITH_REQUESTS` must name at least one aspect not independently verified. A `full` entry, and any entry on a `high-risk` lane, must declare `judge: independent`.
8. **Append durable evidence** — read `docs/playbooks/check-validation.md`, then write the Validation entry by hand in that format. Never self-certify: cite only commands whose real output you captured. REQUEST_CHANGES entries may cite deliberately failing commands.
9. **Declare enforcement honestly** — no verifier runs here. The repository's pre-commit hook is the sole proof guarantee where installed (`bash scripts/install-git-hooks.sh --force`): it re-executes every proof of a newly added APPROVED/APPROVE_WITH_REQUESTS entry before it can commit and rejects a disallowed `same-session` judge; CI re-runs the same guard core where a checked-in workflow extracts it. Without the hook, declare `enforcement: local-only`; never assert enforcement the repository does not have.
10. **Synchronize plan state**:
    - `APPROVED` or `APPROVE_WITH_REQUESTS` on a non-final phase → set the phase status and Current State lifecycle status to `checked`, and route to `handoff` or `git`. On the final phase after `full`, route to closing `handoff`.
    - `REQUEST_CHANGES` → keep the phase and Current State lifecycle status `in-progress`, record findings as blockers/open items, and route back to `work`. IF `docs/evals/failures.md` exists → append one ledger row per finding.
11. **Verify synchronization** — re-read the phase entry and require the statuses you wrote; confirm Current State agrees with the Validation tail. `check` never marks a phase `done`.

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

## What the Guards Cannot Check

Name any that applies in `proof_gaps:`: `judge: independent` is testimony, not proof; an unparsed (misspelled, buried, or unanchored) verdict is ignored, not rejected, so guard silence can mean "unparsed"; a range guard compares endpoints, so an entry added and removed inside one push is out of scope; a `requirements:` line is testimony, and the proof it cites re-runs only when it is also a proof sub-bullet.

## Exit Conditions

- Gate: steps 1–4 and 6–11 ran; the Validation entry landed with judge and model and, IF `same-session`, what was not independently verified; plan statuses match what you wrote.
- Full: every gate condition, plus step 5's complete review and `judge: independent`.
- Bounded: honest proof and verdict in the response; zero plan, report, or markdown writes.
