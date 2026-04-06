---
title: Phase 2 - Validation Pipeline
description: Schema validation, security scanning, and code quality checks
status: draft
created: 2026-04-06
phase: 2
effort: medium
---

# Phase 2: Validation Pipeline

## Objectives

Build comprehensive validation pipeline:
- Schema validation for YAML/JSON
- Security scanning for secrets and malicious code
- Code quality checks for shell scripts and markdown
- Automated test harness
- Exit code standards enforcement

## Tasks

### 2.1 Schema Validation

**Implement cli/lib/validate.sh:**
```bash
#!/usr/bin/env bash

validate_extension() {
  local path="${1:-}"
  local errors=0

  if [[ -z "$path" ]]; then
    echo "Error: Path required" >&2
    return 1
  fi

  if [[ ! -d "$path" ]]; then
    echo "Error: Path does not exist: ${path}" >&2
    return 1
  fi

  echo "=== Validating Extension: ${path} ==="
  echo

  # Detect extension type
  local ext_type
  ext_type=$(detect_extension_type "$path")
  
  case "$ext_type" in
    skill)
      validate_skill "$path" || ((errors++))
      ;;
    agent)
      validate_agent "$path" || ((errors++))
      ;;
    plugin)
      validate_plugin "$path" || ((errors++))
      ;;
    *)
      echo "✗ Unknown extension type" >&2
      return 1
      ;;
  esac

  # Run security checks
  echo
  echo "--- Security Scanning ---"
  scan_security "$path" || ((errors++))

  # Run quality checks
  echo
  echo "--- Code Quality ---"
  check_quality "$path" || ((errors++))

  # Summary
  echo
  if [[ $errors -eq 0 ]]; then
    echo "✓ Validation passed"
    return 0
  else
    echo "✗ Validation failed with ${errors} error(s)" >&2
    return 1
  fi
}

detect_extension_type() {
  local path="$1"
  
  if [[ -f "${path}/SKILL.md" ]]; then
    echo "skill"
  elif [[ -f "${path}/plugin.json" ]]; then
    echo "plugin"
  elif find "$path" -maxdepth 1 -name "*.md" -type f | grep -q .; then
    echo "agent"
  else
    echo "unknown"
  fi
}

validate_skill() {
  local path="$1"
  local skill_file="${path}/SKILL.md"
  local errors=0

  echo "--- Skill Validation ---"

  # Check SKILL.md exists
  if [[ ! -f "$skill_file" ]]; then
    echo "✗ Missing SKILL.md" >&2
    return 1
  fi

  # Extract and validate frontmatter
  if ! validate_yaml_frontmatter "$skill_file" "skill"; then
    echo "✗ Invalid YAML frontmatter" >&2
    ((errors++))
  fi

  # Check required fields
  local name description
  name=$(extract_frontmatter_field "$skill_file" "name")
  description=$(extract_frontmatter_field "$skill_file" "description")

  if [[ -z "$name" ]]; then
    echo "✗ Missing required field: name" >&2
    ((errors++))
  elif [[ ! "$name" =~ ^[a-z0-9-]+$ ]]; then
    echo "✗ Invalid name format (must be kebab-case): ${name}" >&2
    ((errors++))
  fi

  if [[ -z "$description" ]]; then
    echo "✗ Missing required field: description" >&2
    ((errors++))
  elif [[ ${#description} -gt 200 ]]; then
    echo "✗ Description too long (max 200 chars): ${#description}" >&2
    ((errors++))
  fi

  # Check README exists
  if [[ ! -f "${path}/README.md" ]]; then
    echo "✗ Missing README.md" >&2
    ((errors++))
  fi

  # Check file size (instructions should be <5k tokens ~20KB)
  local size
  size=$(wc -c < "$skill_file")
  if [[ $size -gt 20480 ]]; then
    echo "⚠ SKILL.md is large (${size} bytes, recommend <20KB)"
  fi

  if [[ $errors -eq 0 ]]; then
    echo "✓ Skill validation passed"
    return 0
  else
    return 1
  fi
}

validate_agent() {
  local path="$1"
  local errors=0

  echo "--- Agent Validation ---"

  # Find agent markdown file
  local agent_file
  agent_file=$(find "$path" -maxdepth 1 -name "*.md" -type f | head -1)

  if [[ -z "$agent_file" ]]; then
    echo "✗ No agent markdown file found" >&2
    return 1
  fi

  # Validate frontmatter
  if ! validate_yaml_frontmatter "$agent_file" "agent"; then
    echo "✗ Invalid YAML frontmatter" >&2
    ((errors++))
  fi

  # Check required fields
  local name description
  name=$(extract_frontmatter_field "$agent_file" "name")
  description=$(extract_frontmatter_field "$agent_file" "description")

  if [[ -z "$name" ]]; then
    echo "✗ Missing required field: name" >&2
    ((errors++))
  fi

  if [[ -z "$description" ]]; then
    echo "✗ Missing required field: description" >&2
    ((errors++))
  fi

  # Check README
  if [[ ! -f "${path}/README.md" ]]; then
    echo "✗ Missing README.md" >&2
    ((errors++))
  fi

  if [[ $errors -eq 0 ]]; then
    echo "✓ Agent validation passed"
    return 0
  else
    return 1
  fi
}

validate_plugin() {
  local path="$1"
  local plugin_file="${path}/plugin.json"
  local errors=0

  echo "--- Plugin Validation ---"

  if [[ ! -f "$plugin_file" ]]; then
    echo "✗ Missing plugin.json" >&2
    return 1
  fi

  # Validate JSON syntax
  if ! jq empty "$plugin_file" 2>/dev/null; then
    echo "✗ Invalid JSON syntax" >&2
    return 1
  fi

  # Validate against schema
  if command -v ajv &>/dev/null; then
    if ! ajv validate -s "${ORKIT_ROOT}/schemas/plugin.schema.json" -d "$plugin_file"; then
      echo "✗ Schema validation failed" >&2
      ((errors++))
    fi
  else
    echo "⚠ ajv not installed, skipping schema validation"
  fi

  # Check required fields
  local name version description
  name=$(jq -r '.name // empty' "$plugin_file")
  version=$(jq -r '.version // empty' "$plugin_file")
  description=$(jq -r '.description // empty' "$plugin_file")

  if [[ -z "$name" ]]; then
    echo "✗ Missing required field: name" >&2
    ((errors++))
  fi

  if [[ -z "$version" ]]; then
    echo "✗ Missing required field: version" >&2
    ((errors++))
  fi

  if [[ -z "$description" ]]; then
    echo "✗ Missing required field: description" >&2
    ((errors++))
  fi

  if [[ $errors -eq 0 ]]; then
    echo "✓ Plugin validation passed"
    return 0
  else
    return 1
  fi
}

validate_yaml_frontmatter() {
  local file="$1"
  local type="$2"

  # Extract frontmatter (between --- markers)
  local frontmatter
  frontmatter=$(sed -n '/^---$/,/^---$/p' "$file" | sed '1d;$d')

  if [[ -z "$frontmatter" ]]; then
    echo "✗ No YAML frontmatter found" >&2
    return 1
  fi

  # Validate YAML syntax with yq
  if command -v yq &>/dev/null; then
    if ! echo "$frontmatter" | yq eval '.' - >/dev/null 2>&1; then
      echo "✗ Invalid YAML syntax" >&2
      return 1
    fi
  else
    echo "⚠ yq not installed, skipping YAML validation"
  fi

  return 0
}

extract_frontmatter_field() {
  local file="$1"
  local field="$2"

  local frontmatter
  frontmatter=$(sed -n '/^---$/,/^---$/p' "$file" | sed '1d;$d')

  if command -v yq &>/dev/null; then
    echo "$frontmatter" | yq eval ".${field}" - 2>/dev/null
  else
    # Fallback: simple grep
    echo "$frontmatter" | grep "^${field}:" | cut -d: -f2- | xargs
  fi
}
```

