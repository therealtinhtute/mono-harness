#!/bin/bash
# Generate quality dashboard
# Usage: ./generate-dashboard.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILLS_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") [options]

Generate quality dashboard for all skills

Options:
  -h, --help       Show this help message
  -o, --output     Output file (default: .kit/reports/quality/dashboard.md)

Examples:
  ./generate-dashboard.sh
  ./generate-dashboard.sh --output custom-dashboard.md

from therealTINHTUTE with love
USAGE
}

function count_skills() {
  find "$SKILLS_DIR" -maxdepth 2 -name "SKILL.md" | wc -l
}

function count_examples() {
  local skill_file="$1"
  grep -c "^### Example" "$skill_file" 2>/dev/null || echo "0"
}

function count_scripts() {
  local skill_dir="$1"
  if [ -d "$skill_dir/scripts" ]; then
    find "$skill_dir/scripts" -name "*.sh" -type f | wc -l
  else
    echo "0"
  fi
}

function check_references() {
  local skill_dir="$1"
  if [ -d "$skill_dir/references" ]; then
    echo "✅"
  else
    echo "❌"
  fi
}

function get_line_count() {
  local skill_file="$1"
  wc -l < "$skill_file" | tr -d ' '
}

function generate_overview() {
  local total_skills=$(count_skills)
  local skills_with_examples=0
  local skills_with_scripts=0
  local skills_with_references=0
  
  for skill_dir in "$SKILLS_DIR"/*/; do
    if [ ! -f "$skill_dir/SKILL.md" ]; then
      continue
    fi
    
    local examples=$(count_examples "$skill_dir/SKILL.md")
    if [ "$examples" -gt 0 ]; then
      ((skills_with_examples++))
    fi
    
    local scripts=$(count_scripts "$skill_dir")
    if [ "$scripts" -gt 0 ]; then
      ((skills_with_scripts++))
    fi
    
    if [ -d "$skill_dir/references" ]; then
      ((skills_with_references++))
    fi
  done
  
  echo "## Overview"
  echo ""
  echo "**Generated**: $(date)"
  echo ""
  echo "| Metric | Count | Percentage |"
  echo "|--------|-------|------------|"
  echo "| Total Skills | $total_skills | 100% |"
  echo "| With Examples | $skills_with_examples | $((skills_with_examples * 100 / total_skills))% |"
  echo "| With Scripts | $skills_with_scripts | $((skills_with_scripts * 100 / total_skills))% |"
  echo "| With References | $skills_with_references | $((skills_with_references * 100 / total_skills))% |"
  echo ""
}

function generate_skill_table() {
  echo "## Skills Detail"
  echo ""
  echo "| Skill | Lines | Examples | Scripts | References | Status |"
  echo "|-------|-------|----------|---------|------------|--------|"
  
  for skill_dir in "$SKILLS_DIR"/*/; do
    if [ ! -f "$skill_dir/SKILL.md" ]; then
      continue
    fi
    
    local skill_name=$(basename "$skill_dir")
    local skill_file="$skill_dir/SKILL.md"
    local lines=$(get_line_count "$skill_file")
    local examples=$(count_examples "$skill_file")
    local scripts=$(count_scripts "$skill_dir")
    local refs=$(check_references "$skill_dir")
    
    local status="✅"
    if [ "$examples" -eq 0 ]; then
      status="⚠️"
    fi
    
    echo "| $skill_name | $lines | $examples | $scripts | $refs | $status |"
  done
  
  echo ""
}

function generate_quality_metrics() {
  echo "## Quality Metrics"
  echo ""
  
  local total_lines=0
  local total_examples=0
  local total_scripts=0
  local skill_count=0
  
  for skill_dir in "$SKILLS_DIR"/*/; do
    if [ ! -f "$skill_dir/SKILL.md" ]; then
      continue
    fi
    
    ((skill_count++))
    total_lines=$((total_lines + $(get_line_count "$skill_dir/SKILL.md")))
    total_examples=$((total_examples + $(count_examples "$skill_dir/SKILL.md")))
    total_scripts=$((total_scripts + $(count_scripts "$skill_dir")))
  done
  
  local avg_lines=$((total_lines / skill_count))
  local avg_examples=$((total_examples / skill_count))
  
  echo "| Metric | Value |"
  echo "|--------|-------|"
  echo "| Total Lines | $total_lines |"
  echo "| Average Lines per Skill | $avg_lines |"
  echo "| Total Examples | $total_examples |"
  echo "| Average Examples per Skill | $avg_examples |"
  echo "| Total Scripts | $total_scripts |"
  echo ""
}

function generate_recommendations() {
  echo "## Recommendations"
  echo ""
  
  local needs_examples=()
  local needs_scripts=()
  local too_long=()
  
  for skill_dir in "$SKILLS_DIR"/*/; do
    if [ ! -f "$skill_dir/SKILL.md" ]; then
      continue
    fi
    
    local skill_name=$(basename "$skill_dir")
    local skill_file="$skill_dir/SKILL.md"
    local examples=$(count_examples "$skill_file")
    local scripts=$(count_scripts "$skill_dir")
    local lines=$(get_line_count "$skill_file")
    
    if [ "$examples" -eq 0 ]; then
      needs_examples+=("$skill_name")
    fi
    
    if [ "$scripts" -eq 0 ]; then
      needs_scripts+=("$skill_name")
    fi
    
    if [ "$lines" -gt 400 ]; then
      too_long+=("$skill_name ($lines lines)")
    fi
  done
  
  if [ ${#needs_examples[@]} -gt 0 ]; then
    echo "### Skills Needing Examples"
    for item in "${needs_examples[@]}"; do
      echo "- $item"
    done
    echo ""
  fi

  if [ ${#needs_scripts[@]} -gt 0 ]; then
    echo "### Skills That Could Benefit from Scripts"
    for item in "${needs_scripts[@]}"; do
      echo "- $item"
    done
    echo ""
  fi

  if [ ${#too_long[@]} -gt 0 ]; then
    echo "### Skills Exceeding 400 Lines"
    for item in "${too_long[@]}"; do
      echo "- $item"
    done
    echo ""
  fi
  
  if [ ${#needs_examples[@]} -eq 0 ] && [ ${#needs_scripts[@]} -eq 0 ] && [ ${#too_long[@]} -eq 0 ]; then
    echo "✅ All skills meet quality standards"
    echo ""
  fi
}

function main() {
  local output_file=""

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

  if [ -z "$output_file" ]; then
    output_file=".kit/reports/quality/dashboard.md"
    mkdir -p "$(dirname "$output_file")"
  fi

  echo "📊 Generating quality dashboard..."

  {
    echo "---"
    echo "title: Skills Quality Dashboard"
    echo "description: Quality metrics and status for all skills"
    echo "status: active"
    echo "created: $(date +%Y-%m-%d)"
    echo "tags: [quality, dashboard, metrics]"
    echo "---"
    echo ""
    echo "# Skills Quality Dashboard"
    echo ""
    
    generate_overview
    generate_skill_table
    generate_quality_metrics
    generate_recommendations
    
    echo "from therealTINHTUTE with love"
  } > "$output_file"

  echo "✅ Dashboard generated: $output_file"
}

main "$@"
