#!/bin/bash
# Fast pre-commit validation
# Usage: ./pre-commit-check.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0")

Fast pre-commit validation for staged files

Options:
  -h, --help       Show this help message

Examples:
  ./pre-commit-check.sh

Exit codes:
  0: All checks passed
  1: Checks failed

from therealTINHTUTE with love
USAGE
}

function get_staged_files() {
  git diff --cached --name-only --diff-filter=ACM
}

function check_syntax() {
  local file="$1"
  local ext="${file##*.}"
  
  case "$ext" in
    js|jsx|ts|tsx)
      if command -v node &> /dev/null; then
        node --check "$file" 2>&1 && return 0 || return 1
      fi
      ;;
    py)
      if command -v python3 &> /dev/null; then
        python3 -m py_compile "$file" 2>&1 && return 0 || return 1
      fi
      ;;
    sh)
      if command -v bash &> /dev/null; then
        bash -n "$file" 2>&1 && return 0 || return 1
      fi
      ;;
  esac
  
  return 0
}

function scan_secrets() {
  local file="$1"
  
  local patterns=(
    'password\s*=\s*["\047][^"\047]+["\047]'
    'api[_-]?key\s*=\s*["\047][^"\047]+["\047]'
    'secret\s*=\s*["\047][^"\047]+["\047]'
    'token\s*=\s*["\047][^"\047]+["\047]'
    'AWS_ACCESS_KEY'
    'PRIVATE_KEY'
  )
  
  for pattern in "${patterns[@]}"; do
    if grep -qE "$pattern" "$file" 2>/dev/null; then
      return 1
    fi
  done
  
  return 0
}

function check_format() {
  local file="$1"
  local ext="${file##*.}"
  
  case "$ext" in
    js|jsx|ts|tsx|json|md)
      if command -v prettier &> /dev/null; then
        prettier --check "$file" 2>&1 && return 0 || return 1
      fi
      ;;
  esac
  
  return 0
}

function main() {
  # Parse arguments
  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        show_usage
        exit 0
        ;;
      *)
        echo "❌ Unknown option: $1"
        show_usage
        exit 1
        ;;
    esac
  done

  echo "🚀 Pre-commit checks"
  echo ""

  local files=$(get_staged_files)
  
  if [ -z "$files" ]; then
    echo "⚠️  No staged files"
    exit 0
  fi

  local failed=0
  local checked=0

  while IFS= read -r file; do
    if [ ! -f "$file" ]; then
      continue
    fi
    
    ((checked++))
    echo "Checking: $file"
    
    # Syntax check
    if ! check_syntax "$file"; then
      echo "  ❌ Syntax error"
      ((failed++))
      continue
    fi
    
    # Secret scan
    if ! scan_secrets "$file"; then
      echo "  ❌ Secret detected"
      ((failed++))
      continue
    fi
    
    # Format check
    if ! check_format "$file"; then
      echo "  ⚠️  Format issue"
    fi
    
    echo "  ✅ Pass"
  done <<< "$files"

  echo ""
  echo "📊 Summary"
  echo "=========="
  echo "Checked: $checked file(s)"
  
  if [ $failed -eq 0 ]; then
    echo "✅ All checks passed"
    exit 0
  else
    echo "❌ $failed file(s) failed"
    exit 1
  fi
}

main "$@"
