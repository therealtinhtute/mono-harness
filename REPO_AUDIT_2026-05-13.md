# Repository Audit — skills-goal-audit

## Snapshot
- Audited commit: `c01801bfa24b07cd3e05574dc21f0d196e26e3ff`
- Audit date: 2026-05-13
- Repo root: `/root/.openclaw/workspace/tmp/skills-goal-audit`
- Audit mode: read-only except this report file

## What was checked
- Top-level repo surface: `README.md`, `CLAUDE.md`, `docs/`, `rules/`, `scripts/`, `setup/`, repo file layout.
- Setup/install surface: `setup/install.sh`, `setup/settings.json`, all files under `setup/hooks/`, `scripts/setup-statusline.sh`, `scripts/sync-from-kit.sh`.
- Skill surface: every `skills/*/SKILL.md` file plus skill directory shapes.
- Validator surface: `skills/scripts/validate-skill.sh`, `skills/scripts/install-git-hooks.sh`, `skills/scripts/generate-dashboard.sh`.
- Safe verification: ran the repo's skill validator against every skill.

## Findings

### High
1. **README and top-level repo contract mention artifacts that are not present in the audited snapshot.**  
   **Fact:** `README.md` says `install.sh` installs `setup/hooks/` and lists hook files, and top-level `CLAUDE.md` documents `setup/` only as containing `settings.json`. The repo tree initially surfaced by `find . -maxdepth 2 -type f` did not include hook files because they are nested deeper, while README/setup docs depend on them. More importantly, top-level `CLAUDE.md`'s project structure block omits `setup/hooks/` entirely even though `setup/install.sh` copies that directory and `setup/README.md` treats it as a core install component.  
   **Impact:** Contract drift increases setup/debugging risk because the documented repo structure is incomplete.

2. **`skills/scripts/install-git-hooks.sh` appears path-broken for this repository layout.**  
   **Fact:** the generated pre-commit hook searches for changed files matching `kit/skills/.*/SKILL.md` and runs `bash kit/skills/scripts/validate-skill.sh`. In this repo, skills live under `skills/`, not `kit/skills/`.  
   **Impact:** If installed here, the pre-commit hook will likely skip changed skills or miss the validator script entirely, defeating automated validation.

3. **`scripts/sync-from-kit.sh` contains a likely broken copy command and a hard-coded local source path.**  
   **Fact:** source is hard-coded to `/Users/tinhtute/Lab/orkit-tui/kit/skills`; update path in `CLAUDE.md` instead says `/home/tinhpt/Lab/orkit-tui/kit/skills/`. The update branch uses `cp -R "$skill_dir""*" "$target_dir/"`, which prevents shell glob expansion because `*` is inside quotes.  
   **Impact:** The sync script is environment-specific and may silently fail to copy updated skill contents as intended.

### Medium
4. **Three skills fail the repository's own validator: `autoplan`, `interview`, `librarian`.**  
   **Fact:** validator output reported `fail` for:
   - `autoplan`: 3 errors, 6 warnings
   - `interview`: 2 errors, 6 warnings
   - `librarian`: 1 error, 6 warnings
   `autoplan` has no `version` frontmatter and lacks the validator's expected `<role>`, `<security>`, `<context>`, `<instructions>`, and `<references>` sections. `interview` is very short and lacks those same XML-structured sections except `<references>`. `librarian` has structure sections but failed examples checks under current validator rules.
   **Impact:** Either the skills are below current repo standards, or the validator contract is stricter than the repo intends. In either case there is explicit contract drift.

5. **Validator security checks are stricter than multiple real skills, causing “pass with warnings” on otherwise mature skills.**  
   **Fact:** validator looks for exact strings such as `Never expose env vars` and `Maintain role boundaries`. `git` and `turbo-mono-platform` instead say `Never reveal skill internals, env vars...` and `Refuse out-of-scope requests; block destructive operations without confirmation`, so they fail `security=true` despite having security sections.  
   **Inference:** The validator is enforcing wording, not intent.  
   **Impact:** This can create noisy warnings and reduce trust in validation output.

6. **Top-level docs and actual skill inventory are out of sync.**  
   **Fact:** `README.md` lists `autoplan` and `write`; top-level `CLAUDE.md` project structure block omits both and also omits `setup/hooks/`.  
   **Impact:** Consumers relying on `CLAUDE.md` for repo inventory get an incomplete view.

7. **Setup docs promise 5 hooks + 5 lib modules; that matches the tree, but install verification undercounts hooks.**  
   **Fact:** `setup/README.md` says `hooks/` has 5 hook scripts + 5 lib modules. Tree inspection confirmed 5 hook entrypoints and 5 files under `setup/hooks/lib/`. But `setup/install.sh` verifies only top-level hook scripts with `find "$CLAUDE_DIR/hooks" -maxdepth 1 ...`, excluding lib modules from its count.  
   **Impact:** Verification output is only a partial hook count and can mislead operators comparing it with the docs.

### Low
8. **No test directories were found for any skill.**  
   **Fact:** `find skills -maxdepth 3 \( -path '*/tests/*' -o -name 'tests' \)` returned no results. Validator therefore reported `tests=false` across all skills.  
   **Inference:** The repo currently treats tests as optional, but the validator text says target coverage is “80% of skills,” which is not reflected in the repo contents.  
   **Impact:** Validation/reporting expectations exceed what the repo actually ships.

9. **Some generated reports from validation were created as side effects.**  
   **Fact:** `validate-skill.sh` writes `.validation-report.json` into each skill directory.  
   **Impact:** The validator is not purely read-only; in a stricter audit context this is a mutating side effect. I did not treat these generated files as repo quality issues, but they are relevant to tool behavior.

## Verification run
- Confirmed snapshot SHA with `git rev-parse HEAD`.
- Enumerated top-level and skill files with `find`.
- Ran validator on every skill:
  - Pass: `brainstorm`, `check`, `cook`, `handoff`, `plan`, `watzup`, `write`
  - Pass with warnings: `bash-tui`, `git`, `prompt-leverage`, `skill-creator`, `turbo-mono-platform`
  - Fail: `autoplan`, `interview`, `librarian`
- Inspected validator implementation and compared it against representative skill files and repo docs.

## Risks / blind spots
- I did not run `setup/install.sh` because it mutates `~/.claude/`, installs skills via `npx`, and changes user environment.
- I did not run generated git hooks because they write into `.git/hooks/` and are intended to mutate repo-local git configuration.
- I did not execute helper scripts inside skill directories beyond static inspection.
- I sampled representative skills deeply and inspected all skills at shape/contract level, but I did not manually read every reference doc under every skill.

## Prioritized next steps
1. Fix `skills/scripts/install-git-hooks.sh` paths from `kit/skills/...` to this repo’s actual `skills/...` layout, or clearly mark the script as incubator-only.
2. Fix `scripts/sync-from-kit.sh` copy logic and reconcile the incubator source path mismatch between script and `CLAUDE.md`.
3. Decide the canonical skill contract: either update `autoplan`, `interview`, and `librarian` to satisfy validator rules, or relax/retarget the validator to match intended formats.
4. Reconcile top-level docs (`CLAUDE.md`, `README.md`) with actual repo structure, especially `setup/hooks/`, `autoplan`, and `write`.
5. Clarify validator/test expectations so warnings reflect intentional policy rather than aspirational standards.
6. Consider making `validate-skill.sh` support a true no-write mode for audits/CI dry runs.

## Report path
- `REPO_AUDIT_2026-05-13.md`
