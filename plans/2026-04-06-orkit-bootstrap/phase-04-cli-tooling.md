---
title: Phase 4 - CLI Tooling Enhancement
description: Complete CLI commands for search, info, test, and publish
status: draft
created: 2026-04-06
phase: 4
effort: medium
---

# Phase 4: CLI Tooling Enhancement

## Objectives

Complete CLI tooling with full functionality:
- Search extensions by name/tag/category
- Display extension information
- Test extensions locally
- Publish extensions to marketplace
- Update marketplace.json automatically

## Tasks

### 4.1 Search Command

**Implement search functionality in cli/orkit:**

```bash
search_extensions() {
  local query="${1:-}"
  
  if [[ -z "$query" ]]; then
    echo "Error: Search query required" >&2
    echo "Usage: orkit search <query>" >&2
    exit 1
  fi

  echo "Searching for: ${query}"
  echo

  # Search in marketplace.json
  local marketplace="${ORKIT_ROOT}/.claude-plugin/marketplace.json"
  
  if [[ ! -f "$marketplace" ]]; then
    echo "Error: marketplace.json not found" >&2
    exit 1
  fi

  # Search by name, description, tags, category
  local results
  results=$(jq -r --arg q "$query" '
    .plugins[] | 
    select(
      (.name | ascii_downcase | contains($q | ascii_downcase)) or
      (.description | ascii_downcase | contains($q | ascii_downcase)) or
      (.tags[]? | ascii_downcase | contains($q | ascii_downcase)) or
      (.category | ascii_downcase | contains($q | ascii_downcase))
    ) |
    "\(.name)|\(.version)|\(.category)|\(.description)"
  ' "$marketplace")

  if [[ -z "$results" ]]; then
    echo "No extensions found matching: ${query}"
    exit 0
  fi

  # Display results
  echo "Found extensions:"
  echo
  printf "%-25s %-15s %-15s %s\n" "NAME" "VERSION" "CATEGORY" "DESCRIPTION"
  printf "%-25s %-15s %-15s %s\n" "----" "-------" "--------" "-----------"
  
  while IFS='|' read -r name version category description; do
    printf "%-25s %-15s %-15s %s\n" "$name" "$version" "$category" "$description"
  done <<< "$results"
  
  echo
  echo "Use 'orkit info <name>' for details"
}
```

### 4.2 Info Command

**Implement info functionality:**

```bash
show_extension_info() {
  local name="${1:-}"
  
  if [[ -z "$name" ]]; then
    echo "Error: Extension name required" >&2
    echo "Usage: orkit info <name>" >&2
    exit 1
  fi

  local marketplace="${ORKIT_ROOT}/.claude-plugin/marketplace.json"
  
  if [[ ! -f "$marketplace" ]]; then
    echo "Error: marketplace.json not found" >&2
    exit 1
  fi

  # Get extension info
  local info
  info=$(jq -r --arg name "$name" '
    .plugins[] | select(.name == $name)
  ' "$marketplace")

  if [[ -z "$info" ]]; then
    echo "Extension not found: ${name}" >&2
    exit 1
  fi

  # Extract fields
  local ext_name version description category tags source
  ext_name=$(echo "$info" | jq -r '.name')
  version=$(echo "$info" | jq -r '.version')
  description=$(echo "$info" | jq -r '.description')
  category=$(echo "$info" | jq -r '.category // "uncategorized"')
  tags=$(echo "$info" | jq -r '.tags[]? // empty' | paste -sd ',' -)
  source=$(echo "$info" | jq -r '.source')

  # Display info
  cat <<EOF
Extension: ${ext_name}
Version: ${version}
Category: ${category}
Tags: ${tags:-none}

Description:
  ${description}

Source: ${source}

Installation:
  /plugin marketplace add github:tinhtute/orkit
  
Local Testing:
  orkit test plugins/${ext_name}

Documentation:
  ${ORKIT_ROOT}/plugins/${ext_name}/README.md
EOF

  # Show README preview if exists
  local readme="${ORKIT_ROOT}/plugins/${ext_name}/README.md"
  if [[ -f "$readme" ]]; then
    echo
    echo "--- README Preview ---"
    head -20 "$readme"
    echo
    echo "See full README: ${readme}"
  fi
}
```

### 4.3 Test Command

**Implement test functionality:**

