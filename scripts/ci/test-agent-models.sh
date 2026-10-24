#!/usr/bin/env bash
# setup/models.json is the role -> model policy; each setup/agents/*.md must
# carry the model its tier names, and every agent file must have a policy row.
set -euo pipefail
root="$(cd "$(dirname "$0")/../.." && pwd)"
policy="$root/setup/models.json"
fail=0
listed=$(jq -r '.agents[].agent' "$policy" | sort)
present=$(cd "$root/setup/agents" && ls *.md | sed 's/\.md$//' | sort)
if [ "$listed" != "$present" ]; then
  echo "agent files and models.json rows differ:"; diff <(echo "$listed") <(echo "$present") || true; fail=1
fi
for a in $listed; do
  want=$(jq -r --arg a "$a" '.tiers[(.agents[]|select(.agent==$a).tier)]' "$policy")
  jq -e --arg m "$want" '.available|index($m)' "$policy" >/dev/null || { echo "$a: tier model '$want' not in available"; fail=1; }
  f="$root/setup/agents/$a.md"
  [ -f "$f" ] || continue
  got=$(awk '/^---$/{n++;next} n==1 && /^model:/{print $2}' "$f")
  name=$(awk '/^---$/{n++;next} n==1 && /^name:/{print $2}' "$f")
  [ "$got" = "$want" ] || { echo "$a: model '$got', policy wants '$want'"; fail=1; }
  [ "$name" = "$a" ] || { echo "$a: frontmatter name '$name'"; fail=1; }
done
[ "$fail" -eq 0 ] && echo "agent models OK ($(echo "$listed" | wc -l | tr -d ' ') agents)"
exit "$fail"
