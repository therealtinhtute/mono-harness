# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is the personal skills repository for `therealtinhtute` — a `skills.sh`-compatible collection of agent skills for Claude Code and other AI agents.

## Project Structure

```
.
├── README.md              # Skill directory and install instructions
├── CLAUDE.md              # This file
├── skills/                # Installable agent skills
│   ├── bash-tui/
│   ├── brainstorm/
│   ├── git/
│   ├── handoff-manager/
│   ├── interviewer/
│   ├── investigator/
│   ├── media-processing/
│   ├── prompt-leverage/
│   ├── reviewer/
│   ├── skill-creator/
│   ├── strategist/
│   ├── turbo-mono-platform/
│   └── verifier/
└── scripts/               # Utility scripts (sync, validate, etc.)
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
- **Private repo:** Installable via SSH (`git@github.com:therealtinhtute/skills.git`) as long as local SSH keys are configured.
