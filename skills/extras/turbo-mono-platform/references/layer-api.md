# Layer API Reference

Each layer: package location, first import, key export, critical gotcha.

---

## @rp/env

```typescript
import "server-only";                          // MUST be first
import { createEnv } from "@t3-oss/env-nextjs";
import { z } from "zod";

export const env = createEnv({
  server: {
    DATABASE_URL: z.string().url(),
    BETTER_AUTH_SECRET: z.string().min(32),
    UPSTASH_REDIS_REST_URL: z.string().url(),
    UPSTASH_REDIS_REST_TOKEN: z.string().min(1),
  },
  client: {
    NEXT_PUBLIC_APP_URL: z.string().url(),
  },
  runtimeEnv: { /* mirror all keys */ },
});
```

**Gotcha:** `emptyStringAsUndefined: true` prevents silent empty string bugs.

---

## @rp/db

```typescript
import "server-only";                          // MUST be first
import { drizzle } from "drizzle-orm/postgres-js";
import postgres from "postgres";
import { env } from "@rp/env";
import * as schema from "./schema";

const client = postgres(env.DATABASE_URL, { prepare: false });
export const db = drizzle(client, { schema });
export * from "./schema";
```

**Gotcha:** Use `drizzle-orm/postgres-js` NOT `drizzle-orm/node-postgres`. Version **0.40.x** — v1.0 breaks Better Auth adapter.

---

## @rp/auth

```typescript
import "server-only";                          // MUST be first
import { betterAuth } from "better-auth";
import { drizzleAdapter } from "better-auth/adapters/drizzle";
import { db } from "@rp/db";
import { env } from "@rp/env";

export const auth = betterAuth({
  secret: env.BETTER_AUTH_SECRET,
  database: drizzleAdapter(db, { provider: "pg", usePlural: true }),
  experimental: { joins: true },               // REQUIRED for Drizzle adapter
});
```

**getSession() in Server Components:**
```typescript
import { headers } from "next/headers";
export async function getSession() {
  return auth.api.getSession({ headers: await headers() });
}
```

**Route handler** at `apps/web/app/api/auth/[...all]/route.ts`:
```typescript
import { auth } from "@rp/auth";
import { toNextJsHandler } from "better-auth/next-js";
export const { GET, POST } = toNextJsHandler(auth);
```

---

## @rp/trpc

```typescript
import "server-only";
import { initTRPC, TRPCError } from "@trpc/server";
import { cache } from "react";
import { auth } from "@rp/auth";
import { headers } from "next/headers";

export const createTRPCContext = cache(async () => {
  const session = await auth.api.getSession({ headers: await headers() });
  return { session };
});

const t = initTRPC.context<typeof createTRPCContext>().create();
export const router = t.router;
export const publicProcedure = t.procedure;
export const protectedProcedure = t.procedure.use(({ ctx, next }) => {
  if (!ctx.session) throw new TRPCError({ code: "UNAUTHORIZED" });
  return next({ ctx: { session: ctx.session } });
});
```

**Next.js route handler** at `apps/web/app/api/trpc/[trpc]/route.ts`:
```typescript
import { fetchRequestHandler } from "@trpc/server/adapters/fetch";
import { appRouter, createTRPCContext } from "@rp/trpc";
const handler = (req: Request) =>
  fetchRequestHandler({ endpoint: "/api/trpc", req, router: appRouter, createContext: createTRPCContext });
export { handler as GET, handler as POST };
```

**Client hook usage:**
```typescript
"use client";
import { trpc } from "@rp/trpc/client";
const { data } = trpc.user.me.useQuery();
```

---

## @rp/kv

```typescript
import "server-only";                          // MUST be first
import { Redis } from "@upstash/redis";
import { Ratelimit } from "@upstash/ratelimit";
import { env } from "@rp/env";

export const redis = new Redis({
  url: env.UPSTASH_REDIS_REST_URL,
  token: env.UPSTASH_REDIS_REST_TOKEN,
});

export const ratelimit = new Ratelimit({
  redis,
  limiter: Ratelimit.slidingWindow(60, "1 m"),
  analytics: true,
});
```

---

## @rp/ui — shadcn Components

```typescript
import { Button } from "@rp/ui/components/button";
import { cn } from "@rp/ui/lib/utils";
import "@rp/ui/styles/globals.css";          // import in apps/web layout.tsx
```

**Never** import from relative path — always `@rp/ui/components/<name>`.

---

## Package exports pattern (all packages)

```json
{
  "name": "@rp/<n>",
  "private": true,
  "exports": {
    ".": "./src/index.ts"
  }
}
```

```json
{
  "extends": "@rp/typescript-config/base.json",
  "include": ["src"]
}
```
