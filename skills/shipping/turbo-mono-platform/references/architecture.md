# Two-Tier Architecture

## Tier 1 — BASE (always scaffolded, every project)

```
<project>/
├── toolings/
│   ├── typescript-config/     → @rp/typescript-config
│   └── biome-config/          → @rp/biome-config
├── packages/
│   └── ui/                    → @rp/ui (shadcn preset b1tMcUv91, pre-init)
│       ├── src/components/ui/ → button, card, input, badge, dialog pre-added
│       └── src/styles/        → globals.css with @theme + dark mode
└── apps/
    └── web/                   → Next.js 16 (App Router, Turbopack)
        ├── app/layout.tsx     → @rp/ui styles + providers
        ├── app/page.tsx       → demo page with shadcn components
        └── app/(demo)/        → component showcase page
```

## Tier 2 — OPTIONAL (defined by Package Analyzer)

```
packages/
├── env/             → @rp/env        (t3-env + Zod)
├── supabase/        → @rp/supabase   (Supabase SDK + Drizzle queries)
├── auth/            → @rp/auth       (Better Auth 1.4)
├── kv/              → @rp/kv         (Upstash Redis + Ratelimit)
├── trpc/            → @rp/trpc       (tRPC v11 + TanStack Query v5)
├── validators/      → @rp/validators (shared Zod schemas)
└── remote-config/   → @rp/remote-config (feature flags)

apps/
└── api/             → Hono 4 (standalone API, optional)
```
