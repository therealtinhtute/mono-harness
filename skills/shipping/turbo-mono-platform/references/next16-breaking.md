# Next.js 16 — Breaking Changes (Before / After)

## 1. Route Guard: middleware.ts → proxy.ts

**❌ Old (Next.js 15)**
```typescript
// middleware.ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  return NextResponse.next();
}
export const config = { matcher: ["/dashboard/:path*"] };
```

**✅ New (Next.js 16)**
```typescript
// proxy.ts  ← filename changed, export name changed
import { auth } from "@rp/auth";
import { NextRequest, NextResponse } from "next/server";

export async function proxy(request: NextRequest) {
  const session = await auth.api.getSession({ headers: request.headers });
  if (!session) return NextResponse.redirect(new URL("/login", request.url));
  return NextResponse.next();
}
export const config = { matcher: ["/dashboard/:path*", "/settings/:path*"] };
```

---

## 2. next.config.ts: dynamicIO → cacheComponents

**❌ Old**
```typescript
const config: NextConfig = {
  experimental: { dynamicIO: true, ppr: true },
};
```

**✅ New**
```typescript
const config: NextConfig = {
  experimental: { cacheComponents: true },
  transpilePackages: ["@rp/ui"],
};
export default config;
```

---

## 3. Caching: unstable_cache → "use cache" directive

**❌ Old**
```typescript
import { unstable_cache } from "next/cache";
const getData = unstable_cache(async (id: string) => {
  return db.query.items.findFirst({ where: eq(items.id, id) });
}, ["item"], { tags: ["items"] });
```

**✅ New**
```typescript
import {
  unstable_cacheTag as cacheTag,
  unstable_cacheLife as cacheLife,
} from "next/cache";

async function getData(id: string) {
  "use cache";
  cacheTag(`item-${id}`, "items");
  cacheLife("hours");
  return db.query.items.findFirst({ where: eq(items.id, id) });
}
```

---

## 4. revalidateTag — second arg required

**❌ Old**
```typescript
revalidateTag("items");
```

**✅ New**
```typescript
import { revalidateTag } from "next/cache";
await revalidateTag("items", "max");
await revalidateTag(`item-${id}`, "max");
```

**updateTag in Server Actions (read-your-writes):**
```typescript
"use server";
import { updateTag, revalidateTag } from "next/cache";

export async function updateItem(id: string, data: { name: string }) {
  await db.update(items).set(data).where(eq(items.id, id));
  updateTag(`item-${id}`);
  await revalidateTag(`item-${id}`, "max");
}
```

---

## 5. Async params / cookies / headers

**❌ Old**
```typescript
export default function Page({ params }: { params: { id: string } }) {
  const id = params.id;
  const token = cookies().get("token");
  const ua = headers().get("user-agent");
}
```

**✅ New**
```typescript
export default async function Page({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const token = (await cookies()).get("token");
  const ua = (await headers()).get("user-agent");
}
```

---

## cacheLife profiles

| Profile | Stale | Revalidate | Expire |
|---------|-------|------------|--------|
| `seconds` | 0s | 1s | 1m |
| `minutes` | 0s | 1m | 1h |
| `hours` | 5m | 1h | 1d |
| `days` | 1h | 1d | 2w |
| `weeks` | 1d | 1w | 30d |
| `max` | 30d | 30d | 1y |
