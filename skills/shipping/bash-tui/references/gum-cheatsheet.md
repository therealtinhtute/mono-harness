# Gum Cheatsheet — Charm.sh

Install: https://github.com/charmbracelet/gum

```bash
# macOS
brew install gum

# Go
go install github.com/charmbracelet/gum@latest

# Ubuntu/Debian
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://repo.charm.sh/apt/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/charm.gpg
echo "deb [signed-by=/etc/apt/keyrings/charm.gpg] https://repo.charm.sh/apt/ * *" | sudo tee /etc/apt/sources.list.d/charm.list
sudo apt update && sudo apt install gum
```

## Components

### confirm — Yes/No
```bash
gum confirm "Are you sure?"
gum confirm --default=false "Dangerous operation?"
```

### choose — Single/Multi select
```bash
# Single
CHOICE=$(gum choose "Option A" "Option B" "Option C")

# Multi (no limit)
readarray -t CHOICES < <(gum choose --no-limit "Opt A" "Opt B" "Opt C")

# From array
readarray -t PICKS < <(printf '%s\n' "${ITEMS[@]}" | gum choose --no-limit)
```

### input — Text input
```bash
NAME=$(gum input --placeholder "Enter name")
PASS=$(gum input --password --placeholder "Password")
```

### write — Multi-line text
```bash
NOTES=$(gum write --placeholder "Your notes here...")
```

### spin — Spinner
```bash
gum spin --spinner dot --title "Loading..." -- sleep 3
gum spin --spinner line --title "Processing..." -- your-command
# Spinners: dot, line, minidot, jump, pulse, points, globe, moon, monkey, hamburger
```

### style — Styled text
```bash
gum style --foreground 212 "Hello"
gum style --border double --padding "1 2" --border-foreground 212 "Box"
gum style --bold --italic --underline "Formatted"
```

### filter — Fuzzy search
```bash
ITEM=$(echo -e "apple\nbanana\ncherry" | gum filter)
FILE=$(find . -name "*.sh" | gum filter --placeholder "Search files...")
```

### table — Display table
```bash
echo "Name,Age,Role\nAlice,30,Dev\nBob,25,QA" | gum table
```

### log — Structured logging
```bash
gum log --level info "Starting process"
gum log --level warn "Something odd"
gum log --level error "Failed!"
```

## Styling with ENV vars
```bash
export GUM_CHOOSE_CURSOR_FOREGROUND="212"
export GUM_CONFIRM_PROMPT_FOREGROUND="99"
export GUM_INPUT_PROMPT_FOREGROUND="212"
```

## Color palette (256-color codes)
- 212 = Pink/Magenta
- 99  = Purple
- 86  = Teal
- 226 = Yellow
- 196 = Red
- 46  = Green
