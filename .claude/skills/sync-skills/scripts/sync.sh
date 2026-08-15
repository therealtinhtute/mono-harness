#!/usr/bin/env bash
# sync.sh — clean-reinstall this repo's skills into ~/.claude/skills and
# upgrade the zharness CLI to the latest GitHub release.
#
# Usage: bash .claude/skills/sync-skills/scripts/sync.sh [--dry-run]
#
# Derives the managed skill list from skills/*/*/ so adding or removing a skill
# in the repo needs no edit here. Deletes with `trash`, never `rm`. Skips
# symlinked entries so externally managed skills survive.
set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")/../../../.." && pwd)"
SKILLS_DIR="$HOME/.claude/skills"
REMOTE="git@github.com:therealtinhtute/mono-harness.git"

DRY_RUN=0
[ "${1:-}" = "--dry-run" ] && DRY_RUN=1

green() { printf '\033[0;32m%s\033[0m\n' "$*"; }
yellow() { printf '\033[0;33m%s\033[0m\n' "$*"; }
red() { printf '\033[0;31m%s\033[0m\n' "$*"; }

if [ "$DRY_RUN" -eq 0 ] && ! command -v trash >/dev/null 2>&1; then
  red "error: trash is required (this repo never uses rm). install it, then re-run." >&2
  exit 1
fi

# --- Derive the managed skill list from the repo ---
SKILLS=()
while IFS= read -r dir; do
  [ -f "$dir/SKILL.md" ] || continue
  SKILLS+=("$(basename "$dir")")
done < <(find "$REPO_DIR/skills" -mindepth 2 -maxdepth 2 -type d | sort)

if [ "${#SKILLS[@]}" -eq 0 ]; then
  red "error: no skills with a SKILL.md found under $REPO_DIR/skills" >&2
  exit 1
fi
echo "repo:     $REPO_DIR"
echo "managed:  ${#SKILLS[@]} skills — ${SKILLS[*]}"
[ "$DRY_RUN" -eq 1 ] && yellow "mode:     DRY RUN (nothing is changed)"

# --- Clean ---
echo ""
echo "--- clean ---"
for s in "${SKILLS[@]}"; do
  target="$SKILLS_DIR/$s"
  if [ -L "$target" ]; then
    yellow "  skip (symlink, externally managed): $s"
  elif [ -d "$target" ]; then
    if [ "$DRY_RUN" -eq 1 ]; then
      echo "  would trash: $s"
    else
      trash "$target" && green "  trashed: $s"
    fi
  else
    echo "  absent: $s"
  fi
done

# --- Reinstall skills ---
echo ""
echo "--- reinstall skills ---"
if [ "$DRY_RUN" -eq 1 ]; then
  echo "  would run: npx skills add $REMOTE -a claude-code -g -y"
else
  npx skills add "$REMOTE" -a claude-code -g -y
fi

# --- Upgrade CLI ---
echo ""
echo "--- zharness CLI ---"
if [ "$DRY_RUN" -eq 1 ]; then
  echo "  would run: bash $REPO_DIR/scripts/install-zharness.sh"
else
  bash "$REPO_DIR/scripts/install-zharness.sh"
fi

# --- Verify ---
echo ""
echo "--- verify ---"
if [ "$DRY_RUN" -eq 1 ]; then
  yellow "  dry run — no verification to report"
  exit 0
fi

ZH="$(command -v zharness || echo "$HOME/.local/bin/zharness")"
fail=0
printf 'zharness      : %s\n' "$("$ZH" --version 2>/dev/null || echo MISSING)"
for s in "${SKILLS[@]}"; do
  if [ -e "$SKILLS_DIR/$s" ]; then
    printf 'skill ok      : %s\n' "$s"
  else
    printf 'skill MISSING : %s\n' "$s"
    fail=1
  fi
done
printf 'symlinks kept : %s\n' "$(find "$SKILLS_DIR" -maxdepth 1 -type l 2>/dev/null | wc -l | tr -d ' ')"

echo ""
if [ "$fail" -eq 1 ]; then
  red "sync incomplete — see MISSING above"
  exit 1
fi
green "sync complete."
