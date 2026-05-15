#!/usr/bin/env bash
set -e
NAME="${1:?Usage: scaffold-packages.sh <project-name> <pkg1,pkg2,...>}"
PKGS="${2:?Provide comma-separated packages: env,supabase,auth,kv,trpc,validators,remote-config,api}"

IFS=',' read -ra PKG_LIST <<< "$PKGS"

for PKG in "${PKG_LIST[@]}"; do
  PKG=$(echo "$PKG" | tr -d ' ')
  DIR="$NAME/packages/$PKG"
  echo "📦 Scaffolding @rp/$PKG..."

  mkdir -p "$DIR/src"

  cat > "$DIR/tsconfig.json" << TSJSON
{
  "extends": "@rp/typescript-config/base.json",
  "include": ["src"]
}
TSJSON

  case "$PKG" in
    env)
      cat > "$DIR/package.json" << PKGJSON
{
  "name": "@rp/env",
  "private": true,
  "version": "0.0.1",
  "exports": { ".": "./src/index.ts" },
  "dependencies": {
    "@t3-oss/env-nextjs": "^0.11.0",
    "zod": "^3.0.0",
    "server-only": "^0.0.1"
  },
  "devDependencies": { "@rp/typescript-config": "workspace:*" }
}
PKGJSON
      cat > "$DIR/src/index.ts" << 'SRCTS'
import "server-only";
import { createEnv } from "@t3-oss/env-nextjs";
import { z } from "zod";

export const env = createEnv({
  server: {
    NODE_ENV: z.enum(["development", "test", "production"]),
    DATABASE_URL: z.string().url(),
    BETTER_AUTH_SECRET: z.string().min(32),
    BETTER_AUTH_URL: z.string().url(),
    UPSTASH_REDIS_REST_URL: z.string().url().optional(),
    UPSTASH_REDIS_REST_TOKEN: z.string().min(1).optional(),
    SUPABASE_SERVICE_ROLE_KEY: z.string().min(1).optional(),
  },
  client: {
    NEXT_PUBLIC_APP_URL: z.string().url(),
    NEXT_PUBLIC_SUPABASE_URL: z.string().url().optional(),
    NEXT_PUBLIC_SUPABASE_ANON_KEY: z.string().min(1).optional(),
  },
  runtimeEnv: {
    NODE_ENV: process.env.NODE_ENV,
    DATABASE_URL: process.env.DATABASE_URL,
    BETTER_AUTH_SECRET: process.env.BETTER_AUTH_SECRET,
    BETTER_AUTH_URL: process.env.BETTER_AUTH_URL,
    UPSTASH_REDIS_REST_URL: process.env.UPSTASH_REDIS_REST_URL,
    UPSTASH_REDIS_REST_TOKEN: process.env.UPSTASH_REDIS_REST_TOKEN,
    SUPABASE_SERVICE_ROLE_KEY: process.env.SUPABASE_SERVICE_ROLE_KEY,
    NEXT_PUBLIC_APP_URL: process.env.NEXT_PUBLIC_APP_URL,
    NEXT_PUBLIC_SUPABASE_URL: process.env.NEXT_PUBLIC_SUPABASE_URL,
    NEXT_PUBLIC_SUPABASE_ANON_KEY: process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY,
  },
  emptyStringAsUndefined: true,
});
SRCTS
      ;;

    supabase)
      mkdir -p "$DIR/src/schema"
      cat > "$DIR/package.json" << 'PKGJSON'
{
  "name": "@rp/supabase",
  "private": true,
  "version": "0.0.1",
  "exports": {
    ".": "./src/index.ts",
    "./client-safe": "./src/client-safe.ts"
  },
  "scripts": {
    "db:push": "drizzle-kit push",
    "db:generate": "drizzle-kit generate",
    "db:studio": "drizzle-kit studio"
  },
  "dependencies": {
    "@supabase/supabase-js": "^2.0.0",
    "drizzle-orm": "^0.40.0",
    "postgres": "^3.0.0",
    "server-only": "^0.0.1"
  },
  "devDependencies": {
    "@rp/typescript-config": "workspace:*",
    "@rp/env": "workspace:*",
    "drizzle-kit": "^0.28.0"
  }
}
PKGJSON
      cat > "$DIR/src/index.ts" << 'SRCTS'