### 2.2 Security Scanning

**Implement cli/lib/security.sh:**
```bash
#!/usr/bin/env bash

scan_security() {
  local path="$1"
  local errors=0

  echo "Running security scans..."

  # Check for secrets
  if ! scan_secrets "$path"; then
    echo "✗ Secret detection failed" >&2
    ((errors++))
  fi

  # Check for malicious patterns
  if ! scan_malicious_patterns "$path"; then
    echo "✗ Malicious pattern detection failed" >&2
    ((errors++))
  fi

  # Check for dangerous commands
  if ! scan_dangerous_commands "$path"; then
    echo "✗ Dangerous command detection failed" >&2
    ((errors++))
  fi

  if [[ $errors -eq 0 ]]; then
    echo "✓ Security scans passed"
    return 0
  else
    return 1
  fi
}

scan_secrets() {
  local path="$1"

  # Use gitleaks if available
  if command -v gitleaks &>/dev/null; then
    if gitleaks detect --source "$path" --no-git --quiet; then
      echo "  ✓ No secrets detected (gitleaks)"
      return 0
    else
      echo "  ✗ Secrets detected" >&2
      return 1
    fi
  fi

  # Fallback: pattern matching
  local patterns=(
    'AKIA[0-9A-Z]{16}'                    # AWS Access Key
    'AIza[0-9A-Za-z\\-_]{35}'             # Google API Key
    'sk-[a-zA-Z0-9]{48}'                  # OpenAI API Key
    'ghp_[a-zA-Z0-9]{36}'                 # GitHub Personal Access Token
    'xox[baprs]-[0-9a-zA-Z]{10,48}'       # Slack Token
    'password\s*=\s*["\'][^"\']{8,}'      # Password in config
    'api[_-]?key\s*=\s*["\'][^"\']{16,}'  # API key in config
  )

  local found=0
  for pattern in "${patterns[@]}"; do
    if grep -rE "$pattern" "$path" 2>/dev/null | grep -v "\.git" | grep -q .; then
      echo "  ✗ Potential secret found matching pattern: ${pattern}" >&2
      found=1
    fi
  done

  if [[ $found -eq 0 ]]; then
    echo "  ✓ No secrets detected (pattern matching)"
    return 0
  else
    return 1
  fi
}

scan_malicious_patterns() {
  local path="$1"

  # Dangerous patterns to detect
  local patterns=(
    'eval\s*\$'                           # Eval with variable
    'curl.*\|\s*bash'                     # Curl pipe to bash
    'wget.*\|\s*sh'                       # Wget pipe to shell
    'rm\s+-rf\s+/'                        # Recursive delete from root
    'chmod\s+777'                         # Overly permissive chmod
    '>\s*/dev/sd[a-z]'                    # Writing to disk device
    'dd\s+if=.*of=/dev'                   # DD to device
    'mkfs\.'                              # Format filesystem
    ':(){:|:&};:'                         # Fork bomb
  )

  local found=0
  for pattern in "${patterns[@]}"; do
    if grep -rE "$pattern" "$path" 2>/dev/null | grep -v "\.git" | grep -q .; then
      echo "  ✗ Dangerous pattern found: ${pattern}" >&2
      found=1
    fi
  done

  if [[ $found -eq 0 ]]; then
    echo "  ✓ No malicious patterns detected"
    return 0
  else
    return 1
  fi
}

scan_dangerous_commands() {
  local path="$1"

  # Find all shell scripts
  local scripts
  scripts=$(find "$path" -type f \( -name "*.sh" -o -name "*.bash" \) 2>/dev/null)

  if [[ -z "$scripts" ]]; then
    echo "  ✓ No shell scripts to scan"
    return 0
  fi

  # Check for dangerous commands without safeguards
  local dangerous=(
    'rm -rf'
    'rm -fr'
    'sudo rm'
    'sudo dd'
    'sudo mkfs'
  )

  local found=0
  for script in $scripts; do
    for cmd in "${dangerous[@]}"; do
      if grep -q "$cmd" "$script"; then
        # Check if there's a safety check (set -e, error handling, etc.)
        if ! grep -q "set -e" "$script" && ! grep -q "set -u" "$script"; then
          echo "  ⚠ Dangerous command without error handling in: ${script}" >&2
          echo "    Command: ${cmd}" >&2
          found=1
        fi
      fi
    done
  done

  if [[ $found -eq 0 ]]; then
    echo "  ✓ No dangerous commands detected"
    return 0
  else
    echo "  ⚠ Review dangerous commands (warnings only)"
    return 0  # Don't fail, just warn
  fi
}
```

