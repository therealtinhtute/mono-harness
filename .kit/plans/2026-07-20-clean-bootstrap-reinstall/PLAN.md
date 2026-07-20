---
title: Clean Bootstrap and Harness Reinstallation
status: completed
created: 2026-07-20
scope: local-machine-and-project
---

# Clean Bootstrap and Harness Reinstallation

## Outcome

Replace the stale workflow-skill installation with the current `origin/master` versions, install `zharness v0.4.1`, refresh the full Claude Code bootstrap while preserving the local `workflow-core` auto-trigger customization, remove this repository's old harness history, initialize a fresh empty harness, verify the complete setup, and stop before committing.

## Locked decisions

1. Delete this repository's existing `.kit` history and restart from an empty harness.
2. Run the full bootstrap, not a workflow-only reinstall.
3. Preserve `~/.claude/rules/workflow-core.md` and its pointer in `~/.claude/CLAUDE.md`.
4. Refresh the shared `~/.agents/skills/` workflow folders so existing linked agents receive the current versions.
5. Install skills from current `origin/master`; install the CLI from release `v0.4.1`.
6. Use backups and `trash`; never use `rm`.
7. Leave all repository changes uncommitted for review.

## Boundaries

### May change

- Workflow and repository-owned skill folders under `~/.agents/skills/`
- Existing agent-client symlinks that point to those shared folders
- Repository-managed files under `~/.claude/`: `CLAUDE.md`, rules, hooks, statusline, skills, and `.bootstrap-version`
- `~/.local/bin/zharness`
- This repository's `.kit/` contents

### Must not change

- Values or credentials in `~/.claude/settings.json`
- Third-party skills not sourced from this repository
- Project files outside `.kit/`
- Remote branches, tags, releases, commits, or pushes
- The retained `workflow-core.md` rule

## Current facts

- `master` matches `origin/master`.
- The six repository spine skills are thin triggers; the installed copies are older detailed prompts.
- Seven workflow skills are shared from `~/.agents/skills/` across Claude Code and roughly 21 other linked agent clients.
- `to-plan` is an old standalone directory under `~/.claude/skills/`, not a shared symlink.
- `zharness` is absent from `PATH` and `~/.local/bin/zharness` does not exist.
- Latest harness release is `v0.4.1`.
- The 121 old tracked `.kit` artifacts are already absent from the working tree.
- The only known local `CLAUDE.md` customization is the `workflow-core.md` pointer.
- Required tools are available and GitHub authentication is active.

## Execution plan

### Phase 1 — Preflight and recovery map

**Goal:** Prove the cleanup targets before changing anything.

1. Capture the current repository status, `origin/master` commit, global bootstrap stamp, workflow-skill hashes, and cross-agent symlink map.
2. Confirm every existing project deletion is under `.kit/`; pause if any unrelated project file is modified or deleted.
3. Record the exact repository-owned skill names and distinguish shared folders from standalone Claude Code copies.
4. Confirm `~/.claude/settings.json` exists and exclude it from every overwrite or cleanup operation.
5. Confirm release `v0.4.1` contains the expected Linux AMD64 archive before downloading it.

**Verify:** The target list contains only repository-owned skills, bootstrap-managed configuration, the harness binary path, and `.kit/` artifacts.

**Pause if:** unrelated files appear, a target path resolves outside the expected directories, authentication fails, or release assets differ from the expected platform archive.

### Phase 2 — Remove stale workflow installation safely

**Goal:** Eliminate old workflow implementations without affecting unrelated skills or agent configuration.

1. Preserve the preflight inventory and rely on `trash` recovery rather than permanent deletion.
2. Move stale shared workflow folders for `brainstorm`, `work`, `check`, `handoff`, `watzup`, `interview`, and `git` to trash when present.
3. Move any stale shared `to-plan` folder to trash when present.
4. Move the standalone `~/.claude/skills/to-plan` directory to trash.
5. Leave existing cross-agent symlinks in place temporarily; they may be dangling only until the shared folders are reinstalled.
6. Do not hand-edit or delete unrelated entries in `~/.agents/.skill-lock.json`.

**Verify:** Only the eight targeted workflow implementations are absent; unrelated third-party skill folders remain unchanged.

**Pause if:** a workflow target is not repository-owned, contains unrecognized local modifications that were not visible in preflight, or trashing a target would remove an unrelated directory.

### Phase 3 — Run the full bootstrap and preserve local policy

**Goal:** Refresh Claude Code configuration and all repository skills from current `origin/master`.

