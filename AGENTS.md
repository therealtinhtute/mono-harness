# AGENTS.md

This file provides guidance to coding agents (Claude Code, Codex, etc.) working in this repository.

## Repository Overview

This is the personal mono-harness repository for `therealtinhtute`: a `skills.sh`-compatible agent skill set for the SDLC, plus `zharness`, the Rust CLI/protocol that keeps that lifecycle legible and portable across coding agents. See README.md's Goals/Non-goals for the full scope.

## Layout and Pipeline

Skills live in `skills/{workflow,shipping,craft}/<name>/`; the `zharness` crate in `cli/`; global rules in `rules/`; protocol, playbooks, ADRs, and plans in `docs/`. The workflow chain and its stage contracts: `skills/workflow/README.md`.

Plans are committed markdown (`docs/plans/active/{slug}.md`); pre-commit guards (`scripts/install-git-hooks.sh`) enforce proof re-execution, an independent judge on high-risk and `full` checks, and at most one active plan. Editing a playbook: change `cli/docs/embedded/playbooks/<stage>.md`, then copy the same bytes to `docs/playbooks/<stage>.md` — `cd cli && cargo test --test projection_parity` fails if the two drift.

## Gate Commands

`check` runs these before any commit. They sit at two different levels of the
ladder in `docs/patterns/encoding-invariants.md` — declare the level, do not
assert enforcement the repository does not have.

```bash
# [CI] Rust CLI: format, lint, test — .github/workflows/cli-ci.yml, job `build-test`
cd cli && cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test

# [CI] Rust CLI size target — .github/workflows/cli-ci.yml, job `size-gate`
cd cli && cargo build --release --target x86_64-unknown-linux-musl && test "$(wc -c < target/x86_64-unknown-linux-musl/release/zharness)" -le 1850000

# [CI] Guard fixture tests for the pre-commit hook's ZGUARD-CORE block — .github/workflows/cli-ci.yml, job `hook-guard`
bash scripts/test-guards.sh

# [CI] Doc link integrity (broken repo-relative cross-references) — .github/workflows/docs-ci.yml, job `doc-links`, on any tracked *.md change.
# Exceptions live in .claimignore, each one requires a `# reason`.
bash scripts/verify-doc-links.sh
```

`scripts/validate-skill.sh` is `Optional hook`: it runs on changed skills from
the pre-commit hook only, which requires `bash scripts/install-git-hooks.sh --force`.

Run a single Rust test: `cd cli && cargo test --lib installer::registry::tests::registry_update_registers`.

## Architecture Notes

- **Prompt engineering:** before writing or editing skills (SKILL.md), rules (rules/*.md), or any agent instruction file, read `docs/prompt-engineering-principles.md` — context engineering, formatting syntax, language rules, few-shot patterns, anti-patterns, cross-model awareness.
- **Skill format:** All skills follow the `skills.sh` standard — YAML frontmatter with `name` and `description`, imperative instructions, optional `references/` and `scripts/` directories.
- **rules/ directory:** Source-of-truth for rules installed to `~/.claude/rules/`. Keep in sync with installed versions.
- **`site/` is hand-authored, not generated.** A static GitHub Pages site (`.github/workflows/pages.yml` deploys on push to `site/**`) that narrates the same architecture/workflow story as `docs/`; it does not regenerate from `docs/*.md` and can drift — it was last brought in line with v0.24.0.

<!-- ZHARNESS:BEGIN -->
## Harness

Start with the requested outcome and use the repository as the system of record.
Read `docs/WORKFLOW.md` and only relevant product, design, plan, code, and
validation material.

- Answers, explanations, reviews, diagnoses, plans, and status reports are
  read-only. Inspect only what is needed; change nothing.
- For a bounded change, inspect affected behavior and proof, implement, and
  validate. No plan file is required.
- Use one `docs/plans/active/` file when work spans sessions, coordinates
  contributors, has dependencies, or needs recovery. Move it to
  `docs/plans/completed/` only after validation.
- Before editing, identify repository authority for each new externally
  observable policy. If materially different choices remain open, stop before
  edits; configurable defaults are not authority.
- Claim completion only with executable or observable evidence. Report outcome,
  changes, validation, and unresolved risks.

The `zharness` binary is install / update / uninstall only. It does not run
the lifecycle. There is no task database. There is no parallel control-plane state.
<!-- ZHARNESS:END -->
