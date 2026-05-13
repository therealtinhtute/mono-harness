# Repo Audit — TINHTUTE Skills

## Snapshot
- Audited path: `/root/.openclaw/workspace/tmp/skills-latest-origin`
- Snapshot commit: `10f130cb60ac250281224d8328908831df42fcf1`
- Audit date: 2026-05-13
- Repo shape checked: top-level docs, setup, rules, scripts, hooks, and all 14 skill directories.

## What was checked
### Structure and surface area
- Top-level: `README.md`, `CLAUDE.md`, `docs/`, `rules/`, `scripts/`, `setup/`, `skills/`, `assets/`.
- Skills inventory: `bash-tui`, `brainstorm`, `check`, `cook`, `git`, `handoff`, `interview`, `librarian`, `plan`, `prompt-leverage`, `skill-creator`, `turbo-mono-platform`, `watzup`, `write`.
- Supporting assets sampled: references/scripts under multiple skills, setup hooks, installer scripts, and harness architecture docs.

### Representative skill sampling
I inspected every `SKILL.md` at least at the frontmatter/instruction shape level and did deeper reads on a cross-section:
- Core harness skills: `brainstorm`, `plan`, `cook`, `check`, `handoff`, `watzup`
- Utility/domain skills: `git`, `librarian`, `prompt-leverage`, `write`, `interview`, `turbo-mono-platform`, `bash-tui`, `skill-creator`

### Verification run
- Ran `skills/scripts/validate-skill.sh` across all 14 skills via JSON mode.
- Ran direct validations on `brainstorm`, `check`, `write`, `interview`, `librarian`, `prompt-leverage`.
- Ran `python3 skills/skill-creator/scripts/quick_validate.py skills/skill-creator`.
- Read but did not intentionally execute repo-mutating helpers except one accidental installer invocation noted below.

## Findings

### High
1. **Two shipped skills fail the repo's own validator**  
   **Fact:** `interview` and `librarian` return `fail` from `skills/scripts/validate-skill.sh`.
   - `interview`: missing security section, missing `Defer To Instead`, missing XML-structured sections, missing `version`.
   - `librarian`: no examples found by validator, missing `version`, missing explicit env-var security language, overlength description/line count warnings.
   **Why it matters:** the repo presents a quality bar and packaging discipline, but the current snapshot does not meet it uniformly.

2. **Validation policy and repository reality are out of sync**  
   **Fact:** The validator treats missing examples and missing `Defer To Instead` as hard failures, but multiple shipped skills either rely on alternative patterns or omit those fields. Several skills also omit top-level `version:` while still carrying `metadata.version`.  
   **Inference:** the skill contract has evolved faster than the validator or vice versa. This creates false confidence if maintainers assume “repo passes validator” equals “repo is healthy.”

3. **Setup installer is not safe to probe and lacks argument handling/help**  
   **Fact:** `setup/install.sh --help` does not show help; it proceeds to perform a real install into `~/.claude`, copy files, and invoke `npx skills add ...`.  
   **Why it matters:** this is risky for maintainers and reviewers; even a harmless-looking probe mutates the local environment. It also makes CI/dry-run verification harder.

### Medium
4. **Docs drift between README/CLAUDE and actual skill inventory/structure**  
   **Fact:** top-level `CLAUDE.md` still shows a shortened skills tree that omits `write` and some newer repo areas. README is closer, but not fully explicit about validation status or which skills are harness-core vs utility/domain.  
   **Inference:** structural docs are maintained manually and are likely to drift further as more skills are added.

5. **Cross-platform/local-path assumptions are baked into scripts**  
   **Fact:** `scripts/sync-from-kit.sh` hardcodes `KIT_SKILLS="/Users/tinhtute/Lab/orkit-tui/kit/skills"` while repo audit ran on Linux. Top-level `CLAUDE.md` also references a specific local incubator path.  
   **Why it matters:** this is fine for a personal repo, but it lowers portability and makes automation/documentation less trustworthy for anyone else or for future machines.

6. **Validation/test coverage is shallow for helper scripts**  
   **Fact:** no `tests/` directories were present in the skill folders sampled; validator repeatedly reports “No tests directory (target: 80% of skills).” Helper scripts exist in `git`, `skill-creator`, `turbo-mono-platform`, and top-level `skills/scripts`, but there is no visible harness test suite for them.  
   **Inference:** repository confidence currently depends more on convention and manual review than executable regression checks.

