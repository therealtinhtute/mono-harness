#!/bin/bash
# Run all quality checks
# Usage: ./quality-check.sh [--fast]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") [options]

Run all quality checks: tests, types, lint, build

Options:
  -h, --help       Show this help message
  -f, --fast       Skip slow checks (build)
  -p, --parallel   Run checks in parallel

Examples:
  ./quality-check.sh
  ./quality-check.sh --fast
  ./quality-check.sh --parallel

from therealTINHTUTE with love
USAGE
}

function run_tests() {
  echo "🧪 Running tests..."
  
  if command -v npm &> /dev/null && [ -f "package.json" ]; then
    if npm test 2>&1; then
      echo "✅ Tests passed"
      return 0
    else
      echo "❌ Tests failed"
      return 1
    fi
  elif command -v pytest &> /dev/null; then
    if pytest 2>&1; then
      echo "✅ Tests passed"
      return 0
    else
      echo "❌ Tests failed"
      return 1
    fi
  else
    echo "⏭️  No test runner found"
    return 0
  fi
}

function run_type_check() {
  echo "🔍 Running type check..."
  
  if command -v tsc &> /dev/null && [ -f "tsconfig.json" ]; then
    if tsc --noEmit 2>&1; then
      echo "✅ Type check passed"
      return 0
    else
      echo "❌ Type check failed"
      return 1
    fi
  elif command -v mypy &> /dev/null; then
    if mypy . 2>&1; then
      echo "✅ Type check passed"
      return 0
    else
      echo "❌ Type check failed"
      return 1
    fi
  else
    echo "⏭️  No type checker found"
    return 0
  fi
}

function run_lint() {
  echo "🔎 Running lint..."
  
  if command -v eslint &> /dev/null; then
    if eslint . 2>&1; then
      echo "✅ Lint passed"
      return 0
    else
      echo "❌ Lint failed"
      return 1
    fi
  elif command -v biome &> /dev/null; then
    if biome check . 2>&1; then
      echo "✅ Lint passed"
      return 0
    else
      echo "❌ Lint failed"
      return 1
    fi
  else
    echo "⏭️  No linter found"
    return 0
  fi
}

function run_build() {
  echo "🔨 Running build..."
  
  if command -v npm &> /dev/null && [ -f "package.json" ]; then
    if npm run build 2>&1; then
      echo "✅ Build passed"
      return 0
    else
      echo "❌ Build failed"
      return 1
    fi
  else
    echo "⏭️  No build script found"
    return 0
  fi
}

function main() {
  local fast_mode="false"
  local parallel_mode="false"

  # Parse arguments
  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        show_usage
        exit 0
        ;;
      -f|--fast)
        fast_mode="true"
        shift
        ;;
      -p|--parallel)
        parallel_mode="true"
        shift
        ;;
      *)
        echo "❌ Unknown option: $1"
        show_usage
        exit 1
        ;;
    esac
  done

  echo "🚀 Running quality checks"
  echo ""

  local failed=0

  if [ "$parallel_mode" = "true" ]; then
    # Run in parallel
    run_tests &
    local test_pid=$!
    
    run_type_check &
    local type_pid=$!
    
    run_lint &
    local lint_pid=$!
    
    # Wait for all
    wait $test_pid || ((failed++))
    wait $type_pid || ((failed++))
    wait $lint_pid || ((failed++))
  else
    # Run sequentially
    run_tests || ((failed++))
    echo ""
    
    run_type_check || ((failed++))
    echo ""
    
    run_lint || ((failed++))
    echo ""
  fi

  if [ "$fast_mode" = "false" ]; then
    run_build || ((failed++))
    echo ""
  fi

  echo "📊 Summary"
  echo "=========="
  
  if [ $failed -eq 0 ]; then
    echo "✅ All checks passed"
    exit 0
  else
    echo "❌ $failed check(s) failed"
    exit 1
  fi
}

main "$@"
