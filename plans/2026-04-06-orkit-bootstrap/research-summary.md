---
date: 2026-04-06
project: orkit
phase: research
---

# Research Summary: Claude Code Extension System

## Extension Formats

### 1. Skills (SKILL.md)
**Location**: `.claude/skills/<skill-name>/SKILL.md`

**Format**: YAML frontmatter + Markdown body
```yaml
---
name: skill-name
description: Brief description for discovery
allowed-tools: [Read, Grep, Glob]
model: sonnet
context: fork
agent: Explore
disable-model-invocation: true
user-invocable: true
argument-hint: [filename]
effort: medium
paths: src/**/*.ts
shell: bash
---

Detailed instructions here...
```

**Key Features**:
- Progressive disclosure: metadata (~100 tokens) → full instructions (<5k tokens)
- String substitution: `$ARGUMENTS`, `$N`, `${CLAUDE_SESSION_ID}`, `${CLAUDE_SKILL_DIR}`
- Can include `references/`, `scripts/`, `.env.example`
- Supports symlinks for shared skill libraries

### 2. Agents (agent-name.md)
**Location**: `.claude/agents/<agent-name>.md`

**Format**: YAML frontmatter + Markdown body
```yaml
---
name: agent-name
description: Agent purpose
tools: [Read, Grep, Glob, Bash]
disallowedTools: [Write, Edit]
model: sonnet
permissionMode: acceptEdits
maxTurns: 20
skills: [api-conventions, error-handling]
mcpServers: [github, playwright]
hooks: {...}
memory: user
background: false
effort: medium
isolation: worktree
color: blue
initialPrompt: "Review recent changes"
---

You are a specialized agent...
```

**Key Features**:
- Tools allowlist/denylist
- Preloaded skills
- Persistent memory (user/project/local)
- Git worktree isolation
- Lifecycle hooks

### 3. Hooks (hooks.json in settings.json)
**Location**: `.claude/settings.json` or `.claude/hooks/`

**Format**: JSON configuration
```json
{
  "hooks": {
    "PostToolUse": [{
      "matcher": "Write|Edit",
      "hooks": [{
        "type": "command",
        "command": "npx prettier --write",
        "timeout": 60,
        "async": false
      }]
    }]
  }
}
```

**12+ Lifecycle Events**:
- SessionStart, UserPromptSubmit, PreToolUse, PermissionRequest, PermissionDenied
- PostToolUse, PostToolUseFailure, SubagentStart, SubagentStop
- Stop, StopFailure, PreCompact, PostCompact, SessionEnd, Notification
- TaskCreated, TaskCompleted, TeammateIdle, InstructionsLoaded, ConfigChange
- CwdChanged, FileChanged, WorktreeCreate, WorktreeRemove
- Elicitation, ElicitationResult

**Hook Types**:
- `command`: Shell scripts
- `http`: POST requests
- `prompt`: LLM evaluation
- `agent`: Subagent execution

**Exit Codes**:
- 0 = success
- 2 = block operation with stderr sent to Claude
- other = error logged

### 4. Statusline (statusline.sh)
**Location**: Configured in `.claude/settings.json`

**Format**: Executable script + settings.json config
```json
{
  "statusLine": {
    "type": "command",
    "command": "~/.claude/statusline.sh",
    "padding": 0
  }
}
```

**Script receives JSON via stdin**:
```json
{
  "model": {"display_name": "Opus"},
  "workspace": {"current_dir": "/path"},
  "cost": {"total_cost_usd": 0.42},
  "context_window": {
    "used_percentage": 37.5,
    "context_window_size": 200000
  }
}
```

**Output**: Single line to stdout with ANSI color codes

### 5. Plugins (plugin.json)
**Location**: `.claude-plugin/plugin.json`

**Format**: JSON manifest
```json
{
  "name": "plugin-name",
  "version": "1.0.0",
  "description": "Brief description",
  "author": {"name": "Author", "email": "author@example.com"},
  "homepage": "https://docs.example.com",
  "repository": "https://github.com/owner/repo",
  "license": "MIT",
  "keywords": ["keyword1", "keyword2"],
  "commands": "./commands/",
  "agents": "./agents/",
  "skills": "./skills/",
  "outputStyles": "./output-styles/",
  "hooks": "./hooks/hooks.json",
  "mcpServers": "./.mcp.json",
  "lspServers": "./.lsp.json",
  "userConfig": {
    "api_token": {"description": "API token", "sensitive": true}
  },
  "channels": [{"server": "telegram"}]
}
```

**Environment Variables**:
- `${CLAUDE_PLUGIN_ROOT}`: Installation directory
- `${CLAUDE_PLUGIN_DATA}`: Persistent data across updates
- `CLAUDE_PLUGIN_OPTION_<KEY>`: User config values

### 6. Marketplaces (marketplace.json)
**Location**: `.claude-plugin/marketplace.json`

