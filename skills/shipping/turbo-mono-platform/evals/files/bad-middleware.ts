// ❌ This is the OLD Next.js 15 pattern
// Used by eval #3: Claude must identify and fix this

import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

// Wrong: function name, file name should be proxy.ts, no auth check
export function middleware(request: NextRequest) {
  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*"],
};
