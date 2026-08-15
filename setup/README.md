# Claude Code Setup

Bootstrap kit for a new machine. Installs global `~/.claude/` config from this repo.

## Quick start

```bash
bash setup/install.sh
```

After install: edit `~/.claude/settings.json` → set `ANTHROPIC_AUTH_TOKEN`.

## Contents

| File | Installed to | Notes |
| :--- | :--- | :--- |
| `CLAUDE.md` | `~/.claude/CLAUDE.md` | Global rules and workflow |
| `hooks/` | `~/.claude/hooks/` | 5 hook scripts + 5 lib modules |
| `settings.json` | `~/.claude/settings.json` | Template; only copied if file doesn't exist |
| `../rules/*.md` | `~/.claude/rules/` | Karpathy, English coaching, AskUserQuestion |

## Prerequisites

- **Node.js** v18+ (required by hooks)
- **jq** (required by statusline) — `brew install jq` or `apt install jq`
- SSH key configured for `git@github.com:therealtinhtute/mono-harness.git`

## Hooks overview

| Hook | Event | What it does |
| :--- | :--- | :--- |
| `mandatory-instructions.cjs` | UserPromptSubmit | Injects today's date + `.kit/` path convention |
| `question-validator.cjs` | PostToolUse | Detects prose questions, logs violations |
| `privacy-guard.cjs` | PreToolUse | Blocks reads of sensitive files (`.env`, keys, etc.) |
| `scout-block.cjs` | PreToolUse | Blocks heavy directories via `~/.claude/.orkitignore` |
| `post-compact-reminder.sh` | PostCompact | Re-injects Hard Rules after context compaction |

from therealTINHTUTE with love
