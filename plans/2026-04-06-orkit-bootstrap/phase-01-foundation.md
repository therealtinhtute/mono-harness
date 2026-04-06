---
title: Phase 1 - Foundation
description: Repository structure, marketplace catalog, schemas, and CLI scaffolding
status: draft
created: 2026-04-06
phase: 1
effort: medium
---

# Phase 1: Foundation

## Objectives

Establish core infrastructure for orkit marketplace:
- Repository structure with proper organization
- Marketplace.json catalog format
- JSON schemas for validation
- Basic CLI scaffolding tool
- Developer guide foundation

## Tasks

### 1.1 Repository Setup

**Create directory structure:**
```bash
mkdir -p orkit/{.claude-plugin,plugins,cli/{lib,templates},schemas,docs,.github/workflows,tests/{validation,security,integration}}
```

**Initialize git:**
```bash
cd orkit
git init
git branch -M main
```

**Create .gitignore:**
```
# Dependencies
node_modules/
*.log

# Build artifacts
dist/
build/
*.tmp

# IDE
.vscode/
.idea/
*.swp

# OS
.DS_Store
Thumbs.db

# Test coverage
coverage/
.nyc_output/

# Secrets
.env
*.key
*.pem
```

**Create README.md:**
- Project overview
- Installation instructions
- Quick start guide
- Link to full documentation

### 1.2 Marketplace Catalog

**Create .claude-plugin/marketplace.json:**
```json
{
  "name": "orkit",
  "owner": {
    "name": "Orkit Team",
    "email": "team@orkit.dev"
  },
  "metadata": {
    "description": "Curated marketplace for Claude Code extensions - production-grade skills, agents, hooks, and statusline configs",
    "version": "2026-04-06",
    "homepage": "https://github.com/tinhtute/orkit",
    "repository": "https://github.com/tinhtute/orkit",
    "license": "MIT",
    "pluginRoot": "./plugins",
    "keywords": ["claude-code", "extensions", "marketplace", "plugins"]
  },
  "plugins": []
}
```

**Create .claude-plugin/plugin.json (orkit as plugin):**
```json
{
  "name": "orkit",
  "version": "2026-04-06",
  "description": "Orkit marketplace CLI and utilities",
  "author": {
    "name": "Orkit Team",
    "email": "team@orkit.dev"
  },
  "homepage": "https://github.com/tinhtute/orkit",
  "repository": "https://github.com/tinhtute/orkit",
  "license": "MIT",
  "keywords": ["marketplace", "cli", "validation"],
  "commands": "./cli/",
  "skills": "./plugins/"
}
```

### 1.3 JSON Schemas

**Create schemas/skill.schema.json:**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Claude Code Skill",
  "type": "object",
  "required": ["name", "description"],
  "properties": {
    "name": {
      "type": "string",
      "pattern": "^[a-z0-9-]+$",
      "description": "Skill name in kebab-case"
    },
    "description": {
      "type": "string",
      "maxLength": 200,
      "description": "Brief description for discovery"
    },
    "allowed-tools": {
      "type": "array",
      "items": {"type": "string"}
    },
    "model": {
      "type": "string",
      "enum": ["sonnet", "opus", "haiku"]
    },
    "context": {
      "type": "string",
      "enum": ["fork", "inherit"]
    },
    "agent": {
      "type": "string"
    },
    "disable-model-invocation": {
      "type": "boolean"
    },
    "user-invocable": {
      "type": "boolean"
    },
    "argument-hint": {
      "type": "array",
      "items": {"type": "string"}
    },
    "effort": {
      "type": "string",
      "enum": ["low", "medium", "high"]
    },
    "paths": {
      "type": "string"
    },
    "shell": {
      "type": "string"
    }
  }
}
```

**Create schemas/agent.schema.json:**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Claude Code Agent",
  "type": "object",
  "required": ["name", "description"],
  "properties": {
    "name": {
      "type": "string",
      "pattern": "^[a-z0-9-]+$"
    },
    "description": {
      "type": "string",
      "maxLength": 200
    },
    "tools": {
      "type": "array",
      "items": {"type": "string"}
    },
    "disallowedTools": {
      "type": "array",
      "items": {"type": "string"}
    },
    "model": {
      "type": "string",
      "enum": ["sonnet", "opus", "haiku"]
    },
    "permissionMode": {
      "type": "string",
      "enum": ["acceptEdits", "acceptAll", "prompt"]
    },
    "maxTurns": {
      "type": "integer",
      "minimum": 1
    },
    "skills": {
      "type": "array",
      "items": {"type": "string"}
    },
    "mcpServers": {
      "type": "array",
      "items": {"type": "string"}
    },
    "memory": {
      "type": "string",
      "enum": ["user", "project", "local"]
    },
    "background": {
      "type": "boolean"
    },
    "effort": {
      "type": "string",
      "enum": ["low", "medium", "high"]
    },
    "isolation": {
      "type": "string",
      "enum": ["worktree", "none"]
    },
    "color": {
      "type": "string"
    },
    "initialPrompt": {
      "type": "string"
    }
  }
}
```

