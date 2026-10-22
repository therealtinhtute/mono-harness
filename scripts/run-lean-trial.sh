#!/usr/bin/env bash
# Run one lean-workflow trial and write its evidence directory
# (layout: docs/evals/lean-workflow.md, "Evidence record"). Captures only; it
# does not score semantic predicates. PAID: each session calls the model.
# Usage: bash scripts/run-lean-trial.sh <run-dir> <D01|D02> <arm-label> <source-sha> <trial-n>
# Env: LEAN_EFFORT (required, recorded), LEAN_MODEL (default claude-opus-5-5),
#      LEAN_EXECUTOR / LEAN_REVIEWER (recorded identities, default unset -> "unrecorded")
set -euo pipefail

[ $# -eq 5 ] || { echo "usage: run-lean-trial.sh <run-dir> <D01|D02> <arm> <sha> <n>" >&2; exit 2; }
RUN="$1"; CASE="$2"; ARM="$3"; SHA="$4"; N="$5"
: "${LEAN_EFFORT:?set LEAN_EFFORT (recorded in run.env/trial.env)}"
MODEL="${LEAN_MODEL:-claude-opus-5-5}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
EVD="$ROOT/docs/evals/evidence"
SKILLS="brainstorm to-plan work check handoff watzup"
T="$RUN/trials/$CASE-$ARM-$N"
[ ! -e "$T" ] || { echo "trial exists: $T" >&2; exit 1; }
mkdir -p "$T"
WORK="$(mktemp -d)"; FIX="$WORK/fixture"
trap 'rm -rf "$WORK"' EXIT

# 1. fixture, then the arm overlay (source SHA, never a moving ref)
bash "$EVD/lean-fixture-generator-v1.txt" "$CASE" "$FIX" >/dev/null
FIXTURE_HASH="$(git -C "$FIX" rev-parse 'HEAD^{tree}')"
(cd "$ROOT" && git archive "$SHA" AGENTS.md docs/WORKFLOW.md $(for s in $SKILLS; do echo "skills/stages/$s"; done)) | tar -x -C "$FIX"
mkdir -p "$FIX/.claude/skills"
for s in $SKILLS; do mv "$FIX/skills/stages/$s" "$FIX/.claude/skills/$s"; done
rm -rf "$FIX/skills"
INSTR_HASH="$(cd "$ROOT" && git ls-tree -r "$SHA" -- AGENTS.md docs/WORKFLOW.md $(for s in $SKILLS; do echo "skills/stages/$s"; done) | sha256sum | cut -d' ' -f1)"

snap() { (cd "$FIX" && { git status --porcelain -uall; echo "--- sha256 ---"; find . -path ./.git -prune -o -type f -print0 | sort -z | xargs -0 sha256sum; }); }

case "$CASE" in
  D01) REQ="Rename tally.core.count_words to word_count everywhere and keep the tests green." ;;
  D02) REQ="Implement the tally-filters plan." ;;
  *) echo "runner supports D01 and D02 (pilot scope)" >&2; exit 2 ;;
esac
printf '%s\n' "$REQ" > "$T/request.txt"
snap > "$T/before.txt"

# 2. session(s); project-level isolation per lean-workflow.md "Pinned configuration"
claude_run() { # <request> <trace-out>
  (cd "$FIX" && timeout 1200 claude -p "$1" --setting-sources project,local --strict-mcp-config \
    --model "$MODEL" --effort "$LEAN_EFFORT" --permission-mode bypassPermissions \
    --output-format stream-json --verbose) > "$2" 2> "$2.stderr"
}
START=$(date +%s)
INTERRUPTIONS=0
if [ "$CASE" = D02 ]; then
  # stop after R1's acceptance passes and its `done` Log entry is visible, before any R2 edit
  claude_run "$REQ" "$WORK/s1.jsonl" &
  PID=$!
  while kill -0 "$PID" 2>/dev/null; do
    if grep -qE '^\s*-.*R1.*\bdone\b' "$FIX/docs/plans/active/tally-filters.md" 2>/dev/null; then
      sleep 1; kill "$PID" 2>/dev/null || true; INTERRUPTIONS=1; break
    fi
    sleep 1
  done
  wait "$PID" 2>/dev/null || true
  snap > "$WORK/checkpoint.txt"
  cp "$WORK/checkpoint.txt" "$T/checkpoint.txt"
  git -C "$FIX" diff > "$T/checkpoint.diff"
  claude_run "Resume the tally-filters initiative." "$WORK/s2.jsonl" || true
  cat "$WORK/s1.jsonl" "$WORK/s2.jsonl" > "$T/trace.jsonl"
  printf '%s\nResume the tally-filters initiative.\n' "$REQ" > "$T/request.txt.full"
else
  claude_run "$REQ" "$T/trace.jsonl" || true
fi
END=$(date +%s)