### 2.3 Code Quality Checks

**Implement cli/lib/quality.sh:**
```bash
#!/usr/bin/env bash

check_quality() {
  local path="$1"
  local errors=0

  echo "Running quality checks..."

  # Shell script linting
  if ! check_shellcheck "$path"; then
    echo "✗ Shell script linting failed" >&2
    ((errors++))
  fi

  # Markdown linting
  if ! check_markdown "$path"; then
    echo "✗ Markdown linting failed" >&2
    ((errors++))
  fi

  # File naming conventions
  if ! check_naming_conventions "$path"; then
    echo "✗ Naming convention check failed" >&2
    ((errors++))
  fi

  if [[ $errors -eq 0 ]]; then
    echo "✓ Quality checks passed"
    return 0
  else
    return 1
  fi
}

check_shellcheck() {
  local path="$1"

  if ! command -v shellcheck &>/dev/null; then
    echo "  ⚠ shellcheck not installed, skipping"
    return 0
  fi

  # Find all shell scripts
  local scripts
  scripts=$(find "$path" -type f \( -name "*.sh" -o -name "*.bash" \) 2>/dev/null)

  if [[ -z "$scripts" ]]; then
    echo "  ✓ No shell scripts to check"
    return 0
  fi

  local failed=0
  for script in $scripts; do
    if ! shellcheck -x "$script" 2>&1; then
      echo "  ✗ shellcheck failed: ${script}" >&2
      failed=1
    fi
  done

  if [[ $failed -eq 0 ]]; then
    echo "  ✓ Shell scripts passed shellcheck"
    return 0
  else
    return 1
  fi
}

check_markdown() {
  local path="$1"

  if ! command -v markdownlint &>/dev/null; then
    echo "  ⚠ markdownlint not installed, skipping"
    return 0
  fi

  # Find all markdown files
  local mdfiles
  mdfiles=$(find "$path" -type f -name "*.md" 2>/dev/null)

  if [[ -z "$mdfiles" ]]; then
    echo "  ✓ No markdown files to check"
    return 0
  fi

  local failed=0
  for mdfile in $mdfiles; do
    if ! markdownlint "$mdfile" 2>&1; then
      echo "  ✗ markdownlint failed: ${mdfile}" >&2
      failed=1
    fi
  done

  if [[ $failed -eq 0 ]]; then
    echo "  ✓ Markdown files passed linting"
    return 0
  else
    return 1
  fi
}

check_naming_conventions() {
  local path="$1"
  local errors=0

  # Check directory name is kebab-case
  local dirname
  dirname=$(basename "$path")
  if [[ ! "$dirname" =~ ^[a-z0-9-]+$ ]]; then
    echo "  ✗ Directory name not kebab-case: ${dirname}" >&2
    ((errors++))
  fi

  # Check all subdirectories are kebab-case
  while IFS= read -r dir; do
    local name
    name=$(basename "$dir")
    if [[ ! "$name" =~ ^[a-z0-9-]+$ ]]; then
      echo "  ✗ Subdirectory not kebab-case: ${dir}" >&2
      ((errors++))
    fi
  done < <(find "$path" -type d -not -path "*/.*" 2>/dev/null)

  # Check file names (allow .md, .sh, .json, .yaml extensions)
  while IFS= read -r file; do
    local name
    name=$(basename "$file")
    local base="${name%.*}"
    if [[ ! "$base" =~ ^[a-z0-9-]+$ ]] && [[ ! "$name" =~ ^[A-Z]+\. ]]; then
      echo "  ⚠ File name not kebab-case: ${file}"
    fi
  done < <(find "$path" -type f -not -path "*/.*" 2>/dev/null)

  if [[ $errors -eq 0 ]]; then
    echo "  ✓ Naming conventions followed"
    return 0
  else
    return 1
  fi
}
```

