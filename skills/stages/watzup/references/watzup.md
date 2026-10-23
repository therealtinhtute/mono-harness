# Playbook: watzup

Read-only recap from Git and the plan at `.harness/plans/active/{slug}.md`.

## Preconditions

1. Select the active plan by name: exactly one non-empty file may exist under `.harness/plans/active/*.md`; IF more → name every candidate and stop.
2. Remain read-only: no writes, commits, gates, or code changes.

## Steps

1. **Branch state** — resolve the base (do not assume `main`):

   ```sh
   base=$(git symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed 's|^origin/||')
   [ -z "$base" ] || git show-ref --verify --quiet "refs/heads/$base" || base=""
   [ -n "$base" ] || for c in main master; do
     git show-ref --verify --quiet "refs/heads/$c" && base=$c && break
   done
   [ -n "$base" ] || base=$(git branch --show-current)
   ```

   Then `git status -sb`, `git log --oneline "$base..HEAD"`, `git rev-list --left-right --count "$base...HEAD"`. IF base is the current branch → report status only.
2. **Plan by section** — never the whole file (slice as in `work-full.md` step 1): `## Goal` head, Current State, the `## Log` and `## Validation` tails. Read task execution status only from append-only `## Log`, the sole task execution-status source.
3. **WIP** — the working-tree diff, top five files; flag incomplete work or missing proof.
4. **Risks** — blockers verbatim, missing tests, large uncommitted diffs, contract breaks, secrets. No invented reassurance.

## Output

In order: branch and tree state; plan, phase, status; done and WIP; risks and proof gaps; one exact next action (the plan's, IF it still matches Git).