```bash
test_extension() {
  local path="${1:-}"
  
  if [[ -z "$path" ]]; then
    echo "Error: Path required" >&2
    echo "Usage: orkit test <path>" >&2
    exit 1
  fi

  if [[ ! -d "$path" ]]; then
    echo "Error: Path does not exist: ${path}" >&2
    exit 1
  fi

  echo "=== Testing Extension: ${path} ==="
  echo

  # Run validation first
  echo "--- Running Validation ---"
  if ! validate_extension "$path"; then
    echo "✗ Validation failed, cannot proceed with testing" >&2
    exit 1
  fi

  echo
  echo "--- Running Extension Tests ---"

  # Detect extension type
  local ext_type
  ext_type=$(detect_extension_type "$path")

  case "$ext_type" in
    skill)
      test_skill "$path"
      ;;
    agent)
      test_agent "$path"
      ;;
    plugin)
      test_plugin "$path"
      ;;
    *)
      echo "✗ Unknown extension type" >&2
      exit 1
      ;;
  esac
}

test_skill() {
  local path="$1"
  local name
  name=$(basename "$path")

  echo "Testing skill: ${name}"

  # Check if skill can be loaded
  local skill_file="${path}/SKILL.md"
  
  # Validate frontmatter can be parsed
  if ! extract_frontmatter_field "$skill_file" "name" >/dev/null; then
    echo "✗ Cannot parse skill frontmatter" >&2
    return 1
  fi

  # Check for scripts
  if [[ -d "${path}/scripts" ]]; then
    echo "  Testing scripts..."
    for script in "${path}"/scripts/*.sh; do
      if [[ -f "$script" ]]; then
        if bash -n "$script"; then
          echo "  ✓ Script syntax valid: $(basename "$script")"
        else
          echo "  ✗ Script syntax error: $(basename "$script")" >&2
          return 1
        fi
      fi
    done
  fi

  # Test local installation
  echo "  Testing local installation..."
  local test_dir="/tmp/orkit-test-$$"
  mkdir -p "${test_dir}/skills"
  
  if cp -r "$path" "${test_dir}/skills/${name}"; then
    echo "  ✓ Skill can be installed"
    rm -rf "$test_dir"
  else
    echo "  ✗ Skill installation failed" >&2
    rm -rf "$test_dir"
    return 1
  fi

  echo "✓ Skill tests passed"
  return 0
}

test_agent() {
  local path="$1"
  local name
  name=$(basename "$path")

  echo "Testing agent: ${name}"

  # Find agent file
  local agent_file
  agent_file=$(find "$path" -maxdepth 1 -name "*.md" -type f | head -1)

  if [[ -z "$agent_file" ]]; then
    echo "✗ No agent file found" >&2
    return 1
  fi

  # Validate frontmatter
  if ! extract_frontmatter_field "$agent_file" "name" >/dev/null; then
    echo "✗ Cannot parse agent frontmatter" >&2
    return 1
  fi

  # Test local installation
  echo "  Testing local installation..."
  local test_dir="/tmp/orkit-test-$$"
  mkdir -p "${test_dir}/agents"
  
  if cp -r "$path" "${test_dir}/agents/${name}"; then
    echo "  ✓ Agent can be installed"
    rm -rf "$test_dir"
  else
    echo "  ✗ Agent installation failed" >&2
    rm -rf "$test_dir"
    return 1
  fi

  echo "✓ Agent tests passed"
  return 0
}

test_plugin() {
  local path="$1"
  
  echo "Testing plugin..."

  # Validate plugin.json
  local plugin_file="${path}/plugin.json"
  if [[ ! -f "$plugin_file" ]]; then
    echo "✗ plugin.json not found" >&2
    return 1
  fi

  # Check all referenced paths exist
  local commands agents skills hooks
  commands=$(jq -r '.commands // empty' "$plugin_file")
  agents=$(jq -r '.agents // empty' "$plugin_file")
  skills=$(jq -r '.skills // empty' "$plugin_file")
  hooks=$(jq -r '.hooks // empty' "$plugin_file")

  if [[ -n "$commands" ]] && [[ ! -d "${path}/${commands}" ]]; then
    echo "✗ Commands directory not found: ${commands}" >&2
    return 1
  fi

  if [[ -n "$agents" ]] && [[ ! -d "${path}/${agents}" ]]; then
    echo "✗ Agents directory not found: ${agents}" >&2
    return 1
  fi

  if [[ -n "$skills" ]] && [[ ! -d "${path}/${skills}" ]]; then
    echo "✗ Skills directory not found: ${skills}" >&2
    return 1
  fi

  if [[ -n "$hooks" ]] && [[ ! -f "${path}/${hooks}" ]]; then
    echo "✗ Hooks file not found: ${hooks}" >&2
    return 1
  fi

  echo "✓ Plugin tests passed"
  return 0
}
```

### 4.4 Publish Command

**Implement publish functionality:**

