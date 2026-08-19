# Playbook: watzup

## Purpose

Read-only session recap. Combine Git state, DB lifecycle position, and the current initiative at `docs/plans/active/{slug}.md` to answer: what changed, where the initiative stands, what is risky, and what exact action comes next.

## Preconditions

1. Run `zharness preflight watzup --json`. Missing binary: print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop. Otherwise check its `version` field — a `dev` build satisfies the gate; below MIN_ZHARNESS_VERSION (`0.8.1` — see `skills/workflow/README.md`), print the same message and stop. Then follow its status. Its `context` field, when present, is the sole source of lifecycle position for Step 2 below — do not call `resume` separately. Reduced mode may recap Git and any readable active plan when DB state (and so `context`) is absent; it must state that DB position is unavailable.
2. If this session's context was compacted or summarized since the last `preflight` call, re-run it before trusting any earlier-read `context` packet or lifecycle ID — a summarized turn cannot be assumed to have carried exact DB state forward.
3. Remain read-only: do not initialize state, create lifecycle rows, write changesets, edit plans, run quality gates, or modify code.

## Steps

1. **Read branch state** — run `git status -sb`, `git log --oneline main..HEAD`, and `git rev-list --left-right --count main...HEAD`. Capture branch, ahead/behind, and staged/unstaged/untracked scope.
2. **Read lifecycle position from the packet** — read `context` from the same `preflight watzup --json` response captured in Preconditions step 1; render its `position`, `drift`, `readiness`, and latest run/check/handoff IDs 1:1, with no independent prose re-derivation. When `context` is absent, state that DB position is unavailable rather than calling `resume` to fill the gap.
3. **Select the active plan** — confirm exactly one file exists under `docs/plans/active/*.md` by name, without reading it. `query plan` below resolves to it automatically; if more than one exists, it reports `ambiguous_active_plan` — only then read each candidate's Current State to disambiguate.
4. **Recap the plan** — read the plan's `## Outcome` directly (small, fixed size — not the growing history F1 of the ceremony audit identified). For everything else, call `zharness query plan --section current-state --json` (blockers, open items, the recorded exact next action), `zharness query traces --tail 5 --json` and `zharness query decisions --tail 5 --json` (recent Progress/Decisions), and `zharness query checks --tail 1 --json` (latest Validation verdict/proof gaps) instead of reading the rest of the file (the ceremony audit's P3 proposal). If any of these report `degraded: true` or the plan is otherwise unreadable, fall back to a full read. Read task execution status only from append-only `## Progress`, the sole task execution-status source — via `query traces` above, its compressed index, not a full-file read. Task definitions have no status fields.
5. **Read WIP** — inspect the working-tree diff, capped at the most significant five files. Group coherent work themes and identify incomplete implementations or missing proof.
6. **Assess risks** — include DB drift and recovery verbatim, explicit plan blockers, missing tests, large uncommitted diffs, public-contract breaks, secrets, or unsafe migrations. Do not invent reassurance when proof is absent.
7. **Recommend one action** — use drift recovery first; otherwise use the active plan's exact next action when it still matches Git state. If stale, name the concrete reconciliation action.

## Output Shape

Keep the recap concise:

1. Branch and working-tree state.
2. Active plan, phase/status, and latest lifecycle IDs.
3. Completed work and current WIP.
4. Risks/blockers and proof gaps.
5. One exact next action.

## Command Reference

- `zharness preflight watzup --json`
- `zharness query plan --section current-state --json` (step 4)
- `zharness query traces --tail 5 --json`, `zharness query decisions --tail 5 --json`, `zharness query checks --tail 1 --json` (step 4)

## Exit Conditions

Complete only when Git state, DB position or its absence, the selected active plan, current proof, risks, and one exact next action are stated. No command or file operation may mutate repository or harness state.
