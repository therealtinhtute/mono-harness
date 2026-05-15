#!/usr/bin/env bash
# ================================================
# bash-tui lib.sh — Reusable TUI Component Library
# Source this file: source "path/to/lib.sh"
# ================================================

# ── COLORS ──────────────────────────────────────
if command -v tput >/dev/null 2>&1 && tput colors >/dev/null 2>&1 && [ "$(tput colors)" -ge 8 ]; then
  BOLD="$(tput bold)";      DIM="$(tput dim)"
  RED="$(tput setaf 1)";    GREEN="$(tput setaf 2)"
  YELLOW="$(tput setaf 3)"; BLUE="$(tput setaf 4)"
  MAGENTA="$(tput setaf 5)";CYAN="$(tput setaf 6)"
  WHITE="$(tput setaf 7)";  RESET="$(tput sgr0)"
else
  BOLD="" DIM="" RED="" GREEN="" YELLOW="" BLUE="" MAGENTA="" CYAN="" WHITE="" RESET=""
fi

# ── SCREEN ──────────────────────────────────────
clear_screen() { printf "\033[2J\033[H"; }
hide_cursor()  { printf "\033[?25l"; }
show_cursor()  { printf "\033[?25h"; }
save_cursor()  { printf "\033[s"; }
restore_cursor(){ printf "\033[u"; }
erase_line()   { printf "\033[2K\r"; }
move_up()      { printf "\033[%dA" "${1:-1}"; }
sep_line()     { printf "%b────────────────────────────────────────────────────────%b\n" "${DIM}" "${RESET}"; }

# Restore cursor on any exit
trap 'show_cursor' EXIT INT TERM

# ── BANNER ──────────────────────────────────────
# Override show_banner() in your script with your own block-art.
# Default: generic placeholder banner.
show_banner() {
  printf "%b" "${BOLD}${CYAN}"
  printf "▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄\n"
  printf "█  YOUR TOOL NAME — override show_banner()  █\n"
  printf "▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀\n"
  printf "%b" "${RESET}"
}

_render_header() {
  clear_screen
  show_banner
  sep_line
}

# ── HINT BAR ────────────────────────────────────
# hint_bar "↑↓" "move" "Space" "toggle" "Enter" "confirm" "Esc" "cancel"
# Renders: ──────────────────────────────────────────────────────
#           ↑↓  move   •   Space  toggle   •   Enter  confirm   •   Esc  cancel
hint_bar() {
  sep_line
  local out=""
  while [ $# -ge 2 ]; do
    out+="${BOLD}$1${RESET}  ${DIM}$2${RESET}"
    shift 2
    [ $# -ge 2 ] && out+="${DIM}   •   ${RESET}"
  done
  printf " %b\n" "$out"
}

# ── KEYBOARD INPUT ──────────────────────────────
# Returns: up | down | left | right | enter | space | esc | <char>
read_key() {
  local key key2
  IFS= read -rsn1 key < /dev/tty
  case "$key" in
    $'\e')
      IFS= read -rsn2 -t 0.1 key2 < /dev/tty 2>/dev/null
      case "$key2" in
        '[A') printf "up"    ;;
        '[B') printf "down"  ;;
        '[C') printf "right" ;;
        '[D') printf "left"  ;;
        '')   printf "esc"   ;;
        *)    printf "esc"   ;;
      esac ;;
    "")  printf "enter" ;;
    " ") printf "space" ;;
    *)   printf "%s" "$key" ;;
  esac
}