```bash
publish_extension() {
  local path="${1:-}"
  
  if [[ -z "$path" ]]; then
    echo "Error: Path required" >&2
    echo "Usage: orkit publish <path>" >&2
    exit 1
  fi

  if [[ ! -d "$path" ]]; then
    echo "Error: Path does not exist: ${path}" >&2
    exit 1
  fi

  echo "=== Publishing Extension: ${path} ==="
  echo

  # Run full validation and testing
  echo "--- Pre-publish Checks ---"
  if ! validate_extension "$path"; then
    echo "✗ Validation failed" >&2
    exit 1
  fi

  if ! test_extension "$path"; then
    echo "✗ Testing failed" >&2
    exit 1
  fi

  echo
  echo "--- Preparing for Publication ---"

  # Detect extension type and extract metadata
  local ext_type
  ext_type=$(detect_extension_type "$path")
  
  local name version description category tags
  name=$(basename "$path")

  case "$ext_type" in
    skill)
      local skill_file="${path}/SKILL.md"
      description=$(extract_frontmatter_field "$skill_file" "description")
      category="skills"
      ;;
    agent)
      local agent_file
      agent_file=$(find "$path" -maxdepth 1 -name "*.md" -type f | head -1)
      description=$(extract_frontmatter_field "$agent_file" "description")
      category="agents"
      ;;
    plugin)
      local plugin_file="${path}/plugin.json"
      description=$(jq -r '.description' "$plugin_file")
      category=$(jq -r '.category // "plugins"' "$plugin_file")
      ;;
  esac

  # Get version (use date-based)
  version=$(date +%Y-%m-%d)

  # Extract tags from README
  tags=$(grep -i "tags:" "${path}/README.md" 2>/dev/null | cut -d: -f2 | tr ',' '\n' | xargs | tr ' ' ',')

  # Create marketplace entry
  local entry
  entry=$(jq -n \
    --arg name "$name" \
    --arg source "./plugins/${name}" \
    --arg description "$description" \
    --arg version "$version" \
    --arg category "$category" \
    --arg tags "$tags" \
    '{
      name: $name,
      source: $source,
      description: $description,
      version: $version,
      category: $category,
      tags: ($tags | split(",") | map(select(length > 0))),
      strict: true
    }')

  echo "Extension metadata:"
  echo "$entry" | jq .
  echo

  # Update marketplace.json
  local marketplace="${ORKIT_ROOT}/.claude-plugin/marketplace.json"
  
  # Check if extension already exists
  if jq -e --arg name "$name" '.plugins[] | select(.name == $name)' "$marketplace" >/dev/null 2>&1; then
    echo "Updating existing extension: ${name}"
    # Update existing entry
    jq --argjson entry "$entry" \
      'del(.plugins[] | select(.name == $entry.name)) | .plugins += [$entry]' \
      "$marketplace" > "${marketplace}.tmp"
  else
    echo "Adding new extension: ${name}"
    # Add new entry
    jq --argjson entry "$entry" \
      '.plugins += [$entry]' \
      "$marketplace" > "${marketplace}.tmp"
  fi

  mv "${marketplace}.tmp" "$marketplace"

  echo "✓ Extension published to marketplace"
  echo
  echo "Next steps:"
  echo "  1. Review changes: git diff .claude-plugin/marketplace.json"
  echo "  2. Commit changes: git add . && git commit -m 'Add ${name} extension'"
  echo "  3. Push to GitHub: git push"
  echo "  4. Create release: gh release create ${version}"
}
```

### 4.5 List Command

**Add list command to show all extensions:**

```bash
list_extensions() {
  local marketplace="${ORKIT_ROOT}/.claude-plugin/marketplace.json"
  
  if [[ ! -f "$marketplace" ]]; then
    echo "Error: marketplace.json not found" >&2
    exit 1
  fi

  echo "=== Orkit Extensions ==="
  echo

  # Get total count
  local total
  total=$(jq '.plugins | length' "$marketplace")
  echo "Total extensions: ${total}"
  echo

  # List by category
  local categories
  categories=$(jq -r '.plugins[].category' "$marketplace" | sort -u)

  for category in $categories; do
    echo "--- ${category} ---"
    jq -r --arg cat "$category" '
      .plugins[] | 
      select(.category == $cat) |
      "  \(.name) (\(.version)) - \(.description)"
    ' "$marketplace"
    echo
  done
}
```

### 4.6 Update CLI Main

**Update cli/orkit to include new commands:**

```bash
main() {
  case "${1:-}" in
    scaffold)
      shift
      scaffold_extension "$@"
      ;;
    validate)
      shift
      validate_extension "$@"
      ;;
    test)
      shift
      test_extension "$@"
      ;;
    publish)
      shift
      publish_extension "$@"
      ;;
    search)
      shift
      search_extensions "$@"
      ;;
    info)
      shift
      show_extension_info "$@"
      ;;
    list)
      list_extensions
      ;;
    version)
      echo "orkit ${ORKIT_VERSION}"
      ;;
    help|--help|-h|"")
      usage
      ;;
    *)
      echo "Error: Unknown command '$1'" >&2
      usage
      exit 1
      ;;
  esac
}
```

