# Playbook: watzup

## Purpose

Read-only session recap. Combine Git state, DB lifecycle position, and the current initiative at `docs/plans/active/{slug}.md` to answer: what changed, where the initiative stands, what is risky, and what exact action comes next.

## Preconditions

1. Run `zharness --version`. A `dev` build satisfies the gate; otherwise require version `0.1.0` or newer. If unavailable or stale, print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop.
2. Run `zharness preflight watzup --json` and follow its status. Reduced mode may recap Git and any readable active plan when DB state is absent; it must state that DB position is unavailable.
3. Remain read-only: do not initialize state, create lifecycle rows, write changesets, edit plans, run quality gates, or modify code.

## Steps

1. **Read branch state** — run `git status -sb`, `git log --oneline main..HEAD`, and `git rev-list --left-right --count main...HEAD`. Capture branch, ahead/behind, and staged/unstaged/untracked scope.
2. **Read lifecycle position** — run `zharness resume --json` once. Capture readiness, drift, current phase/status, and latest run/check/handoff IDs.
3. **Select the active plan** — inspect `docs/plans/active/*.md`. Prefer the plan whose Phases and Verification or Current State matches DB `current_phase`. If more than one remains plausible, state the ambiguity instead of choosing silently.
4. **Recap the plan** — summarize Outcome; active phase/wave/task; recent append-only Progress; material Decisions; latest Validation verdict/proof gaps; blockers/open items; and the exact next action already recorded in Current State. Read task execution status only from append-only `## Progress`, the sole task execution-status source; task definitions have no status fields.
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

- `zharness --version`
- `zharness preflight watzup --json`
- `zharness resume --json`

## Exit Conditions

Complete only when Git state, DB position or its absence, the selected active plan, current proof, risks, and one exact next action are stated. No command or file operation may mutate repository or harness state.