import "server-only";
export { db } from "./db";
export { supabase, supabaseAdmin } from "./client";
export * from "./schema";
SRCTS
      cat > "$DIR/src/db.ts" << 'SRCTS'
import "server-only";
import { drizzle } from "drizzle-orm/postgres-js";
import postgres from "postgres";
import { env } from "@rp/env";
import * as schema from "./schema";

const client = postgres(env.DATABASE_URL, { prepare: false });
export const db = drizzle(client, { schema });
SRCTS
      cat > "$DIR/src/client.ts" << 'SRCTS'
import "server-only";
import { createClient } from "@supabase/supabase-js";
import { env } from "@rp/env";

export const supabase = createClient(
  env.NEXT_PUBLIC_SUPABASE_URL!,
  env.NEXT_PUBLIC_SUPABASE_ANON_KEY!
);

export const supabaseAdmin = createClient(
  env.NEXT_PUBLIC_SUPABASE_URL!,
  env.SUPABASE_SERVICE_ROLE_KEY!
);
SRCTS
      cat > "$DIR/src/client-safe.ts" << 'SRCTS'
import { createClient } from "@supabase/supabase-js";
export const createBrowserClient = () =>
  createClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!
  );
SRCTS
      echo "export {};" > "$DIR/src/schema/index.ts"
      ;;

    auth)
      cat > "$DIR/package.json" << 'PKGJSON'
{
  "name": "@rp/auth",
  "private": true,
  "version": "0.0.1",
  "exports": {
    ".": "./src/index.ts",
    "./client": "./src/client.ts"
  },
  "dependencies": {
    "better-auth": "^1.4.0",
    "server-only": "^0.0.1"
  },
  "devDependencies": {
    "@rp/typescript-config": "workspace:*",
    "@rp/env": "workspace:*",
    "@rp/supabase": "workspace:*"
  }
}
PKGJSON
      cat > "$DIR/src/index.ts" << 'SRCTS'
import "server-only";
import { betterAuth } from "better-auth";
import { drizzleAdapter } from "better-auth/adapters/drizzle";
import { nextCookies } from "better-auth/next-js";
import { db } from "@rp/supabase";
import { env } from "@rp/env";

export const auth = betterAuth({
  baseURL: env.BETTER_AUTH_URL,
  secret: env.BETTER_AUTH_SECRET,
  database: drizzleAdapter(db, { provider: "pg", usePlural: true }),
  experimental: { joins: true },
  plugins: [nextCookies()],
  emailAndPassword: { enabled: true },
});

export type Session = typeof auth.$Infer.Session;
export type User = typeof auth.$Infer.Session.user;
SRCTS
      cat > "$DIR/src/client.ts" << 'SRCTS'
import { createAuthClient } from "better-auth/react";
export const authClient = createAuthClient({
  baseURL: process.env.NEXT_PUBLIC_APP_URL,
});
export const { signIn, signOut, signUp, useSession } = authClient;
SRCTS
      ;;

    kv)
      cat > "$DIR/package.json" << 'PKGJSON'
{
  "name": "@rp/kv",
  "private": true,
  "version": "0.0.1",
  "exports": { ".": "./src/index.ts" },
  "dependencies": {
    "@upstash/redis": "^1.34.0",
    "@upstash/ratelimit": "^2.0.0",
    "server-only": "^0.0.1"
  },
  "devDependencies": {
    "@rp/typescript-config": "workspace:*",
    "@rp/env": "workspace:*"
  }
}
PKGJSON
      cat > "$DIR/src/index.ts" << 'SRCTS'
import "server-only";
import { Redis } from "@upstash/redis";
import { Ratelimit } from "@upstash/ratelimit";
import { env } from "@rp/env";

export const redis = new Redis({
  url: env.UPSTASH_REDIS_REST_URL!,
  token: env.UPSTASH_REDIS_REST_TOKEN!,
});

export const ratelimit = new Ratelimit({
  redis,
  limiter: Ratelimit.slidingWindow(60, "1 m"),
  analytics: true,
});
SRCTS
      ;;

    trpc)
      mkdir -p "$DIR/src/routers"
      cat > "$DIR/package.json" << 'PKGJSON'
{
  "name": "@rp/trpc",
  "private": true,
  "version": "0.0.1",
  "exports": {
    ".": "./src/index.ts",
    "./client": "./src/client.ts",
    "./context": "./src/context.ts"
  },
  "dependencies": {
    "@trpc/server": "^11.0.0",
    "@trpc/client": "^11.0.0",
    "@trpc/react-query": "^11.0.0",
    "@tanstack/react-query": "^5.0.0",
    "zod": "^3.0.0"
  },
  "devDependencies": {
    "@rp/typescript-config": "workspace:*",
    "@rp/auth": "workspace:*"
  }
}
PKGJSON
      cat > "$DIR/src/context.ts" << 'SRCTS'
