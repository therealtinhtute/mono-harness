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
│   ├── bash-tui/
│   ├── brainstorm/
│   ├── git/
│   ├── handoff/
│   ├── interview/
│   ├── librarian/
│   ├── plan/
│   ├── prompt-leverage/
│   ├── review/
│   ├── skill-creator/
│   ├── spec/
│   ├── turbo-mono-platform/
│   └── watzup/
├── rules/                  # Source for global Claude Code rules (installed to ~/.claude/rules/)
│   ├── ask-user-question.md   # AskUserQuestion enforcement
│   ├── english.md             # English coaching
│   └── karpathy-guidelines.md # Karpathy coding principles
├── scripts/                # Repo utility scripts
│   ├── sync-from-kit.sh    # Incubator → repo sync
│   └── setup-statusline.sh # Statusline installer
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

Canonical order for complex work:
```
brainstorm → spec → plan → interview → implement → review/check → handoff/watzup
```
- `brainstorm` — explore options, evaluate trade-offs
- `spec` — lock requirements into `.planning/SPEC.md`
- `plan` — generate executable phase plans from spec
- `interview` — validate plan before implementation
- implement — actual coding
- `review` / `check` — pre-commit gate
- `handoff` / `watzup` — session close-out

## Architecture Notes

- **Stable vs. incubator:** This repo is the stable release. The incubator is at `/home/tinhpt/Lab/orkit-tui/kit/skills/` (local) or the equivalent on macOS.
- **Skill format:** All skills follow the `skills.sh` standard — YAML frontmatter with `name` and `description`, imperative instructions, optional `references/` and `scripts/` directories.
- **rules/ directory:** Source-of-truth for rules installed to `~/.claude/rules/`. Keep in sync with installed versions.
- **Private repo:** Installable via SSH (`git@github.com:therealtinhtute/skills.git`) as long as local SSH keys are configured.