# ── YES / NO ────────────────────────────────────
# Usage: yes_no "Question?" && on_yes || on_no
# Returns: 0 = Yes, 1 = No/Esc
yes_no() {
  local prompt="$1"
  local selected=0   # 0 = Yes, 1 = No
  hide_cursor
  trap 'show_cursor' RETURN

  while true; do
    _render_header
    printf "\n %b%s%b\n\n" "${BOLD}${BLUE}" "$prompt" "${RESET}"

    if [ "$selected" -eq 0 ]; then
      printf " %b➤ ● Yes%b\n" "${GREEN}" "${RESET}"
      printf "   %b○ No%b\n"  "${DIM}"   "${RESET}"
    else
      printf "   %b○ Yes%b\n" "${DIM}"   "${RESET}"
      printf " %b➤ ● No%b\n"  "${GREEN}" "${RESET}"
    fi

    printf "\n"
    hint_bar "↑↓" "move" "Enter" "confirm" "Esc" "cancel"

    case "$(read_key)" in
      up|down) selected=$(( 1 - selected )) ;;
      enter)   return "$selected" ;;
      esc)     show_cursor; return 1 ;;
    esac
  done
}

# ── MULTI-SELECT ────────────────────────────────
# Usage: multi_select "Title" "Item A" "Item B" "Item C"
# Result: SELECTED_ITEMS array
declare -a SELECTED_ITEMS=()

multi_select() {
  local title="$1"; shift
  local items=("$@")
  local n=${#items[@]}
  local checked=() selected=0
  for (( i=0; i<n; i++ )); do checked[i]=1; done  # all checked by default

  hide_cursor
  trap 'show_cursor' RETURN

  while true; do
    _render_header
    printf "\n %b%s%b\n\n" "${BOLD}${BLUE}" "$title" "${RESET}"

    for (( i=0; i<n; i++ )); do
      local mark
      [ "${checked[$i]}" -eq 1 ] \
        && mark="${GREEN}●${RESET}" \
        || mark="${DIM}○${RESET}"

      if [ "$i" -eq "$selected" ]; then
        printf " %b➤%b %b %s%b\n" "${GREEN}" "${RESET}" "$mark" "${items[$i]}" "${RESET}"
      else
        printf "    %b %s%b\n" "$mark" "${items[$i]}" "${RESET}"
      fi
    done

    printf "\n"
    hint_bar "↑↓" "move" "Space" "toggle" "Enter" "confirm" "Esc" "cancel"

    case "$(read_key)" in
      up)    selected=$(( (selected - 1 + n) % n )) ;;
      down)  selected=$(( (selected + 1) % n )) ;;
      space)
        [ "${checked[$selected]}" -eq 1 ] \
          && checked[$selected]=0 \
          || checked[$selected]=1 ;;
      enter)
        SELECTED_ITEMS=()
        for (( i=0; i<n; i++ )); do
          [ "${checked[$i]}" -eq 1 ] && SELECTED_ITEMS+=("${items[$i]}")
        done
        show_cursor; return 0 ;;
      esc) show_cursor; return 1 ;;
    esac
  done
}

# ── SINGLE SELECT ───────────────────────────────
# Usage: single_select "Title" "Opt A" "Opt B"
# Result: SELECTED_ITEM (string)
SELECTED_ITEM=""

single_select() {
  local title="$1"; shift
  local items=("$@")
  local n=${#items[@]}
  local selected=0
  hide_cursor
  trap 'show_cursor' RETURN

  while true; do
    _render_header
    printf "\n %b%s%b\n\n" "${BOLD}${BLUE}" "$title" "${RESET}"

    for (( i=0; i<n; i++ )); do
      if [ "$i" -eq "$selected" ]; then
        printf " %b➤ ● %s%b\n" "${GREEN}" "${items[$i]}" "${RESET}"
      else
        printf "   %b○ %s%b\n" "${DIM}"   "${items[$i]}" "${RESET}"
      fi
    done

    printf "\n"
    hint_bar "↑↓" "move" "Enter" "select" "Esc" "cancel"

    case "$(read_key)" in
      up)    selected=$(( (selected - 1 + n) % n )) ;;
      down)  selected=$(( (selected + 1) % n )) ;;
      enter) SELECTED_ITEM="${items[$selected]}"; show_cursor; return 0 ;;
      esc)   show_cursor; return 1 ;;
    esac
  done
}

