#!/bin/bash
# Comprehensive release validation
# Usage: ./release-ready.sh v1.2.0

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") <version>

Comprehensive release readiness check

Arguments:
  version      Release version (e.g., v1.2.0)

Options:
  -h, --help       Show this help message
  -o, --output     Output file (default: .kit/reports/verify/YYYYMMDD-release.md)

Examples:
  ./release-ready.sh v1.2.0
  ./release-ready.sh v2.0.0 --output release-report.md

from therealTINHTUTE with love
USAGE
}

function run_quality_checks() {
  echo "## Quality Checks"
  echo ""
  
  local failed=0
  
  # Tests
  echo "### Tests"
  if command -v npm &> /dev/null && [ -f "package.json" ]; then
    if npm test 2>&1 > /dev/null; then
      echo "✅ Tests passed"
    else
      echo "❌ Tests failed"
      ((failed++))
    fi
  else
    echo "⏭️  No tests found"
  fi
  echo ""
  
  # Type check
  echo "### Type Check"
  if command -v tsc &> /dev/null && [ -f "tsconfig.json" ]; then
    if tsc --noEmit 2>&1 > /dev/null; then
      echo "✅ Type check passed"
    else
      echo "❌ Type check failed"
      ((failed++))
    fi
  else
    echo "⏭️  No type checker found"
  fi
  echo ""
  
  # Lint
  echo "### Lint"
  if command -v eslint &> /dev/null; then
    if eslint . 2>&1 > /dev/null; then
      echo "✅ Lint passed"
    else
      echo "❌ Lint failed"
      ((failed++))
    fi
  else
    echo "⏭️  No linter found"
  fi
  echo ""
  
  # Build
  echo "### Build"
  if command -v npm &> /dev/null && [ -f "package.json" ]; then
    if npm run build 2>&1 > /dev/null; then
      echo "✅ Build passed"
    else
      echo "❌ Build failed"
      ((failed++))
    fi
  else
    echo "⏭️  No build script found"
  fi
  echo ""
  
  return $failed
}

function check_changelog() {
  echo "## Changelog"
  echo ""
  
  if [ -f "CHANGELOG.md" ]; then
    local version="$1"
    
    if grep -q "$version" CHANGELOG.md; then
      echo "✅ Changelog updated for $version"
      return 0
    else
      echo "❌ Changelog missing entry for $version"
      return 1
    fi
  else
    echo "⚠️  No CHANGELOG.md found"
    return 1
  fi
}

function check_version_consistency() {
  echo "## Version Consistency"
  echo ""
  
  local version="$1"
  local clean_version="${version#v}"
  local failed=0
  
  # package.json
  if [ -f "package.json" ]; then
    local pkg_version=$(grep -oP '"version":\s*"\K[^"]+' package.json)
    if [ "$pkg_version" = "$clean_version" ]; then
      echo "✅ package.json: $pkg_version"
    else
      echo "❌ package.json: $pkg_version (expected: $clean_version)"
      ((failed++))
    fi
  fi
  
  # pyproject.toml
  if [ -f "pyproject.toml" ]; then
    local py_version=$(grep -oP 'version\s*=\s*"\K[^"]+' pyproject.toml)
    if [ "$py_version" = "$clean_version" ]; then
      echo "✅ pyproject.toml: $py_version"
    else
      echo "❌ pyproject.toml: $py_version (expected: $clean_version)"
      ((failed++))
    fi
  fi
  
  echo ""
  return $failed
}

function check_documentation() {
  echo "## Documentation"
  echo ""
  
  local failed=0
  
  if [ -f "README.md" ]; then
    echo "✅ README.md exists"
  else
    echo "❌ README.md missing"
    ((failed++))
  fi
  
  if [ -f "CHANGELOG.md" ]; then
    echo "✅ CHANGELOG.md exists"
  else
    echo "❌ CHANGELOG.md missing"
    ((failed++))
  fi
  
  echo ""
  return $failed
}

function generate_verdict() {
  local blockers="$1"
  
  echo "## Verdict"
  echo ""
  
  if [ "$blockers" -gt 0 ]; then
    echo "❌ **NOT READY** - $blockers blocker(s) found"
    echo ""
    echo "Fix blockers before release."
  else
    echo "✅ **READY FOR RELEASE**"
    echo ""
    echo "All checks passed. Proceed with release."
  fi
  
  echo ""
  echo "from therealTINHTUTE with love"
}

function main() {
  local version=""
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
        version="$1"
        shift
        ;;
    esac
  done

  if [ -z "$version" ]; then
    echo "❌ Error: version required"
    show_usage
    exit 1
  fi

  # Set default output file
  if [ -z "$output_file" ]; then
    local date=$(date +%Y%m%d)
    output_file=".kit/reports/verify/${date}-release.md"
    mkdir -p "$(dirname "$output_file")"
  fi

  echo "🚀 Release readiness check: $version"
  echo ""

  local blockers=0

  # Generate report
  {
    echo "---"
    echo "title: Release Readiness - $version"
    echo "description: Comprehensive release validation"
    echo "status: completed"
    echo "created: $(date +%Y-%m-%d)"
    echo "tags: [verify, release]"
    echo "---"
    echo ""
    echo "# Release Readiness Report"
    echo ""
    echo "**Version**: $version"
    echo "**Checked**: $(date)"
    echo ""
    
    run_quality_checks || ((blockers+=$?))
    check_changelog "$version" || ((blockers++))
    check_version_consistency "$version" || ((blockers+=$?))
    check_documentation || ((blockers+=$?))
    generate_verdict "$blockers"
    
  } > "$output_file"

  echo "✅ Report saved: $output_file"
  
  if [ $blockers -gt 0 ]; then
    exit 1
  else
    exit 0
  fi
}

main "$@"