**Create schemas/hook.schema.json:**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Claude Code Hook",
  "type": "object",
  "required": ["type"],
  "properties": {
    "type": {
      "type": "string",
      "enum": ["command", "http", "prompt", "agent"]
    },
    "command": {
      "type": "string"
    },
    "timeout": {
      "type": "integer",
      "minimum": 1
    },
    "async": {
      "type": "boolean"
    },
    "url": {
      "type": "string",
      "format": "uri"
    },
    "method": {
      "type": "string",
      "enum": ["GET", "POST", "PUT", "DELETE"]
    },
    "headers": {
      "type": "object"
    },
    "prompt": {
      "type": "string"
    },
    "agent": {
      "type": "string"
    }
  }
}
```

**Create schemas/plugin.schema.json:**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Claude Code Plugin",
  "type": "object",
  "required": ["name", "version", "description"],
  "properties": {
    "name": {
      "type": "string",
      "pattern": "^[a-z0-9-]+$"
    },
    "version": {
      "type": "string",
      "pattern": "^\\d{4}-\\d{2}-\\d{2}$|^\\d+\\.\\d+\\.\\d+$"
    },
    "description": {
      "type": "string",
      "maxLength": 200
    },
    "author": {
      "type": "object",
      "required": ["name"],
      "properties": {
        "name": {"type": "string"},
        "email": {"type": "string", "format": "email"}
      }
    },
    "homepage": {
      "type": "string",
      "format": "uri"
    },
    "repository": {
      "type": "string"
    },
    "license": {
      "type": "string"
    },
    "keywords": {
      "type": "array",
      "items": {"type": "string"}
    },
    "commands": {
      "type": "string"
    },
    "agents": {
      "type": "string"
    },
    "skills": {
      "type": "string"
    },
    "outputStyles": {
      "type": "string"
    },
    "hooks": {
      "type": "string"
    },
    "mcpServers": {
      "type": "string"
    },
    "lspServers": {
      "type": "string"
    }
  }
}
```

### 1.4 CLI Scaffolding Tool

**Create cli/orkit (main entry point):**
```bash
#!/usr/bin/env bash
set -euo pipefail

ORKIT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ORKIT_VERSION="2026-04-06"

# Source libraries
source "${ORKIT_ROOT}/cli/lib/scaffold.sh"
source "${ORKIT_ROOT}/cli/lib/validate.sh"
source "${ORKIT_ROOT}/cli/lib/security.sh"
source "${ORKIT_ROOT}/cli/lib/quality.sh"

# Usage information
usage() {
  cat <<EOF
Orkit CLI - Claude Code Extensions Marketplace

Usage:
  orkit <command> [options]

Commands:
  scaffold <type> <name>  Generate extension boilerplate
  validate <path>         Run validation pipeline
  test <path>             Run test suite
  publish <path>          Prepare for marketplace
  search <query>          Search extensions
  info <name>             Show extension details
  version                 Show version

Types:
  skill                   Create a new skill
  agent                   Create a new agent
  hook                    Create a new hook
  plugin                  Create a new plugin bundle

Examples:
  orkit scaffold skill my-skill
  orkit validate plugins/my-skill
  orkit publish plugins/my-skill

EOF
}

# Main command router
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

main "$@"
```

