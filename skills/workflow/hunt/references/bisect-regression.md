# hunt: bisect and regression modes

Loaded by `hunt` for `bisect` and `regression`, alongside the loop in `SKILL.md`.

## Bisect

- Protect the worktree first: `git status --short --branch -uall`. Any modified, staged, or untracked file means bisect runs in a temporary detached worktree, removed afterwards.
- If last-good is only a few releases back, read `git diff <good>..HEAD -- <suspect path>` first; bisect only when the diff is too large or the culprit is not obvious.
- Bisect with a non-interactive pass/fail command defined up front (`git bisect run <loop>`). Read the culprit diff down to the line, then `git bisect reset`.

## Regression

Treat the reference (last-good commit, old build, fixture, screenshot, described expected state) as the oracle, not decoration. Define the pass/fail check against it before editing, then name the exact current-vs-reference delta. Do not generalize a broken render, race, or state path into "style polish".
