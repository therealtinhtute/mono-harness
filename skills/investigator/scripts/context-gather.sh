#!/bin/bash
# Gather context for a feature
# Usage: ./context-gather.sh "user authentication"

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") <feature>

Gather context for a feature or topic

Arguments:
  feature     Feature or topic to gather context for

Options:
  -h, --help       Show this help message
  -o, --output     Output directory (default: .kit/context/YYYYMMDD-slug/)

Examples:
  ./context-gather.sh "user authentication"
  ./context-gather.sh "payment processing" --output .kit/context/custom/

from therealTINHTUTE with love
USAGE
}

function slugify() {
  echo "$1" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -cd '[:alnum:]-'
}

function find_related_files() {
  local feature="$1"
  
  echo "## Related Files"
  echo ""
  
  # Search by filename
  local files=$(find . -type f -iname "*${feature}*" 2>/dev/null)
  
  if [ -n "$files" ]; then
    echo "### By Filename"
    echo "$files" | while read -r file; do
      echo "- \`$file\`"
    done
    echo ""
  fi
  
  # Search by content
  local content_files=$(grep -rl --include="*.js" --include="*.ts" --include="*.py" --include="*.go" \
    -i "$feature" . 2>/dev/null | head -n 20)
  
  if [ -n "$content_files" ]; then
    echo "### By Content"
    echo "$content_files" | while read -r file; do
      echo "- \`$file\`"
    done
    echo ""
  fi
}

function extract_key_functions() {
  local feature="$1"
  
  echo "## Key Functions"
  echo ""
  
  local functions=$(grep -rn --include="*.js" --include="*.ts" --include="*.py" --include="*.go" \
    -E "(function|class|def|func).*${feature}" . 2>/dev/null | head -n 10)
  
  if [ -z "$functions" ]; then
    echo "No functions found"
  else
    echo "$functions" | while IFS=: read -r file line content; do
      echo "### \`$file:$line\`"
      echo "\`\`\`"
      echo "$(echo "$content" | sed 's/^[[:space:]]*//')"
      echo "\`\`\`"
      echo ""
    done
  fi
  
  echo ""
}

function build_dependency_graph() {
  local feature="$1"
  
  echo "## Dependencies"
  echo ""
  
  # Find imports in related files
  local related_files=$(grep -rl --include="*.js" --include="*.ts" --include="*.py" \
    -i "$feature" . 2>/dev/null | head -n 5)
  
  if [ -z "$related_files" ]; then
    echo "No dependencies found"
  else
    echo "$related_files" | while read -r file; do
      echo "### \`$file\`"
      grep -E "(^import |^from .* import|require\()" "$file" 2>/dev/null | head -n 10 | while read -r import; do
        echo "- $(echo "$import" | sed 's/^[[:space:]]*//')"
      done
      echo ""
    done
  fi
  
  echo ""
}

function generate_summary() {
  local feature="$1"
  
  echo "## Summary"
  echo ""
  echo "Context gathered for: **$feature**"
  echo ""
  echo "**Generated**: $(date)"
  echo ""
  echo "### Next Steps"
  echo "1. Review related files"
  echo "2. Understand key functions"
  echo "3. Map dependencies"
  echo "4. Plan implementation"
  echo ""
}

function main() {
  local feature=""
  local output_dir=""

  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        show_usage
        exit 0
        ;;
      -o|--output)
        output_dir="$2"
        shift 2
        ;;
      *)
        feature="$1"
        shift
        ;;
    esac
  done

  if [ -z "$feature" ]; then
    echo "❌ Error: feature required"
    show_usage
    exit 1
  fi

  local slug=$(slugify "$feature")
  
  if [ -z "$output_dir" ]; then
    local date=$(date +%Y%m%d)
    output_dir=".kit/context/${date}-${slug}"
  fi

  echo "🔍 Gathering context: $feature"
  echo ""

  mkdir -p "$output_dir"

  # Generate context report
  {
    echo "---"
    echo "title: Context - $feature"
    echo "description: Context gathering for $feature"
    echo "status: completed"
    echo "created: $(date +%Y-%m-%d)"
    echo "tags: [context, $slug]"
    echo "---"
    echo ""
    echo "# Context Report: $feature"
    echo ""
    
    find_related_files "$feature"
    extract_key_functions "$feature"
    build_dependency_graph "$feature"
    generate_summary "$feature"
    
    echo "from therealTINHTUTE with love"
  } > "$output_dir/context.md"

  echo "✅ Context report saved: $output_dir/context.md"
}

main "$@"
