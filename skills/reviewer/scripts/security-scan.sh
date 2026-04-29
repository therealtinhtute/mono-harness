#!/bin/bash
# Automated security vulnerability scan
# Usage: ./security-scan.sh [files...]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") [files...]

Scan files for common security vulnerabilities

Arguments:
  files...    Files to scan (default: all staged files)

Options:
  -h, --help       Show this help message
  -o, --output     Output file (default: stdout)
  -f, --format     Output format: text|json (default: text)

Examples:
  ./security-scan.sh
  ./security-scan.sh src/**/*.js
  ./security-scan.sh --format json --output scan.json

from therealTINHTUTE with love
USAGE
}

function scan_sql_injection() {
  local file="$1"
  local findings=()
  
  # SQL injection patterns
  local patterns=(
    'execute.*\+.*\$'
    'query.*\+.*\$'
    'SELECT.*\+.*\$'
    'INSERT.*\+.*\$'
    'UPDATE.*\+.*\$'
    'DELETE.*\+.*\$'
  )
  
  for pattern in "${patterns[@]}"; do
    if grep -nE "$pattern" "$file" 2>/dev/null; then
      findings+=("SQL Injection: $pattern")
    fi
  done
  
  printf '%s\n' "${findings[@]}"
}

function scan_xss() {
  local file="$1"
  local findings=()
  
  # XSS patterns
  local patterns=(
    'innerHTML\s*='
    'dangerouslySetInnerHTML'
    'document\.write'
    'eval\('
    '\$\(.*\)\.html\('
  )
  
  for pattern in "${patterns[@]}"; do
    if grep -nE "$pattern" "$file" 2>/dev/null; then
      findings+=("XSS: $pattern")
    fi
  done
  
  printf '%s\n' "${findings[@]}"
}

function scan_auth_issues() {
  local file="$1"
  local findings=()
  
  # Auth boundary patterns
  local patterns=(
    'auth.*=.*false'
    'authenticated.*=.*false'
    'bypass.*auth'
    'skip.*auth'
  )
  
  for pattern in "${patterns[@]}"; do
    if grep -nE "$pattern" "$file" 2>/dev/null; then
      findings+=("Auth Issue: $pattern")
    fi
  done
  
  printf '%s\n' "${findings[@]}"
}

function scan_secrets() {
  local file="$1"
  local findings=()
  
  # Secret patterns
  local patterns=(
    'password\s*=\s*["\047][^"\047]+["\047]'
    'api[_-]?key\s*=\s*["\047][^"\047]+["\047]'
    'secret\s*=\s*["\047][^"\047]+["\047]'
    'token\s*=\s*["\047][^"\047]+["\047]'
    'AWS_ACCESS_KEY'
    'PRIVATE_KEY'
  )
  
  for pattern in "${patterns[@]}"; do
    if grep -nE "$pattern" "$file" 2>/dev/null; then
      findings+=("Hardcoded Secret: $pattern")
    fi
  done
  
  printf '%s\n' "${findings[@]}"
}

function scan_file() {
  local file="$1"
  
  if [ ! -f "$file" ]; then
    return
  fi
  
  local sql_issues=$(scan_sql_injection "$file")
  local xss_issues=$(scan_xss "$file")
  local auth_issues=$(scan_auth_issues "$file")
  local secret_issues=$(scan_secrets "$file")
  
  if [ -n "$sql_issues" ] || [ -n "$xss_issues" ] || [ -n "$auth_issues" ] || [ -n "$secret_issues" ]; then
    echo "File: $file"
    
    if [ -n "$sql_issues" ]; then
      echo "  🔴 SQL Injection:"
      echo "$sql_issues" | sed 's/^/    /'
    fi
    
    if [ -n "$xss_issues" ]; then
      echo "  🔴 XSS:"
      echo "$xss_issues" | sed 's/^/    /'
    fi
    
    if [ -n "$auth_issues" ]; then
      echo "  🔴 Auth Issue:"
      echo "$auth_issues" | sed 's/^/    /'
    fi
    
    if [ -n "$secret_issues" ]; then
      echo "  🔴 Hardcoded Secret:"
      echo "$secret_issues" | sed 's/^/    /'
    fi
    
    echo ""
  fi
}

function main() {
  local files=()
  local output_file=""
  local format="text"

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
        files+=("$1")
        shift
        ;;
    esac
  done

  # If no files specified, scan staged files
  if [ ${#files[@]} -eq 0 ]; then
    mapfile -t files < <(git diff --cached --name-only)
  fi

  if [ ${#files[@]} -eq 0 ]; then
    echo "❌ No files to scan"
    exit 1
  fi

  echo "🔍 Scanning ${#files[@]} file(s) for security issues..."
  echo ""

  local output=""
  local issue_count=0

  for file in "${files[@]}"; do
    local result=$(scan_file "$file")
    if [ -n "$result" ]; then
      output+="$result"
      ((issue_count++))
    fi
  done

  if [ -n "$output_file" ]; then
    echo "$output" > "$output_file"
    echo "✅ Scan complete: $output_file"
  else
    echo "$output"
  fi

  if [ $issue_count -eq 0 ]; then
    echo "✅ No security issues found"
    exit 0
  else
    echo "❌ Found security issues in $issue_count file(s)"
    exit 1
  fi
}

main "$@"
