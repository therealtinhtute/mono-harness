# Best Practices & Anti-Patterns

## Best Practices

1. `set -uo pipefail` — never `set -e` (breaks UI return codes)
2. `exec < /dev/tty > /dev/tty 2>&1` — when script is piped/curled
3. `trap 'show_cursor' EXIT INT TERM` — always restore terminal
4. Check root **before** any UI renders
5. ESC = cancel on all components — never use `q`
6. `[[ -t 0 ]]` — detect no-TTY and skip interactive mode
7. `COLS=$(tput cols 2>/dev/null || echo 80)` — never hardcode width
8. `printf` over `echo` — portable, handles escape sequences correctly

## Anti-Patterns

| ❌ Don't | ✅ Do |
|---|---|
| `echo -e` for colors | `printf "%b"` |
| `set -e` with UI loops | `set -uo pipefail` |
| `q` to quit | ESC to cancel |
| `read` without `/dev/tty` in pipes | `< /dev/tty` always |
| `clear` bare | `clear 2>/dev/null \|\| printf "\033[2J\033[H"` |
