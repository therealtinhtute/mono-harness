# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is the personal skills repository for `therealtinhtute` — a `skills.sh`-compatible collection of agent skills for Claude Code and other AI agents.

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
│   │   └── watzup/
│   ├── shipping/           # Build & ship code
│   │   ├── create-cli/
│   │   └── turbo-mono-platform/
│   └── craft/              # Research, writing, meta-skills
│       ├── write/
│       ├── librarian/
│       ├── create-skill/
│       └── prompt-leverage/
├── rules/                  # Source for global Claude Code rules (installed to ~/.claude/rules/)
│   ├── ask-user-question.md   # AskUserQuestion enforcement
│   ├── english.md             # English coaching
│   └── karpathy-guidelines.md # Karpathy coding principles
├── docs/                   # Repo-wide reference docs
│   └── prompt-engineering-principles.md  # Prompting principles for writing skills/rules
├── scripts/                # Repo utility scripts
│   ├── sync-from-kit.sh       # Incubator → repo sync
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
npx skills add git@github.com:therealtinhtute/skills.git --list

# Install all skills globally for Claude Code
npx skills add git@github.com:therealtinhtute/skills.git -a claude-code -g -y

# Sync changes from local incubator
bash scripts/sync-from-kit.sh
```

## Skill Pipeline

Two entry points:
```
watzup → work → check → git → handoff          (resume existing work)
brainstorm → to-plan → work → check → git → handoff  (new work)
```
- `watzup` — recap branch state, committed + uncommitted changes, handoff context, recommend next action (session start)
- `brainstorm` — explore options and lock requirements into `.kit/planning/SPEC.md` (4 modes: explore, lock-from-idea, lock-from-files, refine)
- `to-plan` — generate executable phase plans from the locked spec
- `work` — execute the plan wave-by-wave, verify per task, route to `check` as the phase gate
- `check` — pre-commit gate and post-implementation review (also invoked per phase by `work`)
- `git` / `handoff` — session close-out

`interview` is optional — use to grill fuzzy intent into a clear goal, or to validate an existing plan before `work`. Can sit between `brainstorm` and `to-plan`, or between `to-plan` and `work`.

## Prompt Engineering Reference

When writing or editing skills (SKILL.md), rules (rules/*.md), or any agent instruction file, read `docs/prompt-engineering-principles.md` first. It covers: context engineering principles, formatting syntax (XML vs Markdown, bullets vs paragraphs), language rules, few-shot patterns, anti-patterns, and cross-model awareness.

## Architecture Notes

- **Stable vs. incubator:** This repo is the stable release. The incubator is at `/home/tinhpt/Lab/orkit-tui/kit/skills/` (local) or the equivalent on macOS.
- **Skill format:** All skills follow the `skills.sh` standard — YAML frontmatter with `name` and `description`, imperative instructions, optional `references/` and `scripts/` directories.
- **rules/ directory:** Source-of-truth for rules installed to `~/.claude/rules/`. Keep in sync with installed versions.
- **Private repo:** Installable via SSH (`git@github.com:therealtinhtute/skills.git`) as long as local SSH keys are configured.