# ── SPINNER ─────────────────────────────────────
# Usage: spinner "Message..." command [args...]
# Runs command in background, animates while waiting.
spinner() {
  local msg="$1"; shift
  local frames=('⠋' '⠙' '⠹' '⠸' '⠼' '⠴' '⠦' '⠧' '⠇' '⠏')
  local pid delay=0.08 i=0

  hide_cursor
  "$@" &
  pid=$!
  while kill -0 "$pid" 2>/dev/null; do
    printf "\r %b%s%b  %s  " "${CYAN}" "${frames[$i]}" "${RESET}" "$msg"
    i=$(( (i + 1) % ${#frames[@]} ))
    sleep "$delay"
  done
  wait "$pid"; local rc=$?
  if [ $rc -eq 0 ]; then
    printf "\r %b✔%b  %s\n" "${GREEN}" "${RESET}" "$msg"
  else
    printf "\r %b✖%b  %s\n" "${RED}"   "${RESET}" "$msg"
  fi
  show_cursor
  return $rc
}

# ── PROGRESS BAR ────────────────────────────────
# Usage: progress_bar <current> <total> "Label"
# Call in a loop; adds newline when current == total.
progress_bar() {
  local current="$1" total="$2" label="${3:-Progress}"
  local width=40
  local filled=$(( current * width / total ))
  local empty=$(( width - filled ))
  local pct=$(( current * 100 / total ))
  local bar_filled bar_empty
  bar_filled=$(printf '█%.0s' $(seq 1 "$filled"))
  bar_empty=$(printf  '░%.0s' $(seq 1 "$empty"))
  printf "\r %s %b%s%b%b%s%b %3d%%" \
    "$label" \
    "${GREEN}" "$bar_filled" "${RESET}" \
    "${DIM}"   "$bar_empty"  "${RESET}" \
    "$pct"
  [ "$current" -eq "$total" ] && printf "\n"
}

# ── NOTIFICATION HELPERS ─────────────────────────
ok()   { printf " %b✔%b  %s\n" "${GREEN}"   "${RESET}" "$1"; }
warn() { printf " %b⚠%b  %s\n" "${YELLOW}"  "${RESET}" "$1"; }
fail() { printf " %b✖%b  %s\n" "${RED}"     "${RESET}" "$1"; }
info() { printf " %b→%b  %s\n" "${CYAN}"    "${RESET}" "$1"; }

# ── STEP COUNTER ─────────────────────────────────
# Set STEP_TOTAL before your loop, then call step() each iteration.
STEP_CURRENT=0
STEP_TOTAL=1
step() {
  STEP_CURRENT=$(( STEP_CURRENT + 1 ))
  printf "\n %b◆ [%d/%d]  %s%b\n" \
    "${BOLD}${MAGENTA}" "$STEP_CURRENT" "$STEP_TOTAL" "$1" "${RESET}"
}

# ── RUN HELPERS ──────────────────────────────────
# run_cmd "Label" command [args...]   — shows → label, then ✔/✖
run_cmd() {
  local label="$1"; shift
  info "$label"
  if "$@" >/dev/null 2>&1; then
    ok "$label"
  else
    fail "$label"; return 1
  fi
}

# run_spinner "Label" command [args...] — spinner animation while running
run_spinner() {
  local label="$1"; shift
  spinner "$label" "$@"
}

# ── GUARDS ───────────────────────────────────────
require_root() {
  if [[ $EUID -ne 0 ]]; then
    warn "Root required."
    printf " %bRun:%b  sudo bash <(curl -fsSL YOUR_URL)\n" "${CYAN}" "${RESET}"
    exit 1
  fi
}

require_tty() {
  # Redirect I/O to TTY when script is piped/curled
  [[ -t 0 ]] || exec < /dev/tty > /dev/tty 2>&1
}

# ── UTILITY ──────────────────────────────────────
get_free_space() { df -h / | awk 'NR==2 {print $4}'; }
get_os()         { . /etc/os-release 2>/dev/null && echo "${PRETTY_NAME:-unknown}" || uname -s; }
command_exists() { command -v "$1" >/dev/null 2>&1; }
