# Turbo Tasks — turbo.json Configuration

## Canonical turbo.json (Turborepo 2.9 + futureFlags)

```json
{
  "$schema": "https://turbo.build/schema.json",
  "ui": "tui",
  "futureFlags": {
    "globalConfiguration": true
  },
  "global": {
    "inputs": ["tsconfig.json", ".env.*local"],
    "env": ["NODE_ENV"]
  },
  "tasks": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": [".next/**", "dist/**"]
    },
    "transit": {
      "dependsOn": ["^transit"]
    },
    "lint": {
      "dependsOn": ["transit"]
    },
    "type-check": {
      "dependsOn": ["transit"]
    },
    "dev": {
      "cache": false,
      "persistent": true
    },
    "db:push": { "cache": false },
    "db:generate": { "cache": false },
    "db:studio": { "cache": false, "persistent": true }
  }
}
```

## Transit Task Pattern — Parallel lint + type-check

**Problem:** `dependsOn: ["^lint"]` forces lint to run sequentially.

**Solution:** `transit` — a virtual task with no script that creates dependency relationships without running anything. Packages declare it, so lint and type-check know when their upstream deps are ready.

```json
{
  "tasks": {
    "transit": { "dependsOn": ["^transit"] },
    "lint": { "dependsOn": ["transit"] },
    "type-check": { "dependsOn": ["transit"] }
  }
}
```

Each package `package.json` does NOT need a `transit` script — Turborepo skips packages without the script. It only acts as a graph node for ordering.

## futureFlags.globalConfiguration

Available in Turbo 2.9. Moves global settings under a `global` key:

```json
{
  "futureFlags": { "globalConfiguration": true },
  "global": {
    "inputs": ["tsconfig.json"],
    "env": ["CI", "NODE_ENV"]
  },
  "tasks": { ... }
}
```

Without `futureFlags`, use `globalDependencies` and `globalEnv` at root level.

## Task inputs — Precise Cache Invalidation

```json
{
  "tasks": {
    "build": {
      "inputs": ["$TURBO_DEFAULT$", "!README.md", "!**/*.test.ts"],
      "outputs": [".next/**", "dist/**"]
    },
    "lint": {
      "inputs": ["src/**", "biome.json", "tsconfig.json"]
    }
  }
}
```

`$TURBO_DEFAULT$` = all package files. Use `!pattern` to exclude.

## Task-specific package — `shared#build`

```json
{
  "tasks": {
    "build": {
      "dependsOn": ["^build", "@rp/ui#build"]
    }
  }
}
```

Forces `@rp/ui` to build before any consumer, even without a dependency in `package.json`.

## Remote Cache Setup (Vercel)

```bash
bunx turbo login
bunx turbo link
```

Add to CI env:
```
TURBO_TOKEN=<secret>
TURBO_TEAM=<team-slug>
```

## Debugging

```bash
bun run build -- --dry=json   # what would run, as JSON
bun run build -- --summarize  # cache hit rate report
bun run build -- --graph      # open dependency graph in browser
bun run build -- --force      # ignore cache, re-run everything
```
