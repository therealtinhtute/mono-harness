#!/bin/bash
# Generate structured review report
# Usage: ./review-report.sh [commit-range]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") [commit-range]

Generate structured code review report

Arguments:
  commit-range    Git commit range (default: HEAD)

Options:
  -h, --help       Show this help message
  -o, --output     Output file (default: .kit/reports/review/YYYYMMDD-review.md)
  -f, --format     Output format: markdown|json (default: markdown)

Examples:
  ./review-report.sh
  ./review-report.sh HEAD~5..HEAD
  ./review-report.sh main..feature-branch
  ./review-report.sh --output custom-review.md

from therealTINHTUTE with love
USAGE
}

function get_changed_files() {
  local range="$1"
  git diff --name-only "$range" 2>/dev/null || git diff --name-only HEAD
}

function analyze_security() {
  local files="$1"
  
  echo "## Security Analysis"
  echo ""
  
  local issues=0
  
  # Check for common security patterns
  while IFS= read -r file; do
    if [ ! -f "$file" ]; then continue; fi
    
    # SQL injection patterns
    if grep -nE "(execute|query).*\+.*\$|\".*\$.*\"" "$file" 2>/dev/null; then
      echo "🔴 **CRITICAL**: Potential SQL injection in \`$file\`"
      ((issues++))
    fi
    
    # XSS patterns
    if grep -nE "innerHTML|dangerouslySetInnerHTML" "$file" 2>/dev/null; then
      echo "🟠 **MAJOR**: Potential XSS in \`$file\`"
      ((issues++))
    fi
    
    # Hardcoded secrets
    if grep -nE "password\s*=|api[_-]?key\s*=|secret\s*=" "$file" 2>/dev/null; then
      echo "🔴 **CRITICAL**: Potential hardcoded secret in \`$file\`"
      ((issues++))
    fi
  done <<< "$files"
  
  if [ $issues -eq 0 ]; then
    echo "✅ No security issues detected"
  fi
  
  echo ""
}

function analyze_performance() {
  local files="$1"
  
  echo "## Performance Analysis"
  echo ""
  
  local issues=0
  
  while IFS= read -r file; do
    if [ ! -f "$file" ]; then continue; fi
    
    # N+1 query patterns
    if grep -nE "for.*in.*\n.*query|while.*\n.*query" "$file" 2>/dev/null; then
      echo "🟠 **MAJOR**: Potential N+1 query in \`$file\`"
      ((issues++))
    fi
    
    # Blocking operations
    if grep -nE "sleep|setTimeout.*0|while.*true" "$file" 2>/dev/null; then
      echo "🟡 **MINOR**: Blocking operation in \`$file\`"
      ((issues++))
    fi
  done <<< "$files"
  
  if [ $issues -eq 0 ]; then
    echo "✅ No performance issues detected"
  fi
  
  echo ""
}

function analyze_architecture() {
  local files="$1"
  
  echo "## Architecture Analysis"
  echo ""
  
  local issues=0
  
  # Check file count
  local file_count=$(echo "$files" | wc -l)
  if [ "$file_count" -gt 20 ]; then
    echo "🟡 **MINOR**: Large changeset ($file_count files) - consider splitting"
    ((issues++))
  fi
  
  if [ $issues -eq 0 ]; then
    echo "✅ Architecture looks good"
  fi
  
  echo ""
}

function analyze_quality() {
  local files="$1"
  
  echo "## Code Quality"
  echo ""
  
  local issues=0
  
  while IFS= read -r file; do
    if [ ! -f "$file" ]; then continue; fi
    
    # Long functions
    if grep -nE "function.*\{" "$file" 2>/dev/null | wc -l | grep -qE "[5-9][0-9]|[0-9]{3,}"; then
      echo "🟡 **MINOR**: Long functions in \`$file\`"
      ((issues++))
    fi
    
    # TODO/FIXME comments
    if grep -nE "TODO|FIXME" "$file" 2>/dev/null; then
      echo "💡 **SUGGESTION**: Unresolved TODOs in \`$file\`"
    fi
  done <<< "$files"
  
  if [ $issues -eq 0 ]; then
    echo "✅ Code quality looks good"
  fi
  
  echo ""
}

function generate_verdict() {
  local critical_count="$1"
  
  echo "## Verdict"
  echo ""
  
  if [ "$critical_count" -gt 0 ]; then
    echo "❌ **REQUEST CHANGES** - $critical_count critical issue(s) must be fixed"
  else
    echo "✅ **APPROVE** - No critical issues found"
  fi
  
  echo ""
  echo "from therealTINHTUTE with love"
}

function main() {
  local commit_range="HEAD"
  local output_file=""
  local format="markdown"

  # Parse arguments
  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        show_usage
        exit 0
        ;;
      -o|--output)
        output_file="$2"
        shift 2
        ;;
      -f|--format)
        format="$2"
        shift 2
        ;;
      *)
        commit_range="$1"
        shift
        ;;
    esac
  done

  # Set default output file
  if [ -z "$output_file" ]; then
    local date=$(date +%Y%m%d)
    output_file=".kit/reports/review/${date}-review.md"
    mkdir -p "$(dirname "$output_file")"
  fi

  echo "🔍 Generating review report..."
  echo ""

  # Get changed files
  local files=$(get_changed_files "$commit_range")
  
  if [ -z "$files" ]; then
    echo "❌ No files changed in range: $commit_range"
    exit 1
  fi

  # Generate report
  {
    echo "---"
    echo "title: Code Review Report"
    echo "description: Review for $commit_range"
    echo "status: completed"
    echo "created: $(date +%Y-%m-%d)"
    echo "tags: [review, automated]"
    echo "---"
    echo ""
    echo "# Code Review Report"
    echo ""
    echo "**Range**: \`$commit_range\`"
    echo "**Files changed**: $(echo "$files" | wc -l)"
    echo "**Generated**: $(date)"
    echo ""
    
    analyze_security "$files"
    analyze_performance "$files"
    analyze_architecture "$files"
    analyze_quality "$files"
    
    # Count critical issues
    local critical_count=$(grep -c "🔴" <<< "$(analyze_security "$files")" || echo "0")
    generate_verdict "$critical_count"
    
  } > "$output_file"

  echo "✅ Review report generated: $output_file"
}

main "$@"
