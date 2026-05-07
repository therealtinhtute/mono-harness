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

# Copy src→dest only if contents differ. Backs up dest before overwrite.
sync_file() {
  local src="$1"
  local dest="$2"
  if [ -e "$dest" ] && cmp -s "$src" "$dest"; then
    return 0
  fi
  backup "$dest"
  cp "$src" "$dest"
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
sync_file "$REPO_DIR/setup/CLAUDE.md" "$CLAUDE_DIR/CLAUDE.md"
green "✓ CLAUDE.md installed"

# --- Step 3: Rules ---
for f in "$REPO_DIR/rules/"*.md; do
  [ -f "$f" ] || continue
  sync_file "$f" "$CLAUDE_DIR/rules/$(basename "$f")"
done
green "✓ Rules installed ($(ls "$CLAUDE_DIR/rules/"*.md 2>/dev/null | wc -l | tr -d ' ') files)"

# --- Step 4: Hooks ---
for f in "$REPO_DIR/setup/hooks/"*; do
  [ -f "$f" ] || continue
  sync_file "$f" "$CLAUDE_DIR/hooks/$(basename "$f")"
done
for f in "$REPO_DIR/setup/hooks/lib/"*; do
  [ -f "$f" ] || continue
  sync_file "$f" "$CLAUDE_DIR/hooks/lib/$(basename "$f")"
done
# Make all executable hook scripts runnable
find "$CLAUDE_DIR/hooks" -maxdepth 1 -type f \( -name "*.sh" -o -name "*.cjs" \) -exec chmod +x {} +
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

# --- Step 8: Version stamp ---
if git -C "$REPO_DIR" rev-parse HEAD >/dev/null 2>&1; then
  REPO_REV="$(git -C "$REPO_DIR" rev-parse HEAD)"
  printf '%s\n' "$REPO_REV" > "$CLAUDE_DIR/.bootstrap-version"
  green "✓ Bootstrap version stamped ($REPO_REV)"
fi

# --- Step 9: Verify ---
echo ""
echo "Verification"
echo "------------"
RULES_COUNT="$(ls "$CLAUDE_DIR/rules/"*.md 2>/dev/null | wc -l | tr -d ' ')"
HOOKS_COUNT="$(find "$CLAUDE_DIR/hooks" -maxdepth 1 -type f \( -name "*.sh" -o -name "*.cjs" \) 2>/dev/null | wc -l | tr -d ' ')"
SKILLS_COUNT="$(ls -d "$CLAUDE_DIR/skills/"*/ 2>/dev/null | wc -l | tr -d ' ')"
echo "  CLAUDE.md     : $([ -f "$CLAUDE_DIR/CLAUDE.md" ] && echo present || echo MISSING)"
echo "  rules/        : $RULES_COUNT files"
echo "  hooks/        : $HOOKS_COUNT scripts"
echo "  skills/       : $SKILLS_COUNT installed"
echo "  settings.json : $([ -f "$CLAUDE_DIR/settings.json" ] && echo present || echo MISSING)"
echo "  version       : $([ -f "$CLAUDE_DIR/.bootstrap-version" ] && cat "$CLAUDE_DIR/.bootstrap-version" || echo unknown)"
echo ""
green "Bootstrap complete."
echo ""
echo "Next steps:"
echo "  1. Set ANTHROPIC_AUTH_TOKEN in ~/.claude/settings.json"
echo "  2. Restart Claude Code"
echo ""
