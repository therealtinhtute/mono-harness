#!/bin/bash
# Safe merge with conflict detection
# Usage: ./safe-merge.sh feature-branch

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") <branch>

Safe merge with conflict detection and rollback

Arguments:
  branch       Branch to merge (required)

Options:
  -h, --help       Show this help message
  --no-ff          Force merge commit (no fast-forward)
  --squash         Squash commits before merge

Examples:
  ./safe-merge.sh feature-branch
  ./safe-merge.sh feature-branch --no-ff
  ./safe-merge.sh feature-branch --squash

from therealTINHTUTE with love
USAGE
}

function fetch_latest() {
  echo "📥 Fetching latest changes..."
  
  if git fetch origin; then
    echo "✅ Fetch complete"
    return 0
  else
    echo "❌ Fetch failed"
    return 1
  fi
}

function check_conflicts() {
  local branch="$1"
  
  echo "🔍 Checking for conflicts..."
  
  # Try merge with --no-commit to detect conflicts
  if git merge --no-commit --no-ff "$branch" 2>&1 | tee /tmp/merge-check.log; then
    # Abort the test merge
    git merge --abort 2>/dev/null || true
    echo "✅ No conflicts detected"
    return 0
  else
    # Check if it's a conflict or other error
    if grep -q "CONFLICT" /tmp/merge-check.log; then
      echo "❌ Conflicts detected:"
      grep "CONFLICT" /tmp/merge-check.log
      git merge --abort 2>/dev/null || true
      return 1
    else
      git merge --abort 2>/dev/null || true
      echo "❌ Merge check failed"
      return 1
    fi
  fi
}

function perform_merge() {
  local branch="$1"
  local no_ff="$2"
  local squash="$3"
  
  echo "🔀 Performing merge..."
  
  local merge_args=("merge")
  
  if [ "$no_ff" = "true" ]; then
    merge_args+=("--no-ff")
  fi
  
  if [ "$squash" = "true" ]; then
    merge_args+=("--squash")
  fi
  
  merge_args+=("$branch")
  
  if git "${merge_args[@]}"; then
    echo "✅ Merge successful"
    return 0
  else
    echo "❌ Merge failed"
    return 1
  fi
}

function verify_merge() {
  echo "🔎 Verifying merge..."
  
  # Check if working tree is clean
  if git diff --quiet && git diff --cached --quiet; then
    echo "✅ Working tree clean"
  else
    echo "⚠️  Uncommitted changes present"
  fi
  
  # Show merge commit
  local last_commit=$(git log -1 --oneline)
  echo "  Last commit: $last_commit"
  
  echo "✅ Merge verified"
  return 0
}

function rollback_merge() {
  echo "⏪ Rolling back merge..."
  
  if git merge --abort 2>/dev/null; then
    echo "✅ Merge aborted"
  elif git reset --hard HEAD 2>/dev/null; then
    echo "✅ Reset to HEAD"
  else
    echo "⚠️  Manual cleanup may be required"
  fi
}

function main() {
  local branch=""
  local no_ff="false"
  local squash="false"

  # Parse arguments
  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        show_usage
        exit 0
        ;;
      --no-ff)
        no_ff="true"
        shift
        ;;
      --squash)
        squash="true"
        shift
        ;;
      *)
        branch="$1"
        shift
        ;;
    esac
  done

  if [ -z "$branch" ]; then
    echo "❌ Error: branch name required"
    show_usage
    exit 1
  fi

  echo "🚀 Starting safe merge: $branch"
  echo ""

  # Step 1: Fetch latest
  fetch_latest || exit 1
  echo ""

  # Step 2: Check for conflicts
  if ! check_conflicts "$branch"; then
    echo ""
    echo "❌ Merge aborted due to conflicts"
    echo "Resolve conflicts manually and try again"
    exit 1
  fi
  echo ""

  # Step 3: Perform merge
  if ! perform_merge "$branch" "$no_ff" "$squash"; then
    echo ""
    rollback_merge
    exit 1
  fi
  echo ""

  # Step 4: Verify merge
  verify_merge || exit 1
  echo ""

  echo "🎉 Safe merge complete"
}

main "$@"