### 2.4 Test Harness

**Create tests/validation/test-skill-validation.sh:**
```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ORKIT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

source "${ORKIT_ROOT}/cli/lib/validate.sh"

test_valid_skill() {
  echo "Test: Valid skill passes validation"
  
  local test_dir="${ORKIT_ROOT}/tests/fixtures/valid-skill"
  mkdir -p "$test_dir"
  
  cat > "${test_dir}/SKILL.md" <<'EOF'
---
name: test-skill
description: A test skill for validation
user-invocable: true
---

# Test Skill

This is a test skill.
EOF

  cat > "${test_dir}/README.md" <<'EOF'
# Test Skill

Test skill for validation.
EOF

  if validate_skill "$test_dir"; then
    echo "✓ Test passed"
    rm -rf "$test_dir"
    return 0
  else
    echo "✗ Test failed"
    rm -rf "$test_dir"
    return 1
  fi
}

test_invalid_skill_missing_name() {
  echo "Test: Skill without name fails validation"
  
  local test_dir="${ORKIT_ROOT}/tests/fixtures/invalid-skill"
  mkdir -p "$test_dir"
  
  cat > "${test_dir}/SKILL.md" <<'EOF'
---
description: A test skill without name
---

# Test Skill
EOF

  cat > "${test_dir}/README.md" <<'EOF'
# Test Skill
EOF

  if validate_skill "$test_dir" 2>/dev/null; then
    echo "✗ Test failed (should have failed validation)"
    rm -rf "$test_dir"
    return 1
  else
    echo "✓ Test passed (correctly failed validation)"
    rm -rf "$test_dir"
    return 0
  fi
}

# Run tests
echo "=== Running Validation Tests ==="
echo

test_valid_skill
test_invalid_skill_missing_name

echo
echo "=== Tests Complete ==="
```

