#!/usr/bin/env bash
# ================================================
# Bash TUI Starter Template
# Uses: scripts/lib.sh
# Usage: copy this file + scripts/lib.sh to your project
# ================================================

set -uo pipefail

VERSION="1.0.0"

# ── Load component library ──────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../scripts/lib.sh"

# ── Override banner with YOUR tool name ─────────
# Generate at: https://patorjk.com/software/taag/ (font: Block)
# See: assets/banner-generator.md
show_banner() {
  printf "%b" "${BOLD}${CYAN}"
  printf "▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄\n"
  printf "█░▄▄░█░█▀█░██░█░▄▄▀███▀▄▀█░██░▄▄█░▄▄▀█░▄▄▀██\n"
  printf "█░▀▀░█░▄▀█░██░█░▄▄▀███░█▀█░██░▄▄█░▀▀░█░██░██\n"
  printf "████░█▄█▄██▄▄▄█▄▄▄▄████▄██▄▄█▄▄▄█▄██▄█▄██▄██\n"
  printf "▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀\n"
  printf "%b" "${RESET}"
}

# ── Your task functions ─────────────────────────
do_task_a() {
  step "Task A — System update"
  run_cmd   "Updating package lists..."  sleep 1
  run_spinner "Upgrading packages..."    sleep 2
}

do_task_b() {
  step "Task B — Log cleanup"
  run_spinner "Vacuuming journals..."    sleep 1
  ok "Logs cleaned"
}

# ── Main ────────────────────────────────────────
main() {
  require_tty
  # require_root  # uncomment if tasks need sudo

  _render_header
  printf "\n %bSystem: %s%b\n" "${DIM}" "$(get_os)" "${RESET}"
  printf " %bFree:   %s%b\n\n" "${DIM}" "$(get_free_space)" "${RESET}"

  # Step 1: Select tasks
  local items=("Task A — System update" "Task B — Log cleanup")
  multi_select "Select tasks to run:" "${items[@]}" || { warn "Cancelled."; exit 0; }

  if [ ${#SELECTED_ITEMS[@]} -eq 0 ]; then
    warn "Nothing selected."; exit 0
  fi

  # Step 2: Confirm
  yes_no "🚀  Start now?" || { warn "Cancelled."; exit 0; }

  # Step 3: Execute
  STEP_TOTAL=${#SELECTED_ITEMS[@]}
  STEP_CURRENT=0
  sep_line

  for item in "${SELECTED_ITEMS[@]}"; do
    case "$item" in
      "Task A — System update") do_task_a ;;
      "Task B — Log cleanup")   do_task_b ;;
    esac
  done

  # Step 4: Summary
  sep_line
  printf "\n %b✔  Done!%b  Free space now: %b%s%b\n\n" \
    "${BOLD}${GREEN}" "${RESET}" \
    "${CYAN}" "$(get_free_space)" "${RESET}"
}

main "$@"