# 3. final state, diff, oracle
snap > "$T/after.txt"
(cd "$FIX" && git add -N . && git diff) > "$T/diff.patch"
[ -s "$T/diff.patch" ] || echo "# no changes" > "$T/diff.patch"
(cd "$FIX" && python3 -I "$EVD/lean-oracle-v1.txt" "$CASE" "$FIX") > "$T/predicates.tsv" || true
[ -s "$T/predicates.tsv" ] || printf 'oracle_ran\tfail\n' > "$T/predicates.tsv"
{
  printf 'python3 -m unittest -q\t'
  (cd "$FIX" && python3 -m unittest -q >/dev/null 2>&1; echo $?)
} > "$T/exits.tsv"

# 4. metrics and isolation from the trace
python3 -I - "$T" "$FIX" "$((END - START))" > "$WORK/metrics.env" <<'PY'
import json, os, sys
t, fix, elapsed = sys.argv[1], sys.argv[2], sys.argv[3]
ev = []
for line in open(os.path.join(t, "trace.jsonl"), encoding="utf-8", errors="replace"):
    try: ev.append(json.loads(line))
    except Exception: pass
inits = [e for e in ev if e.get("type") == "system" and e.get("subtype") == "init"]
results = [e for e in ev if e.get("type") == "result"]
global_skills = set(os.listdir(os.path.expanduser("~/.claude/skills"))) if os.path.isdir(os.path.expanduser("~/.claude/skills")) else set()
arm_skills = set(os.listdir(os.path.join(fix, ".claude", "skills")))
loaded = set()
mcp = False
for i in inits:
    loaded |= set(i.get("skills") or []) | set(s.lstrip("/") for s in i.get("slash_commands") or [])
    mcp = mcp or bool(i.get("mcp_servers"))
leak = (loaded & global_skills) - arm_skills
isolated = bool(inits) and not leak and not mcp
instr = 0; clar = 0; redundant = 0
last_pass = {}; since_change = {}
pending = {}
for e in ev:
    if e.get("type") == "assistant":
        for c in e.get("message", {}).get("content", []):
            if c.get("type") != "tool_use": continue
            n, inp = c.get("name"), c.get("input", {})
            if n == "AskUserQuestion": clar += 1
            if n == "Read":
                p = inp.get("file_path", "")
                if any(k in p for k in ("AGENTS.md", "WORKFLOW.md", "/.claude/skills/")) and os.path.isfile(p):
                    instr += os.path.getsize(p)
            if n in ("Edit", "Write", "NotebookEdit"):
                last_pass.clear()
            if n == "Bash":
                pending[c["id"]] = inp.get("command", "")
    elif e.get("type") == "user":
        for c in (e.get("message", {}).get("content") or []):
            if isinstance(c, dict) and c.get("type") == "tool_result" and c.get("tool_use_id") in pending:
                cmd = pending.pop(c["tool_use_id"])
                if not c.get("is_error"):
                    if cmd in last_pass: redundant += 1
                    last_pass[cmd] = True
cost = sum(r.get("total_cost_usd") or 0 for r in results)
tin = sum(((r.get("usage") or {}).get("input_tokens") or 0) + ((r.get("usage") or {}).get("cache_read_input_tokens") or 0) + ((r.get("usage") or {}).get("cache_creation_input_tokens") or 0) for r in results)
tout = sum((r.get("usage") or {}).get("output_tokens") or 0 for r in results)
have = bool(results)
sids = ",".join(i.get("session_id", "") for i in inits)
print(f"global_config_loaded={'no' if isolated else 'yes'}")
print(f"clarifications={clar}\ninstr_bytes_read={instr}\nredundant_checks={redundant}\nelapsed_s={elapsed}")
print(f"tokens_in={tin if have else 'null:no result event'}")
print(f"tokens_out={tout if have else 'null:no result event'}")
print(f"cost_usd={round(cost, 4) if have else 'null:no result event'}")
print(f"session_ids={sids}")
PY

# 5. trial.env and checksums
{
  echo "case=$CASE"; echo "arm=$ARM"; echo "trial=$N"
  echo "status=$([ -s "$T/trace.jsonl" ] && grep -q '"type":"result"' "$T/trace.jsonl" && echo scored || echo blocked)"
  echo "source_sha=$SHA"; echo "instr_hash=$INSTR_HASH"; echo "fixture_hash=$FIXTURE_HASH"
  echo "request_sha256=$(sha256sum "$T/request.txt" | cut -d' ' -f1)"
  echo "runtime=claude-code $(claude --version | cut -d' ' -f1)"; echo "model=$MODEL"; echo "effort=$LEAN_EFFORT"
  echo "instruction_author=unrecorded"; echo "fixture_author=lead"
  echo "executor=${LEAN_EXECUTOR:-unrecorded}"; echo "reviewer=${LEAN_REVIEWER:-unrecorded}"
  echo "exposure=none"; echo "interruptions=$INTERRUPTIONS"
  cat "$WORK/metrics.env"
} > "$T/trial.env"
(cd "$T" && sha256sum request.txt trace.jsonl before.txt after.txt diff.patch exits.tsv predicates.tsv trial.env > SHA256SUMS)
echo "trial written: $T ($(sed -n 's/^status=//p' "$T/trial.env"), cost $(sed -n 's/^cost_usd=//p' "$T/trial.env"))"
