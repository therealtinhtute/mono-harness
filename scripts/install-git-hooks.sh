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
# Guards implemented here (zharness v0.15 p1-hook-guard, review-hardened v2):
#   R2: re-execute every nested proof command of newly added ## Validation
#       entries whose anchored verdict is APPROVED or APPROVE_WITH_REQUESTS
#       (`sh -c`, 5-minute timeout each where timeout/gtimeout is available,
#       unbounded otherwise); any non-zero exit rejects, naming
#       the failing command and its output tail. REQUEST_CHANGES proof is
#       never re-executed. No pass marker is read.
#   R3: reject `judge: same-session` on any newly added entry of a plan whose
#       frontmatter carries `lane: high-risk`. Matching strips backticks, so
#       the repository's canonical judge: `same-session` form is caught too.
#   Hardening from independent review (v2): "new entry" means its FULL TEXT
#   differs from every old-side Validation entry — timestamp-set diffing was
#   removed because replaying an existing timestamp made fresh APPROVED text
#   invisible; verdict detection is anchored to a `verdict:` token so prose
#   mentions cannot shadow it; proof bullets accept any indent >= 2 spaces;
#   an approvable verdict citing zero proofs is rejected as malformed instead
#   of being waved through.
#   Hardening v3 (guard-v3 plan R1/R2): the verdict token is read from the
#   entry's FIRST LINE only — a sub-bullet quoting another verdict can never
#   select or shadow the entry's own; and an entry STARTS at any unindented
#   `- ` line, the leading timestamp being optional, so undated entries are
#   visible to both guards.

zharness_lane_of() {                       # <content-file>
  awk '/^---$/{n++; next} n==1 && /^lane:/{sub(/^lane:[ \t]*/, ""); print; exit}' "$1"
}

zharness_dump_entries() {                  # <content-file> <outdir>
  OUT="$2" awk '
    BEGIN { out = ENVIRON["OUT"] }
    $0 == "## Validation" {inv = 1; next}
    /^## / && inv         {exit}
    inv {
      if ($0 ~ /^- /) { n++; buf[n] = $0 }
      else if (n > 0) { buf[n] = buf[n] "\n" $0 }
    }
    END { for (i = 1; i <= n; i++) print buf[i] > (out "/e" sprintf("%04d", i) ".txt") }
  ' "$1"
}

zharness_anchored_verdict() {              # <entry-body>
  # R1: the entry's verdict lives on its own first line — the unindented
  # bullet that starts the entry. A verdict token quoted deeper in the body
  # (sub-bullet, prose, quoted prior review) must not select or shadow it.
  printf '%s\n' "$1" | head -1 \
    | grep -oE '`?verdict`?[[:blank:]]*[:`][[:blank:]]*`?(APPROVED|APPROVE_WITH_REQUESTS|REQUEST_CHANGES)' \
    | grep -oE '(APPROVED|APPROVE_WITH_REQUESTS|REQUEST_CHANGES)' | head -1 || true
}

zharness_entry_proofs() {                  # <entry-body>
  printf '%s\n' "$1" \
    | grep -oE '^[[:space:]]{2,}- `[^`]+`' \
    | sed -E 's/^[[:space:]]*-[[:space:]]*`//; s/`[[:space:]]*$//' \
    | grep -v '^$' || true
}

zharness_entry_has_same_session() {        # <entry-body>, backticks stripped first
  printf '%s\n' "$1" | sed 's/[`]//g' | grep -q 'judge:[[:blank:]]*same-session'
}

zharness_run_proof() {                     # <command>
  # The 300s bound is defensive, not load-bearing. Resolve the wrapper at call
  # time so the guard's verdict depends on the proof's own exit code: GNU
  # `timeout` where present, coreutils `gtimeout` on macOS, otherwise run the
  # command unwrapped. Without this, every proof on a stock macOS exits 127
  # ("timeout: command not found") and the guard rejects each honest entry.
  if command -v timeout >/dev/null 2>&1; then
    timeout 300 sh -c "$1" 2>&1
  elif command -v gtimeout >/dev/null 2>&1; then
    gtimeout 300 sh -c "$1" 2>&1
  else
    echo "⚠️  no timeout/gtimeout on PATH — running proof unbounded (interrupt with Ctrl+C if it hangs)" >&2
    sh -c "$1" 2>&1
  fi
}

