# Playbook: watzup

## Purpose

Read-only session recap built from Git state and the committed plan at `docs/plans/active/{slug}.md`. It answers: what changed, where the initiative stands, what is risky, and what exact action comes next.

## Preconditions

1. Select the active plan by name: exactly one non-empty file may exist under `docs/plans/active/*.md`. With more than one, stop the recap and name every candidate so the user resolves it; do not pick silently.
2. Remain read-only: no file writes, no commits, no quality gates, no code modification.

## Steps

1. **Read branch state** — run `git status -sb`, `git log --oneline main..HEAD`, and `git rev-list --left-right --count main...HEAD`. Capture branch, ahead/behind, and staged/unstaged/untracked scope.
2. **Read the plan by section** — never the whole file. Prefer `bash scripts/plan-slice.sh docs/plans/active/{slug}.md "{heading}"`. Read `## Outcome` head, the entire `## Current State and Next Action` (blockers, open items, the recorded exact next action), and the last few bullet lines of `## Progress`, `## Decisions`, and `## Validation`. Read task execution status only from append-only `## Progress`, the sole task execution-status source; task definitions carry no status fields.
3. **Read WIP** — inspect the working-tree diff, capped at the most significant five files. Group coherent work themes and identify incomplete implementations or missing proof.
4. **Assess risks** — report explicit plan blockers verbatim, missing tests, large uncommitted diffs, public-contract breaks, secrets, unsafe migrations. Never invent reassurance when proof is absent.
5. **Recommend one action** — the active plan's exact next action when it still matches Git state; otherwise name the concrete reconciliation action.

## Output Shape

Keep the recap concise:

1. Branch and working-tree state.
2. Active plan, phase/status, and latest lifecycle anchors named in the plan.
3. Completed work and current WIP.
4. Risks/blockers and proof gaps.
5. One exact next action.

## Exit Conditions

Complete only when Git state, the selected active plan, current proof, risks, and one exact next action are stated. Recap stays response-only: nothing in this playbook mutates repository state, and the optional block above never runs its own writes.
