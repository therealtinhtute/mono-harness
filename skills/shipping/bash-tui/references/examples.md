# Examples

### Example 1: Simple Yes/No Confirmation Dialog
**Scenario**: Add a confirmation prompt before running a destructive operation like deleting files.

**Input**:
```bash
# User wants to add confirmation before cleanup
# Existing script: rm -rf ./temp/*
```

**Output**:
```bash
#!/bin/bash
source "$(dirname "$0")/scripts/lib.sh"

clear_screen
show_banner "Cleanup Tool"

if yes_no "Delete all files in ./temp/? This cannot be undone."; then
  info "Deleting files..."
  rm -rf ./temp/*
  ok "Cleanup complete"
else
  warn "Cleanup cancelled"
fi
```

**Explanation**: The `yes_no` component returns 0 for yes and 1 for no/escape, making it perfect for if statements. This prevents accidental destructive operations while keeping the script simple.

---

### Example 2: Multi-Select Checklist for File Processing
**Scenario**: Let users select multiple log files to archive from a list.

**Input**:
```bash
# User wants to select which log files to compress
```

**Output**:
```bash
#!/bin/bash
source "$(dirname "$0")/scripts/lib.sh"

clear_screen
show_banner "Log Archiver"

# Find all log files
mapfile -t log_files < <(find ./logs -name "*.log" -type f)

if [ ${#log_files[@]} -eq 0 ]; then
  fail "No log files found"
  exit 1
fi

# Let user select files
multi_select "Select logs to archive:" "${log_files[@]}"

if [ ${#SELECTED_ITEMS[@]} -eq 0 ]; then
  warn "No files selected"
  exit 0
fi

# Archive selected files
for file in "${SELECTED_ITEMS[@]}"; do
  info "Archiving: $file"
  gzip "$file"
done

ok "Archived ${#SELECTED_ITEMS[@]} files"
```

**Explanation**: `multi_select` populates the `SELECTED_ITEMS` array with user choices. Users navigate with arrow keys, toggle with space, and confirm with enter.

---

### Example 3: Progress Bar for Batch Processing
**Scenario**: Show progress while processing image files with visual feedback.

**Input**:
```bash
# User wants to resize images with progress indicator
```

**Output**:
```bash
#!/bin/bash
source "$(dirname "$0")/scripts/lib.sh"

clear_screen
show_banner "Image Processor"

mapfile -t images < <(find ./images -name "*.jpg")
total=${#images[@]}

info "Found $total images to process"
echo ""

for i in "${!images[@]}"; do
  current=$((i + 1))
  progress_bar "$current" "$total" "Processing images"
  convert "${images[$i]}" -resize 800x600 "./output/$(basename "${images[$i]}")"
done

echo ""
ok "Processed $total images"
```

**Explanation**: `progress_bar` takes current count, total count, and a label. It updates in place, showing percentage and a visual bar.

---

### Example 4: Full Wizard Flow with Multiple Steps
**Scenario**: Create an interactive installer that guides users through configuration with multiple prompts.

**Input**:
```bash
# User wants to build a project setup wizard
```

**Output**:
```bash
#!/bin/bash
source "$(dirname "$0")/scripts/lib.sh"

clear_screen
show_banner "Project Setup Wizard"

# Step 1: Project type
STEP_TOTAL=4
step "Select project type"
single_select "Choose framework:" "React" "Vue" "Svelte" "Vanilla JS"
PROJECT_TYPE="$SELECTED_ITEM"

# Step 2: Features
step "Select features"
multi_select "Choose features to include:" \
  "TypeScript" \
  "ESLint" \
  "Prettier" \
  "Testing (Vitest)" \
  "Git hooks (Husky)"
FEATURES=("${SELECTED_ITEMS[@]}")

# Step 3: Package manager
step "Select package manager"
single_select "Choose package manager:" "npm" "yarn" "pnpm"
PKG_MANAGER="$SELECTED_ITEM"

# Step 4: Confirmation
step "Review and confirm"
echo ""
info "Project Type: $PROJECT_TYPE"
info "Features: ${FEATURES[*]}"
info "Package Manager: $PKG_MANAGER"
echo ""

if ! yes_no "Create project with these settings?"; then
  warn "Setup cancelled"
  exit 0
fi

# Execute setup
clear_screen
show_banner "Installing..."

run_spinner "Creating project structure" mkdir -p ./my-project/{src,public,tests}
run_spinner "Installing dependencies" sleep 2

ok "Project created successfully!"
info "Run: cd my-project && $PKG_MANAGER run dev"
```

**Explanation**: Combines multiple components into a cohesive wizard flow. `STEP_TOTAL` and `step` show progress through the wizard. State is captured in variables and confirmed before execution.
