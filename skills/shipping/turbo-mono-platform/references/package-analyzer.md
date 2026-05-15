# Package Analyzer — Interview & Decision

## ⛔ HARD GATE
DO NOT scaffold until ALL 5 questions answered + package list confirmed.
Ask ONE question at a time. Wait for full answer before next.

---

## Phase 1 — Auto-detect existing structure

```
!`ls apps/ packages/ toolings/ 2>/dev/null || echo "fresh project"`
!`cat package.json 2>/dev/null | python3 -c "import sys,json; d=json.load(sys.stdin); print('name:', d.get('name'), '| pm:', d.get('packageManager','?'))" 2>/dev/null || echo "no package.json"`
```

Skip Q5 if packageManager already set. Skip scaffold if structure exists.

---

## Phase 2 — Sequential Interview

### Q1 — Project type
> "What are you building?"

| Type | Base | Required Optional | Common Optional |
|------|------|-------------------|-----------------|
| **SaaS** | base | env, supabase, auth | trpc, kv, validators |
| **Internal tool** | base | env, supabase, auth | trpc, validators |
| **Marketplace** | base | env, supabase, auth, kv | trpc, validators, remote-config |
| **E-commerce** | base | env, supabase, auth, kv | trpc, validators |
| **API-only** | base (no apps/web UI) | env, supabase, auth, kv | validators |
| **Content platform** | base | env, supabase | auth, trpc, remote-config |
| **Internal dashboard** | base | env, supabase, auth | remote-config |

### Q2 — Features needed
> "Which features? (pick all that apply)"
> auth / payments / email / realtime / file upload / feature flags / rate limiting / search / background jobs

| Feature | Package(s) | Notes |
|---------|-----------|-------|
| Auth (email/pw + OAuth) | `@rp/auth` | Better Auth 1.4 |
| DB + queries | `@rp/supabase` | Drizzle adapter inside |
| Realtime | `@rp/supabase` | Supabase Realtime SDK |
| Rate limiting | `@rp/kv` | Upstash Ratelimit |
| Caching | `@rp/kv` | Upstash Redis |
| Feature flags | `@rp/remote-config` | |
| Shared validation | `@rp/validators` | Zod schemas |
| RPC layer | `@rp/trpc` | tRPC v11 + TanStack Query |
| Payments | Custom `@rp/billing` | Stripe SDK |
| Email | Custom `@rp/email` | Resend + react-email |
| File upload | Custom `@rp/storage` | Uploadthing |
| Background jobs | Custom `@rp/jobs` | Trigger.dev |

### Q3 — Domain entities
> "Main things your app manages? (e.g. users, posts, orders)"

Each entity maps to:
- `scripts/gen-schema.sh <entity>` → Drizzle schema in `@rp/supabase`
- `scripts/gen-trpc-router.sh <entity>` → tRPC router in `@rp/trpc`

### Q4 — API surface
> "Do you need a standalone API (mobile clients, third-party) or are Next.js API routes enough?"

- **Next.js routes only** → no `apps/api`
- **Standalone API** → add `apps/api` (Hono 4) → run `gen-hono-api.sh`

### Q5 — Deployment
> "Vercel, VPS, or Docker?"

| Target | Action |
|--------|--------|
| Vercel | Add `TURBO_TOKEN` + `TURBO_TEAM` to env |
| VPS/Docker | Note: add Dockerfile after scaffold |

---

## Phase 3 — Generate Package Plan

After all 5 answered, output:

```
📦 Package Plan: <project-name>
Type: <type>

Tier 1 — BASE (always)
  ✅ toolings/typescript-config
  ✅ toolings/biome-config
  ✅ packages/ui       (@rp/ui, shadcn preset b1tMcUv91)
  ✅ apps/web          (Next.js 16, demo page included)

Tier 2 — OPTIONAL for this project
  ✅ packages/env      (@rp/env)
  ✅ packages/supabase (@rp/supabase)
  ✅ packages/auth     (@rp/auth)
  ✅ packages/trpc     (@rp/trpc)
  ✅ packages/kv       (@rp/kv)
  [ ] packages/validators
  [ ] packages/remote-config
  [ ] apps/api

Domain entities: users, posts, ...

Scaffold command:
  bash scripts/scaffold-base.sh <n>
  bash scripts/scaffold-packages.sh <n> env,supabase,auth,trpc,kv

Confirm? (yes / adjust / cancel)
```

---

## Decision Rules

**Always include `@rp/env`** — every project needs env validation, no exception.

**`@rp/supabase` vs direct Supabase SDK:**
- Always use `@rp/supabase` — never import Supabase client directly in apps
- DB queries → Drizzle ORM inside `@rp/supabase`
- Realtime → Supabase SDK inside `@rp/supabase`

**`@rp/trpc` only if:**
- Full-stack Next.js app with client-side data fetching
- Multiple client components need server data
- Otherwise Next.js Server Components + Server Actions are sufficient

**`apps/api` (Hono) only if:**
- Mobile clients (iOS/Android)
- Third-party integrations needing REST
- Microservice that other services call
