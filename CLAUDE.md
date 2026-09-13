# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is the personal mono-harness repository for `therealtinhtute`: a `skills.sh`-compatible agent skill set for the SDLC, plus `zharness`, the Go CLI/protocol that keeps that lifecycle legible and portable across coding agents. See README.md's Goals/Non-goals for the full scope.

## Project Structure

```
.
├── README.md               # Skill directory and install instructions
├── CLAUDE.md               # This file
├── skills/                 # Installable agent skills (npx skills add)
│   ├── workflow/           # Agentic orchestration chain
│   │   ├── brainstorm/
│   │   ├── to-plan/
│   │   ├── work/
│   │   ├── interview/
│   │   ├── check/
│   │   ├── git/
│   │   ├── handoff/
│   │   ├── watzup/
│   │   ├── encode-invariant/
│   │   └── improve-harness/
│   ├── shipping/           # Build & ship code
│   │   ├── create-cli/
│   │   └── turbo-mono-platform/
│   └── craft/              # Research, writing, meta-skills
│       ├── write/
│       ├── librarian/
│       ├── create-skill/
│       └── prompt-leverage/
├── cli/                    # zharness Go binary — install/update/uninstall only
│   ├── cmd/zharness/           # main package (cobra)
│   ├── internal/               # embedded/, installer/, interfaces/
│   └── docs/embedded/          # go:embed source for playbooks/templates
├── rules/                  # Source for global Claude Code rules (installed to ~/.claude/rules/)
│   ├── ask-user-question.md   # AskUserQuestion enforcement
│   ├── english.md             # English coaching
│   ├── execution-discipline.md # Lean tool-call economy, check-in cadence, stop-don't-guess
│   └── karpathy-guidelines.md # Karpathy coding principles
├── docs/                   # Workflow protocol, architecture, decisions, plans
│   ├── WORKFLOW.md             # Protocol entrypoint — read this first
│   ├── ARCHITECTURE.md         # Current design + historical (pre-v0.15) cuts
│   ├── playbooks/              # Operating logic for the 6 spine skills
│   ├── decisions/              # ADRs for hard-to-reverse calls
│   ├── plans/{active,completed}/  # Durable plan lifecycle (see Skill Pipeline)
│   └── prompt-engineering-principles.md  # Prompting principles for skills/rules
├── scripts/                # Repo utility scripts
│   ├── setup-statusline.sh    # Statusline installer
│   ├── generate-dashboard.sh  # Dashboard generation
│   ├── validate-skill.sh      # Skill validation
│   └── install-git-hooks.sh   # Git hooks installer
├── setup/                  # Example configs
│   └── settings.json       # Claude Code settings template
└── assets/                 # README visuals
```

Each skill directory contains:
- `SKILL.md` — Required. Frontmatter + instructions for the agent.
- `references/` — Optional. Detail docs loaded on-demand.
- `scripts/` — Optional. Executable helpers.

## Development Commands

```bash
# List skills without installing
npx skills add git@github.com:therealtinhtute/mono-harness.git --list

# Install all skills globally for Claude Code
npx skills add git@github.com:therealtinhtute/mono-harness.git -a claude-code -g -y
```

## Gate Commands

`check` runs these before any commit. They sit at two different levels of the
ladder in `docs/patterns/encoding-invariants.md` — declare the level, do not
assert enforcement the repository does not have.

