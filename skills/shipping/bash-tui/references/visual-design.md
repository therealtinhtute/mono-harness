# Visual Design System

## Banner — Block ASCII (`▄█▀`)

Always `${BOLD}${CYAN}` or `${BOLD}${MAGENTA}`. See `assets/banner-generator.md` to generate for your tool name.

```
▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄
█░▄▄░█░█▀█░██░█░▄▄▀███▀▄▀█░██░▄▄█░▄▄▀█░▄▄▀██
...
▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀
```

**Separator**: `────────────────────────────────────────────────────────` (always `${DIM}`)

## Icon Set (MANDATORY)

| Context | Icon |
|---|---|
| Cursor / selected | `➤` |
| Checked | `●` |
| Unchecked | `○` |
| Success | `✔` |
| Warning | `⚠` |
| Error | `✖` |
| Info / step | `→` |
| Running step | `◆` |
| Spinner | `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏` |

## Bottom Hint Bar (MANDATORY on every interactive screen)

```bash
hint_bar() {
  sep_line
  local out=""
  while [ $# -ge 2 ]; do
    out+="${BOLD}$1${RESET}  ${DIM}$2${RESET}"
    shift 2; [ $# -ge 2 ] && out+="${DIM}   •   ${RESET}"
  done
  printf " %b\n" "$out"
}
# Usage: hint_bar "↑↓" "move" "Space" "toggle" "Enter" "confirm" "Esc" "cancel"
```
