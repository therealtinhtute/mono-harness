#!/bin/bash
# Generate implementation plan template
# Usage: ./plan-generator.sh "Feature name"

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(dirname "$SCRIPT_DIR")"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") <feature-name>

Generate structured implementation plan template

Arguments:
  feature-name    Name of the feature or project

Options:
  -h, --help       Show this help message
  -d, --dir        Plan directory (default: .kit/plans/YYYYMMDD-slug/)

Examples:
  ./plan-generator.sh "User Authentication"
  ./plan-generator.sh "Payment Integration" --dir .kit/plans/custom/

from therealTINHTUTE with love
USAGE
}

function slugify() {
  echo "$1" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -cd '[:alnum:]-'
}

function create_plan_file() {
  local feature="$1"
  local slug="$2"
  local output="$3"
  
  cat > "$output" << 'PLAN_EOF'
---
title: FEATURE_NAME - Implementation Plan
description: Implementation plan for FEATURE_NAME
status: draft
created: DATE_PLACEHOLDER
tags: [plan, SLUG_PLACEHOLDER]
---

# FEATURE_NAME - Implementation Plan

## Executive Summary

**Goal**: [What are we building and why?]

**Timeline**: [Duration estimate]
**Effort**: [Hours/days estimate]
**Risk**: [LOW/MEDIUM/HIGH]

## Current State

### What Exists
- [List existing relevant code/features]

### What's Missing
- [List gaps this plan addresses]

## Approach

### Option 1: [Approach name]
**Pros**:
- [Benefit 1]

**Cons**:
- [Drawback 1]

**Effort**: [Estimate]

### Recommendation
[Which option and why]

## Implementation Phases

### Phase 1: [Phase name]
**Duration**: [Time estimate]

**Tasks**:
- [ ] Task 1
- [ ] Task 2

## Success Metrics

- [ ] [Measurable outcome 1]
- [ ] [Measurable outcome 2]

## Rollback Plan

[How to undo changes if needed]

from therealTINHTUTE with love
PLAN_EOF

  sed -i.bak "s/FEATURE_NAME/$feature/g" "$output"
  sed -i.bak "s/SLUG_PLACEHOLDER/$slug/g" "$output"
  sed -i.bak "s/DATE_PLACEHOLDER/$(date +%Y-%m-%d)/g" "$output"
  rm -f "$output.bak"
}

function main() {
  local feature=""
  local plan_dir=""

  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        show_usage
        exit 0
        ;;
      -d|--dir)
        plan_dir="$2"
        shift 2
        ;;
      *)
        feature="$1"
        shift
        ;;
    esac
  done

  if [ -z "$feature" ]; then
    echo "❌ Error: feature name required"
    show_usage
    exit 1
  fi

  local slug=$(slugify "$feature")
  
  if [ -z "$plan_dir" ]; then
    local date=$(date +%Y%m%d)
    plan_dir=".kit/plans/${date}-${slug}"
  fi

  echo "📋 Generating plan: $feature"

  mkdir -p "$plan_dir"
  create_plan_file "$feature" "$slug" "$plan_dir/plan.md"
  
  echo "✅ Created: $plan_dir/plan.md"
  echo ""
  echo "🎉 Plan template generated: $plan_dir"
}

main "$@"