### 4.7 Update Usage Documentation

**Update usage() function:**

```bash
usage() {
  cat <<EOF
Orkit CLI - Claude Code Extensions Marketplace

Usage:
  orkit <command> [options]

Commands:
  scaffold <type> <name>  Generate extension boilerplate
  validate <path>         Run validation pipeline
  test <path>             Run test suite
  publish <path>          Prepare and add to marketplace
  search <query>          Search extensions
  info <name>             Show extension details
  list                    List all extensions
  version                 Show version
  help                    Show this help

Types:
  skill                   Create a new skill
  agent                   Create a new agent
  hook                    Create a new hook
  plugin                  Create a new plugin bundle

Examples:
  # Create new skill
  orkit scaffold skill my-skill
  
  # Validate extension
  orkit validate plugins/my-skill
  
  # Test extension
  orkit test plugins/my-skill
  
  # Publish to marketplace
  orkit publish plugins/my-skill
  
  # Search extensions
  orkit search "code review"
  
  # Show extension info
  orkit info code-reviewer
  
  # List all extensions
  orkit list

For more information, visit:
  https://github.com/tinhtute/orkit

EOF
}
```

### 4.8 Add Bash Completion

**Create cli/orkit-completion.bash:**

```bash
#!/usr/bin/env bash

_orkit_completion() {
  local cur prev commands types
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  commands="scaffold validate test publish search info list version help"
  types="skill agent hook plugin"

  case "${prev}" in
    orkit)
      COMPREPLY=( $(compgen -W "${commands}" -- "${cur}") )
      return 0
      ;;
    scaffold)
      COMPREPLY=( $(compgen -W "${types}" -- "${cur}") )
      return 0
      ;;
    validate|test|publish)
      COMPREPLY=( $(compgen -d -- "${cur}") )
      return 0
      ;;
    info)
      # Complete with extension names from marketplace
      local names
      names=$(jq -r '.plugins[].name' .claude-plugin/marketplace.json 2>/dev/null)
      COMPREPLY=( $(compgen -W "${names}" -- "${cur}") )
      return 0
      ;;
  esac
}

complete -F _orkit_completion orkit
```

### 4.9 Installation Script

**Create install.sh:**

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "Installing Orkit CLI..."

# Detect shell
SHELL_NAME=$(basename "$SHELL")
RC_FILE=""

case "$SHELL_NAME" in
  bash)
    RC_FILE="$HOME/.bashrc"
    ;;
  zsh)
    RC_FILE="$HOME/.zshrc"
    ;;
  *)
    echo "Warning: Unsupported shell: ${SHELL_NAME}"
    echo "Please manually add orkit to your PATH"
    ;;
esac

# Get orkit directory
ORKIT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Make CLI executable
chmod +x "${ORKIT_DIR}/cli/orkit"

# Add to PATH if not already there
if [[ -n "$RC_FILE" ]]; then
  if ! grep -q "orkit/cli" "$RC_FILE" 2>/dev/null; then
    echo "" >> "$RC_FILE"
    echo "# Orkit CLI" >> "$RC_FILE"
    echo "export PATH=\"\$PATH:${ORKIT_DIR}/cli\"" >> "$RC_FILE"
    echo "source \"${ORKIT_DIR}/cli/orkit-completion.bash\"" >> "$RC_FILE"
    echo "✓ Added orkit to ${RC_FILE}"
    echo "  Run: source ${RC_FILE}"
  else
    echo "✓ Orkit already in ${RC_FILE}"
  fi
fi

# Install dependencies check
echo
echo "Checking dependencies..."

check_dep() {
  if command -v "$1" &>/dev/null; then
    echo "  ✓ $1"
  else
    echo "  ✗ $1 (optional: $2)"
  fi
}

check_dep "jq" "JSON processing"
check_dep "yq" "YAML processing"
check_dep "shellcheck" "Shell linting"
check_dep "markdownlint" "Markdown linting"
check_dep "gitleaks" "Secret scanning"

echo
echo "Installation complete!"
echo "Run 'orkit help' to get started"
```

## Acceptance Criteria

- [ ] Search command finds extensions by name/tag/category
- [ ] Info command displays detailed extension information
- [ ] Test command validates and tests extensions
- [ ] Publish command adds extensions to marketplace.json
- [ ] List command shows all extensions by category
- [ ] Bash completion works
- [ ] Installation script sets up CLI
- [ ] All commands have proper error handling
- [ ] Documentation updated with examples

## Dependencies

- Phase 1: Foundation
- Phase 2: Validation pipeline
- Phase 3: Initial extensions

## Estimated Effort

3-4 days

## Next Phase

Phase 5: CI/CD Pipeline
