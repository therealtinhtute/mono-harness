# Playbook: watzup

## Purpose

Read-only session recap from Git state and the committed plan at `docs/plans/active/{slug}.md`: what changed, where the initiative stands, what is risky, and the exact next action.

## Preconditions

1. Select the active plan by name: exactly one non-empty file may exist under `docs/plans/active/*.md`. IF more than one → stop and name every candidate; never pick silently.
2. Remain read-only: no file writes, commits, quality gates, or code changes.

## Steps

1. **Read branch state** — resolve the base branch; `main` is not universal. Use the `origin/HEAD` target only if a local branch confirms it (a PR-only clone has `origin/HEAD -> origin/main` but no local `main`, and a bare `main` makes `git log` exit 128); else the first local `main`/`master`; else `git branch --show-current` (works on an unborn branch, where `git rev-parse --abbrev-ref HEAD` fails):

   ```sh
   base=$(git symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed 's|^origin/||')
   [ -z "$base" ] || git show-ref --verify --quiet "refs/heads/$base" || base=""
   [ -n "$base" ] || for c in main master; do
     git show-ref --verify --quiet "refs/heads/$c" && base=$c && break
   done
   [ -n "$base" ] || base=$(git branch --show-current)
   ```

   Then run `git status -sb`, `git log --oneline "$base..HEAD"`, and `git rev-list --left-right --count "$base...HEAD"`; capture branch, ahead/behind, and staged/unstaged/untracked scope. IF base equals the current branch → report from `git status -sb` alone and say the base comparison was empty; never present zero ahead/behind as a finding.
2. **Read the plan by section** — never the whole file. Prefer `bash scripts/plan-slice.sh docs/plans/active/{slug}.md "{heading}"` IF the script exists; else `awk '/^## {heading}$/{on=1;next} on&&/^## /{exit} on' docs/plans/active/{slug}.md`. Read the `## Outcome` head, all of `## Current State and Next Action` (blockers, open items, recorded next action), and the last few bullets of `## Progress`, `## Decisions`, and `## Validation`. Read task execution status only from append-only `## Progress`, the sole task execution-status source; task definitions carry no status fields.
3. **Read WIP** — inspect the working-tree diff, capped at the five most significant files; group work themes and flag incomplete implementations or missing proof.
4. **Assess risks** — report plan blockers verbatim, missing tests, large uncommitted diffs, public-contract breaks, secrets, unsafe migrations. Never invent reassurance when proof is absent.
5. **Recommend one action** — the plan's exact next action IF it still matches Git state; otherwise the concrete reconciliation action.

## Output and Exit

Complete only when the concise recap states, in order:

1. Branch and working-tree state.
2. Active plan, phase/status, and latest lifecycle anchors named in the plan.
3. Completed work and current WIP.
4. Risks/blockers and proof gaps.
5. One exact next action.
