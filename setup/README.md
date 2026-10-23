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
| `CLAUDE.md` | `~/.claude/CLAUDE.md` | Voice, hard rules, working style |
| `hooks/` | `~/.claude/hooks/` | 2 hook scripts + 2 lib modules |
| `settings.json` | `~/.claude/settings.json` | Template; only copied if file doesn't exist |
| `statusline/` | `~/.claude/statusline.{sh,py}` | Statusline installers, run by `install.sh` (slim) or by hand (`setup-pulse.sh`) |
| `install-git-hooks.sh` | `.git/hooks/` of this repo | Pre-commit/pre-push guards; not run by `install.sh` |
| `rules/*.md` | `~/.claude/rules/` | English coaching; retires rules folded into `CLAUDE.md` |

## Prerequisites

- **Node.js** v18+ (required by hooks)
- **jq** (required by statusline) — `brew install jq` or `apt install jq`
- SSH key configured for `git@github.com:therealtinhtute/mono-harness.git`

## Hooks overview

| Hook | Event | What it does |
| :--- | :--- | :--- |
| `privacy-guard.cjs` | PreToolUse | Blocks reads of sensitive files (`.env`, keys, etc.); uses `lib/privacy-checker.cjs`, `lib/hook-logger.cjs` |
| `post-compact-reminder.sh` | PostCompact | Re-injects Hard Rules after context compaction |
| inline in `settings.json` | PreToolUse | Blocks `rm` in tool input; use `trash` |
| inline in `settings.json` | Stop | Reminds to use AskUserQuestion when a reply ends with `?` |

from therealTINHTUTE with love
