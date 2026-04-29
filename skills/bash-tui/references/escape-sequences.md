# ANSI Escape Sequences Reference

> Quick reference for bash TUI development. All codes use `\033` (octal) = `\x1b` (hex) = `\e` (bash).
> Use `printf "%b"` to interpret escape sequences — never `echo -e`.

## Table of Contents
1. [Text Formatting (SGR)](#text-formatting)
2. [8/16 Colors](#816-colors)
3. [256-Color Mode](#256-color-mode)
4. [True Color (RGB)](#true-color)
5. [Cursor Control](#cursor-control)
6. [Screen Control](#screen-control)
7. [Keyboard Input Sequences](#keyboard-input)
8. [tput Equivalents](#tput-equivalents)
9. [Compatibility Notes](#compatibility)

---

## Text Formatting (SGR)

`\033[<code>m`

| Code | Effect | Reset |
|---|---|---|
| `0` | Reset all | — |
| `1` | Bold | `22` |
| `2` | Dim / faint | `22` |
| `3` | Italic | `23` |
| `4` | Underline | `24` |
| `5` | Blink slow | `25` |
| `6` | Blink fast | `25` |
| `7` | Reverse (swap fg/bg) | `27` |
| `8` | Hidden / invisible | `28` |
| `9` | Strikethrough | `29` |

```bash
printf "\033[1mBold\033[0m\n"
printf "\033[1;4mBold + Underline\033[0m\n"
printf "\033[2;3mDim + Italic\033[0m\n"
```

---

## 8/16 Colors

### Foreground (text)

| Color | Normal | Bright |
|---|---|---|
| Black | `30` | `90` |
| Red | `31` | `91` |
| Green | `32` | `92` |
| Yellow | `33` | `93` |
| Blue | `34` | `94` |
| Magenta | `35` | `95` |
| Cyan | `36` | `96` |
| White | `37` | `97` |
| Default | `39` | — |

### Background

| Color | Normal | Bright |
|---|---|---|
| Black | `40` | `100` |
| Red | `41` | `101` |
| Green | `42` | `102` |
| Yellow | `43` | `103` |
| Blue | `44` | `104` |
| Magenta | `45` | `105` |
| Cyan | `46` | `106` |
| White | `47` | `107` |
| Default | `49` | — |

```bash
# Combine: \033[<style>;<fg>;<bg>m
printf "\033[1;32mGreen bold\033[0m\n"
printf "\033[33;41mYellow on red\033[0m\n"
printf "\033[1;97;44mWhite bold on blue\033[0m\n"
```

---

## 256-Color Mode

`\033[38;5;<n>m` — foreground  
`\033[48;5;<n>m` — background  

| Range | Description |
|---|---|
| 0–7 | Standard colors (same as 30–37) |
| 8–15 | Bright colors (same as 90–97) |
| 16–231 | 6×6×6 RGB cube |
| 232–255 | Grayscale ramp (dark → light) |

```bash
# Print color palette
for i in {0..255}; do
  printf "\033[38;5;%dm %3d\033[0m" "$i" "$i"
  [ $(( (i+1) % 16 )) -eq 0 ] && printf "\n"
done

# Use specific color
printf "\033[38;5;214mOrange text\033[0m\n"
printf "\033[48;5;57mPurple background\033[0m\n"
```

**Useful 256-color indices:**
| Color | Index |
|---|---|
| Orange | 214 |
| Pink | 213 |
| Purple | 135 |
| Teal | 43 |
| Gold | 220 |
| Sky blue | 117 |
| Dark gray | 236 |
| Light gray | 250 |

---

## True Color (RGB)

`\033[38;2;<r>;<g>;<b>m` — foreground  
`\033[48;2;<r>;<g>;<b>m` — background  

```bash
printf "\033[38;2;255;127;0mOrange (255,127,0)\033[0m\n"
printf "\033[48;2;30;30;30m\033[38;2;200;200;200mDark bg, light fg\033[0m\n"
```

> **Compatibility**: Requires `COLORTERM=truecolor` or `COLORTERM=24bit`. Check with:
> ```bash
> [[ "$COLORTERM" =~ ^(truecolor|24bit)$ ]] && echo "True color supported"
> ```

---

## Cursor Control

All sequences: `\033[<params><command>`

| Sequence | Effect |
|---|---|
| `\033[H` | Move to top-left (home) |
| `\033[<r>;<c>H` | Move to row r, column c |
| `\033[<n>A` | Move up n lines |
| `\033[<n>B` | Move down n lines |
| `\033[<n>C` | Move right n columns |
| `\033[<n>D` | Move left n columns |
| `\033[<n>E` | Move to start of next n lines |
| `\033[<n>F` | Move to start of prev n lines |
| `\033[<n>G` | Move to column n |
| `\033[s` | Save cursor position |
| `\033[u` | Restore cursor position |
| `\033[?25l` | **Hide cursor** |
| `\033[?25h` | **Show cursor** |
| `\033[?1049h` | Save screen + enter alt buffer |
| `\033[?1049l` | Restore screen + exit alt buffer |

```bash
# Common patterns in TUI
printf "\033[?25l"          # hide cursor
printf "\033[2J\033[H"      # clear screen + home
printf "\033[%dA" 3         # move up 3 lines
printf "\033[2K\r"          # erase current line
printf "\033[?25h"          # show cursor
```

---

## Screen Control

| Sequence | Effect |
|---|---|
| `\033[2J` | Clear entire screen |
| `\033[3J` | Clear screen + scrollback buffer |
| `\033[2J\033[H` | Clear + move to home (full clear) |
| `\033[0J` | Clear from cursor to end of screen |
| `\033[1J` | Clear from cursor to start of screen |
| `\033[2K` | Clear entire current line |
| `\033[0K` | Clear from cursor to end of line |
| `\033[1K` | Clear from cursor to start of line |

```bash
clear_screen() { printf "\033[2J\033[H"; }
erase_line()   { printf "\033[2K\r"; }
erase_to_end() { printf "\033[0J"; }
```

---

## Keyboard Input Sequences

When reading raw input (`read -rsn1`), arrow keys and special keys send multi-byte sequences:

| Key | Byte sequence (decimal) | Bash pattern |
|---|---|---|
| Up arrow | `27 91 65` | `$'\e[A'` → `[A` after ESC |
| Down arrow | `27 91 66` | `$'\e[B'` → `[B` |
| Right arrow | `27 91 67` | `$'\e[C'` → `[C` |
| Left arrow | `27 91 68` | `$'\e[D'` → `[D` |
| ESC (bare) | `27` | `$'\e'` + timeout |
| Enter | `10` | empty string with `-n1` |
| Space | `32` | `" "` |
| Backspace | `127` | `$'\x7f'` |
| Delete | `27 91 51 126` | `[3~` |
| Home | `27 91 72` or `27 91 49 126` | `[H` or `[1~` |
| End | `27 91 70` or `27 91 52 126` | `[F` or `[4~` |
| Page Up | `27 91 53 126` | `[5~` |
| Page Down | `27 91 54 126` | `[6~` |
| F1 | `27 79 80` | `OP` |
| F2 | `27 79 81` | `OQ` |
| F3 | `27 79 82` | `OR` |
| F4 | `27 79 83` | `OS` |

**Reliable read_key pattern** (from lib.sh):
```bash
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
        '')   printf "esc"   ;;  # bare ESC (timeout, no follow-up)
        *)    printf "esc"   ;;
      esac ;;
    "") printf "enter" ;;
    " ") printf "space" ;;
    *) printf "%s" "$key" ;;
  esac
}
```

**Debug key sequences**:
```bash
# Print raw bytes for any key press
while IFS= read -rsn1 -t 3 key; do
  printf "Key: %q  (hex: %s)\n" "$key" "$(printf '%s' "$key" | xxd -p)"
done
```

---

## tput Equivalents

`tput` queries the terminfo database — more portable than raw codes.

| Raw ANSI | tput | Effect |
|---|---|---|
| `\033[1m` | `tput bold` | Bold |
| `\033[2m` | `tput dim` | Dim |
| `\033[4m` | `tput smul` | Underline on |
| `\033[24m` | `tput rmul` | Underline off |
| `\033[7m` | `tput rev` | Reverse |
| `\033[0m` | `tput sgr0` | Reset all |
| `\033[3Xm` | `tput setaf X` | Foreground color |
| `\033[4Xm` | `tput setab X` | Background color |
| `\033[2J\033[H` | `tput clear` | Clear screen |
| `\033[?25l` | `tput civis` | Hide cursor |
| `\033[?25h` | `tput cnorm` | Show cursor |
| `\033[s` | `tput sc` | Save cursor |
| `\033[u` | `tput rc` | Restore cursor |
| — | `tput cols` | Terminal width |
| — | `tput lines` | Terminal height |

**tput color initialization** (preferred for portability):
```bash
if command -v tput >/dev/null 2>&1 && tput colors >/dev/null 2>&1; then
  BOLD="$(tput bold)"; DIM="$(tput dim)"; RESET="$(tput sgr0)"
  RED="$(tput setaf 1)"; GREEN="$(tput setaf 2)"; YELLOW="$(tput setaf 3)"
  BLUE="$(tput setaf 4)"; MAGENTA="$(tput setaf 5)"; CYAN="$(tput setaf 6)"
else
  BOLD="" DIM="" RESET="" RED="" GREEN="" YELLOW="" BLUE="" MAGENTA="" CYAN=""
fi
```

---

## Compatibility

| Feature | Support level |
|---|---|
| 8 basic colors | Universal — all terminals |
| 16 bright colors (`90`–`97`) | Very wide — most modern terminals |
| Bold, dim, underline, reverse | Wide — most terminals |
| 256 colors | Modern terminals (`xterm-256color`) |
| True color (24-bit) | Modern GUI terminals (check `$COLORTERM`) |
| Italic | Terminal-dependent — not all support |
| Cursor hide/show | Wide — xterm, gnome-terminal, iTerm2, etc |
| Alt screen buffer (`?1049h`) | Wide — xterm-derived terminals |

**Safe detection pattern:**
```bash
# Colors supported?
COLORS=$(tput colors 2>/dev/null || echo 0)
[ "$COLORS" -ge 8 ]   && HAS_COLOR=true
[ "$COLORS" -ge 256 ] && HAS_256=true
[[ "$COLORTERM" =~ ^(truecolor|24bit)$ ]] && HAS_TRUECOLOR=true

# tmux note: some sequences don't work inside tmux
# Test: printf "\033[?25l" may not hide cursor in some tmux configs
```

**Avoid inside tmux/screen:**
- `\033[?1049h/l` (alternate screen) — tmux manages this itself
- Cursor blink sequences — often ignored
- Some OSC sequences (window title, clipboard)