```bash
# [CI] Go CLI: build, vet, test — .github/workflows/cli-ci.yml, job `build-test`
cd cli && CGO_ENABLED=0 go build ./... && go vet ./... && go test ./...

# [CI] Guard fixture tests for the pre-commit hook's ZGUARD-CORE block
# .github/workflows/cli-ci.yml, job `hook-guard`
bash scripts/test-guards.sh

# [CI] Doc link integrity — fails on broken repo-relative cross-references.
# .github/workflows/docs-ci.yml, job `doc-links`, on any tracked *.md change.
# Exceptions live in .claimignore, each one requires a `# reason`.
bash scripts/verify-doc-links.sh
```

`scripts/validate-skill.sh` is `Optional hook`: it runs on changed skills from
the pre-commit hook only, which requires `bash scripts/install-git-hooks.sh --force`.

Run a single Go test: `cd cli && go test ./internal/installer/... -run TestName`.

## Skill Pipeline

Two entry points:
```
watzup → work → check → git → handoff          (resume existing work)
brainstorm → to-plan → work → check → git → handoff  (new work)
```
- `watzup` — recap branch state, committed + uncommitted changes, handoff context, recommend next action (session start)
- `brainstorm` — explore options and lock requirements into `docs/plans/active/{slug}.md` (4 modes: explore, lock-from-idea, lock-from-files, refine)
- `to-plan` — generate executable phase plans from the locked plan's requirements
- `work` — execute the plan wave-by-wave, verify per task, route to `check` as the phase gate
- `check` — pre-commit gate and post-implementation review (also invoked per phase by `work`)
- `git` / `handoff` — session close-out

`interview` is optional — use to grill fuzzy intent into a clear goal, or to validate an existing plan before `work`. Can sit between `brainstorm` and `to-plan`, or between `to-plan` and `work`.

State underneath this pipeline is committed markdown: the plan documents under `docs/plans/active/{slug}.md` (moved to `docs/plans/completed/` on closure) are the record, and fail-closed pre-commit guards in `scripts/install-git-hooks.sh` enforce proof re-execution, an independent judge on high-risk and on `full` checks, and at most one active plan. There is no database — the SQLite store and the whole lifecycle command surface were deleted in v0.15 (see `docs/ARCHITECTURE.md`). The 6 spine `SKILL.md` files (`watzup`, `brainstorm`, `to-plan`, `work`, `check`, `handoff`) are thin triggers (≤30 lines) that route straight to `docs/playbooks/<stage>.md`; the operating logic lives there, not in the skill files, so any agent that can read a file and run git can execute the same lifecycle with no binary installed. `zharness` itself is now three verbs — `install` / `update` / `uninstall` — which scaffold that managed doc set, fresh-overwriting playbooks/WORKFLOW.md on update and three-way-merging only `docs/PROJECT.md` and the `AGENTS.md` block. See `skills/workflow/README.md` for the full model and `docs/workflow-harness/migration.md` for the historical 0.14.x adoption path. Editing a playbook: change `cli/docs/embedded/playbooks/<stage>.md`, then copy the same bytes to `docs/playbooks/<stage>.md` — `cd cli && go test ./...` fails (`TestProjectionParity`) if the two drift.

## Prompt Engineering Reference

When writing or editing skills (SKILL.md), rules (rules/*.md), or any agent instruction file, read `docs/prompt-engineering-principles.md` first. It covers: context engineering principles, formatting syntax (XML vs Markdown, bullets vs paragraphs), language rules, few-shot patterns, anti-patterns, and cross-model awareness.

## Architecture Notes

- **Stable release:** This repo is the stable, standalone release of the skills. The former `orkit-tui` incubator has been archived — this repo is no longer synced from it; edit skills here directly.
- **Skill format:** All skills follow the `skills.sh` standard — YAML frontmatter with `name` and `description`, imperative instructions, optional `references/` and `scripts/` directories.
- **rules/ directory:** Source-of-truth for rules installed to `~/.claude/rules/`. Keep in sync with installed versions.
- **Private repo:** Installable via SSH (`git@github.com:therealtinhtute/mono-harness.git`) as long as local SSH keys are configured.
- **`site/` is hand-authored, not generated.** A static GitHub Pages site (`.github/workflows/pages.yml` deploys on push to `site/**`) that narrates the same architecture/workflow story as `docs/`; it does not regenerate from `docs/*.md` and can drift — it currently describes v0.16 behavior.

<!-- ZHARNESS:BEGIN -->
@AGENTS.md
<!-- ZHARNESS:END -->