1. Run `bash setup/install.sh` from the synchronized repository.
2. Capture and inspect every timestamped backup created by the installer.
3. Confirm `~/.claude/settings.json` was reported as existing and was not overwritten.
4. Read the refreshed `~/.claude/CLAUDE.md`; restore only the accepted `workflow-core.md` pointer if the bootstrap removed it.
5. Confirm `~/.claude/rules/workflow-core.md` still exists and was not modified.
6. Confirm the eight workflow skills now resolve through the current shared installation, including `to-plan`.
7. Re-check all previously recorded cross-agent workflow links; the shared targets must resolve again.

**Verify:** Bootstrap verification reports the expected global files, the bootstrap stamp matches current `master`, and local workflow-core behavior remains present.

**Pause if:** settings credentials change, the bootstrap sources a different revision, shared links remain broken, or an unrelated local customization would be overwritten without an accepted recovery step.

### Phase 4 — Install `zharness v0.4.1`

**Goal:** Install the CLI required by the thin workflow triggers without invoking a cleanup path that uses `rm`.

1. Query release `v0.4.1` and download `zharness_linux_amd64.tar.gz` into a temporary directory.
2. Extract the `zharness` binary and install it as `~/.local/bin/zharness` with mode `0755`.
3. Move the temporary download directory to trash after successful installation.
4. Do not run `scripts/install-zharness.sh` unchanged because its temporary-directory trap uses `rm -rf`, which violates the active deletion rule.

**Verify:** `command -v zharness` resolves to `~/.local/bin/zharness`, the file is executable, and `zharness --version` reports `v0.4.1`.

**Pause if:** the downloaded version, platform, archive contents, executable path, or reported version differs.

### Phase 5 — Initialize a fresh project harness

**Goal:** Start a new empty lifecycle without restoring any legacy harness history.

1. Preserve this plan under `.kit/plans/`; treat it as new planning evidence, not legacy history.
2. Confirm the old tracked `.kit/changesets`, planning artifacts, run logs, reports, handoff, and legacy `workflow-state.yml` remain absent.
3. Create `.kit/` if necessary and run `zharness init --json`.
4. Confirm generated `.kit/docs/playbooks/` and local `harness.db` are created according to current ignore rules.
5. Do not import or replay old changesets.
6. Run fresh state and continuity queries to establish the empty starting condition.

**Verify:** Initialization succeeds; generated docs are current; no legacy state file or old artifact is recreated; the harness reports no active historical phase, run, check, or handoff.

**Pause if:** initialization attempts to import legacy content, generated docs are stale, ignored files are unexpectedly tracked, or old history reappears.

### Phase 6 — Final proof and review stop

**Goal:** Demonstrate that the installation and empty harness are coherent before any commit decision.

1. Compare installed workflow `SKILL.md` hashes with the repository sources.
2. Confirm `brainstorm`, `to-plan`, `work`, `check`, `handoff`, and `watzup` are each at most 30 lines and delegate to their matching `.kit/docs/playbooks/*.md` file.
3. Confirm `interview` and `git` match their current repository versions.
4. Confirm every pre-existing linked agent path resolves to the refreshed shared skill target.
5. Run `zharness query state --json`, `zharness resume --json`, `zharness validate --json`, and `zharness audit --json`; inspect both exit status and structured output.
6. Re-run the bootstrap verification checks for CLAUDE.md, rules, hooks, skills, settings, statusline, and bootstrap stamp.
7. Confirm `workflow-core.md` and its CLAUDE.md pointer remain present.
8. Inspect repository status and prove that changes are limited to the intentional `.kit` reset plus this new plan and generated ignored harness files.
9. Stop without staging, committing, pushing, or creating a pull request.

**Done when:** every verification above passes, no old workflow implementation or project harness history remains active, the new empty harness is queryable, shared agent links resolve, and the repository is ready for explicit review.

**Pause if:** any verification fails, a skill differs from `origin/master`, the CLI is not exactly `v0.4.1`, validation reports unexplained drift or contract violations, unrelated files changed, or a commit is required to proceed.

## Final evidence checklist

- [ ] Preflight target and recovery inventory captured
- [ ] Only stale workflow implementations moved to trash
- [ ] Full bootstrap completed from current `master`
- [ ] `~/.claude/settings.json` unchanged
- [ ] `workflow-core.md` rule and pointer preserved
- [ ] Shared agent links resolve
- [ ] Installed skill hashes match repository sources
- [ ] Six spine skills are thin playbook triggers
- [ ] `zharness --version` reports `v0.4.1`
- [ ] Fresh harness initialized without import
- [ ] State, resume, validate, and audit checks inspected
- [ ] Repository changes limited to the accepted `.kit` reset
- [ ] No staging, commit, push, or PR performed
