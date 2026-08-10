# Playbook: git

## Purpose

Git operations with conventional commits: staging, committing, pushing, pull requests, and merges. `git` owns no harness entity (`skills/workflow/README.md`'s skill-to-command mapping) — a missing, stale, or broken harness never blocks it; harness enrichment (Step 0) is a warn-only extra, not a gate.

## Preconditions

1. Run `zharness preflight git --json`. Missing binary: print `harness unavailable: zharness not found or out of date (bash scripts/install-zharness.sh for gate-verdict warnings)`, skip Step 0 below, and proceed to Core Workflow regardless. Otherwise check its `version` field — a `dev` build or any build at or above MIN_ZHARNESS_VERSION (`0.8.1` — see `skills/workflow/README.md`) unlocks Step 0; below it, print the same one-line warning, skip Step 0, and proceed. Any `stop` `preflight` returns (including a corrupted database) is noted the same way and never blocks — Git operations remain non-mutating to harness state.

## Arguments

- `cm` — stage files & create commit(s)
- `cp` — stage, commit, and push
- `pr [to-branch] [from-branch]` — create a pull request (defaults: `main`, current branch)
- `merge [to-branch] [from-branch]` — merge branches (defaults: `main`, current branch)

## Core Workflow

### Step 0: Check latest gate verdict (warn, never block)

Run `zharness query check --latest --json`. If it returns `verdict: REQUEST_CHANGES`, or the command fails (no binary, `db_unreadable`, or no check recorded yet), print a one-line warning naming the verdict or the reason it's unavailable, then proceed anyway. Only `APPROVED` or `APPROVE_WITH_REQUESTS` proceeds silently.

### Step 1: Stage + analyze

```bash
git add -A && git diff --cached --stat && git diff --cached --name-only
```

### Step 2: Security check

Scan staged changes for secrets before committing:

```bash
git diff --cached | grep -iE "(AKIA|api[_-]?key|token|password|secret|credential|private[_-]?key|mongodb://|postgres://|mysql://|redis://|-----BEGIN)"
```

Also warn on staged files matching `.env`/`.env.*` (except `.env.example`), `*.key`, `*.pem`, `*.p12`, `credentials.json`, `secrets.json`, `config/private.*`. **If anything matches: STOP, show the matching lines (`git diff --cached | grep -B2 -A2 <pattern>`), suggest adding to `.gitignore` or moving to environment variables, and offer to unstage (`git reset HEAD <file>`). Do not commit.**

### Step 3: Split decision

Group staged files by kind (`docs:` for `.md`/`.txt`, `test:` for test/spec paths, `config:` for `.claude/` files, `deps:` for `package.json`/lockfiles, `code:` for everything else).

**Single commit:** same type/scope, files ≤ 3, lines ≤ 50.
**Multiple commits:** mixed types/scopes — one commit per group (`chore(config)`, `chore(deps)`, `test`, `feat`/`fix` for `code:`, `docs`). Reset and re-stage per group: `git reset && git add file1 file2 && git commit -m "type(scope): desc"`.

Only use `feat`, `fix`, or `perf` prefixes for `.claude/` directory files (never `docs`). Search for related GitHub issues and note them in the commit/PR body.

### Step 4: Commit

```bash
git commit -m "type(scope): description"
```

**Format:** `type(scope): description`, under 72 characters, present tense/imperative ("add" not "added"), no trailing period, focused on WHAT not HOW.

**Types (priority order):** `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `build`, `ci`.

**Never include AI attribution** ("Generated with Claude", "Co-Authored-By: Claude", or any AI reference) in the commit message.

### Step 5: Push (`cp`, or `cm` + explicit push request only)

```bash
git status && git log origin/$(git rev-parse --abbrev-ref HEAD)..HEAD --oneline 2>/dev/null || echo "NO_UPSTREAM"
git push origin HEAD   # or: git push -u origin HEAD if NO_UPSTREAM
```

**Never force push to `main`, `master`, `production`, `prod`, or `release/*`.** On a feature branch, force push only on explicit user request, and warn: "Force push rewrites history. Collaborators may lose work."

### `pr`: Create pull request

PRs are based on **remote** diffs, not local — local diff includes unpushed changes.

```bash
git fetch origin && git push -u origin HEAD 2>/dev/null || true
BASE=${TO_BRANCH:-main}
HEAD=$(git rev-parse --abbrev-ref HEAD)
git log origin/$BASE...origin/$HEAD --oneline
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

**Console output:**
```
✓ staged: N files (+X/-Y lines)
✓ security: passed
✓ commit: HASH type(scope): description
✓ pushed: yes/no
```

**For `pr`/`merge`:** save a report to `.kit/cache/reports/git/{YYYYMMDD-HHmm}-{operation}.md` (gitignored local scratch — `git` is a sidecar skill and does not own harness lifecycle artifacts), with frontmatter `title`, `description`, `status: completed`, `created`, `tags: [git, {operation}]`.

## Error Handling

| Error | Action |
|---|---|
| Secrets detected | Block commit, show files |
| No changes | Exit cleanly |
| `rejected - non-fast-forward` / push rejected | Suggest `git pull --rebase`, resolve, push again |
| No upstream branch | `git push -u origin HEAD` |
| Merge conflicts | Suggest manual resolution |
| Authentication failed | Check `gh auth status` or SSH keys |

## Anti-Patterns

- Staging everything with `git add -A` instead of specific files — catches `.env`, secrets, `node_modules`.
- Single commit when changes span multiple types/scopes — "one commit is cleaner" produces an un-reviewable diff, impossible to revert selectively.
- Skipping the security scan because "it's just config" — config files often contain secrets or tokens.
- Force pushing without explicit user confirmation — overwrites upstream work silently.

## Command Reference

- `zharness preflight git --json`
- `zharness query check --latest --json` (Step 0, warn-only)
- `gh pr create --base {branch} --head {branch} --title "..." --body "..."`

## Exit Conditions

- `cm`/`cp`: changes staged, security-scanned, committed (single or split by group); `cp` additionally pushed.
- `pr`: remote branch pushed, PR created from the remote diff with a conventional title and summary/test-plan body.
- `merge`: target branch fetched and merged from the remote source branch with `--no-ff`, conflicts resolved or reported, result pushed.