import { cache } from "react";
import { headers } from "next/headers";
import { auth } from "@rp/auth";
export const createTRPCContext = cache(async () => {
  const session = await auth.api.getSession({ headers: await headers() });
  return { session };
});
export type TRPCContext = Awaited<ReturnType<typeof createTRPCContext>>;
SRCTS
      cat > "$DIR/src/index.ts" << 'SRCTS'
import "server-only";
import { initTRPC, TRPCError } from "@trpc/server";
import type { TRPCContext } from "./context";
const t = initTRPC.context<TRPCContext>().create();
export const router = t.router;
export const publicProcedure = t.procedure;
export const protectedProcedure = t.procedure.use(({ ctx, next }) => {
  if (!ctx.session) throw new TRPCError({ code: "UNAUTHORIZED" });
  return next({ ctx: { session: ctx.session } });
});
export const appRouter = router({});
export type AppRouter = typeof appRouter;
SRCTS
      cat > "$DIR/src/client.ts" << 'SRCTS'
import { createTRPCReact } from "@trpc/react-query";
import type { AppRouter } from "./index";
export const trpc = createTRPCReact<AppRouter>();
SRCTS
      ;;

    validators)
      cat > "$DIR/package.json" << 'PKGJSON'
{
  "name": "@rp/validators",
  "private": true,
  "version": "0.0.1",
  "exports": { ".": "./src/index.ts" },
  "dependencies": { "zod": "^3.0.0" },
  "devDependencies": { "@rp/typescript-config": "workspace:*" }
}
PKGJSON
      echo "export {};" > "$DIR/src/index.ts"
      ;;

    remote-config)
      cat > "$DIR/package.json" << 'PKGJSON'
{
  "name": "@rp/remote-config",
  "private": true,
  "version": "0.0.1",
  "exports": { ".": "./src/index.ts" },
  "dependencies": { "server-only": "^0.0.1" },
  "devDependencies": { "@rp/typescript-config": "workspace:*", "@rp/kv": "workspace:*" }
}
PKGJSON
      cat > "$DIR/src/index.ts" << 'SRCTS'
import "server-only";
import { redis } from "@rp/kv";

type FeatureFlag = Record<string, boolean>;

export async function getFlags(): Promise<FeatureFlag> {
  const flags = await redis.get<FeatureFlag>("feature-flags");
  return flags ?? {};
}

export async function isEnabled(flag: string): Promise<boolean> {
  const flags = await getFlags();
  return flags[flag] ?? false;
}

export async function setFlag(flag: string, value: boolean): Promise<void> {
  const flags = await getFlags();
  await redis.set("feature-flags", { ...flags, [flag]: value });
}
SRCTS
      ;;

    api)
      mkdir -p "$NAME/apps/api/src/middleware"
      cat > "$NAME/apps/api/package.json" << 'PKGJSON'
{
  "name": "api",
  "private": true,
  "scripts": {
    "dev": "bun run --watch src/index.ts",
    "start": "bun run src/index.ts",
    "type-check": "tsc --noEmit"
  },
  "dependencies": {
    "hono": "^4.0.0",
    "@hono/trpc-server": "^0.3.0"
  },
  "devDependencies": {
    "@rp/typescript-config": "workspace:*",
    "@rp/env": "workspace:*"
  }
}
PKGJSON
      cat > "$NAME/apps/api/src/index.ts" << 'SRCTS'
import { Hono } from "hono";
import { cors } from "hono/cors";
import { logger } from "hono/logger";

const app = new Hono();
app.use("*", logger());
app.use("/api/*", cors({ origin: process.env.NEXT_PUBLIC_APP_URL ?? "*" }));
app.get("/health", (c) => c.json({ ok: true, ts: Date.now() }));

export default { port: 3001, fetch: app.fetch };
SRCTS
      ;;
  esac

  echo "  ✅ @rp/$PKG"
done

echo ""
echo "✅ Optional packages scaffolded: $PKGS"
echo "Run: cd $NAME && bun install"
