#!/bin/bash
# Install git hooks for skill validation + v0.15 fail-closed guards
# Usage: ./install-git-hooks.sh [--force]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || echo ".")"
HOOKS_DIR="$REPO_ROOT/.git/hooks"

# ZGUARD-CORE-BEGIN
# Core guard logic shared by the local pre-commit hook and the CI re-check.
# Two wrappers feed it pairs of plan-file contents:
#   - pre-commit: staged blob vs HEAD blob ("reads staged bytes, trusts no marker")
#   - CI job:     pushed HEAD blob vs parent blob
# Guards implemented here (zharness v0.15 p1-hook-guard):
#   R2: re-execute every nested proof command of newly added ## Validation
#       entries whose verdict is APPROVED or APPROVE_WITH_REQUESTS
#       (`sh -c`, 5-minute timeout each); any non-zero exit rejects, naming
#       the failing command and its output tail. REQUEST_CHANGES proof is
#       never re-executed. No pass marker is read.
#   R3: reject `judge: same-session` on any plan whose frontmatter carries
#       `lane: high-risk`.

zharness_lane_of() {                       # <content-file>
  awk '/^---$/{n++; next} n==1 && /^lane:/{sub(/^lane:[ \t]*/, ""); print; exit}' "$1"
}

zharness_added_validation_timestamps() {   # <old-file> <new-file>
  comm -13 \
    <(grep -oE '^- `[0-9]{4}-[0-9]{2}-[0-9]{2}T[^`]+' "$1" 2>/dev/null | sort -u) \
    <(grep -oE '^- `[0-9]{4}-[0-9]{2}-[0-9]{2}T[^`]+' "$2" | sort -u) || true
}

zharness_extract_entry() {                 # <file> <entry-prefix>
  ZWANT="$2" awk '
    BEGIN { want = ENVIRON["ZWANT"] }
    $0 == "## Validation"           {inv = 1; next}
    /^## / && inv                   {exit}
    inv && index($0, want) == 1     {grab = 1}
    grab                            {print}
  ' "$1"
}

zharness_entry_verdict() {                 # <entry-body>
  printf '%s\n' "$1" | grep -oE '(APPROVED|APPROVE_WITH_REQUESTS|REQUEST_CHANGES)' | head -1 || true
}

zharness_entry_proofs() {                  # <entry-body>
  printf '%s\n' "$1" \
    | grep -oE '^[[:space:]]{2,6}- `[^`]+`' \
    | sed -E 's/^[[:space:]]*-[[:space:]]*`//; s/`[[:space:]]*$//' \
    | grep -v '^$' || true
}

zharness_validation_same_session_lines() {   # <content-file>
  awk '
    $0 == "## Validation"             {inv = 1; next}
    /^## / && inv                     {exit}
    inv && /judge:[ \t]*same-session/ {print}
  ' "$1"
}

zharness_guards_file() {                   # <path> <old-file> <new-file>
  local path="$1" old="$2" new="$3"
  if [ "$(zharness_lane_of "$new")" != "high-risk" ]; then return 0; fi
  local added
  added=$(comm -13 \
    <(zharness_validation_same_session_lines "$old" | sort -u) \
    <(zharness_validation_same_session_lines "$new" | sort -u))
  if [ -n "$added" ]; then
    echo "" >&2
    echo "❌ R3 JUDGE GUARD REJECTED: $path" >&2
    echo "   frontmatter sets lane: high-risk and a newly added ## Validation" >&2
    echo "   entry declares a same-session judge. An independent judge is required." >&2
    echo "   offending lines:" >&2
    printf '%s\n' "$added" >&2
    return 1
  fi
  return 0
}

zharness_proof_reexec_file() {             # <path> <old-file> <new-file>
  local path="$1" old="$2" new="$3" failed=0 rc_cmd cmd out ts body verdict cmds
  local timestamps
  timestamps=$(zharness_added_validation_timestamps "$old" "$new")
  [ -z "$timestamps" ] && return 0
  while IFS= read -r ts; do
    [ -z "$ts" ] && continue
    body=$(zharness_extract_entry "$new" "$ts")
    verdict=$(zharness_entry_verdict "$body")
    case "$verdict" in
      APPROVED|APPROVE_WITH_REQUESTS) ;;
      *) continue ;;
    esac
    cmds=$(zharness_entry_proofs "$body")
    [ -z "$cmds" ] && continue
    while IFS= read -r cmd; do
      echo "🧪 re-executing proof [$ts]: $cmd"
      out=$(timeout 300 sh -c "$cmd" 2>&1) && rc_cmd=0 || rc_cmd=$?
      if [ "${rc_cmd:-0}" -ne 0 ]; then
        failed=$((failed + 1))
        echo "" >&2
        echo "❌ R2 PROOF GUARD REJECTED: $path" >&2
        echo "   failing command: $cmd" >&2
        echo "   exit: $rc_cmd" >&2
        echo "   output tail:" >&2
        printf '%s\n' "$out" | tail -10 >&2
      fi
    done <<< "$cmds"
  done <<< "$timestamps"
  [ "$failed" -eq 0 ]
}

