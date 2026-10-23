#!/usr/bin/env bash
# Print one ATX section from a markdown file. Convenience; holds no guarantee.
# Usage: plan-slice.sh <path> <heading>
# Heading is the text after hashes, e.g. Outcome or "Current State and Next Action".
set -euo pipefail
file="${1:?Usage: plan-slice.sh <path> <heading>}"
heading="${2:?Usage: plan-slice.sh <path> <heading>}"
heading="${heading#\#}"
heading="${heading#"${heading%%[![:space:]]*}"}"
awk -v h="$heading" '
  BEGIN { re = "^#+[ \t]*" h "([ \t]|$)" }
  $0 ~ re { p = 1; print; next }
  p && /^#+[ \t]/ { exit }
  p { print }
' "$file"
