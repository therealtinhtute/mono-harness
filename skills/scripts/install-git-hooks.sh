#!/bin/bash
# Install git hooks for skill validation
# Usage: ./install-git-hooks.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || echo ".")"
HOOKS_DIR="$REPO_ROOT/.git/hooks"

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") [options]

Install git hooks for automated skill validation

Options:
  -h, --help       Show this help message
  -f, --force      Overwrite existing hooks

Examples:
  ./install-git-hooks.sh
  ./install-git-hooks.sh --force

from therealTINHTUTE with love
USAGE
}

function create_pre_commit_hook() {
  local hook_file="$HOOKS_DIR/pre-commit"
  local force="$1"
  
  if [ -f "$hook_file" ] && [ "$force" != "true" ]; then
    echo "⚠️  pre-commit hook already exists"
    echo "Use --force to overwrite"
    return 1
  fi
  
  cat > "$hook_file" << 'HOOK_EOF'
#!/bin/bash
# Pre-commit hook: Validate changed skills

set -e

echo "🔍 Validating changed skills..."

# Get changed SKILL.md files
changed_skills=$(git diff --cached --name-only | grep "kit/skills/.*/SKILL.md" || true)

if [ -z "$changed_skills" ]; then
  echo "✅ No skill files changed"
  exit 0
fi

# Validate each changed skill
failed=0
while IFS= read -r skill_file; do
  if [ -f "$skill_file" ]; then
    echo "Checking: $skill_file"
    
    skill_dir=$(dirname "$skill_file")
    
    # Run validation script if it exists
    if [ -f "kit/skills/scripts/validate-skill.sh" ]; then
      if ! bash kit/skills/scripts/validate-skill.sh "$skill_file"; then
        echo "❌ Validation failed: $skill_file"
        ((failed++))
      fi
    else
      # Basic validation
      if ! grep -q "^---" "$skill_file"; then
        echo "❌ Missing frontmatter: $skill_file"
        ((failed++))
      fi
      
      if ! grep -q "<role>" "$skill_file"; then
        echo "❌ Missing <role> tag: $skill_file"
        ((failed++))
      fi
      
      if ! grep -q "<security>" "$skill_file"; then
        echo "❌ Missing <security> tag: $skill_file"
        ((failed++))
      fi
    fi
  fi
done <<< "$changed_skills"

if [ $failed -gt 0 ]; then
  echo ""
  echo "❌ $failed skill(s) failed validation"
  echo "Fix issues before committing"
  exit 1
fi

echo "✅ All skills validated"
exit 0
HOOK_EOF

  chmod +x "$hook_file"
  echo "✅ Created: $hook_file"
}

function create_commit_msg_hook() {
  local hook_file="$HOOKS_DIR/commit-msg"
  local force="$1"
  
  if [ -f "$hook_file" ] && [ "$force" != "true" ]; then
    echo "⚠️  commit-msg hook already exists"
    return 0
  fi
  
  cat > "$hook_file" << 'HOOK_EOF'
#!/bin/bash
# Commit-msg hook: Validate conventional commit format

set -e

commit_msg_file="$1"
commit_msg=$(cat "$commit_msg_file")

# Check conventional commit format
if ! echo "$commit_msg" | grep -qE '^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\(.+\))?: .+'; then
  echo "❌ Invalid commit message format"
  echo ""
  echo "Expected: type(scope): description"
  echo "Types: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert"
  echo ""
  echo "Your message:"
  echo "$commit_msg"
  exit 1
fi

exit 0
HOOK_EOF

  chmod +x "$hook_file"
  echo "✅ Created: $hook_file"
}

function main() {
  local force="false"

  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help)
        show_usage
        exit 0
        ;;
      -f|--force)
        force="true"
        shift
        ;;
      *)
        echo "❌ Unknown option: $1"
        show_usage
        exit 1
        ;;
    esac
  done

  if [ ! -d "$HOOKS_DIR" ]; then
    echo "❌ Not a git repository"
    exit 1
  fi

  echo "🔧 Installing git hooks..."
  echo ""

  create_pre_commit_hook "$force"
  create_commit_msg_hook "$force"

  echo ""
  echo "🎉 Git hooks installed successfully"
  echo ""
  echo "Hooks installed:"
  echo "  - pre-commit: Validates changed skills"
  echo "  - commit-msg: Validates commit message format"
}

main "$@"
