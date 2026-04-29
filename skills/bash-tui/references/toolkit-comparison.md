# Toolkit Comparison

## Decision Matrix

| Toolkit | When | Pros | Cons |
|---|---|---|---|
| **Pure bash** | VPS, embedded, zero-dep | No install | More code |
| **gum** | Dev tools, local, modern look | Beautiful, fast | Needs install |
| **dialog** | Servers, Ubuntu built-in | Stable | Old-school |
| **whiptail** | Minimal servers | Lighter | Limited |

## Auto-detect Gum

```bash
USE_GUM=false
command -v gum >/dev/null 2>&1 && USE_GUM=true
```

## Gum Equivalents

See `gum-cheatsheet.md` for full Charm.sh reference.

Quick mapping:
```bash
gum confirm "Proceed?"                              # → yes_no
gum choose "A" "B" "C"                             # → single_select
gum choose --no-limit "A" "B" "C"                  # → multi_select
gum spin --spinner dot --title "Loading..." -- cmd  # → spinner
gum input --placeholder "Name"                     # → read -r -p
```
