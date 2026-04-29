# Claude Code Setup

Example configuration files for Claude Code.

## Files

- **`settings.json`** — Global settings template. Copy to `~/.claude/settings.json` and edit:
  - Replace `YOUR_AUTH_TOKEN_HERE` with your Anthropic API key.
  - Remove or adjust the `hooks` block if you haven't installed the required `.cjs` hook files.
  - Review `env` values (timeouts, model defaults) to match your preferences.

## Prerequisites

- **jq** — Required by the statusline script. Install via `brew install jq`.
- **Node.js** — v18+ required if using the hooks.

## Install

```bash
# Backup current settings
cp ~/.claude/settings.json ~/.claude/settings.json.backup

# Copy template
curl -fsSL https://raw.githubusercontent.com/therealtinhtute/skills/main/setup/settings.json \
  -o ~/.claude/settings.json

# Edit before restarting Claude Code
```

from therealTINHTUTE with love
