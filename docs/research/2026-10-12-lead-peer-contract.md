# Lead → Peer task contract (draft v1)

Design note, 2026-10-12. Target: `therealtinhtute/quaqua` (Principal → CoS → Lead → Peers, Auditor). It adds what that model lacks today: a contract that is **fixed before the build** and that the Auditor can **grade line by line** without trusting anyone's word.

**Verdict:** quaqua already has the hard parts: a brief header check (`scripts/worker/prompt-worker.sh --check-brief`), the watchdog, the build ⇄ review pipeline, `Verdict:` lines, telemetry, and auditor pre-checks. Four gaps keep the brief from being a contract:

1. **No executable acceptance.** `Goal:` is prose and `Evidence:` is "what to return", so the worker ends up choosing its own proof. That is the self-grading failure N3/S2 warn about (`docs/research/2026-10-12-agent-sdlc-practice.md`).
2. **No owned-scope field.** `PEER.md` says "owned scope", but the brief template (`rules/_all/brief.md`) has no field for it, so nobody can check scope creep mechanically.
3. **Two handback formats.** There is the 6-field handback in `PEER.md` (written to `/tmp/<proj>-<task>.md`) and the `runs/<name>.md` conventions that the pipeline parses (`Outcome:`, `## Blockers`, `## Challenge`, `## Lessons`, `Verdict:`). The Auditor cannot grade one shape.
4. **Acceptance is implicit.** The Lead "accepts / merges", but no record says which checks it re-ran and on which SHA.

The contract below closes those four gaps and nothing else. It keeps the roles, the lanes, the pipeline flow, and the at-most-4 handoffs (CoS → Lead → builder → reviewer).

## 1. Brief: the Lead writes it before spawning

```text
Task: <slug> · Source: projects/<slug>.md <item or Rn>
Disposition: Engineer | Architect | Reviewer | Scout
Goal: <what must be true at the end>
Acceptance:
  - A1: `<command>` → <expected result>
  - A2: `<command>` → <expected result>
Owned scope: <paths the Peer may write> · Read-only: <paths it may read>
Hard constraints: <non-negotiables>
Current design (open to challenge): <hypothesis, not the answer>
Unknowns: <what to verify first, and how>
Time budget: <N> min · Review: auto|always|skip · Rounds: <n>
```

- Acceptance commands are **fixed by the Lead** before the build. The Peer may challenge them (`## Challenge`), but may not edit or loosen them. Loosening an assertion changes the contract (`PEER.md`, *Contract before test*).
- Each A-id is executable from the worktree root on a bare checkout, and depends on no prior `cd`, alias, or environment variable.
- `--check-brief` exits 1 when `Acceptance:` (at least one A-id with a backticked command) or `Owned scope:` is missing.
- No pre-solve, as before: the brief states what must be true, never the fix.

## 2. Handback: the Peer writes `runs/<name>.md`, one shape for every disposition

```text
Outcome: COMPLETE | REOPEN_REQUEST | DEPENDENCY_REQUEST | BLOCKED
Snapshot: <sha> · <branch> · <worktree>
Scope: <files changed>                      (must be ⊆ Owned scope)
Acceptance:
  - A1: `<same command>` → <real output, ≤3 lines> — pass | fail
  - A2: ...
Verification: <other commands + real output; what was skipped and why>
Unknown / risk: <assumptions, decisions that need the Principal>
## Challenge        (optional: premise + evidence + one alternative)
## Blockers         (required when Outcome ≠ COMPLETE)
## Lessons          (≤2 lines, or "none")
```

- The A-ids and commands are copied **verbatim** from the brief; a missing or renamed A-id makes the handback invalid.
- `Outcome: COMPLETE` requires every A-id to be `pass`. Claiming done when it isn't is worse than `BLOCKED` (`PEER.md`).
- This replaces the `/tmp/<proj>-<task>.md` path in `PEER.md`: one file, the one the pipeline already reads, published atomically (`rules/_all/brief.md`).

