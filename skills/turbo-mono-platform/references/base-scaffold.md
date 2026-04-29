# Base Scaffold — Tier 1

Every project starts with this. No exceptions.

## Structure

```
<project>/
├── apps/
│   └── web/                         → Next.js 16 (--turbopack)
│       ├── biome.json               → { "extends": "//" }  only
│       ├── src/app/
│       │   ├── globals.css          → @import + @source
│       │   ├── layout.tsx
│       │   └── page.tsx             → demo page
│       └── next.config.ts
├── packages/
│   └── ui/                          → @rp/ui
│       ├── biome.json               → { "extends": "//" }  only
│       ├── components.json          → shadcn preset b1tMcUv91
│       └── src/
│           ├── components/ui/       → shadcn components here
│           ├── lib/utils.ts         → cn()
│           └── styles/
│               └── default.css     → @import "tailwindcss" + @theme
└── toolings/
    ├── typescript-config/
    │   ├── base.json
    │   ├── nextjs.json
    │   └── package.json            → @rp/typescript-config
    └── biome-config/
        └── package.json            → @rp/biome-config
```

---

## Critical CSS Pattern

### `packages/ui/src/styles/default.css` — theme source

```css
@import "tailwindcss";

@theme {
  --font-sans: "Geist", sans-serif;
  --radius: 0.5rem;
}

@layer base {
  :root { --background: oklch(1 0 0); ... }
  .dark { --background: oklch(0.145 0 0); ... }
}
```

### `apps/web/src/app/globals.css` — the consuming app

```css
@import "tailwindcss";
@import "@rp/ui/styles/default.css";
@source "../../../../packages/ui/src";
```

**Why `@source`?**
Tailwind v4 only scans files it can find from the app's directory. Components in `packages/ui/src` are outside `apps/web`, so Tailwind misses their classes. `@source` explicitly tells Tailwind to also scan that path.

**Path math:** `apps/web/src/app/globals.css` → `../../../../` = project root → `packages/ui/src`
If globals.css is at `apps/web/src/app/globals.css`, path is `../../../../packages/ui/src`.
If globals.css is at `apps/web/globals.css`, path is `../../packages/ui/src`.

**Always adjust path depth based on actual nesting.**

---

## Biome Setup — Root Task Pattern

Root `biome.json` has full config. Every package only extends:

```json
{ "extends": "//" }
```

`//` = project root. Biome resolves it automatically.

Root `turbo.json`:
```json
{
  "tasks": {
    "//#check": {
      "outputs": [],
      "inputs": ["**/*.ts", "**/*.tsx", "**/*.js", "**/*.json", "biome.json"]
    },
    "//#check:fix": { "cache": false }
  }
}
```

Root `package.json` scripts:
```json
{
  "check": "biome check .",
  "check:fix": "biome check . --write"
}
```

**Why root task?** Biome is so fast it's cheaper to run once at root than per-package. Cache invalidation is also simpler. Per-package biome.json still exists for IDE integration (VS Code picks it up).

---

## `nursery.noUndeclaredEnvVars` Pattern

Root `biome.json`:
```json
{
  "linter": {
    "rules": {
      "nursery": { "noUndeclaredEnvVars": "error" }
    }
  }
}
```

When enabled, any `process.env.X` in code must be declared in `turbo.json`:
```json
{
  "tasks": {
    "build": {
      "env": ["DATABASE_URL", "NEXT_PUBLIC_APP_URL", "BETTER_AUTH_SECRET"]
    }
  }
}
```

**Rule:** Declare vars in the task that actually uses them, not in all tasks.

---

## `packages/ui` Exports — Wildcard Pattern

```json
{
  "name": "@rp/ui",
  "version": "0.0.1",
  "exports": {
    ".": "./src/index.ts",
    "./components/*": "./src/components/*.tsx",
    "./styles/*": "./src/styles/*"
  }
}
```

Wildcard exports allow:
```typescript
import { Button } from "@rp/ui";                    // barrel from src/index.ts
import { Button } from "@rp/ui/components/button";  // direct component
import "@rp/ui/styles/default.css";                  // css file
```

---

## turbo.json — Correct Outputs

```json
{
  "tasks": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": [".next/**", "dist/**", "!.next/cache/**"]
    }
  }
}
```

`!.next/cache/**` excludes Next.js build cache from Turborepo cache — prevents large unnecessary cache uploads.

---

## Pre-added shadcn Components

```bash
# Init (from apps/web directory or repo root)
bunx shadcn@latest init --preset b1tMcUv91 --template next --monorepo

# Add base components (from apps/web directory)
bunx shadcn@latest add button -c apps/web
bunx shadcn@latest add card -c apps/web
bunx shadcn@latest add input -c apps/web
bunx shadcn@latest add badge -c apps/web
bunx shadcn@latest add dialog -c apps/web
bunx shadcn@latest add separator -c apps/web
bunx shadcn@latest add skeleton -c apps/web
```

Components land in `packages/ui/src/components/ui/` because of `components.json` aliases.

---

## Internal Package Conventions

Every internal package must have:
```json
{
  "name": "@rp/<n>",
  "private": true,
  "version": "0.0.1",
  "devDependencies": {
    "@rp/typescript-config": "workspace:*"
  }
}
```

- `"private": true` — never published to npm
- `"version": "0.0.1"` — required by some tooling
- `workspace:*` — internal references always use this protocol

## Shared Root Deps

```bash
# Add to root (shared across all apps/packages)
bun add -w <pkg>

# Examples:
bun add -w zod
bun add -w typescript

# NOT inside individual packages unless package-specific
```
