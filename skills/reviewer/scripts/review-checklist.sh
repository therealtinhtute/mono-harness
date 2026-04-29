#!/bin/bash
# Interactive review checklist
# Usage: ./review-checklist.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0")

Interactive code review checklist

Options:
  -h, --help       Show this help message
  -o, --output     Output file (default: .kit/reports/review/YYYYMMDD-checklist.md)

Examples:
  ./review-checklist.sh
  ./review-checklist.sh --output custom-checklist.md

from therealTINHTUTE with love
USAGE
}

function prompt_check() {
  local category="$1"
  local item="$2"
  
  echo ""
  echo "[$category] $item"
  read -p "Status (✅ pass / ⚠️ warn / ❌ fail / ⏭️ skip): " -r response
  
  case "$response" in
    "✅"|"pass"|"p"|"y"|"yes")
      echo "✅"
      ;;
    "⚠️"|"warn"|"w")
      echo "⚠️"
      ;;
    "❌"|"fail"|"f"|"n"|"no")
      echo "❌"
      ;;
    *)
      echo "⏭️"
      ;;
  esac
}

function security_checklist() {
  echo "## Security Checklist"
  echo ""
  
  local checks=(
    "Input validation on all user inputs"
    "Authentication checks on protected routes"
    "Authorization boundaries enforced"
    "No SQL injection vulnerabilities"
    "No XSS vulnerabilities"
    "No hardcoded secrets or credentials"
    "Sensitive data not logged"
  )
  
  local results=()
  
  for check in "${checks[@]}"; do
    local result=$(prompt_check "Security" "$check")
    results+=("- [$result] $check")
  done
  
  printf '%s\n' "${results[@]}"
  echo ""
}

function performance_checklist() {
  echo "## Performance Checklist"
  echo ""
  
  local checks=(
    "No N+1 query patterns"
    "Database queries optimized"
    "No blocking operations in hot paths"
    "Memory leaks addressed"
    "Caching strategy appropriate"
  )
  
  local results=()
  
  for check in "${checks[@]}"; do
    local result=$(prompt_check "Performance" "$check")
    results+=("- [$result] $check")
  done
  
  printf '%s\n' "${results[@]}"
  echo ""
}

function architecture_checklist() {
  echo "## Architecture Checklist"
  echo ""
  
  local checks=(
    "YAGNI: No unnecessary features"
    "KISS: Solution is simple"
    "DRY: No code duplication"
    "Separation of concerns maintained"
    "API contracts correct"
    "Backward compatibility preserved"
  )
  
  local results=()
  
  for check in "${checks[@]}"; do
    local result=$(prompt_check "Architecture" "$check")
    results+=("- [$result] $check")
  done
  
  printf '%s\n' "${results[@]}"
  echo ""
}

function quality_checklist() {
  echo "## Code Quality Checklist"
  echo ""
  
  local checks=(
    "Naming is clear and descriptive"
    "Error handling at boundaries"
    "Type safety maintained"
    "Tests cover new behavior"
    "Documentation updated"
  )
  
  local results=()
  
  for check in "${checks[@]}"; do
    local result=$(prompt_check "Quality" "$check")
    results+=("- [$result] $check")
  done
  
  printf '%s\n' "${results[@]}"
  echo ""
}

function generate_summary() {
  local security="$1"
  local performance="$2"
  local architecture="$3"
  local quality="$4"
  
  echo "## Summary"
  echo ""
  
  local fail_count=$(echo -e "$security\n$performance\n$architecture\n$quality" | grep -c "❌" || echo "0")
  
  if [ "$fail_count" -gt 0 ]; then
    echo "❌ **REQUEST CHANGES** - $fail_count item(s) failed"
  else
    echo "✅ **APPROVE** - All checks passed or acceptable"
  fi
  
  echo ""
  echo "from therealTINHTUTE with love"
}

function main() {
  local output_file=""

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
      *)
        echo "❌ Unknown option: $1"
        show_usage
        exit 1
        ;;
    esac
  done

  # Set default output file
  if [ -z "$output_file" ]; then
    local date=$(date +%Y%m%d)
    output_file=".kit/reports/review/${date}-checklist.md"
    mkdir -p "$(dirname "$output_file")"
  fi

  echo "🔍 Interactive Review Checklist"
  echo "================================"
  echo ""
  echo "Press Enter to start..."
  read -r

  # Run checklists
  local security=$(security_checklist)
  local performance=$(performance_checklist)
  local architecture=$(architecture_checklist)
  local quality=$(quality_checklist)

  # Generate report
  {
    echo "---"
    echo "title: Review Checklist"
    echo "description: Interactive review checklist"
    echo "status: completed"
    echo "created: $(date +%Y-%m-%d)"
    echo "tags: [review, checklist]"
    echo "---"
    echo ""
    echo "# Review Checklist"
    echo ""
    echo "**Generated**: $(date)"
    echo ""
    echo "$security"
    echo "$performance"
    echo "$architecture"
    echo "$quality"
    generate_summary "$security" "$performance" "$architecture" "$quality"
  } > "$output_file"

  echo ""
  echo "✅ Checklist saved: $output_file"
}

main "$@"
