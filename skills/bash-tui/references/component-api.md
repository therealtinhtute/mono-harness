# Component API Reference

Source `scripts/lib.sh` at the top of any script — all components ready to use:

```bash
source "$(dirname "$0")/scripts/lib.sh"
# or when distributing as single file, copy lib.sh content directly
```

## Colors & Screen
```bash
# After sourcing lib.sh — colors auto-initialized
clear_screen; hide_cursor; show_cursor; sep_line
```

## read_key → returns: up | down | left | right | enter | space | esc | char
```bash
case "$(read_key)" in
  up) ... ;; down) ... ;; enter) ... ;; esc) ... ;;
esac
```

## yes_no
```bash
# Returns 0=yes, 1=no/esc
yes_no "Proceed?" && do_yes || do_no
```

## multi_select → SELECTED_ITEMS[]
```bash
multi_select "Title" "Item A" "Item B" "Item C"
for item in "${SELECTED_ITEMS[@]}"; do echo "$item"; done
```

## single_select → SELECTED_ITEM
```bash
single_select "Pick one:" "Opt A" "Opt B" "Opt C"
echo "$SELECTED_ITEM"
```

## spinner
```bash
spinner "Installing..." apt-get install -y curl
```

## progress_bar
```bash
for i in $(seq 1 100); do
  progress_bar "$i" 100 "Downloading"
  sleep 0.05
done
```

## Notification helpers
```bash
ok "Done"    # ✔ green
warn "Note"  # ⚠ yellow
fail "Error" # ✖ red
info "→ msg" # → cyan
step "Phase" # ◆ [1/3] magenta (auto-increments STEP_CURRENT)
```

## run_cmd / run_spinner
```bash
run_cmd "Update packages" apt-get update
run_spinner "Building..." make all
```
