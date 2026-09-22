#!/bin/bash
# Independent end-to-end smoke of the three symlink write sites, driven through
# the real binary (not the unit tests). Scratch lives in /tmp only.
set -u
REPO=$(git rev-parse --show-toplevel)
BIN=$REPO/cli/target/debug/zharness
(cd "$REPO/cli" && cargo build > /dev/null 2>&1) || { echo "SMOKE_FAIL=1 (build)"; exit 1; }
S=$(mktemp -d /tmp/p2gate-smoke.XXXXXX)
export HOME=$S/home
export XDG_CONFIG_HOME=$S/xdg
mkdir -p "$HOME" "$XDG_CONFIG_HOME" "$S/outside"
fail=0
chk() { # chk <label> <expected> <actual>
  if [ "$2" = "$3" ]; then echo "PASS  $1"; else echo "FAIL  $1 (expected [$2] got [$3])"; fail=1; fi
}

# ---------------------------------------------------------------- A: AGENTS.md
A=$S/a; mkdir -p "$A"; (cd "$A" && git init -q .)
ln -s "$S/outside/A-AGENTS.md" "$A/AGENTS.md"
"$BIN" install --root "$A" > "$S/a.out" 2>&1; rc=$?
chk "A install exit 0" 0 "$rc"
chk "A link target never created" "no" "$([ -e "$S/outside/A-AGENTS.md" ] && echo YES || echo no)"
chk "A AGENTS.md is a regular file" "regular file" "$(stat -c %F "$A/AGENTS.md")"

# ------------------------------------------------- B: predictable temp path
B=$S/b; mkdir -p "$B"; (cd "$B" && git init -q .)
"$BIN" install --root "$B" > /dev/null 2>&1
printf 'outside bytes\n' > "$S/outside/B-tmp.txt"
ln -s "$S/outside/B-tmp.txt" "$B/docs/WORKFLOW.md.tmp-zharness"
"$BIN" update --root "$B" > "$S/b.out" 2>&1; rc=$?
chk "B update exit 0" 0 "$rc"
chk "B temp link target untouched" "outside bytes" "$(cat "$S/outside/B-tmp.txt")"
chk "B WORKFLOW.md rewritten as regular file" "regular file" "$(stat -c %F "$B/docs/WORKFLOW.md")"
chk "B no temp entry left" "no" "$([ -e "$B/docs/WORKFLOW.md.tmp-zharness" ] && echo YES || echo no)"

# ------------------------------------------- C: uninstall restore of original
C=$S/c; mkdir -p "$C/docs"; (cd "$C" && git init -q .)
printf '# my workflow, written before zharness\n' > "$C/docs/WORKFLOW.md"
"$BIN" install --root "$C" > /dev/null 2>&1
cp "$C/docs/WORKFLOW.md" "$S/outside/C-base.md"   # the recorded base bytes
cp "$C/docs/WORKFLOW.md" "$S/base-copy.md"        # independent copy to compare against
rm "$C/docs/WORKFLOW.md"
ln -s "$S/outside/C-base.md" "$C/docs/WORKFLOW.md"
"$BIN" uninstall --root "$C" > "$S/c.out" 2>&1; rc=$?
chk "C uninstall exit 0" 0 "$rc"
chk "C link target kept the base bytes" "$(cat "$S/base-copy.md")" "$(cat "$S/outside/C-base.md")"
chk "C link target not overwritten with the original" "no" "$(grep -q 'written before zharness' "$S/outside/C-base.md" && echo YES || echo no)"
chk "C managed path is a regular file" "regular file" "$(stat -c %F "$C/docs/WORKFLOW.md")"
chk "C managed path holds the restored original" "# my workflow, written before zharness" "$(cat "$C/docs/WORKFLOW.md")"

echo "--- scratch: $S ---"
echo "SMOKE_FAIL=$fail"
