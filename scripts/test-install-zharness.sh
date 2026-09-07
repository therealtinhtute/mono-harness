#!/usr/bin/env bash
set -euo pipefail

SCRIPT_UNDER_TEST="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/install-zharness.sh"
ROOT="$(mktemp -d)"
trap 'rm -rf "$ROOT"' EXIT

BIN_DIR="$ROOT/bin"
PASSED=0
FAILED=0

build_stub_gh() {
  mkdir -p "$BIN_DIR"
  cat >"$BIN_DIR/gh" <<'STUB_EOF'
#!/usr/bin/env bash
set -euo pipefail
if [ -z "${STUB_CALLS_LOG:-}" ] || [ -z "${STUB_RELEASES_TSV:-}" ]; then
  echo "stub gh: STUB_CALLS_LOG and STUB_RELEASES_TSV must be set" >&2
  exit 2
fi
printf 'invoked: gh %s\n' "$*" >>"$STUB_CALLS_LOG"
case "${1:-}/${2:-}" in
  release/list)
    cat "$STUB_RELEASES_TSV"
    ;;
  release/download)
    if [ -z "${3:-}" ]; then
      echo "stub gh: expected: gh release download <tag> ..." >&2
      exit 2
    fi
    printf 'downloaded-tag: %s\n' "$3" >>"$STUB_CALLS_LOG"
    target_dir="" target_asset="" prev_arg=""
    for arg in "$@"; do
      case "$prev_arg" in
        --dir) target_dir="$arg" ;;
        --pattern) target_asset="$arg" ;;
      esac
      prev_arg="$arg"
    done
    if [ -z "$target_dir" ] || [ -z "$target_asset" ]; then
      echo "stub gh: --dir and --pattern are required" >&2
      exit 2
    fi
    payload_dir="$(mktemp -d)"
    printf '#!/usr/bin/env bash\necho "zharness 9.9.9-test"\n' >"$payload_dir/zharness"
    chmod +x "$payload_dir/zharness"
    tar -czf "${target_dir}/${target_asset}" -C "$payload_dir" zharness
    rm -rf "$payload_dir"
    ;;
  *)
    echo "stub gh: unsupported invocation: $*" >&2
    exit 2
    ;;
esac
STUB_EOF
  chmod +x "$BIN_DIR/gh"
}

tsv_row() { printf '%s\t%s\t%s\n' "$1" "$2" "$3"; }

count_downloads() { grep -c '^downloaded-tag:' "$1" || true; }

expect_eq() {
  if [ "$2" = "$3" ]; then return 0; fi
  printf '    assert: %s\n      expected: [%s]\n      actual:   [%s]\n' "$1" "$2" "$3" >&2
  return 1
}

expect_ne() {
  if [ "$2" != "$3" ]; then return 0; fi
  printf '    assert: %s\n      value must not equal [%s]\n' "$1" "$3" >&2
  return 1
}

expect_file_has() {
  if grep -Fq -- "$3" "$2"; then return 0; fi
  printf '    assert: %s\n      [%s] missing from %s\n' "$1" "$3" "$2" >&2
  return 1
}

expect_file_lacks() {
  if ! grep -Fq -- "$3" "$2"; then return 0; fi
  printf '    assert: %s\n      [%s] unexpectedly present in %s\n' "$1" "$3" "$2" >&2
  return 1
}

begin_scenario() {
  local sdir="$ROOT/$1"
  mkdir -p "$sdir/install"
  CALLS_LOG="$sdir/calls.log"
  STUB_RELEASES_TSV="$sdir/releases.tsv"
  STUB_CALLS_LOG="$CALLS_LOG"
  INSTALL_OUT="$sdir/out.txt"
  INSTALL_ERR="$sdir/err.txt"
  INSTALL_TARGET_DIR="$sdir/install"
  : >"$CALLS_LOG"
  export STUB_CALLS_LOG STUB_RELEASES_TSV
}

invoke_install() {
  (
    export PATH="$BIN_DIR:$PATH"
    export ZHARNESS_INSTALL_DIR="$INSTALL_TARGET_DIR"
    bash "$SCRIPT_UNDER_TEST" "$@"
  )
}

scenario_s1_race_draft_newer_refuses() {
  local failures=0 exit_code=0
  begin_scenario "s1"
  { tsv_row true  2026-08-20T10:00:00Z v0.14.0
    tsv_row false 2026-08-18T09:00:00Z v0.13.0
    tsv_row false 2026-07-01T09:00:00Z v0.12.0
  } >"$STUB_RELEASES_TSV"
  invoke_install >"$INSTALL_OUT" 2>"$INSTALL_ERR" || exit_code=$?
  expect_ne "exit code must be non-zero" "$exit_code" "0" || failures=1
  expect_file_has "stderr mentions still publishing" "$INSTALL_ERR" "still publishing" || failures=1
  expect_eq "zero release download calls" 0 "$(count_downloads "$CALLS_LOG")" || failures=1
  return "$failures"
}

