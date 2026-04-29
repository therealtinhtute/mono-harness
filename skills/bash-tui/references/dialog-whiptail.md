# dialog / whiptail Reference

Use these when the target machine has no internet access to install `gum` and you want
a more robust UI than pure bash. Both are pre-installed on most Ubuntu/Debian servers.

## dialog vs whiptail

| | dialog | whiptail |
|---|---|---|
| Package | `dialog` | `whiptail` (via `newt`) |
| Pre-installed | Most servers | Ubuntu minimal, Debian |
| Look | ncurses (full color) | Newt (simpler) |
| Use when | Richer UI needed | Ultra-minimal server |

Check availability:
```bash
if command -v dialog >/dev/null 2>&1; then
  TUI=dialog
elif command -v whiptail >/dev/null 2>&1; then
  TUI=whiptail
else
  TUI=pure_bash  # fall back to lib.sh components
fi
```

---

## dialog Patterns

### Yes/No
```bash
dialog --title "Confirm" --yesno "Proceed with cleanup?" 7 50
if [ $? -eq 0 ]; then
  echo "User said yes"
fi
```

### Message box
```bash
dialog --title "Done" --msgbox "Cleanup complete!\nFreed 2.3 GB" 8 50
```

### Info box (no button, auto-dismiss)
```bash
dialog --infobox "Loading, please wait..." 4 40
sleep 2
```

### Input box
```bash
VALUE=$(dialog --title "Input" --inputbox "Enter hostname:" 8 50 "localhost" 2>&1 >/dev/tty)
echo "Entered: $VALUE"
```

### Password input
```bash
PASS=$(dialog --title "Auth" --passwordbox "Enter password:" 8 50 2>&1 >/dev/tty)
```

### Menu (single select)
```bash
CHOICE=$(dialog --title "Select Action" \
  --menu "Choose:" 15 50 5 \
  "1" "APT cleanup" \
  "2" "Log cleanup" \
  "3" "Docker prune" \
  2>&1 >/dev/tty)
echo "Chosen: $CHOICE"
```

### Checklist (multi-select)
```bash
RESULT=$(dialog --title "Select Tasks" \
  --checklist "Space=toggle, Enter=confirm:" 15 60 5 \
  "apt"     "APT system cleanup"     ON  \
  "logs"    "System logs"             ON  \
  "trash"   "Trash & temp files"      ON  \
  "browser" "Browser caches"          OFF \
  "docker"  "Docker cleanup"          OFF \
  2>&1 >/dev/tty)
# RESULT = space-separated quoted items: "apt" "logs" "trash"
```

### Progress gauge
```bash
(
  echo 10; sleep 0.5
  echo 40; sleep 0.5
  echo 70; sleep 0.5
  echo 100
) | dialog --gauge "Installing packages..." 6 50 0
```

### File selector
```bash
FILE=$(dialog --title "Select file" --fselect "$HOME/" 14 60 2>&1 >/dev/tty)
```

---

## whiptail Patterns

whiptail uses the same flags as dialog for most common widgets:

```bash
# Yes/No
whiptail --title "Confirm" --yesno "Proceed?" 8 50

# Menu
CHOICE=$(whiptail --title "Menu" \
  --menu "Choose:" 15 50 5 \
  "1" "Option A" \
  "2" "Option B" \
  3>&1 1>&2 2>&3)

# Checklist
RESULT=$(whiptail --title "Select" \
  --checklist "Choose:" 15 60 5 \
  "a" "Item A" ON \
  "b" "Item B" OFF \
  3>&1 1>&2 2>&3)

# Inputbox
VALUE=$(whiptail --inputbox "Enter value:" 8 50 "default" \
  3>&1 1>&2 2>&3)

# Progress
{
  for i in $(seq 10 10 100); do
    echo "$i"
    sleep 0.3
  done
} | whiptail --gauge "Working..." 6 50 0
```

**Note**: whiptail uses `3>&1 1>&2 2>&3` fd-swap trick instead of `2>&1 >/dev/tty`.

---

## Combining with lib.sh

Use dialog/whiptail for the main interactive flow, lib.sh for progress + notifications:

```bash
source "$(dirname "$0")/scripts/lib.sh"

# Use dialog for selection
ITEMS=$(dialog --checklist "Select:" 15 60 5 \
  "apt" "APT cleanup" ON \
  "logs" "Logs" ON \
  2>&1 >/dev/tty)

# Use lib.sh for execution feedback
clear
STEP_TOTAL=2; STEP_CURRENT=0
for item in $ITEMS; do
  case "$item" in
    '"apt"')
      step "APT cleanup"
      run_spinner "Updating packages..." apt-get update ;;
    '"logs"')
      step "Log cleanup"
      run_cmd "Vacuuming journals..." journalctl --vacuum-time=7d ;;
  esac
done
ok "All done!"
```
