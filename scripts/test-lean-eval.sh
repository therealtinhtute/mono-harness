#!/usr/bin/env bash
# Red/green fixtures for scripts/verify-lean-eval.sh: one complete synthetic run
# must pass; each negative mutation must fail. Checks evidence structure only.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VERIFY="$ROOT/scripts/verify-lean-eval.sh"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

fails=0

# make_run <dir>: a complete 2-case x 2-arm x 2-trial run.
make_run() {
  local d="$1" c a n t
  mkdir -p "$d/trials"
  local set_hash
  set_hash="$(cat "$ROOT/docs/evals/evidence/lean-fixture-generator-v1.txt" "$ROOT/docs/evals/evidence/lean-oracle-v1.txt" | sha256sum | cut -d' ' -f1)"
  cat > "$d/run.env" <<EOF
run_id=synthetic-001
kind=dev
cases=D01 D02
arms=B C
trials_per_cell=2
case_set_hash=$set_hash
D01_fixture_hash=2222
D02_fixture_hash=2222
B_source_sha=da9fed6a09e34d20f1e586577cf936d74905b7bb
B_instr_hash=aaaa
C_source_sha=cccccccccccccccccccccccccccccccccccccccc
C_instr_hash=cccc
runtime=claude-code 0.0.0-synthetic
model=claude-opus-5-5
effort=high
EOF
  for c in D01 D02; do
    for a in B C; do
      for n in 1 2; do
        t="$d/trials/$c-$a-$n"
        mkdir -p "$t"
        printf 'synthetic request for %s\n' "$c" > "$t/request.txt"
        printf '{"seq":1,"tool":"Read","path":"skills/work/SKILL.md"}\n' > "$t/trace.jsonl"
        printf 'src/a.py\n' > "$t/before.txt"
        printf 'src/a.py\nsrc/b.py\n' > "$t/after.txt"
        printf -- '--- a\n+++ b\n' > "$t/diff.patch"
        printf 'python3 -m unittest -q\t0\n' > "$t/exits.tsv"
        printf 'correct_result\tpass\nno_unrelated_edits\tpass\n' > "$t/predicates.tsv"
        local ih sha
        if [ "$a" = B ]; then ih=aaaa; sha=da9fed6a09e34d20f1e586577cf936d74905b7bb
        else ih=cccc; sha=cccccccccccccccccccccccccccccccccccccccc; fi
        cat > "$t/trial.env" <<EOF
case=$c
arm=$a
trial=$n
status=scored
source_sha=$sha
instr_hash=$ih
fixture_hash=2222
request_sha256=$(sha256sum "$t/request.txt" | cut -d' ' -f1)
runtime=claude-code 0.0.0-synthetic
model=claude-opus-5-5
effort=high
instruction_author=lead
fixture_author=lead
executor=exec-1
reviewer=rev-1
exposure=none
global_config_loaded=no
interruptions=0
clarifications=0
instr_bytes_read=1200
redundant_checks=0
elapsed_s=90
tokens_in=null:host export unavailable
tokens_out=null:host export unavailable
cost_usd=null:host export unavailable
EOF
        (cd "$t" && sha256sum request.txt trace.jsonl before.txt after.txt diff.patch exits.tsv predicates.tsv trial.env > SHA256SUMS)
      done
    done
  done
}

# expect <pass|fail> <name> <mutation-command>
expect() {
  local want="$1" name="$2" mut="$3" d got
  d="$WORK/$name"
  make_run "$d"
  (cd "$d" && eval "$mut")
  if bash "$VERIFY" "$d" >"$WORK/$name.out" 2>&1; then got=pass; else got=fail; fi
  if [ "$got" = "$want" ]; then
    echo "ok   $name ($got)"
  else
    echo "FAIL $name: want $want, got $got"; sed 's/^/     /' "$WORK/$name.out"
    fails=$((fails + 1))
  fi
}

resum() { echo "(cd trials/$1 && sha256sum request.txt trace.jsonl before.txt after.txt diff.patch exits.tsv predicates.tsv trial.env > SHA256SUMS)"; }

expect pass complete            ':'
expect fail no-run-env          'rm run.env'
expect fail run-key-empty       'sed -i "s/^model=.*/model=/" run.env'
expect fail missing-trial       'rm -r trials/D02-C-2'
expect fail extra-trial         'cp -r trials/D02-C-2 trials/D02-C-3'
expect fail absent-trace        'rm trials/D01-B-1/trace.jsonl'
expect fail empty-trace         ": > trials/D01-B-1/trace.jsonl; $(resum D01-B-1)"
expect fail tampered-artifact   'echo x >> trials/D01-C-1/diff.patch'
expect fail missing-sums        'rm trials/D01-C-2/SHA256SUMS'
expect fail request-hash        "sed -i 's/^request_sha256=.*/request_sha256=0000/' trials/D02-B-1/trial.env; $(resum D02-B-1)"
expect fail arm-instr-mismatch  "sed -i 's/^instr_hash=.*/instr_hash=aaaa/' trials/D02-C-1/trial.env; $(resum D02-C-1)"
expect fail model-drift         "sed -i 's/^model=.*/model=claude-sonnet-5-5/' trials/D01-C-2/trial.env; $(resum D01-C-2)"
expect fail dir-label-mismatch  "sed -i 's/^arm=.*/arm=C/' trials/D01-B-2/trial.env; $(resum D01-B-2)"
expect fail global-config       "sed -i 's/^global_config_loaded=.*/global_config_loaded=yes/' trials/D02-B-2/trial.env; $(resum D02-B-2)"
expect fail blocked-trial       "sed -i 's/^status=.*/status=blocked/' trials/D02-B-2/trial.env; $(resum D02-B-2)"
expect fail bare-null-metric    "sed -i 's/^tokens_in=.*/tokens_in=null/' trials/D01-B-1/trial.env; $(resum D01-B-1)"
expect fail guessed-metric      "sed -i 's/^elapsed_s=.*/elapsed_s=about 90/' trials/D01-B-1/trial.env; $(resum D01-B-1)"
expect fail bad-predicate       "printf 'correct_result\tmaybe\n' > trials/D01-C-1/predicates.tsv; $(resum D01-C-1)"
expect fail case-set-mismatch   'sed -i "s/^case_set_hash=.*/case_set_hash=1111/" run.env'
expect fail fixture-mismatch    "sed -i 's/^fixture_hash=.*/fixture_hash=3333/' trials/D02-C-1/trial.env; $(resum D02-C-1)"
expect fail missing-trial-key   "sed -i '/^reviewer=/d' trials/D02-C-2/trial.env; $(resum D02-C-2)"

if [ "$fails" -ne 0 ]; then
  echo "test-lean-eval: $fails case(s) failed"; exit 1
fi
echo "test-lean-eval: all cases passed"
