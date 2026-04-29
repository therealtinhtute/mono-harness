#!/bin/bash
# Verify implementation matches plan
# Usage: ./plan-alignment.sh plan.md

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") <plan-file>

Verify implementation matches plan

Arguments:
  plan-file    Path to plan markdown file (required)

Options:
  -h, --help       Show this help message
  -o, --output     Output file (default: stdout)

Examples:
  ./plan-alignment.sh .kit/plans/20260416-week3-quality/plan.md
  ./plan-alignment.sh plan.md --output alignment-report.md

from therealTINHTUTE with love
USAGE
}

function extract_tasks_from_plan() {
  local plan_file="$1"
  
  # Extract checkbox items from plan
  grep -E '^\s*-\s*\[[ x]\]' "$plan_file" | sed 's/^\s*-\s*\[[ x]\]\s*//'
}

function check_file_exists() {
  local file="$1"
  
  if [ -f "$file" ]; then
    echo "✅"
  else
    echo "❌"
  fi
}

function analyze_plan_alignment() {
  local plan_file="$1"
  
  echo "## Plan Alignment Analysis"
  echo ""
  echo "**Plan**: \`$plan_file\`"
  echo "**Checked**: $(date)"
  echo ""
  
  local tasks=$(extract_tasks_from_plan "$plan_file")
  
  if [ -z "$tasks" ]; then
    echo "⚠️  No tasks found in plan"
    return 1
  fi
  
  echo "### Task Status"
  echo ""
  
  local total=0
  local completed=0
  
  while IFS= read -r task; do
    ((total++))
    
    # Check if task mentions files
    if echo "$task" | grep -qE '\.(md|sh|py|js|ts)'; then
      local file=$(echo "$task" | grep -oE '[a-zA-Z0-9/_.-]+\.(md|sh|py|js|ts)' | head -1)
      local status=$(check_file_exists "$file")
      
      if [ "$status" = "✅" ]; then
        ((completed++))
      fi
      
      echo "- [$status] $task"
    else
      echo "- [⏭️] $task"
    fi
  done <<< "$tasks"
  
  echo ""
  echo "### Summary"
  echo ""
  echo "- Total tasks: $total"
  echo "- Completed: $completed"
  echo "- Progress: $((completed * 100 / total))%"
  echo ""
}

function check_modified_files() {
  echo "## Modified Files"
  echo ""
  
  local files=$(git diff --name-only HEAD 2>/dev/null || echo "")
  
  if [ -z "$files" ]; then
    echo "No modified files"
  else
    echo "\`\`\`"
    echo "$files"
    echo "\`\`\`"
  fi
  
  echo ""
}

function generate_gaps() {
  local plan_file="$1"
  
  echo "## Gaps & Missing Items"
  echo ""
  
  local tasks=$(extract_tasks_from_plan "$plan_file")
  local gaps=0
  
  while IFS= read -r task; do
    if echo "$task" | grep -qE '\.(md|sh|py|js|ts)'; then
      local file=$(echo "$task" | grep -oE '[a-zA-Z0-9/_.-]+\.(md|sh|py|js|ts)' | head -1)
      
      if [ ! -f "$file" ]; then
        echo "- ❌ Missing: $task"
        ((gaps++))
      fi
    fi
  done <<< "$tasks"
  
  if [ $gaps -eq 0 ]; then
    echo "✅ No gaps found"
  fi
  
  echo ""
}

function main() {
  local plan_file=""
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
        plan_file="$1"
        shift
        ;;
    esac
  done

  if [ -z "$plan_file" ]; then
    echo "❌ Error: plan file required"
    show_usage
    exit 1
  fi

  if [ ! -f "$plan_file" ]; then
    echo "❌ Error: plan file not found: $plan_file"
    exit 1
  fi

  echo "🔍 Checking plan alignment..."
  echo ""

  # Generate report
  local report=$(cat <<REPORT
---
title: Plan Alignment Report
description: Implementation vs plan comparison
status: completed
created: $(date +%Y-%m-%d)
tags: [verify, alignment]
---

# Plan Alignment Report

$(analyze_plan_alignment "$plan_file")

$(check_modified_files)

$(generate_gaps "$plan_file")

from therealTINHTUTE with love
REPORT
)

  if [ -n "$output_file" ]; then
    echo "$report" > "$output_file"
    echo "✅ Report saved: $output_file"
  else
    echo "$report"
  fi
}

main "$@"