scenario_s2_stale_draft_installs_newest_published() {
  local failures=0 exit_code=0
  begin_scenario "s2"
  { tsv_row false 2026-08-18T09:00:00Z v0.13.0
    tsv_row false 2026-07-01T09:00:00Z v0.12.0
    tsv_row true  2026-06-15T09:00:00Z v0.11.9
  } >"$STUB_RELEASES_TSV"
  invoke_install >"$INSTALL_OUT" 2>"$INSTALL_ERR" || exit_code=$?
  expect_eq "exit code" 0 "$exit_code" || failures=1
  expect_eq "exactly one download" 1 "$(count_downloads "$CALLS_LOG")" || failures=1
  expect_file_has "download targets newest published tag" "$CALLS_LOG" "downloaded-tag: v0.13.0" || failures=1
  expect_file_has "installed binary answers --version" "$INSTALL_OUT" "zharness 9.9.9-test" || failures=1
  return "$failures"
}

scenario_s3_no_draft_installs_newest_published() {
  local failures=0 exit_code=0
  begin_scenario "s3"
  { tsv_row false 2026-07-01T09:00:00Z v0.12.0
    tsv_row false 2026-08-18T09:00:00Z v0.13.0
  } >"$STUB_RELEASES_TSV"
  invoke_install >"$INSTALL_OUT" 2>"$INSTALL_ERR" || exit_code=$?
  expect_eq "exit code" 0 "$exit_code" || failures=1
  expect_eq "exactly one download" 1 "$(count_downloads "$CALLS_LOG")" || failures=1
  expect_file_has "download targets newest published tag regardless of row order" "$CALLS_LOG" "downloaded-tag: v0.13.0" || failures=1
  expect_file_has "installed binary answers --version" "$INSTALL_OUT" "zharness 9.9.9-test" || failures=1
  return "$failures"
}

scenario_s4_explicit_tag_skips_resolution() {
  local failures=0 exit_code=0
  begin_scenario "s4"
  { tsv_row true  2026-08-20T10:00:00Z v0.14.0
    tsv_row false 2026-08-18T09:00:00Z v0.13.0
  } >"$STUB_RELEASES_TSV"
  invoke_install >"$INSTALL_OUT" 2>"$INSTALL_ERR" v0.12.0 || exit_code=$?
  expect_eq "exit code" 0 "$exit_code" || failures=1
  expect_file_lacks "zero release list calls" "$CALLS_LOG" "gh release list" || failures=1
  expect_eq "exactly one download" 1 "$(count_downloads "$CALLS_LOG")" || failures=1
  expect_file_has "downloads requested tag" "$CALLS_LOG" "downloaded-tag: v0.12.0" || failures=1
  expect_file_has "installed binary answers --version" "$INSTALL_OUT" "zharness 9.9.9-test" || failures=1
  return "$failures"
}

run_scenario() {
  local label="$1" scenario_fn="$2" rc=0
  printf -- '-- %s\n' "$label"
  "$scenario_fn" || rc=$?
  if [ "$rc" -eq 0 ]; then
    printf 'PASS: %s\n' "$label"
    PASSED=$((PASSED + 1))
  else
    printf 'FAIL: %s\n' "$label"
    FAILED=$((FAILED + 1))
  fi
  return 0
}

main() {
  # Safety: every scenario installs into ZHARNESS_INSTALL_DIR. If the script
  # under test predates that override, S2-S4 would install into the real
  # ~/.local/bin and clobber the live binary - refuse instead.
  if ! grep -q 'ZHARNESS_INSTALL_DIR' "$SCRIPT_UNDER_TEST"; then
    echo "error: $(basename "$SCRIPT_UNDER_TEST") lacks ZHARNESS_INSTALL_DIR support; refusing to run (it would install into the real ~/.local/bin)" >&2
    exit 2
  fi
  build_stub_gh
  printf '== regression suite for %s ==\n' "$(basename "$SCRIPT_UNDER_TEST")"
  run_scenario "S1 race: newest draft newer than newest published -> refuse before download" scenario_s1_race_draft_newer_refuses
  run_scenario "S2 stale draft older than newest published -> install proceeds" scenario_s2_stale_draft_installs_newest_published
  run_scenario "S3 no drafts -> install newest published" scenario_s3_no_draft_installs_newest_published
  run_scenario "S4 explicit tag -> skip resolution, download requested tag" scenario_s4_explicit_tag_skips_resolution
  printf 'summary: %d passed, %d failed\n' "$PASSED" "$FAILED"
  [ "$FAILED" -eq 0 ]
}

main
