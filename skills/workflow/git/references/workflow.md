# git skill: workflow reference

> This is the `git` skill's own procedure, not a harness-projected playbook. It
> is not part of the embedded doc set and `zharness install` never writes it.
> Edit it here — `cli/docs/embedded/playbooks/` holds no `git` entry upstream
> to change. A stale `git.md` under a repository's `docs/playbooks/` is a
> leftover projection from before the deprojection; see
> `docs/workflow-harness/migration.md`.

## Purpose

Git operations with conventional commits: staging, committing, pushing, pull requests, and merges. `git` owns no harness entity — a missing, stale, or broken harness never blocks it; Git operations remain non-mutating to harness state.

Loaded by `SKILL.md` for `pr` and `merge` only; `cm`/`cp` steps, commit output, error handling, anti-patterns, and exit condition live in `SKILL.md`.

### `pr`: Create pull request

PRs are based on **remote** diffs, not local — local diff includes unpushed changes.

```bash
git fetch origin && git push -u origin HEAD 2>/dev/null || true
BASE=$TO_BRANCH
if [ -z "$BASE" ]; then
  BASE=$(git symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed 's|^origin/||')
  [ -z "$BASE" ] || git show-ref --verify --quiet "refs/remotes/origin/$BASE" || git show-ref --verify --quiet "refs/heads/$BASE" || BASE=""
  [ -n "$BASE" ] || for c in main master; do
    { git show-ref --verify --quiet "refs/remotes/origin/$c" || git show-ref --verify --quiet "refs/heads/$c"; } && BASE=$c && break
  done
  [ -n "$BASE" ] || BASE=$(git branch --show-current)
fi
HEAD=$(git rev-parse --abbrev-ref HEAD)
git log "origin/$BASE...origin/$HEAD" --oneline
git diff origin/$BASE...origin/$HEAD --stat
```

If the branch isn't on remote yet, push first and retry. Title: conventional-commit format, <72 chars, no version numbers. Body: summary bullets + test-plan checklist. Create with:

```bash
gh pr create --base $BASE --head $HEAD --title "..." --body "$(cat <<'EOF'
## Summary
- Bullet points

## Test plan
- [ ] Test item
EOF
)"
```

Do not use local-comparison commands (`git diff main...HEAD`, `git diff --cached`, `git status`) to describe PR scope.

### `merge`: Merge branches

```bash
git fetch origin
git checkout {TO_BRANCH}
git pull origin {TO_BRANCH}
git merge origin/{FROM_BRANCH} --no-ff -m "merge: {FROM_BRANCH} into {TO_BRANCH}"
```

Merge from `origin/{FROM_BRANCH}`, never the local branch — this ensures only committed+pushed changes are merged, not local WIP. Before merging, check for conflicts: `git merge --no-commit --no-ff origin/{FROM_BRANCH}` then abort. On conflicts: resolve manually, `git add . && git commit`; report to the caller if clarification is needed. Push the result: `git push origin {TO_BRANCH}`.

## Output Format

**For `pr`/`merge`:** save a report to `.zharness/cache/reports/git/{YYYYMMDD-HHmm}-{operation}.md` (gitignored local scratch — `git` is a sidecar skill and does not own harness lifecycle artifacts), with frontmatter `title`, `description`, `status: completed`, `created`, `tags: [git, {operation}]`.

## Exit Conditions

- `pr`: remote branch pushed, PR created from the remote diff with a conventional title and summary/test-plan body.
- `merge`: target branch fetched and merged from the remote source branch with `--no-ff`, conflicts resolved or reported, result pushed.