**Create cli/lib/scaffold.sh:**
```bash
#!/usr/bin/env bash

scaffold_extension() {
  local type="${1:-}"
  local name="${2:-}"

  if [[ -z "$type" ]] || [[ -z "$name" ]]; then
    echo "Error: Type and name required" >&2
    echo "Usage: orkit scaffold <type> <name>" >&2
    exit 1
  fi

  # Validate name format (kebab-case)
  if [[ ! "$name" =~ ^[a-z0-9-]+$ ]]; then
    echo "Error: Name must be kebab-case (lowercase, numbers, hyphens only)" >&2
    exit 1
  fi

  case "$type" in
    skill)
      scaffold_skill "$name"
      ;;
    agent)
      scaffold_agent "$name"
      ;;
    hook)
      scaffold_hook "$name"
      ;;
    plugin)
      scaffold_plugin "$name"
      ;;
    *)
      echo "Error: Unknown type '$type'" >&2
      echo "Valid types: skill, agent, hook, plugin" >&2
      exit 1
      ;;
  esac
}

scaffold_skill() {
  local name="$1"
  local skill_dir="${ORKIT_ROOT}/plugins/${name}"

  if [[ -d "$skill_dir" ]]; then
    echo "Error: Skill '${name}' already exists" >&2
    exit 1
  fi

  echo "Creating skill: ${name}"
  mkdir -p "${skill_dir}"/{references,scripts}

  cat > "${skill_dir}/SKILL.md" <<EOF
---
name: ${name}
description: Brief description of what this skill does
user-invocable: true
model: sonnet
effort: medium
---

# ${name}

Detailed instructions for this skill.

## Purpose

Explain what this skill does and when to use it.

## Usage

Provide usage examples and guidelines.

## Implementation

Describe how the skill works internally.
EOF

  cat > "${skill_dir}/README.md" <<EOF
# ${name}

Brief description of the skill.

## Installation

\`\`\`bash
# Via orkit marketplace
/plugin marketplace add github:tinhtute/orkit
\`\`\`

## Usage

\`\`\`bash
/${name}
\`\`\`

## Configuration

Describe any configuration options.

## Examples

Provide usage examples.

## License

MIT
EOF

  echo "✓ Skill scaffolded at: ${skill_dir}"
  echo "  Edit ${skill_dir}/SKILL.md to customize"
}

scaffold_agent() {
  local name="$1"
  local agent_dir="${ORKIT_ROOT}/plugins/${name}"

  if [[ -d "$agent_dir" ]]; then
    echo "Error: Agent '${name}' already exists" >&2
    exit 1
  fi

  echo "Creating agent: ${name}"
  mkdir -p "${agent_dir}"

  cat > "${agent_dir}/${name}.md" <<EOF
---
name: ${name}
description: Brief description of what this agent does
model: sonnet
permissionMode: acceptEdits
maxTurns: 20
tools: [Read, Grep, Glob, Bash]
effort: medium
---

# ${name} Agent

You are a specialized agent for [purpose].

## Responsibilities

- Responsibility 1
- Responsibility 2
- Responsibility 3

## Workflow

1. Step 1
2. Step 2
3. Step 3

## Guidelines

- Guideline 1
- Guideline 2
- Guideline 3
EOF

  cat > "${agent_dir}/README.md" <<EOF
# ${name}

Brief description of the agent.

## Installation

\`\`\`bash
# Via orkit marketplace
/plugin marketplace add github:tinhtute/orkit
\`\`\`

## Usage

\`\`\`bash
# Spawn the agent
/agent ${name}
\`\`\`

## Configuration

Describe any configuration options.

## Examples

Provide usage examples.

## License

MIT
EOF

  echo "✓ Agent scaffolded at: ${agent_dir}"
  echo "  Edit ${agent_dir}/${name}.md to customize"
}

scaffold_hook() {
  echo "Hook scaffolding not yet implemented"
  exit 1
}

scaffold_plugin() {
  echo "Plugin scaffolding not yet implemented"
  exit 1
}
```