zharness_guard_entries_of_file() {         # <path> <old-file> <new-file>
  local path="$1" old="$2" new="$3" failed=0 rc_cmd cmd out body verdict cmds scratch h efile header line
  # Old-side entry hashes live in a file, not an associative array: `local -A`
  # needs bash 4+, and macOS ships bash 3.2 as /bin/bash. There the declaration
  # fails, the hex subscript is then evaluated as arithmetic, and the shell dies
  # with "value too great for base" on the first old-side entry — taking the
  # whole guard down with it. A file plus `grep -Fxq` is exact-match membership
  # with identical semantics on every bash.
  local oldhashes

  scratch=$(mktemp -d)
  mkdir -p "$scratch/o" "$scratch/n"
  zharness_dump_entries "$old" "$scratch/o"
  zharness_dump_entries "$new" "$scratch/n"

  oldhashes="$scratch/oldhashes.txt"
  : > "$oldhashes"
  for efile in "$scratch"/o/e*.txt; do
    [ -f "$efile" ] || continue
    sha256sum "$efile" | cut -d' ' -f1 >> "$oldhashes"
  done

  for efile in "$scratch"/n/e*.txt; do
    [ -f "$efile" ] || continue
    h=$(sha256sum "$efile" | cut -d' ' -f1)
    grep -Fxq "$h" "$oldhashes" && continue
    body=$(cat "$efile")
    header=$(head -1 "$efile")

    if [ "$(zharness_lane_of "$new")" = "high-risk" ] && zharness_entry_has_same_session "$body"; then
      failed=$((failed + 1))
      echo "" >&2
      echo "❌ R3 JUDGE GUARD REJECTED: $path" >&2
      echo "   frontmatter sets lane: high-risk and a newly added ## Validation" >&2
      echo "   entry declares a same-session judge. An independent judge is required." >&2
      echo "   offending entry starts: $header" >&2
      printf '%s\n' "$body" | grep 'judge:' | sed 's/^/     /' >&2
      continue
    fi

    verdict=$(zharness_anchored_verdict "$body")
    case "$verdict" in
      APPROVED|APPROVE_WITH_REQUESTS) ;;
      *) continue ;;
    esac

    cmds=$(zharness_entry_proofs "$body")
    if [ -z "$cmds" ]; then
      failed=$((failed + 1))
      echo "" >&2
      echo "❌ R2 PROOF GUARD REJECTED: $path" >&2
      echo "   malformed entry: verdict \`$verdict\` cites no proof commands at all" >&2
      echo "   offending entry starts: $header" >&2
      continue
    fi

    while IFS= read -r cmd; do
      echo "🧪 re-executing proof [$header]: $cmd"
      out=$(zharness_run_proof "$cmd") && rc_cmd=0 || rc_cmd=$?
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
  done

  rm -rf "$scratch"
  [ "$failed" -eq 0 ]
}

zhuards_guard_plans() {                    # <path-list-space-separated> <staged|head> <dir-for-temp>
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
    if ! zharness_guard_entries_of_file "$f" "$tmp/old.md" "$tmp/new.md"; then rc=1; fi
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
# Enforce from the staged installer bytes when the installer itself is part of
# this commit; otherwise fall back to the worktree copy.
GUARD_SRC="$ZHARNESS_HOOK_SOURCE"
if git diff --cached --name-only -- scripts/install-git-hooks.sh | grep -q . 2>/dev/null; then
  git show ":scripts/install-git-hooks.sh" > "$_zhtmp/src.sh" || {
    echo "❌ cannot resolve staged guard source"; exit 1;
  }
  GUARD_SRC="$_zhtmp/src.sh"
fi
awk '$0=="# ZGUARD-CORE-BEGIN"{on=1;next} $0=="# ZGUARD-CORE-END"{on=0} on' "$GUARD_SRC" > "$_zhtmp/guard.sh" || {
  echo "❌ guard core extraction failed from $GUARD_SRC"; exit 1;
}
grep -q '^zharness_guard_entries_of_file()' "$_zhtmp/guard.sh" || { echo "❌ guard core incomplete"; exit 1; }
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

# Skills live at skills/<category>/<name>/SKILL.md. This grep used to say
# "kit/skills/.*/SKILL.md" — a path this repo has never had — so the whole block
# below was dead and every commit printed "No skill files changed".
changed_skills=$(git diff --cached --name-only | grep -E '(^|/)skills/[^/]+/[^/]+/SKILL\.md$' || true)

if [ -z "$changed_skills" ]; then
  echo "✅ No skill files changed"
  exit 0
fi

failed=0
while IFS= read -r skill_file; do
  if [ -f "$skill_file" ]; then
    echo "Checking: $skill_file"
    skill_dir=$(dirname "$skill_file")
    if [ -f "scripts/validate-skill.sh" ]; then
      if ! bash scripts/validate-skill.sh "$skill_file"; then
        echo "❌ Validation failed: $skill_file"
        ((failed++))
      fi
    else
      # Fallback when the validator is absent. Frontmatter only — the <role> and
      # <security> tags this used to demand belong to the heavyweight skill
      # format, not to the <=30-line thin triggers.
      if ! grep -q "^---" "$skill_file"; then echo "❌ Missing frontmatter: $skill_file"; ((failed++)); fi
      if ! grep -q "^name:" "$skill_file"; then echo "❌ Missing name: $skill_file"; ((failed++)); fi
      if ! grep -q "^description:" "$skill_file"; then echo "❌ Missing description: $skill_file"; ((failed++)); fi
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
