# NEVER Rules

```
NEVER use npm / yarn / pnpm                    → Bun only
NEVER use ESLint / Prettier                    → Biome only
NEVER use NextAuth or Clerk                    → Better Auth 1.4
NEVER use Prisma                               → Drizzle 0.40.x (NOT v1.0)
NEVER use Drizzle v1.0 with Better Auth        → use 0.40.x
NEVER use @repo/ as package scope              → always @rp/
NEVER import ../../packages/...                → always @rp/<n>
NEVER use process.env.VAR directly             → env.VAR from @rp/env
NEVER put secrets in NEXT_PUBLIC_*             → server-only vars
NEVER run shadcn add without -c flag           → bunx shadcn@latest add <n> -c apps/web
NEVER import @rp/ui/components/button          → import { Button } from '@rp/ui'
NEVER use middleware.ts                        → proxy.ts (Next.js 16)
NEVER use experimental.dynamicIO               → cacheComponents: true
NEVER call revalidateTag(tag) alone            → revalidateTag(tag, "max")
NEVER read params/cookies()/headers() sync     → always await
NEVER write DB queries outside @rp/supabase    → always in the package
NEVER write tRPC routers outside @rp/trpc      → always in the package
NEVER use fetch() on client                    → tRPC + TanStack Query
NEVER skip server-only in supabase/auth/kv     → must be first import
NEVER use tailwind.config.js                   → CSS-first @theme directive
NEVER use dependsOn: ["^lint"] for lint        → transit task pattern
NEVER use space-x-* / space-y-* in UI         → flex + gap-*
NEVER override component colors in className   → semantic tokens only
```