7. **Quality/structure consistency varies noticeably across skills**  
   **Fact:** most newer skills follow the full XML-ish shape (`<role>`, `<security>`, `<context>`, `<instructions>`, `<references>`), but `interview` is materially more lightweight; some skills are near or above the 150-line guideline (`librarian` 153 lines, `prompt-leverage` 151).  
   **Why it matters:** auto-activation and maintainability benefit from consistent contracts; mixed generations of authoring style make the repo feel partially migrated.

8. **Overlap exists across workflow-facing skills and could cause routing ambiguity**  
   **Fact:** `brainstorm`, `plan`, `cook`, `check`, `handoff`, and `watzup` are intentionally composed; separately, `prompt-leverage`, `skill-creator`, `write`, and `interview` all touch “improve or transform text/instructions.”  
   **Inference:** the overlap is mostly deliberate, but trigger boundaries are still soft in places, especially between prompt improvement vs skill creation vs writing/editing.

### Low
9. **Strong positive structure in the harness-core docs**  
   **Fact:** `docs/SKILLS_HARNESS_ARCHITECTURE.md` and `docs/workflow-state-dogfood.md` are specific, coherent, and align well with the core six-skill pipeline described in README. This is one of the repo’s clearest strengths.

10. **Hook stack is thoughtfully modular but not obviously tested end-to-end**  
   **Fact:** `setup/hooks/lib/*` is split into focused helpers (`instruction-builder`, `privacy-checker`, `path-matcher`, `question-detector`, `hook-logger`) with fail-open behavior.  
   **Inference:** the design is sound, but without tests it is easy for behavior drift to hide inside regex/pattern logic.

11. **Some script assumptions look stale or repo-relative paths look wrong**  
   **Fact:** `skills/scripts/install-git-hooks.sh` writes a pre-commit hook that searches for `kit/skills/.*/SKILL.md` and invokes `kit/skills/scripts/validate-skill.sh`, which does not match this repository layout (`skills/...`).  
   **Why it matters:** this likely came from an earlier path convention and may not work if used as-is.

## Verification run
### Commands/evidence
- `git rev-parse HEAD` → `10f130cb60ac250281224d8328908831df42fcf1`
- Repo inventory via `find` confirmed 14 skill folders and top-level docs/setup/scripts.
- Validator summary across all skills:
  - `pass`: `brainstorm`, `check`, `cook`, `handoff`, `plan`, `watzup`, `write`
  - `pass_with_warnings`: `bash-tui`, `git`, `prompt-leverage`, `skill-creator`, `turbo-mono-platform`
  - `fail`: `interview`, `librarian`
- `python3 skills/skill-creator/scripts/quick_validate.py skills/skill-creator` → `Skill is valid!`

### What was not fully verifiable
- I did not run all helper scripts because several are install/scaffold/mutation oriented.
- I did not run external-network-dependent flows beyond the accidental `setup/install.sh` invocation.
- I did not benchmark actual skill activation behavior in Claude/skills.sh runtime; this audit is repo/static-structure focused.

## Risks / blind spots
- This is a snapshot audit, not a runtime UX study inside Claude Code.
- Some findings about overlap/duplication are interpretive; the concrete evidence is in trigger descriptions and adjacent scopes, but actual misrouting frequency was not tested.
- Accidental execution of `setup/install.sh --help` mutated the local `~/.claude` environment during audit; that itself is a finding, but it also means the verification step was not fully side-effect-free.

## Prioritized next steps
1. **Fix or deliberately exempt failing skills**: bring `interview` and `librarian` up to the current contract, or relax the validator with explicit documented exceptions.
2. **Make installer scripts safe**: add `--help`, `--dry-run`, and argument parsing to `setup/install.sh`; avoid side effects on probe.
3. **Repair stale helper paths**: update `skills/scripts/install-git-hooks.sh` and review other scripts for old `kit/skills` or machine-specific path assumptions.
4. **Define the actual skill contract in one place**: reconcile `SKILL.md` conventions, validator expectations, and README/CLAUDE docs.
5. **Add lightweight automated checks**: at minimum, a repo-wide validation script and tests for hooks/path matchers/validators.
6. **Tighten routing boundaries for overlapping text/prompt skills**: especially `write` vs `prompt-leverage` vs `skill-creator` vs `interview`.

## Report path
`/root/.openclaw/workspace/tmp/skills-latest-origin/REPO_AUDIT_2026-05-13.md`
