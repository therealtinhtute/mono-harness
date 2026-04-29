#!/bin/bash
# Intelligent codebase search
# Usage: ./smart-search.sh "authentication"

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") <query>

Intelligent multi-strategy codebase search

Arguments:
  query       Search query (required)

Options:
  -h, --help       Show this help message
  -l, --limit      Max results per strategy (default: 10)

Examples:
  ./smart-search.sh "authentication"
  ./smart-search.sh "user login" --limit 5

from therealTINHTUTE with love
USAGE
}

function search_filenames() {
  local query="$1"
  local limit="$2"
  
  echo "## File Name Matches"
  echo ""
  
  local results=$(find . -type f -iname "*${query}*" 2>/dev/null | head -n "$limit")
  
  if [ -z "$results" ]; then
    echo "No matches"
  else
    echo "$results" | while read -r file; do
      echo "- \`$file\`"
    done
  fi
  
  echo ""
}

function search_content() {
  local query="$1"
  local limit="$2"
  
  echo "## Content Matches"
  echo ""
  
  local results=$(grep -rn --include="*.js" --include="*.ts" --include="*.py" --include="*.go" \
    -i "$query" . 2>/dev/null | head -n "$limit")
  
  if [ -z "$results" ]; then
    echo "No matches"
  else
    echo "$results" | while IFS=: read -r file line content; do
      echo "- \`$file:$line\`"
      echo "  \`\`\`"
      echo "  $(echo "$content" | sed 's/^[[:space:]]*//')"
      echo "  \`\`\`"
    done
  fi
  
  echo ""
}

function search_symbols() {
  local query="$1"
  local limit="$2"
  
  echo "## Symbol Matches"
  echo ""
  
  # Search for function/class definitions
  local results=$(grep -rn --include="*.js" --include="*.ts" --include="*.py" --include="*.go" \
    -E "(function|class|def|func).*${query}" . 2>/dev/null | head -n "$limit")
  
  if [ -z "$results" ]; then
    echo "No matches"
  else
    echo "$results" | while IFS=: read -r file line content; do
      echo "- \`$file:$line\`"
      echo "  \`\`\`"
      echo "  $(echo "$content" | sed 's/^[[:space:]]*//')"
      echo "  \`\`\`"
    done
  fi
  
  echo ""
}

function main() {
  local query=""
  local limit=10

  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        show_usage
        exit 0
        ;;
      -l|--limit)
        limit="$2"
        shift 2
        ;;
      *)
        query="$1"
        shift
        ;;
    esac
  done

  if [ -z "$query" ]; then
    echo "❌ Error: search query required"
    show_usage
    exit 1
  fi

  echo "🔍 Smart Search: $query"
  echo ""

  search_filenames "$query" "$limit"
  search_content "$query" "$limit"
  search_symbols "$query" "$limit"

  echo "✅ Search complete"
}

main "$@"