**Format**: JSON catalog
```json
{
  "name": "marketplace-name",
  "owner": {"name": "Team", "email": "team@example.com"},
  "metadata": {
    "description": "Marketplace description",
    "version": "1.0.0",
    "pluginRoot": "./plugins"
  },
  "plugins": [
    {
      "name": "plugin-name",
      "source": "./plugins/plugin-name",
      "description": "Plugin description",
      "version": "2.1.0",
      "category": "productivity",
      "tags": ["formatting"],
      "strict": true
    }
  ]
}
```

**Source Types**:
- Relative path: `./plugins/name`
- GitHub: `{"source": "github", "repo": "owner/repo", "ref": "v2.0.0", "sha": "a1b2c3d..."}`
- NPM: `{"source": "npm", "package": "@org/plugin", "version": "^2.0.0"}`
- Git subdir: `{"source": "git-subdir", "url": "https://...", "path": "tools/plugin", "ref": "main"}`

## Directory Structure

### Project-level (.claude/)
```
.claude/
├── settings.json              # Team config (committed)
├── settings.local.json        # Personal overrides (gitignored)
├── .mcp.json                  # MCP servers
├── rules/                     # Modular instructions
├── commands/                  # Legacy slash commands
├── skills/                    # Auto-invoked workflows
│   └── <skill-name>/
│       ├── SKILL.md
│       ├── references/
│       └── scripts/
├── agents/                    # Specialized subagents
├── hooks/                     # Event automation
├── output-styles/             # Output formatting
├── agent-memory/              # Agent persistent memory (committed)
└── agent-memory-local/        # Local agent memory (gitignored)
```

### User-level (~/.claude/)
```
~/.claude/
├── CLAUDE.md                  # Global instructions
├── settings.json              # Global config
├── commands/
├── skills/
├── agents/
├── hooks/
├── plugins/
│   ├── installed_plugins.json
│   ├── known_marketplaces.json
│   ├── blocklist.json
│   └── cache/                 # Downloaded plugins
│       └── <marketplace>/
│           └── <plugin>/
│               └── <version>/
│                   ├── .claude-plugin/
│                   ├── skills/
│                   ├── commands/
│                   ├── agents/
│                   └── hooks/
└── projects/                  # Session history
    └── <project-hash>/
        ├── memory/MEMORY.md
        └── subagents/
```

## Settings Hierarchy (5 Scopes)
1. **Managed** (highest): System-wide IT policies
2. **Command line**: `--model` flag
3. **Local**: `.claude/settings.local.json` (gitignored)
4. **Project**: `.claude/settings.json` (committed)
5. **User** (lowest): `~/.claude/settings.json`

## Best Practices

### Naming Conventions
- Use kebab-case for all file and directory names
- Skill names become `/slash-commands`
- Plugin names must be unique within marketplace

### Progressive Disclosure
- Metadata: ~100 tokens (frontmatter)
- Full instructions: <5k tokens (markdown body)
- Bundled resources: references/, scripts/

### Security
- Validate all JSON configs
- Scan for secrets and malicious code
- Use exit code 2 in hooks to block operations
- Restrict tools with `allowed-tools` or `disallowedTools`

### Versioning
- Semantic versioning (MAJOR.MINOR.PATCH)
- Pin to specific refs or SHAs for reproducibility
- Support update detection

### Distribution
- Support multiple hosting (GitHub, GitLab, npm, URLs)
- Use git-subdir for monorepos (sparse checkout)
- Cache plugins with versioned directories
- Support private repos with tokens

## Key Insights for Orkit

1. **Plugin System is the Distribution Model**: Claude Code already has a plugin system with marketplace support
2. **Marketplace.json is the Catalog Format**: We should use this standard format
3. **Multiple Extension Types in One Plugin**: Plugins can bundle skills, agents, hooks, statusline
4. **Installation is Built-in**: `/plugin marketplace add` command exists
5. **Validation is Critical**: Schema validation, security scanning, code quality checks
6. **Symlinks Enable Sharing**: Skills can be symlinked from shared locations
7. **Date-based Versioning Fits**: Aligns with snapshot-based stability model
8. **Bash-only is Feasible**: All tooling can be shell scripts
9. **Progressive Disclosure Matters**: Keep metadata small, full content lazy-loaded
10. **Security is Paramount**: Exit code 2 blocking, token scanning, validation pipeline

## Recommended Architecture for Orkit

### Option 1: Pure Marketplace (Recommended)
- Create `marketplace.json` catalog
- Host on GitHub
- Users add with `/plugin marketplace add github:tinhtute/orkit`
- Leverage existing Claude Code plugin system
- Focus on curating high-quality extensions

### Option 2: Enhanced CLI + Marketplace
- Build bash CLI for enhanced discovery/search
- Generate marketplace.json from CLI
- Provide validation and scaffolding tools
- Still use Claude Code's plugin system for installation

### Option 3: Standalone Registry
- Build custom registry separate from plugin system
- Requires custom installation mechanism
- More control but more complexity
- May not integrate as smoothly

**Recommendation**: Start with Option 1 (pure marketplace), add Option 2 enhancements if needed.
