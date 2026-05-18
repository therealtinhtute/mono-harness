# CLI Spec Template

Fill this in during Phase 2. Delete sections that don't apply.

---

## {NAME}

> {One-liner: what it does in ≤10 words}

**Target user**: {developer | ops | end-user | CI/automation}
**Language**: {Go | Rust | Node.js | Bash}
**Framework**: {Cobra | Clap | Commander | manual}

---

## USAGE

```
{name} [global-flags] <command> [command-flags] [args...]
```

---

## Commands

| Command | Description | Aliases |
|---------|-------------|---------|
| `{name} init` | ... | |
| `{name} run` | ... | `r` |
| `{name} config` | ... | |

### `{name} init`

```
{name} init [--template <name>] [path]
```

**Arguments**:
- `path` — (optional) directory to initialize. Default: current directory.

**Flags**:
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--template` | `-t` | string | `default` | Template to use |
| `--force` | `-f` | bool | false | Overwrite existing files |

**Output**: Creates project structure, prints summary to stdout.
**Errors**: Exit 1 if path exists and `--force` not set.

---

## Global Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--help` | `-h` | bool | | Show help |
| `--version` | | bool | | Show version |
| `--json` | | bool | false | Output as JSON |
| `--quiet` | `-q` | bool | false | Suppress non-essential output |
| `--verbose` | `-v` | bool | false | Verbose output |
| `--debug` | | bool | false | Debug-level output |
| `--no-input` | | bool | false | Disable interactive prompts |
| `--no-color` | | bool | false | Disable color (also: NO_COLOR env) |
| `--config` | `-c` | string | | Path to config file |

---

## I/O Contract

**stdout**: Primary data output. Pipeable. JSON when `--json` is set.
**stderr**: Progress, warnings, errors. Human-readable.
**stdin**: {Accepts piped input? Describe format.}

### JSON Output Schema

```json
{
  "version": "1",
  "data": { ... },
  "error": null
}
```

### Error JSON Schema

```json
{
  "version": "1",
  "data": null,
  "error": {
    "code": "E001",
    "message": "human-readable message",
    "hint": "how to fix"
  }
}
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Usage error (bad flags, missing args) |
| 126 | Permission denied |
| 127 | Required dependency not found |
| 130 | Interrupted (SIGINT) |

---

## Configuration

**File location**: `$XDG_CONFIG_HOME/{name}/config.{toml|yaml|json}`
**Default**: `~/.config/{name}/config.toml`

**Precedence**: flags > env > project config (`.{name}rc`) > user config > defaults

### Config File Schema

```toml
# ~/.config/{name}/config.toml
[defaults]
output = "json"
verbose = false

[auth]
token = "..."
```

---

## Environment Variables

| Variable | Description | Equivalent flag |
|----------|-------------|-----------------|
| `{NAME}_CONFIG` | Config file path | `--config` |
| `{NAME}_TOKEN` | Auth token | N/A |
| `{NAME}_DEBUG` | Enable debug output | `--debug` |
| `NO_COLOR` | Disable color output | `--no-color` |

---

## Shell Completions

```bash
# Bash
{name} completion bash > /etc/bash_completion.d/{name}

# Zsh
{name} completion zsh > "${fpath[1]}/_{name}"

# Fish
{name} completion fish > ~/.config/fish/completions/{name}.fish
```

---

## Examples

```bash
# Basic usage
{name} run --verbose

# Pipe JSON output
{name} list --json | jq '.data[]'

# Non-interactive (CI)
{name} deploy --no-input --force

# Use config file
{name} --config ./custom.toml run
```

---

## Platform Support

| Platform | Supported | Notes |
|----------|-----------|-------|
| macOS (arm64) | yes | Primary development |
| macOS (amd64) | yes | |
| Linux (amd64) | yes | CI/servers |
| Linux (arm64) | yes | |
| Windows (amd64) | {yes/no} | {notes} |

---

## Distribution

**Primary**: {Homebrew tap | npm | GitHub releases | curl install}
**Secondary**: {others}
**CI release trigger**: git tag `v*`

See `shipping-checklist.md` for full release process.
