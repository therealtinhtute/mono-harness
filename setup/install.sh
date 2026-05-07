#!/usr/bin/env bash
# install.sh — Bootstrap Claude Code global config from this repo.
# Safe to run multiple times. Backs up files before overwriting.
#
# Usage:
#   bash setup/install.sh
#
# What this does:
#   1. Copies CLAUDE.md, rules/, hooks/ to ~/.claude/
#   2. Installs statusline
#   3. Handles settings.json (copy if new, show diff if existing)
#   4. Installs all skills via npx skills add

set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CLAUDE_DIR="$HOME/.claude"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"

green() { echo -e "\033[0;32m$*\033[0m"; }
yellow() { echo -e "\033[0;33m$*\033[0m"; }
red() { echo -e "\033[0;31m$*\033[0m"; }

backup() {
  local src="$1"
  if [ -e "$src" ]; then
    cp -r "$src" "${src}.backup.${TIMESTAMP}"
    yellow "  ↩ backed up: $src"
  fi
}

echo ""
echo "Claude Code Bootstrap"
echo "====================="
echo "Repo:   $REPO_DIR"
echo "Target: $CLAUDE_DIR"
echo ""

# --- Step 1: Create directories ---
mkdir -p "$CLAUDE_DIR/rules" "$CLAUDE_DIR/hooks/lib"
green "✓ Directories ready"

# --- Step 2: CLAUDE.md ---
backup "$CLAUDE_DIR/CLAUDE.md"
cp "$REPO_DIR/setup/CLAUDE.md" "$CLAUDE_DIR/CLAUDE.md"
green "✓ CLAUDE.md installed"

# --- Step 3: Rules ---
for f in "$REPO_DIR/rules/"*.md; do
  [ -f "$f" ] || continue
  name="$(basename "$f")"
  cp "$f" "$CLAUDE_DIR/rules/$name"
done
green "✓ Rules installed ($(ls "$CLAUDE_DIR/rules/"*.md 2>/dev/null | wc -l | tr -d ' ') files)"

# --- Step 4: Hooks ---
for f in "$REPO_DIR/setup/hooks/"*; do
  [ -f "$f" ] || continue
  name="$(basename "$f")"
  cp "$f" "$CLAUDE_DIR/hooks/$name"
done
for f in "$REPO_DIR/setup/hooks/lib/"*; do
  [ -f "$f" ] || continue
  name="$(basename "$f")"
  cp "$f" "$CLAUDE_DIR/hooks/lib/$name"
done
chmod +x "$CLAUDE_DIR/hooks/post-compact-reminder.sh" 2>/dev/null || true
green "✓ Hooks installed"

# --- Step 5: settings.json ---
if [ -f "$CLAUDE_DIR/settings.json" ]; then
  yellow ""
  yellow "⚠  settings.json already exists — not overwriting."
  yellow "   Review the diff below and manually merge new fields:"
  yellow ""
  diff "$REPO_DIR/setup/settings.json" "$CLAUDE_DIR/settings.json" || true
  yellow ""
  yellow "   Template: $REPO_DIR/setup/settings.json"
  yellow "   Live:     $CLAUDE_DIR/settings.json"
else
  cp "$REPO_DIR/setup/settings.json" "$CLAUDE_DIR/settings.json"
  green "✓ settings.json installed"
  yellow ""
  yellow "  → Edit $CLAUDE_DIR/settings.json"
  yellow "     Set ANTHROPIC_AUTH_TOKEN to your actual token."
  if grep -q "kypicc.spendchai.com" "$REPO_DIR/setup/settings.json" 2>/dev/null; then
    : # custom base url already set
  else
    yellow "     Set ANTHROPIC_BASE_URL if you use a custom proxy."
  fi
fi

# --- Step 6: Statusline ---
if [ -f "$REPO_DIR/scripts/setup-statusline.sh" ]; then
  bash "$REPO_DIR/scripts/setup-statusline.sh"
  green "✓ Statusline installed"
fi

# --- Step 7: Skills ---
echo ""
echo "Installing skills..."
npx skills add "git@github.com:therealtinhtute/skills.git" -a claude-code -g -y
green "✓ Skills installed"

echo ""
green "Bootstrap complete."
echo ""
echo "Next steps:"
echo "  1. Set ANTHROPIC_AUTH_TOKEN in ~/.claude/settings.json"
echo "  2. Restart Claude Code"
echo ""