**Create cli/lib/validate.sh:**
```bash
#!/usr/bin/env bash

validate_extension() {
  local path="${1:-}"

  if [[ -z "$path" ]]; then
    echo "Error: Path required" >&2
    echo "Usage: orkit validate <path>" >&2
    exit 1
  fi

  if [[ ! -d "$path" ]]; then
    echo "Error: Path does not exist: ${path}" >&2
    exit 1
  fi

  echo "Validating extension: ${path}"
  echo "TODO: Implement validation logic"
  # Will be implemented in Phase 2
}

test_extension() {
  local path="${1:-}"
  echo "Testing extension: ${path}"
  echo "TODO: Implement test logic"
  # Will be implemented in Phase 2
}

publish_extension() {
  local path="${1:-}"
  echo "Publishing extension: ${path}"
  echo "TODO: Implement publish logic"
  # Will be implemented in Phase 5
}
```

**Create cli/lib/security.sh:**
```bash
#!/usr/bin/env bash

# Security scanning functions
# Will be implemented in Phase 2
```

**Create cli/lib/quality.sh:**
```bash
#!/usr/bin/env bash

# Code quality check functions
# Will be implemented in Phase 2
```

**Create placeholder functions:**
```bash
# In cli/orkit, add these stubs:

search_extensions() {
  echo "Search functionality coming in Phase 4"
}

show_extension_info() {
  echo "Info functionality coming in Phase 4"
}
```

### 1.5 Developer Guide Foundation

**Create docs/README.md:**
```markdown
# Orkit Developer Guide

Welcome to the Orkit developer guide. Learn how to create production-grade Claude Code extensions.

## Contents

- [Getting Started](getting-started.md)
- [Creating Skills](creating-skills.md)
- [Creating Agents](creating-agents.md)
- [Creating Hooks](creating-hooks.md)
- [Validation Rules](validation-rules.md)
- [Contribution Guide](contribution-guide.md)

## Quick Start

1. Install orkit CLI
2. Scaffold a new extension
3. Develop and test
4. Submit for review

## Support

- GitHub Issues: https://github.com/tinhtute/orkit/issues
- Documentation: https://github.com/tinhtute/orkit/docs
```

**Create docs/getting-started.md:**
```markdown
# Getting Started

## Installation

### Install Orkit Marketplace

\`\`\`bash
# In Claude Code
/plugin marketplace add github:tinhtute/orkit
\`\`\`

### Install Orkit CLI

\`\`\`bash
git clone https://github.com/tinhtute/orkit.git
cd orkit
chmod +x cli/orkit
export PATH="$PATH:$(pwd)/cli"
\`\`\`

## Creating Your First Extension

### 1. Scaffold a Skill

\`\`\`bash
orkit scaffold skill my-first-skill
\`\`\`

### 2. Edit the Skill

Edit `plugins/my-first-skill/SKILL.md` with your skill logic.

### 3. Validate

\`\`\`bash
orkit validate plugins/my-first-skill
\`\`\`

### 4. Test Locally

\`\`\`bash
# Copy to your local Claude Code directory
cp -r plugins/my-first-skill ~/.claude/skills/
\`\`\`

## Next Steps

- Read [Creating Skills](creating-skills.md)
- Review [Validation Rules](validation-rules.md)
- Check [Contribution Guide](contribution-guide.md)
```

**Create placeholder docs:**
- `docs/creating-skills.md` (detailed in Phase 6)
- `docs/creating-agents.md` (detailed in Phase 6)
- `docs/creating-hooks.md` (detailed in Phase 6)
- `docs/validation-rules.md` (detailed in Phase 2)
- `docs/contribution-guide.md` (detailed in Phase 6)

### 1.6 License and Metadata

**Create LICENSE:**
```
MIT License

Copyright (c) 2026 Orkit Team

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## Acceptance Criteria

- [ ] Repository structure created with all directories
- [ ] marketplace.json catalog format defined
- [ ] JSON schemas created for all extension types
- [ ] CLI tool can scaffold skills and agents
- [ ] Developer guide foundation in place
- [ ] README.md with installation instructions
- [ ] LICENSE file added
- [ ] Git repository initialized

## Dependencies

- bash 4.0+
- git
- Standard Unix tools (mkdir, cat, chmod)

## Estimated Effort

3-4 days

## Next Phase

Phase 2: Validation Pipeline
