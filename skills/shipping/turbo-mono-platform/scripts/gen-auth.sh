#!/usr/bin/env bash
# gen-auth.sh — Better Auth route handler + proxy.ts guard
set -e
NAME="${1:?Usage: gen-auth.sh <project-name>}"

AUTH_ROUTE="$NAME/apps/web/src/app/api/auth/\[...all\]/route.ts"
mkdir -p "$(dirname "$AUTH_ROUTE")"

cat > "$AUTH_ROUTE" << 'EOF'
import { auth } from "@rp/auth";
import { toNextJsHandler } from "better-auth/next-js";
export const { GET, POST } = toNextJsHandler(auth);
EOF

cat > "$NAME/apps/web/proxy.ts" << 'EOF'
import { auth } from "@rp/auth";
import { NextRequest, NextResponse } from "next/server";

export async function proxy(request: NextRequest) {
  const session = await auth.api.getSession({ headers: request.headers });
  if (!session) return NextResponse.redirect(new URL("/login", request.url));
  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*", "/settings/:path*"],
};
EOF

echo "✅ Auth route handler: $AUTH_ROUTE"
echo "✅ Route guard: $NAME/apps/web/proxy.ts"
echo ""
echo "Next:"
echo "  1. Add auth tables: bun run db:push (after setting DATABASE_URL)"
echo "  2. Update proxy.ts matcher for your protected routes"
echo "  3. Import { useSession } from '@rp/auth/client' in client components"
