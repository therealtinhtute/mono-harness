---
name: get-shit-done-repo-research
description: GSD v1 meta-prompting/context-engineering system for Claude Code, v1.39, by TÂCHES
type: reference
originSessionId: 37cf314b-64eb-42de-9eeb-ac17dd2fc5cd
---
# get-shit-done Repository (GSD v1)

**URL:** https://github.com/gsd-build/get-shit-done
**npm:** `get-shit-done-cc` (v1.39.0-rc.4)
**Install:** `npx get-shit-done-cc@latest`

## What it is

GSD v1 is a **meta-prompting and context-engineering system** for AI coding agents (Claude Code, OpenCode, Gemini CLI, Kilo, Codex, Copilot, Cursor, Windsurf, etc.). It solves context rot — the quality degradation that happens as Claude fills its context window.

Built by **TÂCHES**, a solo developer. The philosophy: "complexity is in the system, not in your workflow."

## Architecture

### Workflows (get-shit-done/workflows/)
60+ workflow files covering:
- **Project lifecycle**: `new-project`, `new-milestone`, `new-workspace`, `plan-phase`, `execute-phase`, `complete-milestone`
- **Discussion**: `discuss-phase`, `discuss-phase-power`, `discuss-phase-assumptions`
- **Quality gates**: `verify-phase`, `code-review`, `audit-milestone`, `eval-review`
- **Code actions**: `add-phase`, `add-tests`, `add-todo`, `edit-phase`, `remove-phase`
- **Workspace**: `map-codebase`, `sketch`, `spike`, `health`, `forensics`
- **Ship/import**: `ship`, `import`, `resume-project`

### Sub-packages
- `get-shit-done/bin` — CLI installer
- `get-shit-done/contexts` — context definitions
- `get-shit-done/references` — reference docs
- `get-shit-done/templates` — templates
- `get-shit-done/workflows/` — workflow files (.md)
- `sdk/` — GSD SDK
- `agents/` — agent configs
- `commands/` — slash command implementations

### CLI entry points
```
get-shit-done-cc  → bin/install.js
gsd-sdk           → bin/gsd-sdk.js
```

## Key features
- **Context engineering** with XML prompt formatting
- **Subagent orchestration** (gsd-* subagents)
- **State management** via `.planning/` directory
- **Quality gates** (schema drift, security enforcement, scope reduction detection)
- **Multi-runtime support** — works with Claude Code, OpenCode, Gemini CLI, Codex, Cursor, Windsurf, etc.
- **Workstream config** — project-level planning config with deep-merge inheritance
- **v1.39 highlights**:
  - `--minimal` install profile (≤700 tokens cold-start, ~94% reduction)
  - `/gsd-edit-phase` — modify phase fields in-place
  - Post-merge build & test gate
  - Per-runtime review-model selection
  - Workstream config inheritance
  - Skill consolidation (86 → 59)

## Relationship to GSD-2
- GSD v1 is the original viral prompt framework
- GSD-2 is the standalone CLI evolution built on Pi SDK
- They share the same community (Discord, $GSD token)
- GSD v1 continues to serve its community; GSD-2 is where active development happens

## Domain terms (from CONTEXT.md)
- **Dispatch Policy Module** — dispatch error mapping, fallback policy, timeout classification
- **Command Definition Module** — canonical command metadata
- **Query Runtime Context Module** — projectDir/ws resolution at query time
- **Native Dispatch Adapter Module** — subprocess dispatch adapter
- **Query CLI Output Module** — projects dispatch results to CLI output contract
- **Query Execution Policy Module** — transport routing policy (preferNative)
- **Query Subprocess Adapter Module** — subprocess execution contract
- **Query Command Resolution Module** — command normalization/resolution

## Links
- Discord: https://discord.gg/mYgfVNfA2r
- NPM: https://www.npmjs.com/package/get-shit-done-cc
- $GSD Token: https://dexscreener.com/solana/dwudwjvan7bzkw9zwlbyv6kspdlvhwzrqy6ebk8xzxkv
