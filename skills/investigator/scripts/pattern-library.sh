#!/bin/bash
# Common search patterns
# Usage: ./pattern-library.sh api-endpoints

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") <pattern-name>

Search using pre-defined patterns

Available patterns:
  api-endpoints       Find API route definitions
  database-queries    Find database queries
  env-vars           Find environment variable usage
  imports            Find import statements
  exports            Find export statements
  todos              Find TODO/FIXME comments

Options:
  -h, --help       Show this help message
  -o, --output     Output file (default: stdout)

Examples:
  ./pattern-library.sh api-endpoints
  ./pattern-library.sh database-queries --output queries.txt

from therealTINHTUTE with love
USAGE
}

function search_api_endpoints() {
  echo "## API Endpoints"
  echo ""
  
  grep -rn --include="*.js" --include="*.ts" --include="*.py" --include="*.go" \
    -E "(app\.(get|post|put|delete|patch)|@(Get|Post|Put|Delete|Patch)|router\.|Route)" . 2>/dev/null | \
    while IFS=: read -r file line content; do
      echo "- \`$file:$line\`"
      echo "  \`\`\`"
      echo "  $(echo "$content" | sed 's/^[[:space:]]*//')"
      echo "  \`\`\`"
    done
  
  echo ""
}

function search_database_queries() {
  echo "## Database Queries"
  echo ""
  
  grep -rn --include="*.js" --include="*.ts" --include="*.py" --include="*.go" \
    -E "(SELECT|INSERT|UPDATE|DELETE|FROM|WHERE|\.query\(|\.execute\(|\.find\(|\.findOne\()" . 2>/dev/null | \
    head -n 50 | \
    while IFS=: read -r file line content; do
      echo "- \`$file:$line\`"
      echo "  \`\`\`"
      echo "  $(echo "$content" | sed 's/^[[:space:]]*//')"
      echo "  \`\`\`"
    done
  
  echo ""
}

function search_env_vars() {
  echo "## Environment Variables"
  echo ""
  
  grep -rn --include="*.js" --include="*.ts" --include="*.py" --include="*.go" \
    -E "(process\.env\.|os\.getenv|ENV\[)" . 2>/dev/null | \
    while IFS=: read -r file line content; do
      echo "- \`$file:$line\`"
      echo "  \`\`\`"
      echo "  $(echo "$content" | sed 's/^[[:space:]]*//')"
      echo "  \`\`\`"
    done
  
  echo ""
}

function search_imports() {
  echo "## Import Statements"
  echo ""
  
  grep -rn --include="*.js" --include="*.ts" --include="*.py" \
    -E "(^import |^from .* import|require\()" . 2>/dev/null | \
    head -n 50 | \
    while IFS=: read -r file line content; do
      echo "- \`$file:$line\`: $(echo "$content" | sed 's/^[[:space:]]*//')"
    done
  
  echo ""
}

function search_exports() {
  echo "## Export Statements"
  echo ""
  
  grep -rn --include="*.js" --include="*.ts" \
    -E "(^export |module\.exports)" . 2>/dev/null | \
    head -n 50 | \
    while IFS=: read -r file line content; do
      echo "- \`$file:$line\`: $(echo "$content" | sed 's/^[[:space:]]*//')"
    done
  
  echo ""
}

function search_todos() {
  echo "## TODO/FIXME Comments"
  echo ""
  
  grep -rn --include="*.js" --include="*.ts" --include="*.py" --include="*.go" \
    -E "(TODO|FIXME|HACK|XXX)" . 2>/dev/null | \
    while IFS=: read -r file line content; do
      echo "- \`$file:$line\`"
      echo "  \`\`\`"
      echo "  $(echo "$content" | sed 's/^[[:space:]]*//')"
      echo "  \`\`\`"
    done
  
  echo ""
}

function main() {
  local pattern=""
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
        pattern="$1"
        shift
        ;;
    esac
  done

  if [ -z "$pattern" ]; then
    echo "❌ Error: pattern name required"
    show_usage
    exit 1
  fi

  echo "🔍 Pattern Search: $pattern"
  echo ""

  local result=""
  
  case "$pattern" in
    api-endpoints)
      result=$(search_api_endpoints)
      ;;
    database-queries)
      result=$(search_database_queries)
      ;;
    env-vars)
      result=$(search_env_vars)
      ;;
    imports)
      result=$(search_imports)
      ;;
    exports)
      result=$(search_exports)
      ;;
    todos)
      result=$(search_todos)
      ;;
    *)
      echo "❌ Unknown pattern: $pattern"
      show_usage
      exit 1
      ;;
  esac

  if [ -n "$output_file" ]; then
    echo "$result" > "$output_file"
    echo "✅ Results saved: $output_file"
  else
    echo "$result"
  fi
}

main "$@"