zhuards_guard_plans() {                    # <path-list-space-separated> <old-source|staged|head> <dir-for-temp>
  local plans="$1" mode="$2" tmp="$3" f rc=0
  for f in $plans; do
    case "$mode" in
      staged)
        git show "HEAD:$f" > "$tmp/old.md" 2>/dev/null || : > "$tmp/old.md"
        git show ":$f"     > "$tmp/new.md"
        ;;
      head)
        git show "HEAD~1:$f" > "$tmp/old.md" 2>/dev/null || : > "$tmp/old.md"
        git show "HEAD:$f"   > "$tmp/new.md"
        ;;
    esac
    if ! zharness_guards_file "$f" "$tmp/old.md" "$tmp/new.md"; then rc=1; fi
    if ! zharness_proof_reexec_file "$f" "$tmp/old.md" "$tmp/new.md"; then rc=1; fi
  done
  return $rc
}

# ZGUARD-CORE-END

function show_usage() {
  cat << 'USAGE'
Usage: $(basename "$0") [options]

Install git hooks for automated skill validation plus the v0.15
fail-closed guards (proof re-execution + independent-judge rule).

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
# Pre-commit hook: validate changed skills + v0.15 fail-closed guards (R2/R3)

export ZHARNESS_HOOK_SOURCE
ROOT="$(git rev-parse --show-toplevel)"
export ROOT
ZHARNESS_HOOK_SOURCE="$ROOT/scripts/install-git-hooks.sh"
_zhtmp=$(mktemp -d)
awk '$0=="# ZGUARD-CORE-BEGIN"{on=1;next} $0=="# ZGUARD-CORE-END"{on=0} on' "$ZHARNESS_HOOK_SOURCE" > "$_zhtmp/guard.sh"
# shellcheck disable=SC1090
source "$_zhtmp/guard.sh"

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir" "$_zhtmp"' EXIT

echo "🔍 v0.15 guards on staged plans..."
plans=$(git diff --cached --name-only --diff-filter=ACM -- 'docs/plans/active/*.md')
[ -n "$plans" ] && printf 'staged plans:\n%s\n' "$plans"
guard_failed=0
if [ -n "$plans" ] && ! zhuards_guard_plans "$plans" staged "$tmpdir"; then
  guard_failed=1
fi

if [ "$guard_failed" -gt 0 ]; then
  echo ""
  echo "❌ v0.15 guards rejected this commit."
  echo "Fix the failing proof command(s) or provide an independent judge, then commit again."
  exit 1
fi
echo "✅ v0.15 guards passed"

echo ""
echo "🔍 Validating changed skills..."

changed_skills=$(git diff --cached --name-only | grep "kit/skills/.*/SKILL.md" || true)

if [ -z "$changed_skills" ]; then
  echo "✅ No skill files changed"
  exit 0
fi

failed=0
while IFS= read -r skill_file; do
  if [ -f "$skill_file" ]; then
    echo "Checking: $skill_file"
    skill_dir=$(dirname "$skill_file")
    if [ -f "kit/skills/scripts/validate-skill.sh" ]; then
      if ! bash kit/skills/scripts/validate-skill.sh "$skill_file"; then
        echo "❌ Validation failed: $skill_file"
        ((failed++))
      fi
    else
      if ! grep -q "^---" "$skill_file"; then echo "❌ Missing frontmatter: $skill_file"; ((failed++)); fi
      if ! grep -q "<role>" "$skill_file"; then echo "❌ Missing <role> tag: $skill_file"; ((failed++)); fi
      if ! grep -q "<security>" "$skill_file"; then echo "❌ Missing <security> tag: $skill_file"; ((failed++)); fi
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

  create_pre_commit_hook "$force" || true
  create_commit_msg_hook "$force"

  echo ""
  echo "🎉 Git hooks installed successfully"
  echo ""
  echo "Hooks installed:"
  echo "  - pre-commit: v0.15 R2/R3 guards + changed-skill validation"
  echo "  - commit-msg: Validates commit message format"
}

main "$@"
