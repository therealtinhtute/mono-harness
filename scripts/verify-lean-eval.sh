#!/usr/bin/env bash
# Deterministic completeness check for one lean-workflow eval run directory
# (layout: docs/evals/lean-workflow.md, "Evidence record"). Checks evidence
# structure and hashes only; it does not score model behavior.
# Usage: bash scripts/verify-lean-eval.sh <run-dir>
set -euo pipefail

if [ $# -ne 1 ] || [ ! -d "$1" ]; then
  echo "usage: verify-lean-eval.sh <run-dir>" >&2; exit 2
fi
RUN="$1"
errs=0
err() { echo "verify-lean-eval: $*" >&2; errs=$((errs + 1)); }

# get <file> <key>: value of key=value (first match), empty if absent.
get() { sed -n "s/^$2=//p" "$1" | head -n1; }

RUN_KEYS="run_id kind cases arms trials_per_cell case_set_hash runtime model effort"
TRIAL_KEYS="case arm trial status source_sha instr_hash fixture_hash request_sha256 runtime model effort instruction_author fixture_author executor reviewer exposure global_config_loaded interruptions"
METRIC_KEYS="clarifications instr_bytes_read redundant_checks elapsed_s tokens_in tokens_out cost_usd"
FILES="request.txt trace.jsonl before.txt after.txt diff.patch exits.tsv predicates.tsv trial.env"

ENV="$RUN/run.env"
if [ ! -f "$ENV" ]; then
  err "missing run.env"; echo "verify-lean-eval: FAIL ($errs)"; exit 1
fi
for k in $RUN_KEYS; do [ -n "$(get "$ENV" "$k")" ] || err "run.env: missing or empty $k"; done

cases="$(get "$ENV" cases)"; arms="$(get "$ENV" arms)"; n="$(get "$ENV" trials_per_cell)"
case "$n" in ''|*[!0-9]*|0) err "run.env: trials_per_cell must be a positive integer"; n=0 ;; esac
for c in $cases; do [ -n "$(get "$ENV" "${c}_fixture_hash")" ] || err "run.env: missing ${c}_fixture_hash"; done
# case_set_hash binds the run to the committed fixture generator and oracle.
EVD="$(cd "$(dirname "$0")/.." && pwd)/docs/evals/evidence"
want_set="$(cat "$EVD/lean-fixture-generator-v1.txt" "$EVD/lean-oracle-v1.txt" | sha256sum | cut -d' ' -f1)"
[ "$(get "$ENV" case_set_hash)" = "$want_set" ] || err "run.env: case_set_hash differs from generator+oracle ($want_set)"
for a in $arms; do
  for k in source_sha instr_hash; do [ -n "$(get "$ENV" "${a}_$k")" ] || err "run.env: missing ${a}_$k"; done
done

# Required matrix: exactly cases x arms x 1..n.
expected=""
for c in $cases; do for a in $arms; do for i in $(seq 1 "$n"); do expected="$expected $c-$a-$i"; done; done; done
for id in $expected; do [ -d "$RUN/trials/$id" ] || err "matrix: missing trial $id"; done
if [ -d "$RUN/trials" ]; then
  for t in "$RUN"/trials/*/; do
    id="$(basename "$t")"
    case " $expected " in *" $id "*) ;; *) err "matrix: unexpected trial $id" ;; esac
  done
fi

for id in $expected; do
  t="$RUN/trials/$id"; [ -d "$t" ] || continue
  te="$t/trial.env"
  for f in $FILES; do [ -s "$t/$f" ] || err "$id: missing or empty $f"; done
  if [ -f "$t/SHA256SUMS" ]; then
    for f in $FILES; do grep -q "  $f\$" "$t/SHA256SUMS" || err "$id: SHA256SUMS does not cover $f"; done
    (cd "$t" && sha256sum --quiet -c SHA256SUMS >/dev/null 2>&1) || err "$id: SHA256SUMS mismatch"
  else
    err "$id: missing SHA256SUMS"
  fi
  [ -f "$te" ] || continue
  for k in $TRIAL_KEYS $METRIC_KEYS; do [ -n "$(get "$te" "$k")" ] || err "$id: missing or empty $k"; done

  IFS=- read -r c a i <<<"$id"
  [ "$(get "$te" case)" = "$c" ] || err "$id: case label does not match directory"
  [ "$(get "$te" arm)" = "$a" ] || err "$id: arm label does not match directory"
  [ "$(get "$te" trial)" = "$i" ] || err "$id: trial label does not match directory"
  [ "$(get "$te" status)" = scored ] || err "$id: status $(get "$te" status) is not a complete scored trial"
  [ "$(get "$te" global_config_loaded)" = no ] || err "$id: global config isolation not shown"
  [ "$(get "$te" source_sha)" = "$(get "$ENV" "${a}_source_sha")" ] || err "$id: source_sha differs from run.env ${a}_source_sha"
  [ "$(get "$te" instr_hash)" = "$(get "$ENV" "${a}_instr_hash")" ] || err "$id: instr_hash differs from run.env ${a}_instr_hash"
  [ "$(get "$te" fixture_hash)" = "$(get "$ENV" "${c}_fixture_hash")" ] || err "$id: fixture_hash differs from run.env ${c}_fixture_hash"
  for k in runtime model effort; do
    [ "$(get "$te" "$k")" = "$(get "$ENV" "$k")" ] || err "$id: $k differs from run.env"
  done
  if [ -f "$t/request.txt" ]; then
    [ "$(get "$te" request_sha256)" = "$(sha256sum "$t/request.txt" | cut -d' ' -f1)" ] || err "$id: request_sha256 mismatch"
  fi
  for k in $METRIC_KEYS; do
    v="$(get "$te" "$k")"
    [[ "$v" =~ ^[0-9]+(\.[0-9]+)?$ || "$v" =~ ^null:.+$ ]] || err "$id: $k must be a number or null:<reason>"
  done
  if [ -s "$t/predicates.tsv" ]; then
    awk -F'\t' 'NF!=2 || ($2!="pass" && $2!="fail") {bad=1} END {exit bad}' "$t/predicates.tsv" \
      || err "$id: predicates.tsv rows must be <name>TAB<pass|fail>"
  fi
done

if [ "$errs" -ne 0 ]; then echo "verify-lean-eval: FAIL ($errs)"; exit 1; fi
echo "verify-lean-eval: OK ($(echo $expected | wc -w) trials)"
