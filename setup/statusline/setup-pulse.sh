#!/usr/bin/env bash
# setup-pulse.sh — Install the Python statusline for Claude Code
# Features: session bar, weekly bar, context bar, TPM, heartbeat, git branch

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$SCRIPT_DIR/statusline.py"
DEST_PY="${HOME}/.claude/statusline.py"
SETTINGS="${HOME}/.claude/settings.json"

echo "Installing Python statusline..."

# Copy Python script
cp "$SRC" "$DEST_PY"
chmod +x "$DEST_PY"
echo "  Installed script → ${DEST_PY}"

# Detect python
PYTHON=$(command -v python3 2>/dev/null || command -v python 2>/dev/null || echo "")
if [ -z "$PYTHON" ]; then
  echo "  WARNING: python3 not found — install Python 3.6+."
  PYTHON="python3"
fi

STATUS_CMD="${PYTHON} ${DEST_PY}"

# Update settings.json statusLine
if [ -f "$SETTINGS" ]; then
  if command -v jq &>/dev/null; then
    TMP=$(mktemp)
    jq --arg cmd "$STATUS_CMD" '.statusLine = {"type": "command", "command": $cmd}' "$SETTINGS" > "$TMP" && mv "$TMP" "$SETTINGS"
    echo "  Updated statusLine in ${SETTINGS}"
  else
    echo "  jq not found — add manually to ${SETTINGS}:"
    echo "    \"statusLine\": { \"type\": \"command\", \"command\": \"${STATUS_CMD}\" }"
  fi
else
  echo "  ${SETTINGS} not found — add manually:"
  echo "    \"statusLine\": { \"type\": \"command\", \"command\": \"${STATUS_CMD}\" }"
fi

# Install PostToolUse hook for heartbeat
echo ""
read -r -p "Install PostToolUse hook for live heartbeat? [y/N] " REPLY
if [[ "${REPLY}" =~ ^[Yy]$ ]]; then
  HOOK_CMD="${PYTHON} \"${DEST_PY}\" --hook-refresh"
  if [ -f "$SETTINGS" ] && command -v jq &>/dev/null; then
    TMP=$(mktemp)
    jq --arg cmd "$HOOK_CMD" '
      .hooks.PostToolUse //= [] |
      if (.hooks.PostToolUse | map(select(.hooks[]?.command? | test("statusline.py"))) | length) > 0
      then .
      else .hooks.PostToolUse += [{"matcher": "", "hooks": [{"type": "command", "command": $cmd}]}]
      end
    ' "$SETTINGS" > "$TMP" && mv "$TMP" "$SETTINGS"
    echo "  Installed PostToolUse hook."
  else
    echo "  Add hook manually to ${SETTINGS}:"
    echo "    \"hooks\": { \"PostToolUse\": [{\"matcher\": \"\", \"hooks\": [{\"type\": \"command\", \"command\": \"${HOOK_CMD}\"}]}] }"
  fi
fi

echo ""
echo "Done. Restart Claude Code to activate."
echo "Requires Python 3.6+ and Claude Code v2.1.80+ for session/weekly bars."
