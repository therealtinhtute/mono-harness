# Block Banner Generator Guide

Banners use Unicode block characters (`▄ █ ░ ▀`) for a bold retro-terminal look.
Always render in `${BOLD}${CYAN}` or `${BOLD}${MAGENTA}`.

## Online Generators

| Tool | URL | Recommended Font |
|---|---|---|
| patorjk Text to ASCII | https://patorjk.com/software/taag/ | `Block`, `ANSI Shadow`, `DOS Rebel` |
| fsymbols blocky | https://fsymbols.com/generators/blocky/ | Default |
| FIGlet web | http://www.figlet.org/figlet-cgi.pl | `block`, `banner3` |

**Best settings for patorjk:**
- Font: `Block` or `ANSI Shadow`
- Character width: Full
- Character height: Default
- Horizontal layout: Full

## Local Generation with figlet

```bash
# Install
sudo apt-get install figlet toilet

# Generate block style
figlet -f block "MY TOOL"
figlet -f banner3 "MY TOOL"
toilet -f mono9 "MY TOOL"
toilet -f pagga "MY TOOL"  # box-drawing style

# With color (toilet)
toilet -f mono9 --gay "MY TOOL"
toilet -f pagga -F metal "MY TOOL"
```

## Using in show_banner()

Copy generated art into `show_banner()`. Wrap in `${BOLD}${CYAN}`:

```bash
show_banner() {
  printf "%b" "${BOLD}${CYAN}"
  # Paste your generated art here — one printf per line:
  printf "▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄\n"
  printf "█░▄▄░█░█▀█░██░█░▄▄▀███▀▄▀█░██░▄▄█░▄▄▀█░▄▄▀██\n"
  printf "█░▀▀░█░▄▀█░██░█░▄▄▀███░█▀█░██░▄▄█░▀▀░█░██░██\n"
  printf "████░█▄█▄██▄▄▄█▄▄▄▄████▄██▄▄█▄▄▄█▄██▄█▄██▄██\n"
  printf "▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀\n"
  printf "%b" "${RESET}"
}
```

## Color Variants

| Style | Code |
|---|---|
| Cyan (default) | `${BOLD}${CYAN}` |
| Magenta | `${BOLD}${MAGENTA}` |
| Green (matrix) | `${BOLD}${GREEN}` |
| Yellow (warning) | `${BOLD}${YELLOW}` |
| 256-color orange | `\033[38;5;214m` |

## Separator Line

Always follow the banner with a dim separator:

```bash
sep_line() {
  printf "%b────────────────────────────────────────────────────────%b\n" \
    "${DIM}" "${RESET}"
}
```

## Compact Banner (one-liner)

When terminal width is limited (< 60 cols):

```bash
show_banner_compact() {
  printf "%b[ %s ]%b\n" "${BOLD}${CYAN}" "MY TOOL v1.0" "${RESET}"
}
```

## Version / Subtitle Line

Add below the banner art, before `sep_line()`:

```bash
show_banner() {
  printf "%b" "${BOLD}${CYAN}"
  printf "... art ...\n"
  printf "%b" "${RESET}"
  printf " %bv%s%b  %b%s%b\n" \
    "${DIM}" "${VERSION:-1.0}" "${RESET}" \
    "${DIM}" "Ubuntu Cleanup Tool" "${RESET}"
}
```
