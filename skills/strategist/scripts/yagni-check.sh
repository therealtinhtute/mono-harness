#!/bin/bash
# Validate YAGNI/KISS/DRY principles
# Usage: ./yagni-check.sh [files...]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") [files...]

Check for YAGNI/KISS/DRY violations

Arguments:
  files...    Files to check (default: all modified files)

Options:
  -h, --help       Show this help message

Examples:
  ./yagni-check.sh
  ./yagni-check.sh src/**/*.js

from therealTINHTUTE with love
USAGE
}

function check_yagni() {
  local file="$1"
  local violations=()
  
  # Unused abstractions
  if grep -qE "abstract class|interface.*extends.*extends" "$file" 2>/dev/null; then
    violations+=("Potential over-abstraction detected")
  fi
  
  # Premature optimization
  if grep -qE "cache|memo|optimize" "$file" 2>/dev/null; then
    violations+=("Potential premature optimization")
  fi
  
  # Feature flags for single use
  if grep -qE "feature.*flag|toggle" "$file" 2>/dev/null; then
    violations+=("Feature flag - ensure it's needed")
  fi
  
  printf '%s\n' "${violations[@]}"
}

function check_kiss() {
  local file="$1"
  local violations=()
  
  # Complex conditionals
  local complex_lines=$(grep -cE "if.*&&.*&&|if.*\|\|.*\|\|" "$file" 2>/dev/null || echo "0")
  if [ "$complex_lines" -gt 3 ]; then
    violations+=("Complex conditionals ($complex_lines lines)")
  fi
  
  # Deep nesting
  local indent_depth=$(grep -oE "^[[:space:]]+" "$file" 2>/dev/null | awk '{print length}' | sort -rn | head -1)
  if [ -n "$indent_depth" ] && [ "$indent_depth" -gt 24 ]; then
    violations+=("Deep nesting (${indent_depth} spaces)")
  fi
  
  printf '%s\n' "${violations[@]}"
}

function check_dry() {
  local file="$1"
  local violations=()
  
  # Duplicate code patterns (simple heuristic)
  local duplicate_lines=$(sort "$file" | uniq -d | wc -l)
  if [ "$duplicate_lines" -gt 10 ]; then
    violations+=("Potential code duplication ($duplicate_lines duplicate lines)")
  fi
  
  printf '%s\n' "${violations[@]}"
}

function check_file() {
  local file="$1"
  
  if [ ! -f "$file" ]; then
    return
  fi
  
  local yagni_issues=$(check_yagni "$file")
  local kiss_issues=$(check_kiss "$file")
  local dry_issues=$(check_dry "$file")
  
  if [ -n "$yagni_issues" ] || [ -n "$kiss_issues" ] || [ -n "$dry_issues" ]; then
    echo "File: $file"
    
    if [ -n "$yagni_issues" ]; then
      echo "  ⚠️  YAGNI violations:"
      echo "$yagni_issues" | sed 's/^/    /'
    fi
    
    if [ -n "$kiss_issues" ]; then
      echo "  ⚠️  KISS violations:"
      echo "$kiss_issues" | sed 's/^/    /'
    fi
    
    if [ -n "$dry_issues" ]; then
      echo "  ⚠️  DRY violations:"
      echo "$dry_issues" | sed 's/^/    /'
    fi
    
    echo ""
  fi
}

function main() {
  local files=()

  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        show_usage
        exit 0
        ;;
      *)
        files+=("$1")
        shift
        ;;
    esac
  done

  if [ ${#files[@]} -eq 0 ]; then
    mapfile -t files < <(git diff --name-only HEAD 2>/dev/null || echo "")
  fi

  if [ ${#files[@]} -eq 0 ]; then
    echo "❌ No files to check"
    exit 1
  fi

  echo "🔍 Checking ${#files[@]} file(s) for YAGNI/KISS/DRY violations..."
  echo ""

  local issue_count=0

  for file in "${files[@]}"; do
    local result=$(check_file "$file")
    if [ -n "$result" ]; then
      echo "$result"
      ((issue_count++))
    fi
  done

  if [ $issue_count -eq 0 ]; then
    echo "✅ No violations found"
    exit 0
  else
    echo "⚠️  Found potential violations in $issue_count file(s)"
    exit 0
  fi
}

main "$@"
