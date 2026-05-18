# CLI Design Guidelines

Condensed from [clig.dev](https://clig.dev/). This is the default rubric for all CLI designs.

## Philosophy

- **Human-first design** — easy to learn, hard to misuse
- **Simple parts that compose** — do one thing, pipe to the next
- **Consistency across commands** — same flag means same thing everywhere
- **Say just enough** — no walls of text, no silent failures
- **Easy discovery** — help text is the docs for most users
- **Robustness** — handle bad input, partial failures, signals gracefully
- **Empathy** — error messages that help, not blame

## The Basics

- Use a **subcommand model** if you have 3+ distinct operations
- Keep command names **short, lowercase, no hyphens** when possible
- Prefer `tool verb noun` over `tool noun verb` (e.g., `git add file`)

## Help

- `-h` / `--help` on every command — shows usage, flags, examples
- Group flags: required first, then common, then rare
- Include 2-3 examples showing real use cases
- Show default values inline: `--timeout <seconds> (default: 30)`

## Output

- **stdout** for primary data (pipeable)
- **stderr** for messages, progress, errors (human-readable)
- Detect TTY: color/formatting when interactive, plain when piped
- `--json` for machine-readable output
- `--quiet` suppresses informational output
- `--no-color` or respect `NO_COLOR` env var

## Errors

- Write to stderr, never stdout
- Include: what went wrong, why, how to fix
- Exit codes:
  - `0` — success
  - `1` — general error
  - `2` — usage error (wrong flags, missing args)
  - `126` — permission denied
  - `127` — command not found
  - `130` — interrupted (SIGINT)
- Distinguish user errors from internal errors

## Arguments and Flags

- Positional args: max 2, or use subcommands
- Required args are positional; optional things are flags
- Boolean flags: `--flag` enables, `--no-flag` disables
- Short flags for frequent use (`-v`), long flags for clarity (`--verbose`)
- Accept `-` as stdin, `--` to end flag parsing
- Validate early, fail with clear message

## Standard Flags

Every CLI should support:

| Flag | Purpose |
|------|---------|
| `-h`, `--help` | Show help |
| `--version` | Show version |
| `-q`, `--quiet` | Suppress output |
| `-v`, `--verbose` | Increase verbosity |
| `--debug` | Debug-level output |
| `--json` | Machine-readable output |
| `--no-input` | Disable interactive prompts |
| `--no-color` | Disable color output |
| `-f`, `--force` | Skip confirmations |
| `-n`, `--dry-run` | Preview without executing |

## Interactivity

- Default to interactive when TTY detected
- `--no-input` must work for CI/automation
- Never require interactivity for core functionality
- Confirm before destructive actions (unless `--force`)

## Configuration

**Precedence** (highest to lowest):
1. Command-line flags
2. Environment variables
3. Project-level config (`.toolrc`, `tool.config.json`)
4. User-level config (`$XDG_CONFIG_HOME/tool/config`)
5. System-level config (`/etc/tool/config`)

**XDG Base Directories**:
- Config: `$XDG_CONFIG_HOME` (default `~/.config`)
- Data: `$XDG_DATA_HOME` (default `~/.local/share`)
- Cache: `$XDG_CACHE_HOME` (default `~/.cache`)

## Environment Variables

- Prefix with tool name: `MYTOOL_DEBUG=1`
- Document every env var in `--help`
- Never read generic env vars without namespacing
- `NO_COLOR=1` disables color (standard)
- `DEBUG=1` or `MYTOOL_DEBUG=1` enables debug output

## Signals

- **SIGINT** (Ctrl+C): clean up and exit 130
- **SIGTERM**: clean up and exit gracefully
- **SIGPIPE**: exit silently (piped output closed)
- Clean up temp files, release locks on any signal

## Robustness

- Validate all input before acting
- Atomic writes: write to temp, then rename
- Idempotent operations where possible
- Timeouts on network/IO operations
- Respect filesystem permissions

## Future-proofing

- Use `--flag=value` style for flags that might get values later
- Subcommands can't be renamed without breaking scripts
- Env var names can't change — pick carefully
- Version your config file format
- Reserve exit codes 3-125 for future use
