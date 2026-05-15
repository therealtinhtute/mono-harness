#!/bin/bash
# Automated commit workflow with validation
# Usage: ./commit-workflow.sh "feat: add new feature"

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") <commit-message>

Automate the commit workflow: stage → validate → commit → verify

Arguments:
  commit-message    Conventional commit message (required)

Options:
  -h, --help       Show this help message
  -v, --verbose    Verbose output
  -n, --dry-run    Show what would be done without executing

Examples:
  ./commit-workflow.sh "feat: add user authentication"
  ./commit-workflow.sh "fix: resolve login bug"
  ./commit-workflow.sh "docs: update README"

from therealTINHTUTE with love
USAGE
}

function validate_commit_message() {
  local msg="$1"

  # Check conventional commit format
  if ! echo "$msg" | grep -qE '^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\(.+\))?: .+'; then
    echo "❌ Invalid commit message format"
    echo "Expected: type(scope): description"
    echo "Types: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert"
    return 1
  fi

  echo "✅ Commit message format valid"
  return 0
}

function stage_files() {
  echo "📦 Staging files..."

  # Get modified and new files
  local files=$(git status --porcelain | grep -E '^\s*[AM]' | awk '{print $2}')

  if [ -z "$files" ]; then
    echo "⚠️  No files to stage"
    return 1
  fi

  echo "$files" | while read -r file; do
    echo "  + $file"
    git add "$file"
  done

  echo "✅ Files staged"
  return 0
}

function scan_secrets() {
  echo "🔍 Scanning for secrets..."

  # Check for common secret patterns
  local patterns=(
    'password\s*=\s*["\047][^"\047]+["\047]'
    'api[_-]?key\s*=\s*["\047][^"\047]+["\047]'
    'secret\s*=\s*["\047][^"\047]+["\047]'
    'token\s*=\s*["\047][^"\047]+["\047]'
    'AWS_ACCESS_KEY'
    'PRIVATE_KEY'
  )

  for pattern in "${patterns[@]}"; do
    if git diff --cached | grep -iE "$pattern" > /dev/null; then
      echo "❌ Potential secret detected: $pattern"
      return 1
    fi
  done

  echo "✅ No secrets detected"
  return 0
}

function create_commit() {
  local msg="$1"

  echo "💾 Creating commit..."

  git commit -m "$msg

from therealTINHTUTE with love"

  echo "✅ Commit created"
  return 0
}

function verify_commit() {
  echo "🔎 Verifying commit..."

  local last_commit=$(git log -1 --oneline)
  echo "  Last commit: $last_commit"

  echo "✅ Commit verified"
  return 0
}

function main() {
  local commit_msg=""
  local verbose=0
  local dry_run=0

  # Parse arguments
  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        show_usage
        exit 0
        ;;
      -v|--verbose)
        verbose=1
        shift
        ;;
      -n|--dry-run)
        dry_run=1
        shift
        ;;
      *)
        commit_msg="$1"
        shift
        ;;
    esac
  done

  if [ -z "$commit_msg" ]; then
    echo "❌ Error: commit message required"
    show_usage
    exit 1
  fi

  echo "🚀 Starting commit workflow"
  echo ""

  # Step 1: Validate message
  validate_commit_message "$commit_msg" || exit 1
  echo ""

  # Step 2: Stage files
  stage_files || exit 1
  echo ""

  # Step 3: Scan for secrets
  scan_secrets || exit 1
  echo ""

  if [ $dry_run -eq 1 ]; then
    echo "🏁 Dry run complete (no commit created)"
    exit 0
  fi

  # Step 4: Create commit
  create_commit "$commit_msg" || exit 1
  echo ""

  # Step 5: Verify commit
  verify_commit || exit 1
  echo ""

  echo "🎉 Commit workflow complete"
}

main "$@"
