# pstack follow-ups: multi-model review, arena, recall (retro #5)

Backlog note, 2026-10-22. Source: retro finding #5 of the pstack-lessons session ([pstack-concepts-audit.md](pstack-concepts-audit.md)). Status: **not locked**. The next step is `brainstorm` on these three items; none of them is a plan yet.

Severity S2. Fix mode: human, because the owner must choose the models and the token budget before any design.

## Why

- The owner asked for all three items during the pstack review.
- [check-review.md](../../skills/stages/check/references/check-review.md) uses a single model. Sub-agents split the review by axis (spec, standards), not by model.
- [work-full.md](../../skills/stages/work/references/work-full.md) step 4 runs independent tasks in parallel worktrees and rejects a handback that lacks a SHA and proof output. It has no way to run competing attempts at the same task and pick one.
- On resume, `watzup` reads git and the active plan. It does not read the previous session's transcript.

## 1. Multi-model review (pstack `interrogate`)

- **pstack:** `interrogate` Steps 1–4. One reviewer per model runs on the same diff. The merge step ranks findings that 2+ models report highest. `thermo-nuclear-review` audits the diff before it reads bot comments.
- **mono-harness today:** `check` step 5 has two axes with one judge. The `judge:` declaration in step 7 is unchanged by sub-agents.
- **Candidate shape:** an opt-in review mode in `check` for `full` or high-risk. It runs N model runners on the same `git diff <base>...HEAD`, merges by agreement, and keeps the Spec and Standards headings.
- **Cost:** N× review tokens per run.

## 2. Competing attempts (pstack `swarm` / `arena`)

- **pstack:** `swarm` Phase B writes self-contained briefs (goal, scope, verify, report as `PASS|ISSUES|BLOCKED` with SHA). Phase C drops reports with no SHA. `arena` runs several attempts at one task and synthesizes or picks the winner.
- **mono-harness today:** the brief and the handback rejection already exist (`work-full.md` step 4). The missing part is arena: N worktrees on the *same* task, then a comparison by the task's acceptance check.
- **Candidate shape:** an `arena` option in `work-full` for a task whose design is open. Each attempt gets its own worktree. Winners are picked by the acceptance command output plus diff size, and losing worktrees are removed.
- **Cost:** N× build tokens plus worktree disk.

## 3. Transcript recall (pstack `recall`)

- **pstack:** `recall` searches past session transcripts so a resumed session can recover decisions not written to the repo.
- **mono-harness today:** `retro` already parses transcripts with [session-digest.py](../../skills/stages/retro/scripts/session-digest.py). The repository is the system of record, so a recall that gives transcripts authority would conflict with the Harness rules in `AGENTS.md`.
- **Candidate shape:** a read-only lookup in `watzup` that reuses `session-digest.py`. Its output is a lead to verify against the repo, never evidence.
- **Cost:** low. It is a read of local transcript files.

## Open decisions for `brainstorm`

1. Which models run the reviewers and arena attempts, and which runner calls them (Claude sub-agents only, or other CLIs as well)?
2. What is the token budget per review and per arena run, and which lane turns each feature on (opt-in, `full` only, high-risk only)?
3. Do the three items ship as one plan or as three separate bounded changes?
4. Does recall live in `watzup`, in `retro`, or as a new `tools/` skill?

## Non-goals

These items were rejected in [pstack-concepts-audit.md](pstack-concepts-audit.md):

- Chief of Staff.
- Event-driven orchestration.
- A task database.
- A SessionStart hook.
