#!/bin/bash
# Clean up merged branches
# Usage: ./branch-cleanup.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") [options]

Clean up merged branches (local and remote)

Options:
  -h, --help       Show this help message
  -y, --yes        Skip confirmation prompts
  -r, --remote     Also delete remote branches
  -n, --dry-run    Show what would be deleted

Examples:
  ./branch-cleanup.sh
  ./branch-cleanup.sh --yes --remote
  ./branch-cleanup.sh --dry-run

from therealTINHTUTE with love
USAGE
}

function get_main_branch() {
  if git show-ref --verify --quiet refs/heads/main; then
    echo "main"
  elif git show-ref --verify --quiet refs/heads/master; then
    echo "master"
  else
    echo "main"
  fi
}

function list_merged_branches() {
  local main_branch="$1"
  
  echo "🔍 Finding merged branches..."
  
  # Get merged branches, exclude main/master and current branch
  git branch --merged "$main_branch" | \
    grep -v "^\*" | \
    grep -v "main" | \
    grep -v "master" | \
    sed 's/^[[:space:]]*//'
}

function confirm_deletion() {
  local branch="$1"
  
  read -p "Delete branch '$branch'? (y/N) " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    return 0
  else
    return 1
  fi
}

function delete_local_branch() {
  local branch="$1"
  
  if git branch -d "$branch" 2>&1; then
    echo "  ✅ Deleted local: $branch"
    return 0
  else
    echo "  ❌ Failed to delete: $branch"
    return 1
  fi
}

function delete_remote_branch() {
  local branch="$1"
  
  if git push origin --delete "$branch" 2>&1; then
    echo "  ✅ Deleted remote: $branch"
    return 0
  else
    echo "  ⚠️  Remote branch not found or already deleted: $branch"
    return 0
  fi
}

function main() {
  local skip_confirm="false"
  local delete_remote="false"
  local dry_run="false"

  # Parse arguments
  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        show_usage
        exit 0
        ;;
      -y|--yes)
        skip_confirm="true"
        shift
        ;;
      -r|--remote)
        delete_remote="true"
        shift
        ;;
      -n|--dry-run)
        dry_run="true"
        shift
        ;;
      *)
        echo "❌ Unknown option: $1"
        show_usage
        exit 1
        ;;
    esac
  done

  echo "🚀 Branch cleanup"
  echo ""

  local main_branch=$(get_main_branch)
  echo "Main branch: $main_branch"
  echo ""

  # Get merged branches
  local merged_branches=$(list_merged_branches "$main_branch")
  
  if [ -z "$merged_branches" ]; then
    echo "✅ No merged branches to clean up"
    exit 0
  fi

  echo "Merged branches:"
  echo "$merged_branches" | while read -r branch; do
    echo "  - $branch"
  done
  echo ""

  if [ "$dry_run" = "true" ]; then
    echo "🏁 Dry run complete (no branches deleted)"
    exit 0
  fi

  # Delete branches
  local deleted_count=0
  local skipped_count=0

  echo "$merged_branches" | while read -r branch; do
    if [ "$skip_confirm" = "false" ]; then
      if ! confirm_deletion "$branch"; then
        echo "  ⏭️  Skipped: $branch"
        ((skipped_count++))
        continue
      fi
    fi

    # Delete local branch
    if delete_local_branch "$branch"; then
      ((deleted_count++))
      
      # Delete remote if requested
      if [ "$delete_remote" = "true" ]; then
        delete_remote_branch "$branch"
      fi
    fi
  done

  echo ""
  echo "🎉 Cleanup complete"
  echo "  Deleted: $deleted_count"
  if [ "$skip_confirm" = "false" ]; then
    echo "  Skipped: $skipped_count"
  fi
}

main "$@"
