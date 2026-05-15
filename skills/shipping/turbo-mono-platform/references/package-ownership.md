# Package Ownership

| Task | Package | Import as |
|------|---------|-----------|
| Env vars | `packages/env` | `@rp/env` |
| Supabase SDK + Realtime | `packages/supabase` | `@rp/supabase` |
| DB queries (Drizzle) | `packages/supabase` | `@rp/supabase` |
| Auth / session | `packages/auth` | `@rp/auth` |
| tRPC routers | `packages/trpc` | `@rp/trpc` |
| Redis / rate limit | `packages/kv` | `@rp/kv` |
| UI components | `packages/ui` | `@rp/ui` |
| Feature flags | `packages/remote-config` | `@rp/remote-config` |
| Shared validation | `packages/validators` | `@rp/validators` |
| Next.js app | `apps/web` | — |
| Hono API | `apps/api` | — |
