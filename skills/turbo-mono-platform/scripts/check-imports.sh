#!/usr/bin/env bash
set -e
ERRORS=0

while IFS= read -r -d '' file; do
  matches=$(grep -nE "from ['\"](\.\./)+packages/" "$file" 2>/dev/null || true)
  if [ -n "$matches" ]; then
    echo "❌ $file"
    echo "$matches" | while read -r line; do echo "   $line"; done
    ERRORS=$((ERRORS + 1))
  fi
done < <(find apps packages -name "*.ts" -o -name "*.tsx" 2>/dev/null | tr '\n' '\0')

if [ "$ERRORS" -eq 0 ]; then
  echo "✅ No relative cross-package imports found"
else
  echo ""
  echo "Fix: replace ../../packages/<n>/... with @repo/<n>"
  exit 1
fi