**Create tests/security/test-secret-detection.sh:**
```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ORKIT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

source "${ORKIT_ROOT}/cli/lib/security.sh"

test_detect_aws_key() {
  echo "Test: Detect AWS access key"
  
  local test_dir="${ORKIT_ROOT}/tests/fixtures/secret-test"
  mkdir -p "$test_dir"
  
  echo "AWS_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE" > "${test_dir}/config.sh"

  if scan_secrets "$test_dir" 2>/dev/null; then
    echo "✗ Test failed (should have detected secret)"
    rm -rf "$test_dir"
    return 1
  else
    echo "✓ Test passed (correctly detected secret)"
    rm -rf "$test_dir"
    return 0
  fi
}

test_no_secrets() {
  echo "Test: No secrets in clean code"
  
  local test_dir="${ORKIT_ROOT}/tests/fixtures/clean-test"
  mkdir -p "$test_dir"
  
  echo "echo 'Hello World'" > "${test_dir}/script.sh"

  if scan_secrets "$test_dir" 2>/dev/null; then
    echo "✓ Test passed (no secrets detected)"
    rm -rf "$test_dir"
    return 0
  else
    echo "✗ Test failed (false positive)"
    rm -rf "$test_dir"
    return 1
  fi
}

# Run tests
echo "=== Running Security Tests ==="
echo

test_detect_aws_key
test_no_secrets

echo
echo "=== Tests Complete ==="
```

### 2.5 Documentation

**Create docs/validation-rules.md:**
```markdown
# Validation Rules

Orkit enforces strict validation rules to ensure quality and security.

## Schema Validation

### Skills (SKILL.md)

Required fields:
- `name`: kebab-case, unique
- `description`: max 200 characters

Optional fields:
- `allowed-tools`: array of tool names
- `model`: sonnet, opus, or haiku
- `user-invocable`: boolean
- `effort`: low, medium, or high

### Agents (agent-name.md)

Required fields:
- `name`: kebab-case, unique
- `description`: max 200 characters

Optional fields:
- `tools`: array of allowed tools
- `disallowedTools`: array of denied tools
- `model`: sonnet, opus, or haiku
- `maxTurns`: integer
- `permissionMode`: acceptEdits, acceptAll, or prompt

### Plugins (plugin.json)

Required fields:
- `name`: kebab-case, unique
- `version`: date-based (YYYY-MM-DD) or semver
- `description`: max 200 characters

## Security Rules

### Prohibited Patterns

- AWS/API keys and tokens
- Passwords in plaintext
- Eval with untrusted input
- Curl/wget piped to shell
- Destructive commands without safeguards
- Fork bombs and malicious code

### Required Safeguards

- Shell scripts must use `set -euo pipefail`
- Dangerous commands must have error handling
- No writes to system directories
- No privilege escalation without justification

## Code Quality Rules

### Shell Scripts

- Must pass shellcheck with no errors
- Use proper error handling
- Include usage documentation
- Follow bash best practices

### Markdown

- Must pass markdownlint
- Include proper headings
- No broken links
- Consistent formatting

### Naming Conventions

- Directories: kebab-case
- Files: kebab-case with appropriate extensions
- Skills/Agents: kebab-case names
- No spaces or special characters

## Size Limits

- Skill metadata: ~100 tokens
- Skill instructions: <5k tokens (~20KB)
- Agent instructions: <5k tokens
- README files: <10KB

## Required Files

### Skills
- `SKILL.md`: Main skill definition
- `README.md`: User documentation

### Agents
- `{agent-name}.md`: Agent definition
- `README.md`: User documentation

### Plugins
- `plugin.json`: Plugin manifest
- `README.md`: User documentation
- `LICENSE`: License file

## Validation Process

1. Schema validation (structure and types)
2. Security scanning (secrets and malicious code)
3. Code quality checks (linting and conventions)
4. Manual review (for invite-only contributors)

## Exit Codes

- `0`: Validation passed
- `1`: Validation failed
- `2`: Critical security issue (blocks merge)
```

## Acceptance Criteria

- [ ] Schema validation implemented for all extension types
- [ ] Security scanning detects secrets and malicious patterns
- [ ] Code quality checks for shell scripts and markdown
- [ ] Test harness with passing tests
- [ ] Validation rules documented
- [ ] CLI tool validates extensions end-to-end

## Dependencies

- jq (JSON processing)
- yq (YAML processing)
- shellcheck (shell linting)
- markdownlint (markdown linting)
- gitleaks (secret scanning, optional)

## Estimated Effort

4-5 days

## Next Phase

Phase 3: Initial Extensions
