#!/usr/bin/env bash
set -e
NAME="${1:?Usage: scaffold-base.sh <project-name>}"

echo "🏗️  Scaffolding base tier: $NAME"

# --- Directories ---
mkdir -p "$NAME"/{apps/web/src/app,packages/ui/src/{components/ui,lib,styles},toolings/{typescript-config,biome-config}}

# ─── Root package.json ──────────────────────────────────────────────────────
cat > "$NAME/package.json" << EOF
{
  "name": "$NAME",
  "private": true,
  "workspaces": ["apps/*", "packages/*", "toolings/*"],
  "scripts": {
    "dev": "turbo dev",
    "build": "turbo build",
    "check": "biome check .",
    "check:fix": "biome check . --write",
    "type-check": "turbo type-check"
  },
  "devDependencies": {
    "turbo": "2.9.2",
    "typescript": "^5.8.0",
    "bun-types": "latest",
    "@biomejs/biome": "1.9.4"
  },
  "packageManager": "bun@1.2.0"
}
EOF

# ─── turbo.json ──────────────────────────────────────────────────────────────
# //#check = root-level Biome task (runs once at root, not per-package)
# Each package biome.json only needs { "extends": "//" }
cat > "$NAME/turbo.json" << 'EOF'
{
  "$schema": "https://turbo.build/schema.json",
  "ui": "tui",
  "tasks": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": [".next/**", "dist/**", "!.next/cache/**"]
    },
    "dev": { "cache": false, "persistent": true },
    "typecheck": { "dependsOn": ["^build"] },
    "//#check": {
      "outputs": [],
      "inputs": ["**/*.ts", "**/*.tsx", "**/*.js", "**/*.json", "biome.json"]
    },
    "//#check:fix": {
      "cache": false
    },
    "transit": { "dependsOn": ["^transit"] },
    "db:push": { "cache": false },
    "db:generate": { "cache": false },
    "db:studio": { "cache": false, "persistent": true }
  }
}
EOF

# ─── biome.json (root — full config) ─────────────────────────────────────────
cat > "$NAME/biome.json" << 'EOF'
{
  "$schema": "https://biomejs.dev/schemas/1.9.4/schema.json",
  "organizeImports": { "enabled": true },
  "formatter": {
    "enabled": true,
    "indentStyle": "space",
    "indentWidth": 2,
    "lineWidth": 100
  },
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true,
      "correctness": {
        "noUnusedVariables": "error",
        "noUnusedImports": "error"
      },
      "nursery": {
        "noUndeclaredEnvVars": "error"
      }
    }
  },
  "javascript": {
    "formatter": {
      "quoteStyle": "double",
      "trailingCommas": "es5",
      "semicolons": "always"
    }
  },
  "files": {
    "ignore": ["node_modules", ".next", "dist", ".turbo"]
  }
}
EOF

# ─── .gitignore ───────────────────────────────────────────────────────────────
cat > "$NAME/.gitignore" << 'EOF'
node_modules
.next
dist
.turbo
.env*.local
*.skill
EOF

# ─── toolings/typescript-config ──────────────────────────────────────────────
cat > "$NAME/toolings/typescript-config/base.json" << 'EOF'
{
  "$schema": "https://json.schemastore.org/tsconfig",
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "exactOptionalPropertyTypes": true,
    "noUncheckedIndexedAccess": true,
    "skipLibCheck": true,
    "verbatimModuleSyntax": true,
    "declaration": true,
    "sourceMap": true,
    "esModuleInterop": true,
    "isolatedModules": true
  }
}
EOF

cat > "$NAME/toolings/typescript-config/nextjs.json" << 'EOF'
{
  "extends": "./base.json",
  "compilerOptions": {
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "jsx": "preserve",
    "plugins": [{ "name": "next" }]
  }
}
EOF

cat > "$NAME/toolings/typescript-config/package.json" << 'EOF'
{
  "name": "@rp/typescript-config",
  "private": true,
  "version": "0.0.1",
  "exports": {
    "./base.json": "./base.json",
    "./nextjs.json": "./nextjs.json"
  }
}
EOF

# ─── toolings/biome-config ────────────────────────────────────────────────────
# Packages only extend root — no full config needed
cat > "$NAME/toolings/biome-config/package.json" << 'EOF'
{
  "name": "@rp/biome-config",
  "private": true,
  "version": "0.0.1"
}
EOF

# ─── packages/ui ─────────────────────────────────────────────────────────────
cat > "$NAME/packages/ui/package.json" << 'EOF'
{
  "name": "@rp/ui",
  "private": true,
  "version": "0.0.1",
  "exports": {
    ".": "./src/index.ts",
    "./components/*": "./src/components/*.tsx",
    "./styles/*": "./src/styles/*"
  },
  "scripts": {
    "type-check": "tsc --noEmit"
  },
  "dependencies": {
    "tailwindcss": "^4.0.0",
    "class-variance-authority": "^0.7.0",
    "clsx": "^2.0.0",
    "tailwind-merge": "^2.0.0",
    "lucide-react": "^0.400.0",
    "geist": "^1.3.0"
  },
  "devDependencies": {
    "@rp/typescript-config": "workspace:*"
  }
}
EOF

# Per-package biome.json — only extends root
cat > "$NAME/packages/ui/biome.json" << 'EOF'
{
  "extends": "//"
}
EOF

cat > "$NAME/packages/ui/tsconfig.json" << 'EOF'
{
  "extends": "@rp/typescript-config/base.json",
  "compilerOptions": { "jsx": "react-jsx" },
  "include": ["src"]
}
EOF

cat > "$NAME/packages/ui/components.json" << 'EOF'
{
  "$schema": "https://ui.shadcn.com/schema.json",
  "style": "new-york",
  "preset": "b1tMcUv91",
  "rsc": true,
  "tsx": true,
  "tailwind": {
    "config": "",
    "css": "src/styles/globals.css",
    "baseColor": "neutral",
    "cssVariables": true
  },
  "aliases": {
    "components": "@rp/ui/components",
    "utils": "@rp/ui/lib/utils",
    "ui": "@rp/ui/components/ui"
  }
}
EOF

cat > "$NAME/packages/ui/src/lib/utils.ts" << 'EOF'
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
EOF

# packages/ui has its OWN css that just imports tailwindcss
# apps/web globals.css then imports this + adds @source
cat > "$NAME/packages/ui/src/styles/default.css" << 'EOF'
@import "tailwindcss";

@theme {
  --font-sans: "Geist", sans-serif;
  --font-mono: "Geist Mono", monospace;
  --radius: 0.5rem;
}

@layer base {
  :root {
    --background: oklch(1 0 0);
    --foreground: oklch(0.145 0 0);
    --primary: oklch(0.205 0 0);
    --primary-foreground: oklch(0.985 0 0);
    --secondary: oklch(0.97 0 0);
    --secondary-foreground: oklch(0.205 0 0);
    --muted: oklch(0.97 0 0);
    --muted-foreground: oklch(0.556 0 0);
    --border: oklch(0.922 0 0);
    --ring: oklch(0.708 0 0);
  }
  .dark {
    --background: oklch(0.145 0 0);
    --foreground: oklch(0.985 0 0);
    --primary: oklch(0.985 0 0);
    --primary-foreground: oklch(0.205 0 0);
    --muted: oklch(0.269 0 0);
    --muted-foreground: oklch(0.708 0 0);
    --border: oklch(1 0 0 / 10%);
    --ring: oklch(0.556 0 0);
  }
}
EOF

cat > "$NAME/packages/ui/src/index.ts" << 'EOF'
export { cn } from "./lib/utils";
EOF

# ─── apps/web ─────────────────────────────────────────────────────────────────
cat > "$NAME/apps/web/package.json" << EOF
{
  "name": "web",
  "private": true,
  "version": "0.0.1",
  "scripts": {
    "dev": "next dev --turbopack",
    "build": "next build",
    "start": "next start",
    "type-check": "tsc --noEmit"
  },
  "dependencies": {
    "next": "16.2.0",
    "react": "19.2.0",
    "react-dom": "19.2.0",
    "geist": "^1.3.0",
    "@rp/ui": "workspace:*"
  },
  "devDependencies": {
    "@rp/typescript-config": "workspace:*",
    "typescript": "^5.8.0"
  }
}
EOF

# Per-package biome.json — only extends root
cat > "$NAME/apps/web/biome.json" << 'EOF'
{
  "extends": "//"
}
EOF

cat > "$NAME/apps/web/next.config.ts" << 'EOF'
import type { NextConfig } from "next";
const config: NextConfig = {
  experimental: { cacheComponents: true },
  transpilePackages: ["@rp/ui"],
};
export default config;
EOF

cat > "$NAME/apps/web/tsconfig.json" << 'EOF'
{
  "extends": "@rp/typescript-config/nextjs.json",
  "compilerOptions": {
    "baseUrl": ".",
    "paths": { "@/*": ["./src/*"] }
  },
  "include": ["src", "next-env.d.ts", "next.config.ts"],
  "exclude": ["node_modules"]
}
EOF

# apps/web/globals.css:
# 1. imports tailwindcss (via @rp/ui/styles/default.css)
# 2. @source tells Tailwind to scan packages/ui/src for classes
# Path depth: apps/web/src/app/globals.css → 4 levels up → packages/ui/src
cat > "$NAME/apps/web/src/app/globals.css" << 'EOF'
@import "tailwindcss";
@import "@rp/ui/styles/default.css";
@source "../../../../packages/ui/src";
EOF

cat > "$NAME/apps/web/src/app/layout.tsx" << 'EOF'
import type { Metadata } from "next";
import { GeistSans } from "geist/font/sans";
import { GeistMono } from "geist/font/mono";
import "./globals.css";

export const metadata: Metadata = {
  title: "turbo-mono-platform",
  description: "Next.js 16 + shadcn/ui + @rp/ui",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${GeistSans.variable} ${GeistMono.variable} font-sans antialiased`}>
        {children}
      </body>
    </html>
  );
}
EOF

cat > "$NAME/apps/web/src/app/page.tsx" << 'EOF'
export default function HomePage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-8 p-8">
      <h1 className="text-4xl font-bold tracking-tight">turbo-mono-platform</h1>
      <p className="text-muted-foreground">Next.js 16 · shadcn/ui · @rp/ui</p>
      <p className="text-sm text-muted-foreground">
        Run{" "}
        <code className="bg-muted px-1 rounded">
          bunx shadcn@latest add button -c apps/web
        </code>{" "}
        to add components
      </p>
    </main>
  );
}
EOF

echo ""
echo "✅ Base tier scaffolded: $NAME/"
echo ""
echo "Next steps:"
echo "  1. cd $NAME && bun install"
echo "  2. bunx shadcn@latest init --preset b1tMcUv91 --template next --monorepo"
echo "  3. bunx shadcn@latest add button card input badge dialog separator -c apps/web"
echo "  4. bun run dev"
echo ""
echo "Shared root dep:  bun add -w <pkg>     (adds to root, not per-package)"
echo "Optional pkgs:    bash scripts/scaffold-packages.sh $NAME env,supabase,auth,kv,trpc"
