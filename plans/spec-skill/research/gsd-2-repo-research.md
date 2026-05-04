---
name: gsd-2-repo-research
description: GSD-2 coding agent CLI based on Pi SDK, v2.80, full-stack TypeScript monorepo
type: reference
originSessionId: 37cf314b-64eb-42de-9eeb-ac17dd2fc5cd
---
# GSD-2 Repository

**URL:** https://github.com/gsd-build/gsd-2
**npm:** `gsd-pi` (v2.80.0)
**Install:** `npm install -g gsd-pi@latest`

## What it is

GSD-2 is a standalone CLI coding agent built on the [Pi SDK](https://github.com/badlogic/pi-mono). It wraps Claude Code with real control over context windows, session management, git branches, cost/token tracking, auto-recovery, and milestone-based execution. One command to run an entire milestone.

## Architecture

### Packages (monorepo via workspaces)
- `@gsd/pi-tui` — terminal UI
- `@gsd/pi-ai` — AI harness
- `@gsd/pi-agent-core` — core agent logic
- `@gsd/pi-coding-agent` — coding-specific agent
- `@gsd-build/rpc-client` — RPC client
- `@gsd-build/mcp-server` — MCP server
- `@gsd/native` — native binary provisioning
- `daemon` — background daemon

### Extensions (in `src/resources/extensions/`)
- `gsd` — core GSD extension
- `subagent` — subagent dispatch
- `claude-code-cli` — Claude CLI integration
- `github-sync` — GitHub issue/PR sync
- `voice` — TTS/voice
- `browser-tools` — browser automation
- `mcp-client` — MCP client config
- `ollama` — Ollama provider
- `context7` — Context7 MCP
- `slash-commands` — slash command registry
- `async-jobs` — background jobs
- `universal-config` — config system
- `bg-shell` — background shell commands
- `search-the-web` — web search
- `google-search` — Google search

### Core CLI files (`src/`)
- `cli.ts` — main entry
- `loader.ts` — module loader
- `auto.ts` — auto orchestration
- `headless.ts` — headless mode
- `worktree-cli.ts` — worktree management
- `rtk.ts` — RTK compression
- `mcp-server.ts` — MCP server
- `models-resolver.ts` — model routing

## Key features
- **Auto-mode** with DB-backed runtime state (SQLite)
- **Worktree isolation** per milestone
- **Deep planning mode** (phase 11) with research dispatch units
- **RTK integration** for shell output compression
- **VS Code extension** with checkpoint tree view
- **Docker** support
- **Extension-first** architecture

## Vision/Principles
- Extension-first (core stays lean)
- Simplicity over abstraction
- Tests are the contract
- Ship fast, fix fast
- Provider-agnostic (any LLM)

## What they won't accept (from VISION.md)
- Enterprise patterns (DI containers, abstract factories)
- Framework swaps without clear benefit
- Cosmetic refactors
- Complexity without user value
- Heavy orchestration layers duplicating agent infra

## Links
- Discord: https://discord.invite/nKXTsAcmbT
- NPM: https://www.npmjs.com/package/gsd-pi