## 3. Review: the reviewer writes `runs/<reviewer>.md`

```text
## Spec        (against Goal, Acceptance, Owned scope)
- [blocking|nit] <path:line> — <claim> — <evidence>
## Standards   (repo rules, security, tests that prove behavior)
- [blocking|nit] <path:line> — <claim> — <evidence>
Verdict: clean | <N> findings (<B> blocking)
```

- `clean` means zero blocking findings. Nits never stop the pipeline; they ride the next fix round or are dropped.
- Always blocking: a deleted, skipped, or weakened test; a write outside Owned scope; an A-id that passes without proving the Goal (*if the behavior vanished, would this still pass?*, from `PEER.md`).

## 4. Acceptance record: the Lead appends to `runs/pipeline-<task>.md`

```text
## Acceptance
Accepted: yes | no · by <slug>-lead · at <sha>
- A1: re-ran `<command>` → <real output> — pass
- A2: re-ran `<command>` → <real output> — pass
- Scope: `git diff --name-only <base_sha> <sha>` ⊆ Owned scope — pass
- Review: Verdict clean (round <n>) · Challenge: none | resolved: <decision>
- Merge: <integ sha> into quaqua/<slug>-integ
```

- The Lead **never accepts on the handback's word**: it re-runs every A-id itself on the frozen SHA. The handback SHA is immutable (`PEER.md`, gate 4).
- `Accepted: no` names the failing line and routes back to a fix round or to the CoS.

## 5. Auditor scorecard: deterministic pre-checks, one line per pipeline run

`scripts/auditor/auditor.sh` already pre-checks `runs/pipeline-*.md`. Add these, each printed as `PASS|FAIL <check> — <evidence>`:

| # | Check | How |
|---|---|---|
| C1 | Brief is a contract | `Acceptance:` with ≥1 backticked A-id, and `Owned scope:` present |
| C2 | Handback matches the brief | handback A-ids = brief A-ids; all `pass` when `Outcome: COMPLETE` |
| C3 | Scope held | `git diff --name-only <base_sha> <sha>` ⊆ Owned scope |
| C4 | Independent acceptance | `## Acceptance` exists, re-runs every A-id, and names the same SHA the reviewer saw |
| C5 | Review is honest | `Verdict: clean` ⇒ zero `[blocking]` lines in the reviewer file |
| C6 | Telemetry agrees | the last `runs/telemetry.jsonl` line for the task has the same status as the pipeline file |

A supervisor (the CoS or the Principal) reads the scorecard, not the transcripts: six lines per task, each one checkable.

## Invariants this keeps

- At most 4 handoffs per task. 2026 data: past 4 handoffs, work "almost always failed" (Marmelab survey, via `docs/research/2026-10-12-agent-sdlc-practice.md`).
- One writer per scope; every parallel writer gets its own worktree.
- No worker accepts its own work; one independent reviewer per round. No extra reviewer layers (NLAH: a second reviewer cost 8%).
- Proof is a command anyone can re-run. Prose is testimony.

## Change list for quaqua (if adopted)

| File | Change |
|---|---|
| `rules/_all/brief.md` | add `Task/Source`, `Disposition`, `Acceptance`, `Owned scope`, `Review/Rounds`; keep the no-pre-solve rules |
| `PEER.md` | handback = the §2 shape in `runs/<name>.md`; drop `/tmp/<proj>-<task>.md` |
| `scripts/worker/prompt-worker.sh` | `--check-brief` requires `Acceptance:` + `Owned scope:` |
| `scripts/lead/pipeline.sh` | reviewer prompt asks for the §3 shape; `clean` with a `[blocking]` line → blocked; Lead writes `## Acceptance` before merge |
| `scripts/auditor/auditor.sh` | C1–C6 pre-checks |
| `tests/` | one fixture per check: a pass and a fail case each |
