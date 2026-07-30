#!/usr/bin/env bash
# verify-doc-links.sh - fail on broken repo-relative cross-references in tracked docs.
#
# A claim is a backtick-quoted, path-like token (contains a "/") whose first
# segment is one of the repository's own top-level doc surfaces. It passes if it
# resolves either from the repository root or from the referencing file's own
# directory. Anything outside the allowlist is skipped by default, so
# illustrative example paths never become findings.
#
# Exit 0 = no findings, 1 = broken references, 2 = malformed .claimignore.

set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

IGNORE_FILE=".claimignore"
ALLOWED_PREFIXES="skills docs rules cli setup references"

# --- ignore file: every non-comment line must carry a `# reason` -------------
ignore_patterns=()
if [ -f "$IGNORE_FILE" ]; then
  lineno=0
  while IFS= read -r line || [ -n "$line" ]; do
    lineno=$((lineno + 1))
    case "$line" in
      '' | '#'*) continue ;;
    esac
    case "$line" in
      *'#'*) : ;;
      *)
        printf 'ERROR: %s:%d has no `# reason` comment: %s\n' \
          "$IGNORE_FILE" "$lineno" "$line" >&2
        exit 2
        ;;
    esac
    pattern="${line%%#*}"
    pattern="${pattern%"${pattern##*[![:space:]]}"}"
    if [ -n "$pattern" ]; then
      ignore_patterns+=("$pattern")
    fi
  done <"$IGNORE_FILE"
fi

# --- file list: a real file, never a shell variable --------------------------
# A multi-line variable expanded into grep is silently treated as a pattern by
# some grep implementations (ugrep), which produces a false clean.
LIST="$(mktemp)"
trap 'rm -f "$LIST"' EXIT
# docs/plans/** is excluded by category, not by exception: a plan artifact must be
# able to name a file it will create, and a completed plan is an immutable record
# of paths as they were. Neither is a live cross-reference.
find docs skills rules setup -name '*.md' -type f -not -path 'docs/plans/*' >"$LIST"
for extra in CLAUDE.md README.md; do
  if [ -f "$extra" ]; then
    printf '%s\n' "$extra" >>"$LIST"
  fi
done

# --- scan --------------------------------------------------------------------
findings=0
while IFS= read -r file; do
  dir="$(dirname "$file")"
  claims="$(grep -ohE '`[A-Za-z0-9._/-]+/[A-Za-z0-9._/-]+\.(md|sh|go|json|yml|toml|py)`' \
    "$file" 2>/dev/null | tr -d '`' | sort -u || true)"
  [ -n "$claims" ] || continue

  while IFS= read -r claim; do
    [ -n "$claim" ] || continue

    case "$claim" in
      *'{'* | *'}'* | *'*'*) continue ;;
    esac

    first="${claim%%/*}"
    allowed=0
    for prefix in $ALLOWED_PREFIXES; do
      if [ "$first" = "$prefix" ]; then
        allowed=1
        break
      fi
    done
    if [ "$allowed" -eq 0 ]; then
      continue
    fi

    if [ -e "$claim" ] || [ -e "$dir/$claim" ]; then
      continue
    fi

    skip=0
    for pattern in ${ignore_patterns[@]+"${ignore_patterns[@]}"}; do
      case "$claim" in
        *"$pattern"*)
          skip=1
          break
          ;;
      esac
    done
    if [ "$skip" -eq 1 ]; then
      continue
    fi

    printf '%s -> %s\n' "$file" "$claim"
    findings=$((findings + 1))
  done <<<"$claims"
done <"$LIST"

if [ "$findings" -gt 0 ]; then
  printf '\n%d broken doc cross-reference(s).\n' "$findings" >&2
  exit 1
fi

printf 'doc links OK (0 findings)\n'
