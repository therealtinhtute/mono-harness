#!/usr/bin/env bash
# sync-from-kit.sh — Copy skills from local incubator to this repo

KIT_SKILLS="/Users/tinhtute/Lab/orkit-tui/kit/skills"
REPO_SKILLS="$(cd "$(dirname "$0")/.." && pwd)/skills"

echo "Syncing skills from incubator..."
echo "  Source: $KIT_SKILLS"
echo "  Target: $REPO_SKILLS"
echo ""

# Find skills that exist in kit but not in repo
for skill_dir in "$KIT_SKILLS"/*/; do
  skill_name=$(basename "$skill_dir")
  target_dir="$REPO_SKILLS/$skill_name"

  if [ ! -d "$target_dir" ]; then
    echo "  [NEW] $skill_name — copying..."
    cp -R "$skill_dir" "$target_dir"
  elif [ "$skill_dir/SKILL.md" -nt "$target_dir/SKILL.md" ]; then
    echo "  [UPD] $skill_name — SKILL.md is newer, copying..."
    cp -R "$skill_dir""*" "$target_dir/"
  else
    echo "  [OK ] $skill_name — up to date"
  fi
done

echo ""
echo "Done. Run 'git status' to review changes before committing."
