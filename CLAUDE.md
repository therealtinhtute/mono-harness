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
│   ├── investigator/
│   ├── media-processing/
│   ├── prompt-leverage/
│   ├── reviewer/
│   ├── skill-creator/
│   ├── strategist/
│   ├── turbo-mono-platform/
│   ├── verifier/
│   └── watzup/
├── rules/                  # Claude Code rules
│   └── english.md
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

## Architecture Notes

- **Stable vs. incubator:** This repo is the stable release. The incubator workspace is `/Users/tinhtute/Lab/orkit-tui/kit/skills/`.
- **Skill format:** All skills follow the `skills.sh` standard — YAML frontmatter with `name` and `description`, imperative instructions, and optional `references/` and `scripts/` directories.
- **Current repo shape:** This is a curated skills collection, not a full app/monorepo. Prefer lightweight docs and validation over heavy tooling.
- **Private repo:** Installable via SSH (`git@github.com:therealtinhtute/skills.git`) as long as local SSH keys are configured.
