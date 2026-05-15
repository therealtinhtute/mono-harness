#!/bin/bash
# Create PR with standard template
# Usage: ./create-pr.sh "Feature title" "reviewer1,reviewer2"

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") <title> [reviewers]

Create a pull request with standard template

Arguments:
  title        PR title (required)
  reviewers    Comma-separated list of reviewers (optional)

Options:
  -h, --help       Show this help message
  -b, --base       Base branch (default: master)
  -d, --draft      Create as draft PR

Examples:
  ./create-pr.sh "Add user authentication"
  ./create-pr.sh "Fix login bug" "reviewer1,reviewer2"
  ./create-pr.sh "WIP: New feature" --draft

from therealTINHTUTE with love
USAGE
}

function get_current_branch() {
  git branch --show-current
}

function get_base_branch() {
  # Try to detect main branch
  if git show-ref --verify --quiet refs/heads/main; then
    echo "main"
  elif git show-ref --verify --quiet refs/heads/master; then
    echo "master"
  else
    echo "main"
  fi
}

function generate_pr_body() {
  local current_branch="$1"
  local base_branch="$2"

  echo "## Summary"
  echo ""
  
  # Get commit messages since base
  git log --pretty=format:"- %s" "$base_branch..$current_branch"
  
  echo ""
  echo ""
  echo "## Test plan"
  echo ""
  echo "- [ ] Tests pass locally"
  echo "- [ ] Manual testing completed"
  echo "- [ ] No breaking changes"
  echo ""
  echo "from therealTINHTUTE with love"
}

function push_branch() {
  local branch="$1"
  
  echo "📤 Pushing branch: $branch"
  
  if git push -u origin "$branch" 2>&1; then
    echo "✅ Branch pushed"
    return 0
  else
    echo "❌ Failed to push branch"
    return 1
  fi
}

function create_pull_request() {
  local title="$1"
  local base="$2"
  local reviewers="$3"
  local draft="$4"
  local current_branch=$(get_current_branch)
  
  echo "🔨 Creating pull request..."
  
  local body=$(generate_pr_body "$current_branch" "$base")
  
  local gh_args=(
    "pr" "create"
    "--title" "$title"
    "--body" "$body"
    "--base" "$base"
  )
  
  if [ "$draft" = "true" ]; then
    gh_args+=("--draft")
  fi
  
  if [ -n "$reviewers" ]; then
    gh_args+=("--reviewer" "$reviewers")
  fi
  
  if gh "${gh_args[@]}"; then
    echo "✅ Pull request created"
    return 0
  else
    echo "❌ Failed to create pull request"
    return 1
  fi
}

function main() {
  local title=""
  local reviewers=""
  local base=""
  local draft="false"

  # Parse arguments
  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        show_usage
        exit 0
        ;;
      -b|--base)
        base="$2"
        shift 2
        ;;
      -d|--draft)
        draft="true"
        shift
        ;;
      *)
        if [ -z "$title" ]; then
          title="$1"
        else
          reviewers="$1"
        fi
        shift
        ;;
    esac
  done

  if [ -z "$title" ]; then
    echo "❌ Error: PR title required"
    show_usage
    exit 1
  fi

  # Check if gh CLI is available
  if ! command -v gh &> /dev/null; then
    echo "❌ Error: GitHub CLI (gh) not found"
    echo "Install: brew install gh"
    exit 1
  fi

  # Get base branch if not specified
  if [ -z "$base" ]; then
    base=$(get_base_branch)
  fi

  local current_branch=$(get_current_branch)
  
  echo "🚀 Creating PR: $title"
  echo "  Branch: $current_branch → $base"
  if [ -n "$reviewers" ]; then
    echo "  Reviewers: $reviewers"
  fi
  echo ""

  # Step 1: Push branch
  push_branch "$current_branch" || exit 1
  echo ""

  # Step 2: Create PR
  create_pull_request "$title" "$base" "$reviewers" "$draft" || exit 1
  echo ""

  echo "🎉 PR creation complete"
}

main "$@"
