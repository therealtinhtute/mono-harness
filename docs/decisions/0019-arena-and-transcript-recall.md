# 0019 — Implementation arena and transcript recall: defer both, with limits fixed now

## Status

Accepted 2026-10-24 on merge of the PR that resolves issue #157. Decisions: implementation arena deferred (no new capability); transcript recall deferred (limits fixed now so a later build is bounded).

## Context

The retro backlog lists two candidates (`docs/research/2026-10-22-pstack-retro-5.md`, sections 2 and 3). Current state, checked against the files:

- `skills/reason/arena/SKILL.md` already exists. It is a design bakeoff: sub-agents build one artifact several ways, a rubric picks a base, and candidates live in scratch paths outside the repo (`SKILL.md:20`). It does not run implementation attempts in worktrees.
- `work-fanout.md` already defines a **race**: N workers on one identical brief, a rule declared before spawning (`first pass`, `rank all` or `best-of`), one worktree each, and rejection of any report that does not name the tested source and its acceptance output (`work-fanout.md:8`, `:56`). The missing piece is a tie-break for `best-of` on implementation attempts.
- Recall: `watzup` is read-only over Git and the active plan (`skills/stages/watzup/references/watzup.md`). `retro` already reads transcripts through `skills/stages/retro/scripts/session-digest.py` (`:6`). The lost-decision failure class exists in the log taxonomy (`work-full.md:21`), but the repository holds no recorded instance of it.

## Decision

**1. Implementation arena: defer; no new skill or flag.** The race shape already does the job and `arena` already covers open design. No ledger row shows a task where two attempts would have beaten one. If a task truly has an open implementation choice, the owner can use `race` with `best-of` today. When the first such task appears and the missing tie-break blocks it, add one line to the Aggregate section of `work-fanout.md`. The rule is fixed here so it is not reinvented:

- Eligible: a task whose plan marks the implementation choice open. Not eligible: a task with one obvious approach.
- Attempts: 2 to 3, each in its own worktree, each with the identical brief and the acceptance command in `VERIFY`.
- Budget: the owner names a token budget; the ceiling is N times one attempt. Over budget stops the remaining attempts.
- Winner: only attempts whose report names the tested source and whose acceptance command passed on that source, re-run by the lead, are candidates. Among them: smaller diff by changed lines, then fewer files touched, then the lead's reasoned pick recorded as a `decision`. No candidate left means `blocked`, not a pick.
- Cleanup: removing the losing worktrees and branches is a delete and needs the owner's approval (`work-fanout.md:67`).

**2. Transcript recall: defer; do not split an issue yet.** Reason: no recorded lost decision that `handoff` or the plan Log failed to carry. Rejected now: giving `watzup` a default transcript read, because it would add a read outside the repository to a stage that is read-only by contract and make the transcript look authoritative against `AGENTS.md` ("the repository is the system of record"). IF a `LOST_DECISION` block is recorded, open one bounded issue with these limits:

- Opt-in: only on an explicit `watzup recall` argument; never default, never from a hook.
- Scope: `~/.claude/projects/<cwd-slug>/` only (the path `session-digest.py` already resolves), the newest 3 sessions, read-only.
- Output: at most 40 lines, each labeled `lead (unverified)` with the session ID; secrets and tokens redacted.
- Authority: a lead is checked against repository state and the plan before use. It is never written to the plan, the Log or Validation as evidence.
- Privacy: no transcript text leaves the machine, including to a search or summarizing service.
- Location: reuse `session-digest.py`; no new skill and no index of transcripts.

## Consequences

No code, flag or skill ships, so nothing here changes behavior. The cost of deferring is that a hard design choice still needs the owner to run a manual race, and a lost decision across sessions is recovered by hand. The issue's criteria are resolved by this record: arena eligibility, isolation and winner rule are set in decision 1, recall limits and labeling in decision 2, and each capability has a verdict.

## Authority

`docs/research/2026-10-22-pstack-retro-5.md`; `skills/reason/arena/SKILL.md:20`; `skills/stages/work/references/work-fanout.md:8`, `:56`, `:67`; `skills/stages/work/references/work-full.md:21`; `skills/stages/watzup/references/watzup.md`; `skills/stages/retro/scripts/session-digest.py:6`; `AGENTS.md` Harness section; issue #157.
